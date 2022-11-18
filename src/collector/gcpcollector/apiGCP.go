package gcpcollector

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/api/apikeys/v2"
	cloudasset "google.golang.org/api/cloudasset/v1p1beta1"
	"google.golang.org/api/cloudidentity/v1"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/spanner/v1"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	"google.golang.org/api/storage/v1"
)

type GCPApi interface {
	ListAllResourceGroups(ctx context.Context) ([]*cloudresourcemanager.Project, error)
	ListApiKeys(name string) (*apikeys.V2ListKeysResponse, error)
	ListBackendServices(name string) (*compute.BackendServiceAggregatedList, error)
	ListBuckets(name string) (*storage.Buckets, error)
	ListBucketsIamPolicy(bucketId string) (*storage.Policy, error)
	ListCertificates(name string) (*compute.SslCertificateAggregatedList, error)
	ListCloudSqlDatabases(ctx context.Context, name string) ([]*sqladmin.DatabaseInstance, error)
	ListClustersByZone(name string, zone string) (*container.ListClustersResponse, error)
	ListInstances(name string) (*compute.InstanceAggregatedList, error)
	ListProjectIamPolicy(name string) (*cloudresourcemanager.Policy, error)
	ListRegions(name string) (*compute.RegionList, error)
	ListServiceAccount(name string) (*iam.ListServiceAccountsResponse, error)
	ListServiceAccountKeys(name string) (*iam.ListServiceAccountKeysResponse, error)
	ListSpannerDatabases(ctx context.Context, name string) ([]*spanner.Database, error)
	ListSubNetworksByRegion(name string, region string) (*compute.SubnetworkList, error)
	ListTargetHttpsProxies(name string) (*compute.TargetHttpsProxyAggregatedList, error)
	ListUrlMaps(name string) (*compute.UrlMapsAggregatedList, error)
	ListUsersInGroup(ctx context.Context, group string) ([]string, error)
	ListZones(name string) (*compute.ZoneList, error)
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
	spannerService              *spanner.Service
	sqlAdminService             *sqladmin.Service
	storageService              *storage.Service
}

func detailedGoogleError(e error, apiDetail string) error {
	if gErr, ok := e.(*googleapi.Error); ok {
		gErr.Message = fmt.Sprintf("%s: %s", apiDetail, gErr.Message)
		return gErr
	}
	return e
}

func NewGCPApiReal(ctx context.Context) (GCPApi, error) {
	apiKeyService, err := apikeys.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("apikeys.NewService error: %v", err)
	}
	cloudAssetService, err := cloudasset.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("cloudasset.NewService error: %v", err)
	}
	cloudIdentityService, err := cloudidentity.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("cloudidentity.NewService error: %v", err)
	}
	cloudresourcemanagerService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("cloudresourcemanager.NewService error: %v", err)
	}
	computeService, err := compute.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("compute.NewService error: %v", err)
	}
	containerService, err := container.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("container.NewService error: %v", err)
	}
	iamService, err := iam.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("iam.NewService error: %v", err)
	}
	spannerService, err := spanner.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("spanner.NewService: %v", err)
	}
	sqladminService, err := sqladmin.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("sqladmin.NewService: %v", err)
	}
	storageService, err := storage.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage.NewService error: %v", err)
	}

	return &GCPApiReal{
		apiKeyService:               apiKeyService,
		cloudAssetService:           cloudAssetService,
		cloudIdentityService:        cloudIdentityService,
		cloudresourcemanagerService: cloudresourcemanagerService,
		computeService:              computeService,
		containerService:            containerService,
		iamService:                  iamService,
		spannerService:              spannerService,
		sqlAdminService:             sqladminService,
		storageService:              storageService,
	}, nil
}

func (api *GCPApiReal) ListApiKeys(name string) (*apikeys.V2ListKeysResponse, error) {
	resp, err := api.apiKeyService.Projects.Locations.Keys.List(name).Do()
	if err != nil {
		return nil, detailedGoogleError(err, "ApiKey.List")
	}
	return resp, nil
}

