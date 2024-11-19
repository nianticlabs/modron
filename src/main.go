package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/nianticlabs/modron/src/collector"
	pb "github.com/nianticlabs/modron/src/proto/generated"
)

var args struct {
	AdditionalAdminRoles   []string      `arg:"--additional-admin-roles,env:ADDITIONAL_ADMIN_ROLES" help:"Comma separated list of roles that are considered administrators of a resource group"`
	AdminGroups            []string      `arg:"--admin-groups,env:ADMIN_GROUPS" help:"Comma separated list of groups that are allowed to see all projects"`
	AllowedSccCategories   []string      `arg:"--allowed-scc-categories,env:ALLOWED_SCC_CATEGORIES" help:"Comma separated list of SCC categories that are allowed to create observations"`
	CollectAndScanInterval time.Duration `arg:"--collect-and-scan-interval,env:COLLECT_AND_SCAN_INTERVAL" help:"Interval between collecting and scanning (example: 3h)" default:"6h"`
	// TODO: Collector should be a list, as we might want to support more
	Collector                   collector.Type `arg:"--collector,env:COLLECTOR" help:"Specify which collector to use" default:"gcp"`
	DbBatchSize                 int32          `arg:"--db-batch-size,env:DB_BATCH_SIZE" help:"Number of records to insert in a single batch" default:"32"`
	DbConnectionMaxIdleTime     time.Duration  `arg:"--db-connection-max-idle-time,env:DB_CONNECTION_MAX_IDLE_TIME" help:"Maximum amount of time a connection may be idle" default:"30s"`
	DbConnectionMaxLifetime     time.Duration  `arg:"--db-connection-max-lifetime,env:DB_CONNECTION_MAX_LIFETIME" help:"Maximum amount of time a connection may be reused" default:"1h"`
	DbMaxConnections            int            `arg:"--db-max-connections,env:DB_MAX_CONNECTIONS" help:"Maximum number of connections to the database" default:"10"`
	DbMaxIdleConnections        int            `arg:"--db-max-idle-connections,env:DB_MAX_IDLE_CONNECTIONS" help:"Maximum number of idle connections to the database" default:"10"`
	DisableTelemetry            bool           `arg:"--disable-telemetry,env:DISABLE_TELEMETRY" help:"Disable OTEL telemetry" default:"false"`
	Environment                 string         `arg:"--environment,env:ENVIRONMENT" help:"Environment (development, production)" default:"development"`
	ExcludedRules               []string       `arg:"--excluded-rules,env:EXCLUDED_RULES" help:"Comma separated list of rules to exclude from the scan."`
	ImpactMap                   string         `arg:"--impact-map,env:IMPACT_MAP" help:"JSON map that maps the environment name to the impact level" default:"{}"`
	IsE2EGrpcTest               bool           `arg:"--is-e2e-grpc-test,env:IS_E2E_GRPC_TEST" help:"Is this an end-to-end gRPC test" default:"false"`
	LabelToEmailRegexp          string         `arg:"--label-to-email-regexp,env:LABEL_TO_EMAIL_REGEXP" help:"Regular expression to extract email from labels" default:"(.*)_(.*?)_(.*?)$"`
	LabelToEmailSubst           string         `arg:"--label-to-email-substitution,env:LABEL_TO_EMAIL_SUBSTITUTION" help:"Substitution to apply to the email extracted from labels" default:"$1@$2.$3"`
	ListenAddr                  string         `arg:"--listen-addr,env:LISTEN_ADDR" help:"Address to listen on" default:"127.0.0.1"`
	LogFormat                   LogFormat      `arg:"--log-format,env:LOG_FORMAT" help:"Log format (json,text)" default:"json"`
	LogLevel                    string         `arg:"--log-level,env:LOG_LEVEL" help:"Log level (trace,debug,info,warning,error)" default:"info"`
	LogAllSQLQueries            bool           `arg:"--log-all-sql-queries,env:LOG_ALL_SQL_QUERIES" help:"Log all SQL queries" default:"false"`
	NotificationInterval        time.Duration  `arg:"--notification-interval,env:NOTIFICATION_INTERVAL" help:"Interval between notifications (minimum: 24h)" default:"24h"`
	NotificationService         string         `arg:"--notification-service,env:NOTIFICATION_SERVICE" help:"Address of the notification service"`
	NotificationServiceClientID string         `arg:"--notification-service-client-id,env:NOTIFICATION_SERVICE_CLIENT_ID" help:"Client ID for the notification service"`
	OrgID                       string         `arg:"--org-id,env:ORG_ID,required" help:"Organization ID"`
	OrgSuffix                   string         `arg:"--org-suffix,env:ORG_SUFFIX,required" help:"Organization suffix (e.g: @example.com)"`
	PersistentCache             bool           `arg:"--persistent-cache,env:PERSISTENT_CACHE" help:"Use a persistent ACL cache that will be stored on the temporary directory" default:"false"`
	PersistentCacheTimeout      time.Duration  `arg:"--persistent-cache-timeout,env:PERSISTENT_CACHE_TIMEOUT" help:"Amount of time to keep the ACLs on the filesystem before we fetch them again" default:"5m"`
	Port                        int32          `arg:"--port,env:PORT" help:"Port to listen on" default:"8080"`
	RuleConfigs                 string         `arg:"--rule-configs,env:RULE_CONFIGS" help:"A map of rule names to their JSON configuration" default:"{}"`
	RunAutomatedScans           bool           `arg:"--run-automated-scans,env:RUN_AUTOMATED_SCANS" help:"Run automated scans" default:"true"`
	SelfURL                     string         `arg:"--self-url,env:SELF_URL" help:"URL of Modron - to be used when sending notifications" default:"https://modron"`
	SkipIAP                     bool           `arg:"--skip-iap,env:SKIP_IAP" help:"Skip IAP authentication" default:"false"`
	SQLBackendDriver            string         `arg:"--sql-backend-driver,env:SQL_BACKEND_DRIVER" help:"SQL backend driver (postgres)" default:"postgres"`
	SQLConnectionString         string         `arg:"--sql-connection-string,env:SQL_CONNECT_STRING" help:"SQL connection string" default:""` // TODO: Find where SQL_CONNECT_STRING is used and change it to SQL_CONNECTION_STRING for the future
	Storage                     string         `arg:"--storage,env:STORAGE" help:"Storage type (memory,sql)" default:"sql"`
	TagCustomerData             string         `arg:"--tag-customer-data,env:TAG_CUSTOMER_DATA,required" help:"Tag to use to define the customer data (e.g: 111111111/customer_data)"`
	TagEmployeeData             string         `arg:"--tag-employee-data,env:TAG_EMPLOYEE_DATA,required" help:"Tag to use to define the employee data (e.g: 111111111/employee_data)"`
	TagEnvironment              string         `arg:"--tag-environment,env:TAG_ENVIRONMENT,required" help:"Tag to use to define the environment (e.g: 111111111/environment)"`
}

