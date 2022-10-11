// Binary modron is a Cloud auditing tool.

// Modron compares the existing state with a set of predefined rules
// and provides ways to crowd source the resolution of issues
// to resource owners.
package main

import (
	"context"
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
	"github.com/nianticlabs/modron/src/storage/bigquerystorage"
	"github.com/nianticlabs/modron/src/storage/memstorage"

	"github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	adminGroupsEnvVar         = "ADMIN_GROUPS"
	collectorEnvVar           = "COLLECTOR"
	datasetIdEnvVar           = "DATASET_ID"
	environmentEnvVar         = "ENVIRONMENT"
	gcpProjectIdEnvVar        = "GCP_PROJECT_ID"
	notificationSvcAddrEnvVar = "NOTIFICATION_SERVICE"
	observationTableIdEnvVar  = "OBSERVATION_TABLE_ID"
	operationTableIdEnvVar    = "OPERATION_TABLE_ID"
	portEnvVar                = "PORT"
	resourceTableIdEnvVar     = "RESOURCE_TABLE_ID"
	storageEnvVar             = "STORAGE"

	fakeCollectorEnvironment = "FAKE"
	productionEnvironment    = "PRODUCTION"
	e2eGrpcTestEnvironment   = "E2E_GRPC_TESTING"
	memStorageEnvironment    = "MEM"
)

var (
	port            int64
	adminGroups     = []string{}
	orgSuffix       string
	requiredEnvVars = []string{datasetIdEnvVar, gcpProjectIdEnvVar, observationTableIdEnvVar, resourceTableIdEnvVar, notificationSvcAddrEnvVar}
)

// TODO: Implement paginated API
type modronService struct {
	storage         model.Storage
	ruleEngine      model.RuleEngine
	collector       model.Collector
	checker         model.Checker
	stateManager    model.StateManager
	notificationSvc model.NotificationService
	// Required
	pb.UnimplementedModronServiceServer
	pb.UnimplementedNotificationServiceServer
}

func (modron *modronService) validateResourceGroupNames(ctx context.Context, resourceGroupNames []string) ([]string, error) {
	ownedResourceGroups, err := modron.checker.ListResourceGroupNamesOwned(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed authenticating request")
	}
	if len(resourceGroupNames) == 0 {
		for k := range ownedResourceGroups {
			resourceGroupNames = append(resourceGroupNames, k)
		}
	} else {
		for _, rsgName := range resourceGroupNames {
			if _, ok := ownedResourceGroups[rsgName]; !ok {
				return nil, status.Error(codes.Internal, "resource group(s) is inaccessible")
			}
		}
	}
	return resourceGroupNames, nil
}

func (modron *modronService) scan(ctx context.Context, resourceGroupNames []string, scanId string) error {
	filteredGroups := modron.stateManager.AddScan(scanId, resourceGroupNames)
	if len(filteredGroups) != 0 {
		glog.Infof("starting scan %s", scanId)
		modron.logScanStatus(ctx, scanId, filteredGroups, model.OperationStarted)
		obs, err := modron.ruleEngine.CheckRules(ctx, scanId, filteredGroups)
		if err != nil {
			glog.Errorf("scanId %s: %v", scanId, err)
		}
		modron.stateManager.EndScan(scanId, filteredGroups)
		modron.logScanStatus(ctx, scanId, filteredGroups, model.OperationCompleted)
		glog.Infof("scan %s completed", scanId)
		if err := modron.storage.FlushOpsLog(ctx); err != nil {
			glog.Warningf("flush ops log: %v", err)
		}
		for _, o := range obs {
			notifications, err := modron.notificationsFromObservation(ctx, o)
			if err != nil {
				return fmt.Errorf("notifications: %v", err)
			}
			for _, n := range notifications {
				_, err := modron.notificationSvc.CreateNotification(ctx, n)
				if err != nil {
					if s, ok := status.FromError(err); !ok {
						glog.Warningf("notification creation: %v", err)
					} else {
						if s.Code() == codes.AlreadyExists {
							continue
						} else {
							glog.Warningf("notification creation: %s: %s, %v", s.Code(), s.Message(), s.Err())
						}
					}
				}
			}
		}
	}
	return nil
}