func (api *GCPApiReal) ListAllResourceGroups(ctx context.Context) ([]*cloudresourcemanager.Project, error) {
	projects := []*cloudresourcemanager.Project{}
	err := api.cloudresourcemanagerService.Projects.List().Pages(ctx, func(lpr *cloudresourcemanager.ListProjectsResponse) error {
		for _, p := range lpr.Projects {
			if p.LifecycleState == "ACTIVE" {
				projects = append(projects, p)
			}
		}
		return nil
	})
	if err != nil {
		return nil, detailedGoogleError(err, "Projects.List")
	}
	return projects, nil
}

func (api *GCPApiReal) ListBuckets(name string) (*storage.Buckets, error) {
	resp, err := api.storageService.Buckets.List(name).Do()
	if err != nil {
		return nil, detailedGoogleError(err, "Buckets.List")
	}
	return resp, nil
}

func (api *GCPApiReal) ListBucketsIamPolicy(bucketId string) (*storage.Policy, error) {
	iamPolcies, err := api.storageService.Buckets.GetIamPolicy(bucketId).Do()
	if err != nil {
		return nil, detailedGoogleError(err, "Buckets.GetIamPolicy")
	}
	return iamPolcies, nil
}

func (api *GCPApiReal) ListProjectIamPolicy(name string) (*cloudresourcemanager.Policy, error) {
	resp, err := api.cloudresourcemanagerService.Projects.
		GetIamPolicy(name, new(cloudresourcemanager.GetIamPolicyRequest)).Do()
	if err != nil {
		return nil, detailedGoogleError(err, "Project.GetIamPolicy")
	}
	return resp, nil
}

func (api *GCPApiReal) ListZones(name string) (*compute.ZoneList, error) {
	resZones, err := api.computeService.Zones.List(name).Do()
	if err != nil {
		return nil, detailedGoogleError(err, "Zones.List")
	}
	return resZones, nil
}

func (api *GCPApiReal) ListClustersByZone(name string, zone string) (*container.ListClustersResponse, error) {
	clusters, err := api.containerService.Projects.Zones.Clusters.List(name, zone).Do()
	if err != nil {
		return nil, detailedGoogleError(err, "Clusters.List")
	}
	return clusters, nil
}

func (api *GCPApiReal) ListCertificates(name string) (*compute.SslCertificateAggregatedList, error) {
	certs, err := api.computeService.SslCertificates.AggregatedList(name).Do()
	if err != nil {
		return nil, detailedGoogleError(err, "SslCertificates.AggregatedList")
	}
	return certs, nil
}

func (api *GCPApiReal) ListTargetHttpsProxies(name string) (*compute.TargetHttpsProxyAggregatedList, error) {
	proxies, err := api.computeService.TargetHttpsProxies.AggregatedList(name).Do()
	if err != nil {
		return nil, detailedGoogleError(err, "TargetHttpsProxies.AggregatedList")
	}
	return proxies, nil

}

func (api *GCPApiReal) ListUrlMaps(name string) (*compute.UrlMapsAggregatedList, error) {
	maps, err := api.computeService.UrlMaps.AggregatedList(name).Do()
	if err != nil {
		return nil, detailedGoogleError(err, "UrlMaps.AggregatedList")
	}
	return maps, nil
}

func (api *GCPApiReal) ListBackendServices(name string) (*compute.BackendServiceAggregatedList, error) {
	backendServices, err := api.computeService.BackendServices.AggregatedList(name).Do()
	if err != nil {
		return nil, detailedGoogleError(err, "BackendServices.AggregatedList")
	}
	return backendServices, nil
}

func (api *GCPApiReal) ListRegions(name string) (*compute.RegionList, error) {
	resRegions, err := api.computeService.Regions.List(name).Do()
	if err != nil {
		return nil, detailedGoogleError(err, "Regions.List")
	}
	return resRegions, nil
}

