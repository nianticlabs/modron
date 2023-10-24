// Binary modron is a Cloud auditing tool.

// Modron compares the existing state with a set of predefined rules
// and provides ways to crowd source the resolution of issues
// to resource owners.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/engine/rules"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

// TODO: Implement paginated API
type modronService struct {
	checker         model.Checker
	collector       model.Collector
	notificationSvc model.NotificationService
	ruleEngine      model.RuleEngine
	stateManager    model.StateManager
	storage         model.Storage
	// Required
	pb.UnimplementedModronServiceServer
	pb.UnimplementedNotificationServiceServer
}

func (modron *modronService) validateResourceGroupNames(ctx context.Context, resourceGroupNames []string) ([]string, error) {
	ownedResourceGroups, err := modron.checker.ListResourceGroupNamesOwned(ctx)
	if err != nil {
		glog.Warningf("validate resource groups: %v", err)
		return nil, status.Error(codes.Unauthenticated, "failed authenticating request")
	}
	if len(resourceGroupNames) == 0 {
		for k := range ownedResourceGroups {
			resourceGroupNames = append(resourceGroupNames, k)
		}
	} else {
		for _, rsgName := range resourceGroupNames {
			if _, ok := ownedResourceGroups[rsgName]; !ok {
				return nil, status.Error(codes.PermissionDenied, "resource group(s) is inaccessible")
			}
		}
	}
	return resourceGroupNames, nil
}

func (modron *modronService) scan(ctx context.Context, resourceGroupNames []string, scanId string) {
	glog.V(5).Infof("request scan %s for %+v", scanId, resourceGroupNames)
	filteredGroups := modron.stateManager.AddScan(scanId, resourceGroupNames)
	glog.V(5).Infof("filtered scan %s for %+v", scanId, filteredGroups)
	if len(filteredGroups) < 1 {
		glog.Warningf("no groups to scan, aborting %s", scanId)
		return
	} else {
		glog.Infof("starting scan %s for resource groups %v", scanId, filteredGroups)
	}
	obs, errs := modron.ruleEngine.CheckRules(ctx, scanId, filteredGroups)
	if len(errs) > 0 {
		glog.Errorf("scanId %s: %v", scanId, errors.Join(errs...))
	}
	glog.V(5).Infof("ending scan %s for %+v", scanId, filteredGroups)
	modron.stateManager.EndScan(scanId, filteredGroups)
	glog.Infof("scan %s completed", scanId)
	if err := modron.storage.FlushOpsLog(ctx); err != nil {
		glog.Warningf("flush ops log: %v", err)
	}
	if len(obs) < 1 {
		glog.Warningf("scan %s returned no observations.", scanId)
	}
	for _, o := range obs {
		notifications, err := modron.notificationsFromObservation(ctx, o)
		if err != nil {
			glog.Warningf("notifications: %v", err)
			continue
		}
		for _, n := range notifications {
			_, err := modron.notificationSvc.CreateNotification(ctx, n)
			if err != nil {
				if s, ok := status.FromError(err); !ok {
					glog.Warningf("notification creation: %v", err)
				} else {
					if s.Code() == codes.AlreadyExists || s.Code() == codes.FailedPrecondition {
						// Failed precondition is returned when an exception exist for the notification.
						continue
					} else {
						glog.Warningf("notification %v creation: %s: %s", n, s.Code(), s.Message())
					}
				}
			}
		}
	}
}

func (modron *modronService) collect(ctx context.Context, resourceGroupNames []string, collectId string) {
	glog.V(5).Infof("request collection %s for %+v", collectId, resourceGroupNames)
	filteredGroups := modron.stateManager.AddCollect(collectId, resourceGroupNames)
	glog.V(5).Infof("filtered collection %s for %+v", collectId, filteredGroups)
	if len(filteredGroups) > 0 {
		glog.Infof("collect start id: %s for %+v", collectId, filteredGroups)
		if err := modron.collector.CollectAndStoreAllResourceGroupResources(ctx, collectId, filteredGroups); len(err) != 0 {
			glog.Warningf("collectId %s, errs: %v", collectId, errors.Join(err...))
		} else {
			glog.Infof("collect done %s", collectId)
		}
	}
	modron.stateManager.EndCollect(collectId, filteredGroups)
}