func (modron *modronService) collect(ctx context.Context, resourceGroupNames []string, collectId string) error {
	filteredGroups := modron.stateManager.AddCollect(collectId, resourceGroupNames)
	if len(filteredGroups) != 0 {
		glog.Infof("collect start id: %v for %v", collectId, filteredGroups)
		if err := modron.collector.CollectAndStoreAllResourceGroupResources(ctx, collectId, filteredGroups); len(err) != 0 {
			return fmt.Errorf("error collecting for collectId %v, err: %v", collectId, strings.ReplaceAll(fmt.Sprintf("%v", err), "\n", " "))
		}
		glog.Infof("collect done %v", collectId)
		modron.stateManager.EndCollect(collectId, filteredGroups)
	}
	return nil
}

func (modron *modronService) CollectAndScan(ctx context.Context, in *pb.CollectAndScanRequest) (*pb.CollectAndScanResponse, error) {
	modronCtx := context.Background()
	resourceGroupNames, err := modron.validateResourceGroupNames(ctx, in.ResourceGroupNames)
	if err != nil {
		return nil, err
	}
	collectId, scanId := uuid.NewString(), uuid.NewString()
	in.ResourceGroupNames = resourceGroupNames
	go func() {
		if err := modron.collect(modronCtx, in.ResourceGroupNames, collectId); err != nil {
			glog.Error(err)
			// Do not fail the scan if the collection failed.
		}
		if err := modron.scan(modronCtx, in.ResourceGroupNames, scanId); err != nil {
			glog.Error(err)
		}
	}()
	return &pb.CollectAndScanResponse{
		CollectId: collectId,
		ScanId:    scanId,
	}, nil
}

func (modron *modronService) ListObservations(ctx context.Context, in *pb.ListObservationsRequest) (*pb.ListObservationsResponse, error) {
	modronCtx := context.Background()
	groups, err := modron.validateResourceGroupNames(ctx, in.ResourceGroupNames)
	if err != nil {
		return nil, err
	}

	obsByGroupByRules := map[string]map[string][]*pb.Observation{}
	obs, err := modron.storage.ListObservations(modronCtx, model.StorageFilter{
		ResourceGroupNames: &groups,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed listing observations")
	}
	for _, group := range groups {
		obsByGroupByRules[group] = map[string][]*pb.Observation{}
		for _, rule := range rules.GetRules() {
			obsByGroupByRules[group][rule.Info().Name] = []*pb.Observation{}
		}
	}
	for _, ob := range obs {
		group := ob.Resource.ResourceGroupName
		rule := ob.Name
		obsByGroupByRules[group][rule] = append(
			obsByGroupByRules[group][rule],
			ob,
		)
	}
	res := []*pb.ResourceGroupObservationsPair{}
	for name, ruleObs := range obsByGroupByRules {
		val := []*pb.RuleObservationPair{}
		for rule, obs := range ruleObs {
			val = append(val, &pb.RuleObservationPair{
				Rule:         rule,
				Observations: obs,
			})
		}
		res = append(res, &pb.ResourceGroupObservationsPair{
			ResourceGroupName: name,
			RulesObservations: val,
		})
	}
	return &pb.ListObservationsResponse{
		ResourceGroupsObservations: res,
	}, nil
}

func (modron *modronService) CreateObservation(ctx context.Context, in *pb.CreateObservationRequest) (*pb.Observation, error) {
	if in.Observation == nil {
		return nil, status.Error(codes.InvalidArgument, "observation is nil")
	}
	if in.Observation.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "observation does not have a name")
	}
	if in.Observation.Resource == nil {
		return nil, status.Error(codes.InvalidArgument, "resource to link observation with not defined")
	}
	if in.Observation.Remediation == nil || in.Observation.Remediation.Recommendation == "" {
		return nil, status.Error(codes.InvalidArgument, "cannot create an observation without recommendation")
	}
	in.Observation.Timestamp = timestamppb.Now()
	res, err := modron.storage.ListResources(ctx, model.StorageFilter{ResourceNames: &[]string{in.Observation.Resource.Name}})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "resource to link observation to not found: %v", err)
	}
	if len(res) != 1 {
		return nil, status.Errorf(codes.FailedPrecondition, "found %d resources matching %+v", len(res), in.Observation.Resource)
	}
	in.Observation.Resource = res[0]
	if obs, err := modron.storage.BatchCreateObservations(ctx, []*pb.Observation{in.Observation}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	} else {
		if len(obs) != 1 {
			return nil, status.Errorf(codes.Internal, "creation returned %d items", len(obs))
		} else {
			return obs[0], nil
		}
	}
}

func (modron *modronService) GetStatusCollectAndScan(ctx context.Context, in *pb.GetStatusCollectAndScanRequest) (*pb.GetStatusCollectAndScanResponse, error) {
	collectStatus := modron.stateManager.GetCollectState(in.CollectId)
	scanStatus := modron.stateManager.GetScanState(in.ScanId)
	return &pb.GetStatusCollectAndScanResponse{
		CollectStatus: collectStatus,
		ScanStatus:    scanStatus,
	}, nil
}

