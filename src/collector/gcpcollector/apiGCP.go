package gcpcollector

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/api/apikeys/v2"
	cloudasset "google.golang.org/api/cloudasset/v1"
	"google.golang.org/api/cloudidentity/v1"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/monitoring/v3"
	"google.golang.org/api/spanner/v1"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	"google.golang.org/api/storage/v1"
	"github.com/nianticlabs/modron/src/constants"
)

const (
	projectAssetType        = "cloudresourcemanager.googleapis.com/Project"
	folderAssetType         = "cloudresourcemanager.googleapis.com/Folder"
	organizationAssetType   = "cloudresourcemanager.googleapis.com/Organization"
	excludeSysProjectsQuery = "NOT additionalAttributes.projectId:sys-"
	searchForProject        = "name=//cloudresourcemanager.googleapis.com/%s"
)

type GroupCacheEntry struct {
	Creation time.Time
	Members  []string
}

var (
	listGroupCache = map[string]GroupCacheEntry{}
)

type GCPApi interface {
	ListApiKeys(ctx context.Context, name string) ([]*apikeys.V2Key, error)
	ListBackendServices(ctx context.Context, name string) ([]*compute.BackendService, error)
	ListBuckets(ctx context.Context, name string) ([]*storage.Bucket, error)
	ListBucketsIamPolicy(bucketId string) (*storage.Policy, error)
	ListCertificates(ctx context.Context, name string) ([]*compute.SslCertificate, error)
	ListCloudSqlDatabases(ctx context.Context, name string) ([]*sqladmin.DatabaseInstance, error)
	ListClustersByZone(name string, zone string) ([]*container.Cluster, error)
	ListFoldersIamPolicy(name string) (*cloudresourcemanager.Policy, error)
	ListGroupMembers(ctx context.Context, group string) ([]*cloudidentity.Membership, error)
	ListGroups(ctx context.Context) ([]*cloudidentity.Group, error)
	ListInstances(ctx context.Context, name string) ([]*compute.Instance, error)
	ListOrganizationsIamPolicy(name string) (*cloudresourcemanager.Policy, error)
	ListProjectIamPolicy(name string) (*cloudresourcemanager.Policy, error)
	ListRegions(ctx context.Context, name string) ([]*compute.Region, error)
	ListResourceGroups(ctx context.Context, name string) ([]*cloudasset.ResourceSearchResult, error)
	ListServiceAccount(ctx context.Context, name string) ([]*iam.ServiceAccount, error)
	ListServiceAccountKeys(name string) ([]*iam.ServiceAccountKey, error)
	ListServiceAccountKeyUsage(ctx context.Context, resourceGroup string, request *monitoring.QueryTimeSeriesRequest) *monitoring.ProjectsTimeSeriesQueryCall
	ListSpannerDatabases(ctx context.Context, name string) ([]*spanner.Database, error)
	ListSslPolicies(ctx context.Context, name string) ([]*compute.SslPolicy, error)
	ListSubNetworksByRegion(ctx context.Context, name string, region string) ([]*compute.Subnetwork, error)
	ListTargetHttpsProxies(ctx context.Context, name string) ([]*compute.TargetHttpsProxy, error)
	ListTargetSslProxies(ctx context.Context, name string) ([]*compute.TargetSslProxy, error)
	ListUrlMaps(ctx context.Context, name string) ([]*compute.UrlMap, error)
	ListUsersInGroup(ctx context.Context, group string) ([]string, error)
	ListZones(ctx context.Context, name string) ([]*compute.Zone, error)
	SearchIamPolicy(ctx context.Context, scope string, query string) ([]*cloudasset.IamPolicySearchResult, error)
}

type GCPApiReal struct {
	GCPApi

	apiKeyService               *apikeys.Service
	cloudAssetService           *cloudasset.Service
	cloudIdentityService        *cloudidentity.Service
	cloudresourcemanagerService *cloudresourcemanager.Service
	computeService              *compute.Service
	containerService            *container.Service
	iamService                  *iam.Service
	monitoringService           *monitoring.Service
	spannerService              *spanner.Service
	sqlAdminService             *sqladmin.Service
	storageService              *storage.Service
}

