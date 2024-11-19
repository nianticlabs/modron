package gcpcollector

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/risk"
	"github.com/nianticlabs/modron/src/utils"
)

const (
	maxParallelCollections = 50
	uuidGenRetries         = 3
)

var (
	log                        = logrus.StandardLogger().WithField(constants.LogKeyPkg, "gcpcollector")
	sysGcpProjectRegex         = regexp.MustCompile("^sys-[0-9]+")
	tracer                     = otel.Tracer("github.com/nianticlabs/modron/src/collector/gcpcollector")
	validGcpResourceGroupRegex = regexp.MustCompile(`((organizations|folders)/\d+|(projects/[-a-z0-9.:]+))`)
)

type GCPCollector struct {
	allowedSccCategories    map[string]struct{}
	additionalAdminRolesMap map[constants.Role]struct{}
	api                     GCPApi
	orgID                   string
	orgSuffix               string
	storage                 model.Storage
	tagConfig               risk.TagConfig

	metrics metrics
}

func (collector *GCPCollector) CollectAndStoreAll(ctx context.Context, collectID string, resourceGroupNames []string, preCollectedRgs []*pb.Resource) error {
	ctx, span := tracer.Start(ctx, "CollectAndStoreAll")
	defer span.End()
	resourceGroupNames = filterValidResourceGroupNames(resourceGroupNames)
	collectAndStoreSemaphore := make(chan struct{}, maxParallelCollections)
	errGroup := new(errgroup.Group)
	for _, rgName := range resourceGroupNames {
		collectAndStoreSemaphore <- struct{}{}
		errGroup.Go(func() error {
			ctx, span := tracer.Start(ctx, "CollectAndStoreAllInRg",
				trace.WithNewRoot(),
				trace.WithLinks(trace.Link{SpanContext: trace.SpanContextFromContext(ctx)}),
				trace.WithAttributes(attribute.String(constants.TraceKeyResourceGroup, rgName)),
			)
			defer span.End()
			defer func() { <-collectAndStoreSemaphore }()
			log := log.WithField("resource_group", rgName)
			if err := collector.collectAndStoreAllInRg(ctx, collectID, rgName, preCollectedRgs); err != nil {
				span.RecordError(err)
				log.WithError(err).Errorf("Failed to collect and store for resource group %s: %v", rgName, err)
				return err
			}
			return nil
		})
	}
	return errGroup.Wait()
}

func New(
	ctx context.Context,
	storage model.Storage,
	orgID string,
	orgSuffix string,
	additionalAdminRoles []string,
	tagConfig risk.TagConfig,
	allowedSccCategories []string,
) (model.Collector, error) {
	if strings.HasPrefix(orgID, constants.GCPOrgIDPrefix) {
		strippedOrgID := strings.TrimPrefix(orgID, constants.GCPOrgIDPrefix)
		log.Warnf("orgID \"%s\" is deprecated, use \"%s\" instead", orgID, strippedOrgID)
		orgID = strippedOrgID
	}
	api, err := NewGCPApiReal(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("container.NewService: %w", err)
	}
	m := initMetrics()
	additionalAdminRolesMap := map[constants.Role]struct{}{}
	for _, role := range additionalAdminRoles {
		additionalAdminRolesMap[constants.ToRole(role)] = struct{}{}
	}
	allowedSccCategoriesMap := map[string]struct{}{}
	for _, category := range allowedSccCategories {
		allowedSccCategoriesMap[category] = struct{}{}
	}
	return &GCPCollector{
		allowedSccCategories:    allowedSccCategoriesMap,
		api:                     api,
		additionalAdminRolesMap: additionalAdminRolesMap,
		storage:                 storage,
		orgID:                   orgID,
		orgSuffix:               orgSuffix,
		metrics:                 m,
		tagConfig:               tagConfig,
	}, nil
}

func filterValidResourceGroupNames(resourceGroupNames []string) (filteredNames []string) {
	for _, name := range resourceGroupNames {
		if sysGcpProjectRegex.MatchString(name) {
			continue
		}
		if !validGcpResourceGroupRegex.MatchString(name) {
			log.Warnf("invalid resource group name: %q", name)
			continue
		}
		filteredNames = append(filteredNames, name)
	}
	return filteredNames
}