func (modron *modronService) GetNotificationException(ctx context.Context, req *pb.GetNotificationExceptionRequest) (*pb.NotificationException, error) {
	ex, err := modron.validateUserAndGetException(ctx, req.Uuid)
	if err != nil {
		return nil, err
	}
	return ex.ToProto(), err
}

func (modron *modronService) CreateNotificationException(ctx context.Context, req *pb.CreateNotificationExceptionRequest) (*pb.NotificationException, error) {
	if userEmail, err := modron.checker.GetValidatedUser(ctx); err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed authenticating request")
	} else {
		req.Exception.UserEmail = userEmail
	}
	ex, err := modron.notificationSvc.CreateException(ctx, model.ExceptionFromProto(req.Exception))
	return ex.ToProto(), err
}

func (modron *modronService) UpdateNotificationException(ctx context.Context, req *pb.UpdateNotificationExceptionRequest) (*pb.NotificationException, error) {
	if ex, err := modron.validateUserAndGetException(ctx, req.Exception.Uuid); err != nil {
		return nil, err
	} else {
		req.Exception.UserEmail = ex.UserEmail
	}
	ex, err := modron.notificationSvc.UpdateException(ctx, model.ExceptionFromProto(req.Exception))
	return ex.ToProto(), err
}

func (modron *modronService) DeleteNotificationException(ctx context.Context, req *pb.DeleteNotificationExceptionRequest) (*emptypb.Empty, error) {
	if _, err := modron.validateUserAndGetException(ctx, req.Uuid); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, modron.notificationSvc.DeleteException(ctx, req.Uuid)
}

func (modron *modronService) ListNotificationExceptions(ctx context.Context, req *pb.ListNotificationExceptionsRequest) (*pb.ListNotificationExceptionsResponse, error) {
	if userEmail, err := modron.checker.GetValidatedUser(ctx); err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed authenticating request")
	} else {
		req.UserEmail = userEmail
	}
	ex, err := modron.notificationSvc.ListExceptions(ctx, req.UserEmail, req.PageSize, req.PageToken)
	exList := []*pb.NotificationException{}
	for _, e := range ex {
		exList = append(exList, e.ToProto())
	}
	return &pb.ListNotificationExceptionsResponse{Exceptions: exList}, err
}

func (modron *modronService) logScanStatus(ctx context.Context, scanId string, resourceGroups []string, status model.OperationStatus) {
	ops := []model.Operation{}
	for _, resourceGroup := range resourceGroups {
		ops = append(ops, model.Operation{ID: scanId, ResourceGroup: resourceGroup, OpsType: "scan", StatusTime: time.Now(), Status: status})
	}
	if err := modron.storage.AddOperationLog(ctx, ops); err != nil {
		glog.Warningf("log operation: %v", err)
	}
}

// TODO: Allow security admins to bypass the checks
func (modron *modronService) validateUserAndGetException(ctx context.Context, notificationUuid string) (model.Exception, error) {
	userEmail, err := modron.checker.GetValidatedUser(ctx)
	if err != nil {
		return model.Exception{}, status.Error(codes.Unauthenticated, "failed authenticating request")
	}
	if ex, err := modron.notificationSvc.GetException(ctx, notificationUuid); err != nil {
		return model.Exception{}, err
	} else if ex.UserEmail != userEmail {
		return model.Exception{}, status.Error(codes.Unauthenticated, "failed authenticating request")
	} else {
		return ex, nil
	}
}

