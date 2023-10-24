package gcpcollector

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GCPCollector struct {
	model.Collector
	api     GCPApi
	storage model.Storage
}

const (
	maxGoroutines = 50
)

var (
	sysGcpProjectRegex   = regexp.MustCompile("^sys-[0-9]+")
	validGcpProjectRegex = regexp.MustCompile("^projects/[a-z][-a-z0-9]{4,28}[a-z0-9]{1}$")

	orgId     string
	orgSuffix string
)

func New(ctx context.Context, storage model.Storage) (model.Collector, error) {
	api, err := NewGCPApiReal(ctx)
	if err != nil {
		return nil, fmt.Errorf("container.NewService: %w", err)
	}
	if orgIdEnv := os.Getenv(constants.OrgIdEnvVar); orgIdEnv == "" {
		return nil, fmt.Errorf("environment variable %q is not set", constants.OrgIdEnvVar)
	} else {
		orgId = fmt.Sprintf("%s%s", constants.GCPOrgIdPrefix, orgIdEnv)
	}
	if orgSuffixEnv := os.Getenv(constants.OrgSuffixEnvVar); orgSuffixEnv == "" {
		return nil, fmt.Errorf("environment variable %q is not set", constants.OrgSuffixEnvVar)
	} else {
		orgSuffix = orgSuffixEnv
	}
	return &GCPCollector{
		api:     api,
		storage: storage,
	}, nil
}

func filterValidResourceGroupNames(resourceGroupNames []string) []string {
	filteredNames := []string{}
	for _, name := range resourceGroupNames {
		if sysGcpProjectRegex.MatchString(name) {
			continue
		}
		if !(strings.HasPrefix(name, constants.GCPFolderIdPrefix) || strings.HasPrefix(name, constants.GCPOrgIdPrefix)) &&
			!validGcpProjectRegex.MatchString(name) {
			continue
		}
		filteredNames = append(filteredNames, name)
	}
	return filteredNames
}

func (collector *GCPCollector) CollectAndStoreAllResourceGroupResources(ctx context.Context, collectId string, resourceGroupNames []string) []error {
	resourceGroupNames = filterValidResourceGroupNames(resourceGroupNames)
	errors := make([][]error, len(resourceGroupNames))
	guard := make(chan struct{}, maxGoroutines)
	wg := sync.WaitGroup{}
	for i, rgName := range resourceGroupNames {
		wg.Add(1)
		guard <- struct{}{}
		go func(index int, name string) {
			errors[index] = collector.CollectAndStoreResources(ctx, collectId, name)
			<-guard
			wg.Done()
		}(i, rgName)
	}
	wg.Wait()
	errs := []error{}
	for i, projectError := range errors {
		if len(projectError) != 0 {
			for _, e := range projectError {
				errs = append(errs, fmt.Errorf("project %s: %v", resourceGroupNames[i], e))
			}
		}
	}
	return errs
}

func (collector *GCPCollector) CollectAndStoreResources(ctx context.Context, collectId string, resourceGroupID string) []error {
	collector.logCollectionStatus(ctx, collectId, resourceGroupID, model.OperationStarted, "")
	var resourceGroupCall CollectorResourceGroupCall = collector.GetResourceGroup
	resourceGroup, err := resourceGroupCall.ExponentialBackoffRun(ctx, collectId, resourceGroupID)
	if err != nil {
		collector.logCollectionStatus(ctx, collectId, resourceGroupID, model.OperationFailed, err.Error())
		return []error{fmt.Errorf("GetResourceGroup: %w", err)}
	}
	resources, errors := collector.ListResourceGroupResources(ctx, collectId, resourceGroup)
	if _, err = collector.storage.BatchCreateResources(ctx, append(resources, resourceGroup)); err != nil {
		errors = append(errors, err)
	}
	glog.V(5).Infof("%s found %+v resources", resourceGroup.Name, len(resources))
	if err := collector.storage.FlushOpsLog(ctx); err != nil {
		glog.Warningf("flush ops log: %v", err)
	}
	collector.logCollectionStatus(ctx, collectId, resourceGroupID, model.OperationCompleted, "")
	return errors
}

// TODO: Make sure this is meaningful outside of projects
func formatResourceName(name string, id interface{}) string {
	if name == id {
		return name
	}
	switch id.(type) {
	case []string, []uint64, []uint, []int64, []int:
		return fmt.Sprintf("%s%v", name, id)
	default:
		return fmt.Sprintf("%s[%v]", name, id)
	}

}

func (collector *GCPCollector) ListResourceGroupResources(ctx context.Context, collectId string, resourceGroup *pb.Resource) ([]*pb.Resource, []error) {
	projectCollectors := []CollectorCall{
		collector.ListApiKeys,
		collector.ListBuckets,
		collector.ListCloudSqlDatabases,
		collector.ListKubernetesClusters,
		collector.ListLoadBalancers,
		collector.ListNetworks,
		collector.ListServiceAccounts,
		collector.ListSpannerDatabases,
		collector.ListVmInstances,
	}
	organizationCollectors := []CollectorCall{
		collector.ListGroups,
	}
	var collectors []CollectorCall
	errors := []error{}
	switch {
	case strings.HasPrefix(resourceGroup.Name, constants.GCPFolderIdPrefix):
		collectors = []CollectorCall{}
	case strings.HasPrefix(resourceGroup.Name, constants.GCPOrgIdPrefix):
		collectors = organizationCollectors
	case strings.HasPrefix(resourceGroup.Name, constants.GCPProjectsNamePrefix):
		collectors = projectCollectors
	default:
		errors = append(errors, fmt.Errorf("no collectors for %q", resourceGroup.Name))
		return nil, errors
	}
	res := []*pb.Resource{}
	for i, collector := range collectors {
		collectedResources, err := collector.ExponentialBackoffRun(ctx, resourceGroup)
		if err != nil {
			errors = append(errors, fmt.Errorf("collector.%v: %v", i, err))
		} else {
			for _, r := range collectedResources {
				r.CollectionUid = collectId
				r.Timestamp = timestamppb.Now()
				res = append(res, r)
			}
		}
	}
	return res, errors
}

func (collector *GCPCollector) logCollectionStatus(ctx context.Context, collectId, resourceGroupName string, status model.OperationStatus, reason string) {
	if err := collector.storage.AddOperationLog(ctx,
		[]model.Operation{{
			ID:            collectId,
			ResourceGroup: resourceGroupName,
			OpsType:       "collection",
			StatusTime:    time.Now(),
			Status:        status,
			Reason:        reason,
		}}); err != nil {
		glog.Warningf("add operation log: %v", err)
	}
}