func detailedGoogleError(e error, apiDetail string) error {
	if e == nil {
		return nil
	}
	if gErr, ok := e.(*googleapi.Error); ok {
		gErr.Message = fmt.Sprintf("%s: %s", apiDetail, gErr.Message)
		return gErr
	}
	return e
}

func NewGCPApiReal(ctx context.Context) (GCPApi, error) {
	apiKeyService, err := apikeys.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("apikeys.NewService : %w", err)
	}
	cloudAssetService, err := cloudasset.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("cloudasset.NewService : %w", err)
	}
	cloudIdentityService, err := cloudidentity.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("cloudidentity.NewService : %w", err)
	}
	cloudresourcemanagerService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("cloudresourcemanager.NewService : %w", err)
	}
	computeService, err := compute.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("compute.NewService : %w", err)
	}
	containerService, err := container.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("container.NewService : %w", err)
	}
	iamService, err := iam.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("iam.NewService : %w", err)
	}
	monitoringService, err := monitoring.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("monitoring.NewService: %w", err)
	}
	spannerService, err := spanner.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("spanner.NewService: %w", err)
	}
	sqladminService, err := sqladmin.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("sqladmin.NewService: %w", err)
	}
	storageService, err := storage.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage.NewService : %w", err)
	}

	return &GCPApiReal{
		apiKeyService:               apiKeyService,
		cloudAssetService:           cloudAssetService,
		cloudIdentityService:        cloudIdentityService,
		cloudresourcemanagerService: cloudresourcemanagerService,
		computeService:              computeService,
		containerService:            containerService,
		iamService:                  iamService,
		monitoringService:           monitoringService,
		spannerService:              spannerService,
		sqlAdminService:             sqladminService,
		storageService:              storageService,
	}, nil
}

func (api *GCPApiReal) ListApiKeys(ctx context.Context, name string) (apiKeys []*apikeys.V2Key, err error) {
	err = api.apiKeyService.Projects.Locations.Keys.List(constants.ResourceWithProjectsPrefix(name)).Pages(ctx, func(vlkr *apikeys.V2ListKeysResponse) error {
		apiKeys = append(apiKeys, vlkr.Keys...)
		return nil
	})
	glog.V(15).Infof("%s fetched %d api keys", name, len(apiKeys))
	return apiKeys, detailedGoogleError(err, "ApiKey.List")
}

func (api *GCPApiReal) ListResourceGroups(ctx context.Context, name string) (resourceGroups []*cloudasset.ResourceSearchResult, err error) {
	call := api.cloudAssetService.V1.SearchAllResources(orgId).AssetTypes(projectAssetType, folderAssetType, organizationAssetType)
	if name != "" {
		call = call.Query(fmt.Sprintf(searchForProject, name))
		glog.V(15).Infof("query: %q", fmt.Sprintf(searchForProject, name))
	} else {
		call = call.Query(excludeSysProjectsQuery)
		glog.V(15).Infof("query: %q", excludeSysProjectsQuery)
	}

	err = call.Pages(ctx, func(sarr *cloudasset.SearchAllResourcesResponse) error {
		for _, r := range sarr.Results {
			if r.State == "ACTIVE" {
				resourceGroups = append(resourceGroups, r)
			}
		}
		return nil
	})
	glog.V(15).Infof("%q fetched %d resourcegroups", name, len(resourceGroups))
	return resourceGroups, detailedGoogleError(err, "ListResourceGroups")
}

func (api *GCPApiReal) ListBuckets(ctx context.Context, name string) (buckets []*storage.Bucket, err error) {
	err = api.storageService.Buckets.List(constants.ResourceWithoutProjectsPrefix(name)).Pages(ctx, func(b *storage.Buckets) error {
		buckets = append(buckets, b.Items...)
		return nil
	})
	glog.V(15).Infof("%s fetched %d buckets", name, len(buckets))
	return buckets, detailedGoogleError(err, "Buckets.List")
}

