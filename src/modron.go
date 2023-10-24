// Binary modron is a Cloud auditing tool.

// Modron compares the existing state with a set of predefined rules
// and provides ways to crowd source the resolution of issues
// to resource owners.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/nianticlabs/modron/src/acl/fakeacl"
	"github.com/nianticlabs/modron/src/acl/gcpacl"
	"github.com/nianticlabs/modron/src/collector/gcpcollector"
	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/engine"
	"github.com/nianticlabs/modron/src/engine/rules"
	"github.com/nianticlabs/modron/src/lognotifier"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/nagatha"
	"github.com/nianticlabs/modron/src/pb"
	"github.com/nianticlabs/modron/src/statemanager/reqdepstatemanager"
	"github.com/nianticlabs/modron/src/storage/memstorage"
	"github.com/nianticlabs/modron/src/storage/sqlstorage"

	"github.com/golang/glog"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

const (
	adminGroupsEnvVar        = "ADMIN_GROUPS"
	collectAndScanInterval   = "COLLECT_AND_SCAN_INTERVAL" // given as a parsable duration string
	collectorEnvVar          = "COLLECTOR"
	dbBatchSizeEnvVar        = "DB_BATCH_SIZE"
	dbMaxConnectionsEnvVar   = "DB_MAX_CONNECTIONS"
	e2eGrpcTestEnvironment   = "E2E_GRPC_TESTING"
	environmentEnvVar        = "ENVIRONMENT"
	excludedRulesEnvVar      = "EXCLUDED_RULES"
	fakeCollectorEnvironment = "FAKE"
	gcpProjectIdEnvVar       = "GCP_PROJECT_ID"
	glogLevelEnvVar          = "GLOG_v"
	// TODO: Remove mem storage now that we have in memory SQLite.
	memStorageEnvironment      = "MEM"
	notificationIntervalEnvVar = "NOTIFICATION_INTERVAL_DURATION"
	notificationSvcAddrEnvVar  = "NOTIFICATION_SERVICE"
	observationTableIdEnvVar   = "OBSERVATION_TABLE_ID"
	operationTableIdEnvVar     = "OPERATION_TABLE_ID"
	portEnvVar                 = "PORT"
	productionEnvironment      = "PRODUCTION"
	resourceTableIdEnvVar      = "RESOURCE_TABLE_ID"
	runAutomatedScansEnvVar    = "RUN_AUTOMATED_SCANS"
	sqlBackendEnvVar           = "SQL_BACKEND_DRIVER"
	sqlConnectStringEnvVar     = "SQL_CONNECT_STRING"
	sqlStorageEnvironment      = "SQL"
	storageEnvVar              = "STORAGE"

	defaultCollectAndScanInterval = "12h"
	notificationIntervalDefault   = "720h" // 30d
	dbDefaultBatchSize            = int32(32)
)

var (
	adminGroups          = []string{}
	dbBatchSize          int32
	dbMaxConnections     int64
	excludedRules        = []string{}
	notificationInterval time.Duration
	orgSuffix            string
	port                 int64
	runAutomatedScans    bool

	requiredEnvVars = []string{gcpProjectIdEnvVar, storageEnvVar, observationTableIdEnvVar, resourceTableIdEnvVar, operationTableIdEnvVar, notificationSvcAddrEnvVar}
)