func (api *GCPApiReal) ListSubNetworksByRegion(name string, region string) (*compute.SubnetworkList, error) {
	subNetworks, err := api.computeService.Subnetworks.List(name, region).Do()
	if err != nil {
		return nil, detailedGoogleError(err, "Subnetworks.List")
	}
	return subNetworks, nil
}

func (api *GCPApiReal) ListServiceAccount(name string) (*iam.ListServiceAccountsResponse, error) {
	serviceAccounts, err := api.iamService.Projects.ServiceAccounts.List(name).Do()
	if err != nil {
		return nil, detailedGoogleError(err, "ServiceAccounts.List")
	}
	return serviceAccounts, nil
}

func (api *GCPApiReal) ListInstances(name string) (*compute.InstanceAggregatedList, error) {
	resInstances, err := api.computeService.Instances.AggregatedList(name).Do()
	if err != nil {
		return nil, detailedGoogleError(err, "Instances.AggregatedList")
	}
	return resInstances, nil
}

func (api *GCPApiReal) ListServiceAccountKeys(name string) (*iam.ListServiceAccountKeysResponse, error) {
	resp, err := api.iamService.Projects.ServiceAccounts.Keys.List(name).Do()
	if err != nil {
		return nil, detailedGoogleError(err, "ServiceAccounts.Keys.List")
	}
	return resp, nil
}

func (api *GCPApiReal) ListSpannerDatabases(ctx context.Context, name string) ([]*spanner.Database, error) {
	dbs := []*spanner.Database{}
	instanceSvc := spanner.NewProjectsInstancesService(api.spannerService)
	err := instanceSvc.List(name).Pages(ctx, func(lir *spanner.ListInstancesResponse) error {
		for _, instance := range lir.Instances {
			databaseSvc := spanner.NewProjectsInstancesDatabasesService(api.spannerService)
			return databaseSvc.List(instance.Name).Pages(ctx, func(ldr *spanner.ListDatabasesResponse) error {
				dbs = append(dbs, ldr.Databases...)
				return nil
			})
		}
		return nil
	})
	if err != nil {
		return nil, detailedGoogleError(err, "Spanner.Instances.List")
	}
	return dbs, nil
}

func (api *GCPApiReal) ListCloudSqlDatabases(ctx context.Context, name string) ([]*sqladmin.DatabaseInstance, error) {
	instances := []*sqladmin.DatabaseInstance{}
	instanceSvc := sqladmin.NewInstancesService(api.sqlAdminService)
	err := instanceSvc.List(name).Pages(ctx, func(ilr *sqladmin.InstancesListResponse) error {
		instances = append(instances, ilr.Items...)
		return nil
	})
	if err != nil {
		return nil, detailedGoogleError(err, "SqlAdmin.Instances.List")
	}
	return instances, nil
}

func (api *GCPApiReal) ListUsersInGroup(ctx context.Context, group string) ([]string, error) {
	group = strings.TrimPrefix(group, "group:")
	groupId, err := api.cloudIdentityService.Groups.Lookup().GroupKeyId(group).Do()
	if err != nil {
		return []string{}, fmt.Errorf("group lookup: %v", err)
	}
	groupMembers := []string{}
	if err := api.cloudIdentityService.Groups.Memberships.List(groupId.Name).Pages(ctx, func(lmr *cloudidentity.ListMembershipsResponse) error {
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
	}); err != nil {
		return []string{}, err
	}
	return groupMembers, nil
}

func (api *GCPApiReal) SearchIamPolicy(ctx context.Context, scope string, query string) ([]*cloudasset.IamPolicySearchResult, error) {
	//resp, err := api.cloudAssetService.IamPolicies.SearchAll(scope).Query(query).Do()
	results := []*cloudasset.IamPolicySearchResult{}
	err := api.cloudAssetService.IamPolicies.SearchAll(scope).Query(query).Pages(ctx,
		func(resp *cloudasset.SearchAllIamPoliciesResponse) error {
			results = append(results, resp.Results...)
			return nil
		})
	if err != nil {
		return nil, detailedGoogleError(err, fmt.Sprintf("IamPolicies.SearchAll.Query %q", query))
	}
	return results, nil
}