func (api *GCPApiReal) ListBucketsIamPolicy(bucketId string) (*storage.Policy, error) {
	iamPolcies, err := api.storageService.Buckets.GetIamPolicy(bucketId).Do()
	glog.V(15).Infof("fetched bucket %s iampolicy", bucketId)
	return iamPolcies, detailedGoogleError(err, "Buckets.GetIamPolicy")
}

func (api *GCPApiReal) ListProjectIamPolicy(name string) (*cloudresourcemanager.Policy, error) {
	resp, err := api.cloudresourcemanagerService.Projects.
		GetIamPolicy(constants.ResourceWithProjectsPrefix(name), &cloudresourcemanager.GetIamPolicyRequest{}).Do()
	glog.V(15).Infof("fetched project %s iam policy", name)
	return resp, detailedGoogleError(err, "Project.GetIamPolicy")
}

func (api *GCPApiReal) ListFoldersIamPolicy(name string) (*cloudresourcemanager.Policy, error) {
	resp, err := api.cloudresourcemanagerService.Folders.
		GetIamPolicy(name, new(cloudresourcemanager.GetIamPolicyRequest)).Do()
	glog.V(15).Infof("fetched folder %s iam policy", name)
	return resp, detailedGoogleError(err, "Folders.GetIamPolicy")
}

func (api *GCPApiReal) ListOrganizationsIamPolicy(name string) (*cloudresourcemanager.Policy, error) {
	resp, err := api.cloudresourcemanagerService.Organizations.GetIamPolicy(name, new(cloudresourcemanager.GetIamPolicyRequest)).Do()
	glog.V(15).Infof("fetched organization %s iam policy", name)
	return resp, detailedGoogleError(err, "Organizations.GetIamPolicy")
}

func (api *GCPApiReal) ListZones(ctx context.Context, name string) (zones []*compute.Zone, err error) {
	err = api.computeService.Zones.List(constants.ResourceWithoutProjectsPrefix(name)).Pages(ctx, func(zl *compute.ZoneList) error {
		zones = append(zones, zl.Items...)
		return nil
	})
	glog.V(15).Infof("%s fetched %d zones", name, len(zones))
	return zones, detailedGoogleError(err, "Zones.List")
}

func (api *GCPApiReal) ListClustersByZone(name string, zone string) (clusters []*container.Cluster, err error) {
	resp, err := api.containerService.Projects.Zones.Clusters.List(constants.ResourceWithoutProjectsPrefix(name), zone).Do()
	if err != nil {
		return nil, detailedGoogleError(err, "Clusters.List")
	}
	glog.V(15).Infof("%s fetched %d clusters", name, len(clusters))
	return resp.Clusters, nil
}

func (api *GCPApiReal) ListCertificates(ctx context.Context, name string) (certs []*compute.SslCertificate, err error) {
	err = api.computeService.SslCertificates.AggregatedList(constants.ResourceWithoutProjectsPrefix(name)).Pages(ctx, func(scal *compute.SslCertificateAggregatedList) error {
		for _, c := range scal.Items {
			certs = append(certs, c.SslCertificates...)
		}
		return nil
	})
	glog.V(15).Infof("%s fetched %d certificates", name, len(certs))
	return certs, detailedGoogleError(err, "SslCertificates.AggregatedList")
}

func (api *GCPApiReal) ListTargetHttpsProxies(ctx context.Context, name string) (proxies []*compute.TargetHttpsProxy, err error) {
	err = api.computeService.TargetHttpsProxies.AggregatedList(constants.ResourceWithoutProjectsPrefix(name)).Pages(ctx, func(thpal *compute.TargetHttpsProxyAggregatedList) error {
		for _, p := range thpal.Items {
			proxies = append(proxies, p.TargetHttpsProxies...)
		}
		return nil
	})
	glog.V(15).Infof("%s fetched %d https proxies", name, len(proxies))
	return proxies, detailedGoogleError(err, "TargetHttpsProxies.AggregatedList")
}

