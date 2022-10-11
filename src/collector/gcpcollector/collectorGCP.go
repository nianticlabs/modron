package gcpcollector

import (
	"fmt"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

type GCPCollector struct {
	model.Collector
	api     GCPApi
	storage model.Storage
}

const (
	maxGoroutines = 1000
)

var (
	sysGcpProjectRegex   = regexp.MustCompile("^sys-[0-9]+")
	validGcpProjectRegex = regexp.MustCompile("^[a-z][-a-z0-9]{4,28}[a-z0-9]{1}$")

	orgId     string
	orgSuffix string
)

func New(ctx context.Context, storage model.Storage) (model.Collector, error) {
	api, err := NewGCPApiReal(ctx)
	if err != nil {
		return nil, fmt.Errorf("container.NewService error: %v", err)
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

func filterValidGcpProjectId(resourceGroupNames []string) []string {
	filteredNames := []string{}
	for _, name := range resourceGroupNames {
		if sysGcpProjectRegex.MatchString(name) {
			continue
		}
		if !validGcpProjectRegex.MatchString(name) {
			continue
		}
		filteredNames = append(filteredNames, name)
	}
	return filteredNames
}

func (collector *GCPCollector) CollectAndStoreAllResourceGroupResources(ctx context.Context, collectId string, resourceGroupNames []string) []error {
	resourceGroupNames = filterValidGcpProjectId(resourceGroupNames)
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
	error := []error{}
	for i, projectError := range errors {
		if len(projectError) != 0 {
			for _, e := range projectError {
				error = append(error, fmt.Errorf("ProjectId-%v %s error: %v", i, resourceGroupNames[i], e))
			}
		}
	}
	return error
}

func (collector *GCPCollector) CollectAndStoreResources(ctx context.Context, collectId string, resourceGroupID string) []error {
	collector.logCollectionStatus(ctx, collectId, resourceGroupID, model.OperationStarted)
	var resourceGroupCall CollectorResourceGroupCall = collector.GetResourceGroup
	resourceGroup, err := resourceGroupCall.ExponentialBackoffRun(ctx, collectId, resourceGroupID)
	if err != nil {
		collector.logCollectionStatus(ctx, collectId, resourceGroupID, model.OperationFailed)
		return []error{fmt.Errorf("GetResourceGroup error: %v", err)}
	}
	resources, errors := collector.ListResourceGroupResources(ctx, collectId, resourceGroup)
	if len(errors) > 0 {
		collector.logCollectionStatus(ctx, collectId, resourceGroupID, model.OperationFailed)
		return errors
	}
	if _, err = collector.storage.BatchCreateResources(ctx, append(resources, resourceGroup)); err != nil {
		errors = append(errors, err)
	}
	collector.logCollectionStatus(ctx, collectId, resourceGroupID, model.OperationCompleted)
	if err := collector.storage.FlushOpsLog(ctx); err != nil {
		glog.Warningf("flush ops log: %v", err)
	}
	return errors
}

// TODO: Use `SelfLink` instead of this
func formatResourceName(name string, id interface{}) string {
	if name == id {
		return name
	}
	return fmt.Sprintf("%v[%s]", id, name)
}

func (collector *GCPCollector) getNewUid() string {
	return common.GetUUID(3)
}

func (collector *GCPCollector) ListResourceGroupResources(ctx context.Context, collectId string, resourceGroup *pb.Resource) ([]*pb.Resource, []error) {
	collectors := []CollectorCall{
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

	res := []*pb.Resource{}
	errors := []error{}
	for i, collector := range collectors {
		collectedResources, err := collector.ExponentialBackoffRun(ctx, resourceGroup)
		if err != nil {
			errors = append(errors, fmt.Errorf("collector.%v error: %v", i, err))
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

func (collector *GCPCollector) logCollectionStatus(ctx context.Context, collectId, resourceGroupName string, status model.OperationStatus) {
	if err := collector.storage.AddOperationLog(ctx,
		[]model.Operation{{
			ID:            collectId,
			ResourceGroup: resourceGroupName,
			OpsType:       "collection",
			StatusTime:    time.Now(),
			Status:        status}}); err != nil {
		glog.Warningf("add operation log: %v", err)
	}
}