// collectAndStoreResources should not be used directly, use collectAndStoreAllInRg instead
func (collector *GCPCollector) collectAndStoreResources(ctx context.Context, collectID string, rgName string) []error {
	ctx, span := tracer.Start(ctx, "collectAndStoreResources")
	span.SetAttributes(
		attribute.String(constants.TraceKeyCollectID, collectID),
		attribute.String(constants.TraceKeyResourceGroup, rgName),
	)
	defer span.End()
	rg, err := collector.GetResourceGroupWithIamPolicy(ctx, collectID, rgName)
	if err != nil {
		return []error{err}
	}
	resources, errArr := collector.ListResourceGroupResources(ctx, collectID, rgName)
	resources = append(resources, rg)
	if _, err := collector.storage.BatchCreateResources(ctx, resources); err != nil {
		errArr = append(errArr, err)
	}
	log.Infof("%s found %+v resources", rgName, len(resources))
	return errArr
}

func (collector *GCPCollector) collectAndStoreObservations(ctx context.Context, collectID string, rgName string, preCollectedRgs []*pb.Resource) []error {
	ctx, span := tracer.Start(ctx, "collectAndStoreObservations")
	span.SetAttributes(
		attribute.String(constants.TraceKeyCollectID, collectID),
		attribute.String(constants.TraceKeyResourceGroup, rgName),
	)
	defer span.End()
	rg, err := collector.GetResourceGroupWithIamPolicy(ctx, collectID, rgName)
	if err != nil {
		return []error{err}
	}
	obs, errArr := collector.ListResourceGroupObservations(ctx, collectID, rgName)
	if len(errArr) > 0 {
		return errArr
	}

	rgHierarchy, err := utils.ComputeRgHierarchy(preCollectedRgs)
	if err != nil {
		return []error{fmt.Errorf("computeRgHierarchy: %w", err)}
	}

	impact, reason := risk.GetImpact(collector.tagConfig, rgHierarchy, rgName)
	for k, o := range obs {
		obs[k].Impact = impact
		obs[k].ImpactReason = reason
		obs[k].RiskScore = risk.GetRiskScore(impact, o.Severity)
	}
	if _, err = collector.storage.BatchCreateObservations(ctx, obs); err != nil {
		errArr = append(errArr, err)
	}
	log.Infof("%s found %+v observations", rg.Name, len(obs))
	collector.logCollectionStatus(ctx, collectID, rgName, pb.Operation_COMPLETED, "")
	return errArr
}

func (collector *GCPCollector) ListResourceGroupResources(ctx context.Context, collectID string, rgName string) ([]*pb.Resource, []error) {
	ctx, span := tracer.Start(ctx, "ListResourceGroupResources")
	defer span.End()
	span.SetAttributes(
		attribute.String(constants.TraceKeyCollectID, collectID),
		attribute.String(constants.TraceKeyResourceGroup, rgName),
	)
	projectCollectors := []GenericCollector[*pb.Resource]{
		collector.ListAPIKeys,
		collector.ListBuckets,
		collector.ListCloudSQLDatabases,
		collector.ListKubernetesClusters,
		collector.ListKubernetesNamespaces,
		collector.ListKubernetesPods,
		collector.ListLoadBalancers,
		collector.ListNetworks,
		collector.ListServiceAccounts,
		collector.ListSpannerDatabases,
		collector.ListVMInstances,
	}
	// TODO: Add organization collectors, if needed
	var organizationCollectors []GenericCollector[*pb.Resource]

	collectors, errArr := chooseCollectors(rgName, projectCollectors, organizationCollectors)
	var res []*pb.Resource
	resMutex := sync.Mutex{}
	errArrMutex := sync.Mutex{}
	wg := sync.WaitGroup{}
	for _, collector := range collectors {
		wg.Add(1)
		go func() {
			ctx, span := tracer.Start(ctx, "RunCollector")
			defer wg.Done()
			defer span.End()
			collValue := reflect.ValueOf(collector)
			functionName := runtime.FuncForPC(collValue.Pointer()).Name()
			log := log.WithField(constants.LogKeyCollector, functionName)
			span.SetAttributes(
				attribute.String(constants.TraceKeyCollector, functionName),
			)
			collectedResources, err := collector.ExponentialBackoffRun(ctx, rgName)
			if err != nil {
				log.WithError(err).Errorf("ExponentialBackoffRun: %v", err)
				span.RecordError(err)
				span.SetStatus(codes.Error, "ExponentialBackoffRun failed")
				errArrMutex.Lock()
				errArr = append(errArr, err)
				errArrMutex.Unlock()
			}
			for _, r := range collectedResources {
				r.CollectionUid = collectID
				r.Timestamp = timestamppb.Now()
				resMutex.Lock()
				res = append(res, r)
				resMutex.Unlock()
			}
		}()
	}
	wg.Wait()
	return res, errArr
}