func newServer(ctx context.Context) (*modronService, error) {
	var storage model.Storage
	var err error
	switch os.Getenv(storageEnvVar) {
	case memStorageEnvironment:
		storage = memstorage.New()
	case sqlStorageEnvironment:
		db, err := sql.Open(os.Getenv(sqlBackendEnvVar), os.Getenv(sqlConnectStringEnvVar))
		if err != nil {
			return nil, fmt.Errorf("sql storage: %w", err)
		}
		db.SetMaxOpenConns(int(dbMaxConnections))
		db.SetMaxIdleConns(int(dbMaxConnections))
		db.SetConnMaxLifetime(time.Hour)
		storage, err = sqlstorage.New(db, sqlstorage.Config{
			ResourceTableID:    os.Getenv(resourceTableIdEnvVar),
			ObservationTableID: os.Getenv(observationTableIdEnvVar),
			OperationTableID:   os.Getenv(operationTableIdEnvVar),
			BatchSize:          dbBatchSize,
		})
		if err != nil {
			return nil, fmt.Errorf("sql storage creation: %w", err)
		}
	default:
		return nil, fmt.Errorf("no storage specified.")
	}
	if err := storage.PurgeIncompleteOperations(ctx); err != nil {
		glog.Errorf("purging incomplete operations: %v", err)
	}
	ruleEngine := engine.New(storage, rules.GetRules(), excludedRules)
	var collector model.Collector
	if useFakeCollector() {
		collector = gcpcollector.NewFake(ctx, storage)
	} else {
		var err error
		if collector, err = gcpcollector.New(ctx, storage); err != nil {
			return nil, fmt.Errorf("NewGCPCollector: %w", err)
		}
	}
	var checker model.Checker
	if useFakeCollector() {
		checker = fakeacl.New()
	} else {
		var err error
		if checker, err = gcpacl.New(ctx, collector, gcpacl.Config{AdminGroups: adminGroups, CacheTimeout: 20 * time.Second}); err != nil {
			return nil, fmt.Errorf("NewGcpChecker: %w", err)
		}
	}

	stateManager, err := reqdepstatemanager.New()
	if err != nil {
		return nil, fmt.Errorf("creating reqdepstatemanager: %w", err)
	}

	var notificationSvc model.NotificationService
	notificationSvcAddr := os.Getenv(notificationSvcAddrEnvVar)
	if notificationSvcAddr != "" {
		notificationSvc, err = nagatha.New(ctx, notificationSvcAddr)
		if err != nil {
			return nil, fmt.Errorf("nagatha service: %w", err)
		}
	} else {
		glog.Infof("%s is empty, logging instead", notificationSvcAddrEnvVar)
		notificationSvc = lognotifier.New()
	}

	return &modronService{
		checker:         checker,
		collector:       collector,
		notificationSvc: notificationSvc,
		ruleEngine:      ruleEngine,
		stateManager:    stateManager,
		storage:         storage,
	}, nil
}

func isProduction() bool {
	return os.Getenv(environmentEnvVar) == productionEnvironment
}

func isE2eGrpcTest() bool {
	return os.Getenv(environmentEnvVar) == e2eGrpcTestEnvironment
}

func useFakeCollector() bool {
	return os.Getenv(collectorEnvVar) == fakeCollectorEnvironment
}

func useMemStorage() bool {
	return os.Getenv(storageEnvVar) == memStorageEnvironment
}

func validateEnvironment() (errs []error) {
	var err error
	notificationInterval, err = time.ParseDuration(os.Getenv(notificationIntervalEnvVar))
	if err != nil {
		glog.Infof("Invalid notification interval %q: %v, using default %s", os.Getenv(notificationIntervalEnvVar), err, notificationIntervalDefault)
		notificationInterval, _ = time.ParseDuration(notificationIntervalDefault)
	}
	if orgSuffixEnv := os.Getenv(constants.OrgSuffixEnvVar); orgSuffixEnv == "" {
		errs = append(errs, fmt.Errorf("environment variable %q is not set", constants.OrgSuffixEnvVar))
	} else {
		orgSuffix = orgSuffixEnv
	}
	portStr := os.Getenv(portEnvVar)
	port, err = strconv.ParseInt(portStr, 10, 32)
	if err != nil {
		errs = append(errs, fmt.Errorf("%s contains an invalid port number %s: %w", portEnvVar, portStr, err))
	}
	if os.Getenv(glogLevelEnvVar) != "" {
		if err := flag.Set("v", os.Getenv(glogLevelEnvVar)); err != nil {
			errs = append(errs, fmt.Errorf("%s invalid value %s: %v", glogLevelEnvVar, os.Getenv(glogLevelEnvVar), err))
		}
	}
	runAutomatedScans = true
	if strings.EqualFold("false", os.Getenv(runAutomatedScansEnvVar)) {
		runAutomatedScans = false
	}
	excludedRules = strings.Split(os.Getenv(excludedRulesEnvVar), ",")
	if isProduction() {
		for _, v := range requiredEnvVars {
			if os.Getenv(v) == "" {
				errs = append(errs, fmt.Errorf("%s can't be empty", v))
			}
		}
		if useFakeCollector() {
			errs = append(errs, fmt.Errorf("cannot use fake collector in production"))
		}
		if useMemStorage() {
			errs = append(errs, fmt.Errorf("cannot use memstorage in production"))
		}
		adminGroups = strings.Split(os.Getenv(adminGroupsEnvVar), ",")
		if len(adminGroups) < 1 {
			errs = append(errs, fmt.Errorf("%q has no entries, add at least one admin group", adminGroupsEnvVar))
		}
	}
	if os.Getenv(storageEnvVar) == sqlStorageEnvironment {
		dbMaxConnectionsStr := os.Getenv(dbMaxConnectionsEnvVar)
		dbMaxConnections, err = strconv.ParseInt(dbMaxConnectionsStr, 10, 32)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s contains an invalid int %s: %w", dbMaxConnectionsEnvVar, dbMaxConnectionsStr, err))
		}
		dbBatchSizeStr := os.Getenv(dbBatchSizeEnvVar)
		if dbBatchSizeStr == "" {
			glog.Infof("%s is not defined, using %d as default", dbBatchSizeEnvVar, dbDefaultBatchSize)
			dbBatchSize = dbDefaultBatchSize
		} else {
			dbBatchSize64, err := strconv.ParseInt(dbBatchSizeStr, 10, 32)
			if err != nil && dbBatchSizeStr != "" {
				errs = append(errs, fmt.Errorf("%s contains an invalid int: %s: %w", dbBatchSizeEnvVar, dbBatchSizeStr, err))
			}
			dbBatchSize = int32(dbBatchSize64)
			if dbBatchSize < 1 {
				glog.Infof("%s is %d smaller than 1, using 1", dbBatchSizeEnvVar, dbBatchSize)
				dbBatchSize = 1
			}
		}
	}

	if len(errs) > 0 {
		fmt.Println("invalid environment:")
		for _, e := range errs {
			glog.Error(e)
			fmt.Println(e)
		}
		os.Exit(2)
	}
	return
}