func (api *GCPApiReal) ListTargetSslProxies(ctx context.Context, name string) (proxies []*compute.TargetSslProxy, err error) {
	err = api.computeService.TargetSslProxies.List(constants.ResourceWithoutProjectsPrefix(name)).Pages(ctx, func(tspl *compute.TargetSslProxyList) error {
		proxies = append(proxies, tspl.Items...)
		return nil
	})
	glog.V(15).Infof("%s fetched %d ssl proxies", name, len(proxies))
	return proxies, detailedGoogleError(err, "TargetSslProxies.List")
}

func (api *GCPApiReal) ListSslPolicies(ctx context.Context, name string) (policies []*compute.SslPolicy, err error) {
	err = api.computeService.SslPolicies.AggregatedList(constants.ResourceWithoutProjectsPrefix(name)).Pages(ctx, func(spal *compute.SslPoliciesAggregatedList) error {
		for _, i := range spal.Items {
			policies = append(policies, i.SslPolicies...)
		}
		return nil
	})
	glog.V(15).Infof("%s fetched %d ssl policies", name, len(policies))
	return policies, detailedGoogleError(err, "SslPolicies,AggregatedList")
}

func (api *GCPApiReal) ListUrlMaps(ctx context.Context, name string) (maps []*compute.UrlMap, err error) {
	err = api.computeService.UrlMaps.AggregatedList(constants.ResourceWithoutProjectsPrefix(name)).Pages(ctx, func(umal *compute.UrlMapsAggregatedList) error {
		for _, m := range umal.Items {
			maps = append(maps, m.UrlMaps...)
		}
		return nil
	})
	glog.V(15).Infof("%s fetched %d url maps", name, len(maps))
	return maps, detailedGoogleError(err, "UrlMaps.AggregatedList")
}

func (api *GCPApiReal) ListBackendServices(ctx context.Context, name string) (backendSvcs []*compute.BackendService, err error) {
	err = api.computeService.BackendServices.AggregatedList(constants.ResourceWithoutProjectsPrefix(name)).Pages(ctx, func(bsal *compute.BackendServiceAggregatedList) error {
		for _, be := range bsal.Items {
			backendSvcs = append(backendSvcs, be.BackendServices...)
		}
		return nil
	})
	glog.V(15).Infof("%s fetched %d backend services", name, len(backendSvcs))
	return backendSvcs, detailedGoogleError(err, "BackendServices.AggregatedList")
}

func (api *GCPApiReal) ListRegions(ctx context.Context, name string) (regions []*compute.Region, err error) {
	err = api.computeService.Regions.List(constants.ResourceWithoutProjectsPrefix(name)).Pages(ctx, func(rl *compute.RegionList) error {
		regions = append(regions, rl.Items...)
		return nil
	})
	glog.V(15).Infof("%s fetched %d regions", name, len(regions))
	return regions, detailedGoogleError(err, "Regions.List")
}

func (api *GCPApiReal) ListSubNetworksByRegion(ctx context.Context, name string, region string) (subnetworks []*compute.Subnetwork, err error) {
	err = api.computeService.Subnetworks.List(constants.ResourceWithoutProjectsPrefix(name), region).Pages(ctx, func(sl *compute.SubnetworkList) error {
		subnetworks = append(subnetworks, sl.Items...)
		return nil
	})
	glog.V(15).Infof("%s region %s fetched %d subnetworks", name, region, len(subnetworks))
	return subnetworks, detailedGoogleError(err, "Subnetworks.List")
}