func (collector *GCPCollector) ListResourceGroupObservations(ctx context.Context, collectID string, rgName string) (obs []*pb.Observation, errArr []error) {
	projectCollectors := []GenericCollector[*pb.Observation]{
		collector.ListSccFindings,
	}
	var organizationCollectors []GenericCollector[*pb.Observation]
	var collectors []GenericCollector[*pb.Observation]
	collectors, errArr = chooseCollectors(rgName, projectCollectors, organizationCollectors)
	for _, collector := range collectors {
		cLogger := log.
			WithFields(logrus.Fields{
				constants.LogKeyCollector:     fmt.Sprintf("%T", collector),
				constants.LogKeyResourceGroup: rgName,
			})
		cLogger.Info("Collecting observations")
		collectedObs, err := collector.ExponentialBackoffRun(ctx, rgName)
		if err != nil {
			errArr = append(errArr, err)
			cLogger.WithError(err).Errorf("Failed to collect some observations")
		} else {
			for _, o := range collectedObs {
				o.CollectionId = utils.RefOrNull(collectID)
				if o.Timestamp == nil {
					o.Timestamp = timestamppb.Now()
				}
				obs = append(obs, o)
			}
		}
		cLogger.Infof("Collection complete")
	}
	return
}

func chooseCollectors[T any](
	rgName string,
	projectCollectors []GenericCollector[T],
	orgCollectors []GenericCollector[T],
) (collectors []GenericCollector[T], errors []error) {
	switch {
	case strings.HasPrefix(rgName, constants.GCPFolderIDPrefix):
		collectors = []GenericCollector[T]{}
	case strings.HasPrefix(rgName, constants.GCPOrgIDPrefix):
		collectors = orgCollectors
	case strings.HasPrefix(rgName, constants.GCPProjectsNamePrefix):
		collectors = projectCollectors
	default:
		errors = append(errors, fmt.Errorf("no collectors for %q", rgName))
		return nil, errors
	}
	return
}

func (collector *GCPCollector) logCollectionStatus(ctx context.Context, collectID, resourceGroupName string, status pb.Operation_Status, reason string) {
	ctx, span := tracer.Start(ctx, "logCollectionStatus")
	defer span.End()
	log.
		WithField(constants.LogKeyCollectID, collectID).
		WithField(constants.LogKeyResourceGroup, resourceGroupName).
		Infof("Logging collection status: %s", status.String())
	if err := collector.storage.AddOperationLog(ctx,
		[]*pb.Operation{{
			Id:            collectID,
			ResourceGroup: resourceGroupName,
			Type:          "collection",
			StatusTime:    timestamppb.New(time.Now()),
			Status:        status,
			Reason:        reason,
		}}); err != nil {
		span.RecordError(err)
		log.Warnf("add operation log: %v", err)
	}
}

func (collector *GCPCollector) collectAndStoreAllInRg(ctx context.Context, collectID string, rgName string, preCollectedRgs []*pb.Resource) error {
	ctx, span := tracer.Start(ctx, "collectAndStoreAllInRg")
	defer span.End()
	span.SetAttributes(
		attribute.String(constants.TraceKeyCollectID, collectID),
		attribute.String(constants.TraceKeyResourceGroup, rgName),
	)
	collector.logCollectionStatus(ctx, collectID, rgName, pb.Operation_STARTED, "")
	collectLogger := log.
		WithFields(logrus.Fields{
			constants.LogKeyCollectID:     collectID,
			constants.LogKeyResourceGroup: rgName,
		})
	errGroup := new(errgroup.Group)
	errGroup.Go(func() error {
		collectLogger.Infof("Starting collect resources")
		collectErrs := collector.collectAndStoreResources(ctx, collectID, rgName)
		collectLogger.WithError(errors.Join(collectErrs...)).Infof("Done collecting resources")
		return errors.Join(collectErrs...)
	})

	errGroup.Go(func() error {
		collectLogger.Infof("Starting collect observations")
		collectErrs := collector.collectAndStoreObservations(ctx, collectID, rgName, preCollectedRgs)
		collectLogger.WithError(errors.Join(collectErrs...)).Infof("Done collecting observations")
		return errors.Join(collectErrs...)
	})
	flushOps := func() {
		// Flush the ops log after the collection is done
		if err := collector.storage.FlushOpsLog(ctx); err != nil {
			// If we cannot flush the ops log, it's not a big deal:
			// the next time Modron starts, the pending operations are marked as complete.
			collectLogger.Warnf("flush ops log: %v", err)
		}
	}
	defer flushOps()
	if err := errGroup.Wait(); err != nil {
		collectLogger.Errorf("collectAndStoreAllInRg: %v", err)
		collector.logCollectionStatus(ctx, collectID, rgName, pb.Operation_FAILED, err.Error())
		return err
	}
	collector.logCollectionStatus(ctx, collectID, rgName, pb.Operation_COMPLETED, "")
	return nil
}
