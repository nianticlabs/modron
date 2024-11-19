package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/engine"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/nagatha"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"
)

var (
	tracer = otel.Tracer("github.com/nianticlabs/modron/src/service")
	meter  = otel.Meter("github.com/nianticlabs/modron/src/service")
)

// TODO: Implement paginated API
type Modron struct {
	Checker                model.Checker
	CollectAndScanInterval time.Duration
	Collector              model.Collector
	NotificationInterval   time.Duration
	NotificationSvc        model.NotificationService
	OrgSuffix              string
	RuleEngine             model.Engine
	SelfURL                string
	StateManager           model.StateManager
	Storage                model.Storage

	additionalAdminRolesMap map[constants.Role]struct{}
	labelToEmailRegexp      *regexp.Regexp
	labelToEmailSubst       string

	metrics metrics
	// Required
	pb.UnimplementedModronServiceServer
	pb.UnimplementedNotificationServiceServer
}

var (
	log = logrus.StandardLogger().WithField(constants.LogKeyPkg, "service")
)

type metrics struct {
	CollectDuration metric.Float64Histogram
	Observations    metric.Int64Counter
	ScanDuration    metric.Float64Histogram
}

const (
	oneDay  = time.Hour * 24
	oneWeek = oneDay * 7
)

func (modron *Modron) validateResourceGroupNames(ctx context.Context, resourceGroupNames []string) ([]string, error) {
	ownedResourceGroups, err := modron.Checker.ListResourceGroupNamesOwned(ctx)
	if err != nil {
		log.Warnf("validate resource groups: %v", err)
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

// preCollect retrieves and stores all the Resource Groups (projects, folders, org) _without their IAM policies_.
// We don't collect the IAM policies because we only need those of the RGs that we need to analyze,
// but we need the entire set of RGs to perform cross-environment checks
func (modron *Modron) preCollect(ctx context.Context, resourceGroupNames []string) ([]*pb.Resource, error) {
	ctx, span := tracer.Start(ctx, "preCollect")
	defer span.End()
	collectID, ok := ctx.Value(constants.CollectIDKey).(string)
	if !ok {
		return nil, fmt.Errorf("collectID not found in context")
	}
	pcLog := log.
		WithField("collect_id", collectID).
		WithField("resource_group_names", resourceGroupNames)
	pcLog.Info("starting pre-collect")
	defer pcLog.Info("pre-collect done")

	rgs, err := modron.Collector.ListResourceGroups(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("list resource groups: %w", err)
	}
	for i := range rgs {
		rgs[i].CollectionUid = collectID
	}
	return rgs, nil
}

func (modron *Modron) collect(ctx context.Context, resourceGroupNames []string, preCollectedRgs []*pb.Resource) []*pb.Observation {
	start := time.Now()
	collectID, ok := ctx.Value(constants.CollectIDKey).(string)
	if !ok {
		log.Errorf("collectID not found in context")
		return nil
	}
	ctx, span := tracer.Start(ctx, "collect",
		trace.WithAttributes(
			attribute.String(constants.TraceKeyCollectID, collectID),
			attribute.StringSlice(constants.TraceKeyResourceGroupNames, resourceGroupNames),
		),
	)
	defer span.End()
	collectLogger := log.
		WithField("collect_id", collectID).
		WithField("resource_group_names", resourceGroupNames)
	collectLogger.Debug("request collection")
	filteredGroups := modron.StateManager.AddCollect(collectID, resourceGroupNames)
	collectLogger = collectLogger.WithField("filtered_groups", filteredGroups)
	collectLogger.Debugf("filtered collection")
	if len(filteredGroups) > 0 {
		collectLogger.Infof("collect start")
		if err := modron.Collector.CollectAndStoreAll(ctx, collectID, filteredGroups, preCollectedRgs); err != nil {
			collectLogger.
				WithError(err).
				Warnf("some errors during collect: %v", err)
		} else {
			collectLogger.Info("collect done")
		}
	}
	modron.StateManager.EndCollect(collectID, filteredGroups)
	modron.metrics.CollectDuration.Record(ctx, time.Since(start).Seconds())

	// Create notifications for collected observations
	collectedObs, err := modron.getCollectedObservations(ctx, collectID, filteredGroups)
	if err != nil {
		collectLogger.
			WithError(err).
			Warnf("get collected observations: %v", err)
		return nil
	}
	return collectedObs
}

func (modron *Modron) scan(ctx context.Context, resourceGroupNames []string, preCollectedRgs []*pb.Resource) []*pb.Observation {
	start := time.Now()
	scanID, ok := ctx.Value(constants.ScanIDKey).(string)
	if !ok {
		log.Errorf("scanID not found in context")
		return nil
	}
	ctx, span := tracer.Start(ctx, "scan",
		trace.WithAttributes(
			attribute.String(constants.TraceKeyScanID, scanID),
			attribute.StringSlice(constants.TraceKeyResourceGroupNames, resourceGroupNames),
		),
	)
	defer span.End()
	collectID, ok := ctx.Value(constants.CollectIDKey).(string)
	if !ok {
		log.Errorf("collectID not found in context")
		return nil
	}
	scanLogger := log.
		WithField("collect_id", collectID).
		WithField("scan_id", scanID).
		WithField("resource_group_names", resourceGroupNames)
	scanLogger.Debugf("requested scan")
	filteredGroups := modron.StateManager.AddScan(scanID, resourceGroupNames)
	scanLogger = scanLogger.WithField("filtered_groups", filteredGroups)
	scanLogger.Debug("filtered scan")
	if len(filteredGroups) < 1 {
		scanLogger.Warnf("no groups to scan, aborting")
		return nil
	}
	scanLogger.Info("starting scan")
	obs, errs := modron.RuleEngine.CheckRules(ctx, scanID, collectID, filteredGroups, preCollectedRgs)
	tookSeconds := time.Since(start).Seconds()
	scanLogger = scanLogger.
		WithFields(logrus.Fields{
			"observations": len(obs),
			"took":         tookSeconds,
		})
	if len(errs) > 0 {
		scanLogger.
			WithError(errors.Join(errs...)).
			Errorf("scan completed with errors: %v", errors.Join(errs...))
	}
	scanLogger.Debug("ending scan")
	modron.StateManager.EndScan(scanID, filteredGroups)
	scanLogger.Info("scan completed")
	modron.metrics.ScanDuration.Record(ctx, tookSeconds)
	if err := modron.Storage.FlushOpsLog(ctx); err != nil {
		scanLogger.WithError(err).
			Warnf("flush ops log: %v", err)
	}
	modron.metrics.Observations.Add(ctx, int64(len(obs)))
	if len(obs) < 1 {
		scanLogger.Warnf("scan returned no observations")
	}
	return obs
}

func (modron *Modron) createNotifications(ctx context.Context, obs []*pb.Observation) {
	ctx, span := tracer.Start(ctx, "createNotifications")
	defer span.End()
	var allNotifications []model.Notification
	for _, o := range obs {
		ctx, span := tracer.Start(ctx, "createNotificationFromObservation",
			trace.WithAttributes(
				attribute.String(constants.TraceKeyObservationUID, o.Uid),
			),
		)
		notifications, err := modron.notificationsFromObservation(ctx, o)
		if err != nil {
			log.Warnf("notifications from observation: %v", err)
			continue
		}
		allNotifications = append(allNotifications, notifications...)
		span.End()
	}
	span.SetAttributes(
		attribute.Int(constants.TraceKeyNumNotifications, len(allNotifications)),
	)
	log.Infof("Creating %d notifications in batch", len(allNotifications))
	_, err := modron.NotificationSvc.BatchCreateNotifications(ctx, allNotifications)
	if err != nil {
		log.Warnf("notifications: %v", err)
		span.RecordError(err)
	}
}

func (modron *Modron) collectAndScan(ctx context.Context, rgs []string, scanType pb.ScanType) (*pb.CollectAndScanResponse, error) {
	ctx = context.WithoutCancel(ctx)
	collectID, scanID := uuid.NewString(), uuid.NewString()
	ctx = context.WithValue(ctx, constants.CollectIDKey, collectID)
	ctx = context.WithValue(ctx, constants.ScanIDKey, scanID)
	ctx, span := tracer.Start(ctx, "collectAndScan",
		trace.WithAttributes(
			attribute.String(constants.TraceKeyCollectID, collectID),
			attribute.String(constants.TraceKeyScanID, scanID),
			attribute.String(constants.TraceKeyScanType, scanType.String()),
			attribute.StringSlice(constants.TraceKeyResourceGroupNames, rgs),
		),
	)
	defer span.End()

	switch scanType {
	case pb.ScanType_SCAN_TYPE_PARTIAL:
	// We use rgs
	case pb.ScanType_SCAN_TYPE_FULL:
		var err error
		if rgs, err = modron.Collector.ListResourceGroupNames(ctx); err != nil {
			return nil, fmt.Errorf("list resource groups: %w", err)
		}
	}

	obsChan := make(chan []*pb.Observation)
	go func(resourceGroupNames []string) {
		ctx, span := tracer.Start(ctx, "collectAndScanAsync", trace.WithNewRoot())
		defer span.End()

		preCollectedRgs, err := modron.preCollect(ctx, resourceGroupNames)
		if err != nil {
			log.Errorf("pre-collect failed: %v", err)
			modron.StateManager.EndCollect(collectID, resourceGroupNames)
			modron.StateManager.EndScan(scanID, resourceGroupNames)
			span.SetStatus(otelcodes.Error, err.Error())
			span.End()
			return
		}

		obsChan <- modron.collect(ctx, resourceGroupNames, preCollectedRgs)
		obsChan <- modron.scan(ctx, resourceGroupNames, preCollectedRgs)
		close(obsChan)
	}(rgs)
	go func() {
		// Process observations / notifications
		for v := range obsChan {
			if v == nil {
				continue
			}
			ctx, span := tracer.Start(ctx, "processObservations",
				trace.WithNewRoot(),
				trace.WithLinks(trace.LinkFromContext(ctx)),
				trace.WithAttributes(
					attribute.Int(constants.TraceKeyNumObservations, len(v)),
				),
			)
			modron.createNotifications(ctx, v)
			span.End()
		}
	}()
	return &pb.CollectAndScanResponse{
		CollectId: collectID,
		ScanId:    scanID,
	}, nil
}

// getCollectedObservations retrieves all the observations collected during the collection phase
// so that we can use them to create notifications
func (modron *Modron) getCollectedObservations(ctx context.Context, collectID string, groups []string) ([]*pb.Observation, error) {
	obs, err := modron.Storage.ListObservations(ctx, model.StorageFilter{
		OperationID:        collectID,
		ResourceGroupNames: groups,
	})
	if err != nil {
		return nil, fmt.Errorf("list observations: %w", err)
	}
	return obs, nil
}

func (modron *Modron) CollectAndScan(ctx context.Context, req *pb.CollectAndScanRequest) (*pb.CollectAndScanResponse, error) {
	return modron.collectAndScan(ctx, req.ResourceGroupNames, pb.ScanType_SCAN_TYPE_PARTIAL)
}

func (modron *Modron) CollectAndScanAll(ctx context.Context, _ *pb.CollectAndScanAllRequest) (*pb.CollectAndScanResponse, error) {
	return modron.collectAndScan(ctx, nil, pb.ScanType_SCAN_TYPE_FULL)
}

func (modron *Modron) ScheduledRunner(ctx context.Context) {
	interval := modron.CollectAndScanInterval
	log.Tracef("starting scheduler with interval %v", interval)
	for {
		ctx, span := tracer.Start(ctx, "ScheduledRunner", trace.WithNewRoot())
		log.Infof("scan scheduler: starting")
		if ctx.Err() != nil {
			log.Errorf("scan scheduler: %v", ctx.Err())
			return
		}
		r, err := modron.collectAndScan(ctx, nil, pb.ScanType_SCAN_TYPE_FULL)
		if err != nil {
			log.Errorf("scan scheduler: %v", err)
		}
		log.Infof("scan scheduler done: collectionID: %s, scanID: %s", r.CollectId, r.ScanId)
		span.End()
		time.Sleep(interval)
	}
}

func (modron *Modron) ListObservations(ctx context.Context, in *pb.ListObservationsRequest) (*pb.ListObservationsResponse, error) {
	groups, err := modron.validateResourceGroupNames(ctx, in.ResourceGroupNames)
	if err != nil {
		return nil, err
	}
	obsByGroupByRules := map[string]map[string][]*pb.Observation{}
	oneWeekAgo := time.Now().Add(-oneWeek)
	obs, err := modron.Storage.ListObservations(ctx, model.StorageFilter{
		ResourceGroupNames: groups,
		StartTime:          oneWeekAgo,
		TimeOffset:         time.Since(oneWeekAgo),
	})
	if err != nil {
		log.Warnf("list observations: %v", err)
		return nil, status.Error(codes.Internal, "failed listing observations")
	}
	for _, group := range groups {
		obsByGroupByRules[group] = map[string][]*pb.Observation{}
		for _, rule := range modron.RuleEngine.GetRules() {
			obsByGroupByRules[group][rule.Info().Name] = []*pb.Observation{}
		}
	}
	for _, ob := range obs {
		group := ob.ResourceRef.GroupName
		// TODO: Remove in the future when all of our observations do not use this field anymore:
		ob.DeprecatedResource = nil
		rule := ob.Name
		if obsByGroupByRules[group] == nil {
			obsByGroupByRules[group] = map[string][]*pb.Observation{}
		}
		obsByGroupByRules[group][rule] = append(
			obsByGroupByRules[group][rule],
			ob,
		)
	}
	var res []*pb.ResourceGroupObservationsPair
	keys := maps.Keys(obsByGroupByRules)
	sort.Strings(keys)
	for _, name := range keys {
		ruleObs := obsByGroupByRules[name]
		var val []*pb.RuleObservationPair
		ruleObsKeys := maps.Keys(ruleObs)
		sort.Strings(ruleObsKeys)
		for _, key := range ruleObsKeys {
			obs := ruleObs[key]
			val = append(val, &pb.RuleObservationPair{
				Rule:         key,
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

func (modron *Modron) CreateObservation(ctx context.Context, in *pb.CreateObservationRequest) (*pb.Observation, error) {
	if in.Observation == nil {
		return nil, status.Error(codes.InvalidArgument, "observation is nil")
	}
	if in.Observation.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "observation does not have a name")
	}
	if in.Observation.ResourceRef == nil {
		return nil, status.Error(codes.InvalidArgument, "resource to link observation with not defined")
	}
	if in.Observation.Remediation == nil || in.Observation.Remediation.Recommendation == "" {
		return nil, status.Error(codes.InvalidArgument, "cannot create an observation without recommendation")
	}
	in.Observation.Timestamp = timestamppb.Now()
	externalID := ""
	if in.Observation.ResourceRef.ExternalId != nil {
		externalID = *in.Observation.ResourceRef.ExternalId
	}
	res, err := modron.Storage.ListResources(ctx, model.StorageFilter{ResourceNames: []string{externalID}})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "resource to link observation to not found: %v", err)
	}
	if len(res) != 1 {
		return nil, status.Errorf(codes.FailedPrecondition, "found %d resources matching %+v", len(res), in.Observation.ResourceRef)
	}
	in.Observation.ResourceRef = utils.GetResourceRef(res[0])
	in.Observation.Uid = uuid.NewString()
	obs, err := modron.Storage.BatchCreateObservations(ctx, []*pb.Observation{in.Observation})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if len(obs) != 1 {
		return nil, status.Errorf(codes.Internal, "creation returned %d items", len(obs))
	}
	return obs[0], nil
}

func (modron *Modron) GetStatusCollectAndScan(_ context.Context, in *pb.GetStatusCollectAndScanRequest) (*pb.GetStatusCollectAndScanResponse, error) {
	collectStatus := modron.StateManager.GetCollectState(in.CollectId)
	scanStatus := modron.StateManager.GetScanState(in.ScanId)
	return &pb.GetStatusCollectAndScanResponse{
		CollectStatus: collectStatus,
		ScanStatus:    scanStatus,
	}, nil
}

func (modron *Modron) GetNotificationException(ctx context.Context, req *pb.GetNotificationExceptionRequest) (*pb.NotificationException, error) {
	ex, err := modron.validateUserAndGetException(ctx, req.Uuid)
	if err != nil {
		return nil, err
	}
	return ex.ToProto(), err
}

func (modron *Modron) CreateNotificationException(ctx context.Context, req *pb.CreateNotificationExceptionRequest) (*pb.NotificationException, error) {
	userEmail, err := modron.Checker.GetValidatedUser(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed authenticating request")
	}
	req.Exception.UserEmail = userEmail
	ex, err := modron.NotificationSvc.CreateException(ctx, model.ExceptionFromProto(req.Exception))
	if err != nil {
		return nil, err
	}
	return ex.ToProto(), err
}

func (modron *Modron) UpdateNotificationException(ctx context.Context, req *pb.UpdateNotificationExceptionRequest) (*pb.NotificationException, error) {
	ex, err := modron.validateUserAndGetException(ctx, req.Exception.Uuid)
	if err != nil {
		return nil, err
	}
	req.Exception.UserEmail = ex.UserEmail
	ex, err = modron.NotificationSvc.UpdateException(ctx, model.ExceptionFromProto(req.Exception))
	if err != nil {
		return nil, err
	}
	return ex.ToProto(), err
}

func (modron *Modron) DeleteNotificationException(ctx context.Context, req *pb.DeleteNotificationExceptionRequest) (*emptypb.Empty, error) {
	if _, err := modron.validateUserAndGetException(ctx, req.Uuid); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, modron.NotificationSvc.DeleteException(ctx, req.Uuid)
}

func (modron *Modron) ListNotificationExceptions(ctx context.Context, req *pb.ListNotificationExceptionsRequest) (*pb.ListNotificationExceptionsResponse, error) {
	userEmail, err := modron.Checker.GetValidatedUser(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed authenticating request")
	}
	req.UserEmail = userEmail
	ex, err := modron.NotificationSvc.ListExceptions(ctx, req.UserEmail, req.PageSize, req.PageToken)
	var exList []*pb.NotificationException
	for _, e := range ex {
		exList = append(exList, e.ToProto())
	}
	return &pb.ListNotificationExceptionsResponse{Exceptions: exList}, err
}

// TODO: Allow security admins to bypass the checks
func (modron *Modron) validateUserAndGetException(ctx context.Context, notificationUUID string) (model.Exception, error) {
	userEmail, err := modron.Checker.GetValidatedUser(ctx)
	if err != nil {
		return model.Exception{}, status.Error(codes.Unauthenticated, "failed authenticating request")
	}
	ex, err := modron.NotificationSvc.GetException(ctx, notificationUUID)
	if err != nil {
		return model.Exception{}, err
	}
	if ex.UserEmail != userEmail {
		return model.Exception{}, status.Error(codes.Unauthenticated, "failed authenticating request")
	}
	return ex, nil
}

func (modron *Modron) notificationsFromObservation(ctx context.Context, ob *pb.Observation) ([]model.Notification, error) {
	log := log.WithFields(logrus.Fields{
		constants.LogKeyResourceGroup:   ob.ResourceRef.GroupName,
		constants.LogKeyObservationUID:  ob.Uid,
		constants.LogKeyObservationName: ob.Name,
	})
	ty, err := utils.TypeFromResource(&pb.Resource{Type: &pb.Resource_ResourceGroup{}})
	if err != nil {
		return nil, fmt.Errorf("type from resource: %w", err)
	}
	collectionID, ok := ctx.Value(constants.CollectIDKey).(string)
	if !ok {
		return nil, fmt.Errorf("collectID not found in context")
	}
	rg, err := modron.Storage.ListResources(ctx, model.StorageFilter{
		ResourceNames: []string{ob.ResourceRef.GroupName},
		OperationID:   collectionID,
		ResourceTypes: []string{ty},
		Limit:         1,
	})
	if err != nil {
		return nil, err
	}
	if len(rg) < 1 {
		log.Errorf("no resource found")
		return nil, fmt.Errorf("no resource found %+v", ob.ResourceRef.Uid)
	}
	if len(rg) > 1 {
		log.Warnf("multiple resources group found for %+v, using the first one", ob.ResourceRef.Uid)
	}
	// We can have the same contacts in owners and labels, de-duplicate.
	contacts := modron.contactsFromRG(rg[0])
	if len(contacts) < 1 {
		log.Errorf("no contacts found for observation")
		return nil, fmt.Errorf("no contacts found for observation %q, resource group: %q", ob.Uid, ob.ResourceRef.GroupName)
	}
	notifications := make([]model.Notification, 0)
	for _, c := range contacts {
		if c == "" {
			log.Warnf("empty contact")
			continue
		}
		notifications = append(notifications,
			nagatha.NotificationFromObservation(c, modron.NotificationInterval, ob),
		)
	}
	return notifications, nil
}

func (modron *Modron) contactsFromRG(rg *pb.Resource) []string {
	uniqueContacts := make(map[string]struct{})
	if rg.IamPolicy != nil {
		for _, b := range rg.IamPolicy.Permissions {
			theRole := constants.ToRole(b.Role)
			_, isAdminRole := constants.AdminRoles[theRole]
			_, isAdditionalAdminRole := modron.additionalAdminRolesMap[theRole]
			if isAdminRole || isAdditionalAdminRole {
				for _, m := range b.Principals {
					if strings.HasSuffix(m, modron.OrgSuffix) {
						uniqueContacts[strings.TrimPrefix(m, constants.GCPUserAccountPrefix)] = struct{}{}
					}
				}
			}
		}
	}
	log.Debugf("contacts from IAM policy: %v", uniqueContacts)

	contact1, ok := rg.Labels[constants.LabelContact1]
	if ok && contact1 != "" {
		uniqueContacts[modron.LabelToEmail(contact1)] = struct{}{}
	}
	contact2, ok := rg.Labels[constants.LabelContact2]
	if ok && contact2 != "" {
		uniqueContacts[modron.LabelToEmail(contact2)] = struct{}{}
	}

	contacts := maps.Keys(uniqueContacts)
	return contacts
}

// LabelToEmail converts a contact1,contact2 label into an email address
// these labels are formatted as firstname_lastname_example_com, which is the representation of
// firstname.lastname@example.com.
// Due to the _ replacement, we do not support emails like noreply_test@example.com.
func (modron *Modron) LabelToEmail(contact string) string {
	contact = modron.labelToEmailRegexp.ReplaceAllString(contact, modron.labelToEmailSubst)
	contact = strings.ReplaceAll(contact, "_", ".")
	return contact
}

func (modron *Modron) initMetrics() error {
	collectDurationHist, err := meter.Float64Histogram(
		constants.MetricsPrefix+"collections_duration",
		metric.WithDescription("Duration of the collection process"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}
	observationsCounter, err := meter.Int64Counter(
		constants.MetricsPrefix+"observations_total",
		metric.WithDescription("Total number of observations created"),
	)
	if err != nil {
		return err
	}
	scanDurationHist, err := meter.Float64Histogram(
		constants.MetricsPrefix+"scan_duration",
		metric.WithDescription("Duration of the scan process"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}
	modron.metrics = metrics{
		CollectDuration: collectDurationHist,
		Observations:    observationsCounter,
		ScanDuration:    scanDurationHist,
	}
	return err
}

func New(
	checker model.Checker,
	collectAndScanInterval time.Duration,
	coll model.Collector,
	notificationInterval time.Duration,
	svc model.NotificationService,
	suffix string,
	engine *engine.RuleEngine,
	url string,
	manager model.StateManager,
	st model.Storage,
	additionalAdminRoles map[constants.Role]struct{},
	labelToEmailRegexp *regexp.Regexp,
	labelToEmailSubst string,
) (*Modron, error) {
	s := Modron{
		Checker:                 checker,
		CollectAndScanInterval:  collectAndScanInterval,
		Collector:               coll,
		NotificationInterval:    notificationInterval,
		NotificationSvc:         svc,
		OrgSuffix:               suffix,
		RuleEngine:              engine,
		SelfURL:                 url,
		StateManager:            manager,
		Storage:                 st,
		additionalAdminRolesMap: additionalAdminRoles,
		labelToEmailRegexp:      labelToEmailRegexp,
		labelToEmailSubst:       labelToEmailSubst,
	}
	err := s.initMetrics()
	return &s, err
}