func (api *GCPApiReal) ListServiceAccount(ctx context.Context, name string) (serviceaccounts []*iam.ServiceAccount, err error) {
	err = api.iamService.Projects.ServiceAccounts.List(constants.ResourceWithProjectsPrefix(name)).Pages(ctx, func(lsar *iam.ListServiceAccountsResponse) error {
		serviceaccounts = append(serviceaccounts, lsar.Accounts...)
		return nil
	})
	glog.V(15).Infof("%s fetched %d service accounts", name, len(serviceaccounts))
	return serviceaccounts, detailedGoogleError(err, "ServiceAccounts.List")
}

func (api *GCPApiReal) ListServiceAccountKeyUsage(ctx context.Context, resourceGroup string, request *monitoring.QueryTimeSeriesRequest) *monitoring.ProjectsTimeSeriesQueryCall {
	return api.monitoringService.Projects.TimeSeries.Query(constants.ResourceWithProjectsPrefix(resourceGroup), request)
}

func (api *GCPApiReal) ListInstances(ctx context.Context, name string) (instances []*compute.Instance, err error) {
	err = api.computeService.Instances.AggregatedList(constants.ResourceWithoutProjectsPrefix(name)).Pages(ctx, func(ial *compute.InstanceAggregatedList) error {
		for _, ia := range ial.Items {
			instances = append(instances, ia.Instances...)
		}
		return nil
	})
	glog.V(15).Infof("%s fetched %d instances", name, len(instances))
	return instances, detailedGoogleError(err, "Instances.AggregatedList")
}

func (api *GCPApiReal) ListServiceAccountKeys(name string) (keys []*iam.ServiceAccountKey, err error) {
	resp, err := api.iamService.Projects.ServiceAccounts.Keys.List(constants.ResourceWithProjectsPrefix(name)).Do()
	if err != nil {
		return nil, detailedGoogleError(err, "ServiceAccounts.Keys.List")
	}
	keys = append(keys, resp.Keys...)
	glog.V(15).Infof("%s fetched %d service account keys", name, len(keys))
	return keys, nil
}

func (api *GCPApiReal) ListSpannerDatabases(ctx context.Context, name string) (dbs []*spanner.Database, err error) {
	instanceSvc := spanner.NewProjectsInstancesService(api.spannerService)
	err = instanceSvc.List(constants.ResourceWithoutProjectsPrefix(name)).Pages(ctx, func(lir *spanner.ListInstancesResponse) error {
		for _, instance := range lir.Instances {
			databaseSvc := spanner.NewProjectsInstancesDatabasesService(api.spannerService)
			return databaseSvc.List(instance.Name).Pages(ctx, func(ldr *spanner.ListDatabasesResponse) error {
				dbs = append(dbs, ldr.Databases...)
				return nil
			})
		}
		return nil
	})
	glog.V(15).Infof("%s fetched %d spanner databases", name, len(dbs))
	return dbs, detailedGoogleError(err, "Spanner.Instances.List")
}

func (api *GCPApiReal) ListCloudSqlDatabases(ctx context.Context, name string) (instances []*sqladmin.DatabaseInstance, err error) {
	instanceSvc := sqladmin.NewInstancesService(api.sqlAdminService)
	err = instanceSvc.List(constants.ResourceWithoutProjectsPrefix(name)).Pages(ctx, func(ilr *sqladmin.InstancesListResponse) error {
		instances = append(instances, ilr.Items...)
		return nil
	})
	glog.V(15).Infof("%s fetched %d cloudSql databases", name, len(instances))
	return instances, detailedGoogleError(err, "SqlAdmin.Instances.List")
}

func (api *GCPApiReal) ListGroups(ctx context.Context) (groups []*cloudidentity.Group, err error) {
	err = api.cloudIdentityService.Groups.List().Pages(ctx, func(lgr *cloudidentity.ListGroupsResponse) error {
		groups = append(groups, lgr.Groups...)
		return nil
	})
	glog.V(15).Infof("fetched %d groups", len(groups))
	return groups, detailedGoogleError(err, "CloudIdentity.Groups.List")
}