var (
	start  = time.Now()
	log    = logrus.StandardLogger()
	tracer trace.Tracer
)

const (
	ExitCodeOK = iota
	ExitCodeInvalidArgs
	ExitCodeFailedToListen
	ExitCodeFailedToCreateServer
	ExitCodeFailedToServeGRPC
	ExitCodeFailedToServeHTTP
)

const (
	HTTPHeaderReadTimeout = 5 * time.Second
)

func main() {
	arg.MustParse(&args)
	setLogLevel()
	setLogFormat()

	log.Debugf("validating arguments..")
	if err := validateArgs(); err != nil {
		log.Error(err)
		os.Exit(ExitCodeInvalidArgs)
	}
	log.Debugf("starting")
	doMain()
}

func doMain() {
	mainCtx, cancel := context.WithCancel(context.Background())
	tp, mp := initTracer(mainCtx)
	defer func() {
		_ = tp.Shutdown(mainCtx)
		_ = mp.Shutdown(mainCtx)
	}()

	defer log.Tracef("net.Listen on port %d", args.Port)
	// Handle SIGINT (for Ctrl+C) and SIGTERM (for Cloud Run) signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-c
		log.Infof("received signal: %+v", sig)
		cancel()
	}()

	go func() {
		ctx, span := tracer.Start(mainCtx, "setup-grpc")
		log.Tracef("Setting up GRPC")
		// Use insecure credentials since communication is encrypted and authenticated via
		// HTTPS end-to-end (i.e., from client to Cloud Run ingress).
		var opts = []grpc.ServerOption{
			grpc.Creds(insecure.NewCredentials()),
		}
		opts = append(
			opts,
			grpc.StatsHandler(otelgrpc.NewServerHandler()),
		)
		grpcServer := grpc.NewServer(opts...) // nosemgrep: go.grpc.security.grpc-server-insecure-connection.grpc-server-insecure-connection
		log.Tracef("Creating newServer")
		srv, err := newServer(ctx)
		if err != nil {
			log.Errorf("server creation: %v", err)
			os.Exit(ExitCodeFailedToCreateServer)
		}
		log.Tracef("Registering Modron Service Server")
		pb.RegisterModronServiceServer(grpcServer, srv)
		log.Tracef("Registering Modron Notification Service Server")
		pb.RegisterNotificationServiceServer(grpcServer, srv)
		log.Infof("server starting on port %d", args.Port)
		if args.RunAutomatedScans {
			log.Tracef("Starting scheduled runner")
			go srv.ScheduledRunner(mainCtx)
		}
		span.End()
		if args.IsE2EGrpcTest {
			log.Warnf("E2E gRPC test mode enabled")
			// TODO: Unfortunately we need this as the GRPC-Web is different from the GRPC protocol.
			// This is used only in the integration test that doesn't have a GRPC-Web client.
			// We should look into https://github.com/improbable-eng/grpc-web and check how we can implement a golang GRPC-web client.
			lis, err := net.Listen("tcp", fmt.Sprintf(":%d", args.Port))
			if err != nil {
				log.Errorf("failed to listen: %v", err)
				os.Exit(ExitCodeFailedToListen)
			}
			if err := grpcServer.Serve(lis); err != nil {
				log.Errorf("error while listening: %v", err)
				os.Exit(ExitCodeFailedToServeGRPC)
			}
		} else {
			log.Infof("starting gRPC-Web server")
			grpcWebServer := grpcweb.WrapServer(grpcServer, withCors()...)
			log.Debugf("time until start: %v", time.Since(start))
			httpServer := &http.Server{
				Addr:              fmt.Sprintf("%s:%d", args.ListenAddr, args.Port),
				ReadHeaderTimeout: HTTPHeaderReadTimeout,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					switch req.URL.Path {
					case "/healthz":
						w.WriteHeader(http.StatusOK)
					default:
						otelhttp.NewMiddleware("grpc-web")(grpcWebServer).ServeHTTP(w, req)
					}
				}),
			}
			if err := httpServer.ListenAndServe(); err != nil {
				log.Errorf("error while listening: %v", err)
				os.Exit(ExitCodeFailedToServeHTTP)
			}
		}
	}()

	<-mainCtx.Done()
	log.Infof("server stopped")
}

func newTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("Modron"),
		),
	)

	if err != nil {
		panic(err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(r),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exp)),
	)
}

func newTraceExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	return otlptrace.New(ctx,
		otlptracegrpc.NewClient(otlptracegrpc.WithInsecure()),
	)
}

func newMetricExporter(ctx context.Context) (sdkmetric.Exporter, error) {
	return otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithInsecure(),
	)
}

func initTracer(ctx context.Context) (*sdktrace.TracerProvider, *sdkmetric.MeterProvider) {
	if args.DisableTelemetry {
		log.Warnf("telemetry is disabled!")
		tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.NeverSample()))
		tracer = tp.Tracer("github.com/nianticlabs/modron")
		return tp, sdkmetric.NewMeterProvider()
	}
	traceExp, err := newTraceExporter(ctx)
	if err != nil {
		log.Fatalf("failed to initialize exporter: %v", err)
	}
	metricExp, err := newMetricExporter(ctx)
	if err != nil {
		log.Fatalf("failed to initialize metric exporter: %v", err)
	}
	// Create a new tracer provider with a batch span processor and the given exporter.
	tp := newTraceProvider(traceExp)
	mp := newMeterProvider(ctx, metricExp)

	if err := runtime.Start(); err != nil {
		log.Fatalf("failed to start runtime metrics: %v", err)
	}

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	tracer = tp.Tracer("github.com/nianticlabs/modron")
	return tp, mp
}

func newMeterProvider(ctx context.Context, exp sdkmetric.Exporter) *sdkmetric.MeterProvider {
	rsrc, err := resource.New(ctx)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
		return nil
	}

	return sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(rsrc),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp,
			sdkmetric.WithInterval(time.Second),
			sdkmetric.WithProducer(runtime.NewProducer()),
		)),
	)
}