func withCors() []grpcweb.Option {
	return []grpcweb.Option{
		grpcweb.WithOriginFunc(func(origin string) bool {
			return true
		}),
		grpcweb.WithAllowedRequestHeaders([]string{"*"}),
	}
}

func main() {
	flag.Parse()
	start := time.Now()
	validateEnvironment()
	glog.V(5).Infof("environment validation took %v", time.Since(start))
	ctx, cancel := context.WithCancel(context.Background())
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		glog.Errorf("failed to listen: %v", err)
		os.Exit(1)
	}

	// Handle SIGINT (for Ctrl+C) and SIGTERM (for Cloud Run) signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-c
		glog.Infof("received signal: %+v", sig)
		cancel()
	}()
	go func() {
		// Use insecure credentials since communication is encrypted and authenticated via
		// HTTPS end-to-end (i.e., from client to Cloud Run ingress).
		var opts []grpc.ServerOption = []grpc.ServerOption{
			grpc.Creds(insecure.NewCredentials()),
		}
		grpcServer := grpc.NewServer(opts...) // nosemgrep: go.grpc.security.grpc-server-insecure-connection.grpc-server-insecure-connection
		srv, err := newServer(ctx)
		if err != nil {
			glog.Errorf("server creation: %v", err)
			os.Exit(3)
		}
		pb.RegisterModronServiceServer(grpcServer, srv)
		pb.RegisterNotificationServiceServer(grpcServer, srv)
		glog.Infof("server starting on port %d", port)
		if runAutomatedScans {
			go srv.scheduledRunner(ctx)
		}
		if isE2eGrpcTest() {
			// TODO: Unfortunately we need this as the GRPC-Web is different from the GRPC protocol.
			// This is used only in the integration test that doesn't have a GRPC-Web client.
			// We should look into https://github.com/improbable-eng/grpc-web and check how we can implement a golang GRPC-web client.
			if err := grpcServer.Serve(lis); err != nil {
				glog.Errorf("error while listening: %v", err)
				os.Exit(3)
			}
		} else {
			grpcWebServer := grpcweb.WrapServer(grpcServer, withCors()...)
			glog.V(5).Infof("time until start: %v", time.Since(start))
			if err := http.Serve(lis, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				grpcWebServer.ServeHTTP(w, req)
			})); err != nil {
				glog.Errorf("error while listening: %v", err)
				os.Exit(4)
			}
		}
	}()
	<-ctx.Done()
	glog.Infof("server stopped")
}