func (api *GCPApiReal) ListGroupMembers(ctx context.Context, group string) (groupMembers []*cloudidentity.Membership, err error) {
	group = strings.TrimPrefix(group, "group:")
	groupId, err := api.cloudIdentityService.Groups.Lookup().GroupKeyId(group).Do()
	if err != nil {
		return []*cloudidentity.Membership{}, fmt.Errorf("group lookup: %w", err)
	}
	err = api.cloudIdentityService.Groups.Memberships.List(groupId.Name).Pages(ctx, func(lmr *cloudidentity.ListMembershipsResponse) error {
		groupMembers = append(groupMembers, lmr.Memberships...)
		return nil
	})
	glog.V(15).Infof("%s fetched %d group members", group, len(groupMembers))
	return groupMembers, detailedGoogleError(err, "CloudIdentity.Groups.Memberships.List")
}

func (api *GCPApiReal) ListUsersInGroup(ctx context.Context, group string) (groupMembers []string, err error) {
	group = strings.TrimPrefix(group, "group:")
	if e, ok := listGroupCache[group]; ok && time.Since(e.Creation) < time.Duration(time.Second*300) {
		return e.Members, nil
	}
	groupId, err := api.cloudIdentityService.Groups.Lookup().GroupKeyId(group).Do()
	if err != nil {
		return []string{}, fmt.Errorf("group lookup: %w", err)
	}
	err = api.cloudIdentityService.Groups.Memberships.List(groupId.Name).Pages(ctx, func(lmr *cloudidentity.ListMembershipsResponse) error {
		for _, m := range lmr.Memberships {
			switch m.Type {
			case "GROUP":
				transitiveMembers, err := api.ListUsersInGroup(ctx, m.PreferredMemberKey.Id)
				if err != nil {
					return err
				}
				groupMembers = append(groupMembers, transitiveMembers...)
			default:
				groupMembers = append(groupMembers, m.PreferredMemberKey.Id)
			}
		}
		return nil
	})
	glog.V(15).Infof("%s fetched %d group members", group, len(groupMembers))
	listGroupCache[group] = GroupCacheEntry{Creation: time.Now(), Members: groupMembers}
	return groupMembers, detailedGoogleError(err, "CloudIdentity.Groups.Memberships.List")
}

func (api *GCPApiReal) SearchIamPolicy(ctx context.Context, scope string, query string) (iamPolicies []*cloudasset.IamPolicySearchResult, err error) {
	attemptNumber := 1
	waitSec := 0.
	start := time.Now()
	for {
		err := api.cloudAssetService.V1.SearchAllIamPolicies(scope).Query(query).Pages(ctx,
			func(resp *cloudasset.SearchAllIamPoliciesResponse) error {
				iamPolicies = append(iamPolicies, resp.Results...)
				return nil
			})
		if err == nil {
			break
		}
		if time.Since(start) > maxSecAcrossAttempts {
			return nil, fmt.Errorf("SearchIamPolicy.tooLong : after %v seconds %v ", time.Since(start), detailedGoogleError(err, fmt.Sprintf("IamPolicies.SearchAll.Query %q", query)))
		}
		if attemptNumber >= maxAttemptNumber {
			return nil, fmt.Errorf("SearchIamPolicy.maxAttempt : after %v  attempts %v ", attemptNumber, detailedGoogleError(err, fmt.Sprintf("IamPolicies.SearchAll.Query %q", query)))
		}
		if !isErrorCodeRetryable(getErrorCode(err)) {
			return nil, fmt.Errorf("SearchIamPolicy.notRetryableCode :  %v ", detailedGoogleError(err, fmt.Sprintf("IamPolicies.SearchAll.Query %q", query)))
		}
		waitSec = math.Min((math.Pow(2, float64(attemptNumber)) + rand.Float64()), maxSecBtwAttempts)
		time.Sleep(time.Duration(waitSec) * time.Second)
		attemptNumber += 1
	}
	glog.V(15).Infof("fetched %d iam policies", len(iamPolicies))
	return iamPolicies, nil
}