func (modron *modronService) collectAndScan(ctx context.Context, resourceGroupNames []string) *pb.CollectAndScanResponse {
	modronCtx := context.Background()
	collectId, scanId := uuid.NewString(), uuid.NewString()
	go func(resourceGroupNames []string) {
		modron.collect(modronCtx, resourceGroupNames, collectId)
		modron.scan(modronCtx, resourceGroupNames, scanId)
	}(resourceGroupNames)
	return &pb.CollectAndScanResponse{
		CollectId: collectId,
		ScanId:    scanId,
	}
}

// rpc exposed CollectAndScan with resource group ownership validation
func (modron *modronService) CollectAndScan(ctx context.Context, in *pb.CollectAndScanRequest) (*pb.CollectAndScanResponse, error) {
	resourceGroupNames, err := modron.validateResourceGroupNames(ctx, in.ResourceGroupNames)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "resource group %+v: %v", in.ResourceGroupNames, err)
	}
	return modron.collectAndScan(ctx, resourceGroupNames), nil
}

func (modron *modronService) scheduledRunner(ctx context.Context) {
	intervalS := os.Getenv(collectAndScanInterval)

	interval, err := time.ParseDuration(intervalS)
	if err != nil || interval < time.Hour { // enforce a minimum of 1 hour interval
		interval, _ = time.ParseDuration(defaultCollectAndScanInterval)
		glog.Warningf("scan scheduler: env interval: %v . Keeping default value: %s", err, interval)
	}
	for {
		glog.Infof("scan scheduler: starting")
		if ctx.Err() != nil {
			glog.Errorf("scan scheduler: %v", ctx.Err())
			return
		}
		rgs, err := modron.collector.ListResourceGroupNames(ctx)
		if err != nil {
			glog.Errorf("list resource groups: %v", err)
			return
		}
		r := modron.collectAndScan(ctx, rgs)
		glog.Infof("scan scheduler done: collectionID: %s, scanID: %s", r.CollectId, r.ScanId)
		time.Sleep(interval)
	}
}

func (modron *modronService) ListObservations(ctx context.Context, in *pb.ListObservationsRequest) (*pb.ListObservationsResponse, error) {
	groups, err := modron.validateResourceGroupNames(ctx, in.ResourceGroupNames)
	if err != nil {
		return nil, err
	}
	obsByGroupByRules := map[string]map[string][]*pb.Observation{}
	oneWeekAgo := time.Now().Add(-time.Hour * 24 * 7)
	obs, err := modron.storage.ListObservations(ctx, model.StorageFilter{
		ResourceGroupNames: groups,
		StartTime:          oneWeekAgo,
		TimeOffset:         time.Since(oneWeekAgo),
	})
	if err != nil {
		glog.Warningf("list observations: %v", err)
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
		if obsByGroupByRules[group] == nil {
			obsByGroupByRules[group] = map[string][]*pb.Observation{}
		}
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
	res, err := modron.storage.ListResources(ctx, model.StorageFilter{ResourceNames: []string{in.Observation.Resource.Name}})
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
	rg, err := modron.storage.ListResources(ctx, model.StorageFilter{ResourceNames: []string{ob.Resource.ResourceGroupName}, Limit: 1})
	if err != nil {
		return nil, err
	}
	if len(rg) < 1 {
		return nil, fmt.Errorf("no resource found %+v", ob.Resource.Uid)
	}
	if len(rg) > 1 {
		glog.Warningf("multiple resources group found for %+v, using the first one", ob.Resource.Uid)
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
						uniqueContacts[strings.TrimPrefix(m, constants.GCPUserAccountPrefix)] = struct{}{}
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
				Interval:     notificationInterval,
			})
	}
	return notifications, nil
}