func (modron *modronService) notificationsFromObservation(ctx context.Context, ob *pb.Observation) ([]model.Notification, error) {
	rg, err := modron.storage.ListResources(ctx, model.StorageFilter{ResourceNames: &[]string{ob.Resource.ResourceGroupName}})
	if err != nil {
		return nil, err
	}
	if len(rg) < 1 {
		glog.Warningf("no resource found %+v", ob.Resource.Uid)
	}
	if len(rg) > 1 {
		glog.Warningf("multiple resources found for %+v", ob.Resource.Uid)
	}
	if rg[0].IamPolicy == nil {
		glog.Warningf("no iam policy found for %s", rg[0].Name)
	}
	// We can have the same contacts in owners and labels, de-duplicate.
	uniqueContacts := make(map[string]struct{}, 0)
	for _, b := range rg[0].IamPolicy.Permissions {
		for r := range constants.AdminRoles {
			if strings.EqualFold(b.Role, r) {
				for _, m := range b.Principals {
					if strings.HasSuffix(m, orgSuffix) {
						uniqueContacts[strings.TrimPrefix(m, "user:")] = struct{}{}
					}
				}
			}
		}
	}

	contacts := maps.Keys(uniqueContacts)
	if len(contacts) < 1 {
		return nil, fmt.Errorf("no contacts found for observation %q, resource group: %q", ob.Uid, ob.Resource.ResourceGroupName)
	}
	notifications := make([]model.Notification, 0)
	for _, c := range contacts {
		notifications = append(notifications,
			model.Notification{
				SourceSystem: "modron",
				Name:         ob.Name,
				Recipient:    c,
				Content:      ob.Remediation.Recommendation,
				Interval:     time.Duration(7 * time.Hour * 24),
			})
	}
	return notifications, nil
}

func newServer(ctx context.Context) (*modronService, error) {
	var storage model.Storage
	var err error
	if useMemStorage() {
		storage = memstorage.New()
	} else {
		storage, err = bigquerystorage.New(
			ctx,
			os.Getenv(gcpProjectIdEnvVar),
			os.Getenv(datasetIdEnvVar),
			os.Getenv(resourceTableIdEnvVar),
			os.Getenv(observationTableIdEnvVar),
			os.Getenv(operationTableIdEnvVar),
		)
		if err != nil {
			return nil, fmt.Errorf("bigquerystorage creation error: %v", err)
		}
	}
	ruleEngine := engine.New(storage, rules.GetRules())
	var collector model.Collector
	if useFakeCollector() {
		collector = gcpcollector.NewFake(ctx, storage)
	} else {
		var err error
		if collector, err = gcpcollector.New(ctx, storage); err != nil {
			return nil, fmt.Errorf("NewGCPCollector error: %v", err)
		}
	}
	var checker model.Checker
	if useFakeCollector() {
		checker = fakeacl.New()
	} else {
		var err error
		if checker, err = gcpacl.New(ctx, collector, gcpacl.Config{AdminGroups: adminGroups, CacheTimeout: 20 * time.Second}); err != nil {
			return nil, fmt.Errorf("NewGcpChecker error: %v", err)
		}
	}

	stateManager, err := reqdepstatemanager.New()
	if err != nil {
		return nil, fmt.Errorf("creating reqdepstatemanager error: %v", err)
	}

	var notificationSvc model.NotificationService
	if isProduction() {
		notificationSvc, err = nagatha.New(ctx, os.Getenv(notificationSvcAddrEnvVar))
		if err != nil {
			return nil, fmt.Errorf("nagatha service: %v", err)
		}
	} else {
		notificationSvc = lognotifier.New()
	}

	return &modronService{
		storage:         storage,
		ruleEngine:      ruleEngine,
		collector:       collector,
		checker:         checker,
		stateManager:    stateManager,
		notificationSvc: notificationSvc,
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
	if orgSuffixEnv := os.Getenv(constants.OrgSuffixEnvVar); orgSuffixEnv == "" {
		errs = append(errs, fmt.Errorf("environment variable %q is not set", constants.OrgSuffixEnvVar))
	} else {
		orgSuffix = orgSuffixEnv
	}
	portStr := os.Getenv(portEnvVar)
	var err error
	port, err = strconv.ParseInt(portStr, 10, 32)
	if err != nil {
		errs = append(errs, fmt.Errorf("%s contains an invalid port number %s: %v", portEnvVar, portStr, err))
	}
	if len(errs) > 0 {
		fmt.Println("invalid environment:")
		for _, e := range errs {
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
	validateEnvironment()
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
		glog.Infof("server starting on port %d", port)
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

		if isE2eGrpcTest() {
			// TODO: Unfortunately we need this as the GRPC-Web is different from the GRPC protocol.
			// This is used only in the integration test that doesn't have a GRPC-Web client.
			// We should look into https://github.com/improbable-eng/grpc-web and check how we can implement a golang GRPC-web client.
			if err := grpcServer.Serve(lis); err != nil {
				glog.Errorf("error while listening: %v", err)
				os.Exit(2)
			}
		} else {
			grpcWebServer := grpcweb.WrapServer(grpcServer, withCors()...)
			if err := http.Serve(lis, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				grpcWebServer.ServeHTTP(w, req)
			})); err != nil {
				glog.Errorf("error while listening: %v", err)
				os.Exit(2)
			}
		}
	}()
	<-ctx.Done()
	glog.Infof("server stopped")
}
