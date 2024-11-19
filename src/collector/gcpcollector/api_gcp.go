package gcpcollector

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/net/context"

	// TODO: Use cloud.google.com packages
	"google.golang.org/api/apikeys/v2"
	"google.golang.org/api/cloudasset/v1"
	"google.golang.org/api/cloudidentity/v1"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/monitoring/v3"
	"google.golang.org/api/securitycenter/v1"
	"google.golang.org/api/spanner/v1"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	"google.golang.org/api/storage/v1"

	"github.com/nianticlabs/modron/src/constants"
)

const (
	projectAssetType        = "cloudresourcemanager.googleapis.com/Project"
	folderAssetType         = "cloudresourcemanager.googleapis.com/Folder"
	organizationAssetType   = "cloudresourcemanager.googleapis.com/Organization"
	k8sPodAssetType         = "k8s.io/Pod"
	k8sNamespaceAssetType   = "k8s.io/Namespace"
	excludeSysProjectsQuery = "NOT additionalAttributes.projectId:sys-"
	searchForProject        = `name="//cloudresourcemanager.googleapis.com/%s"`
)

type GroupCacheEntry struct {
	Creation time.Time
	Members  []string
}

var (
	listGroupCache = map[string]GroupCacheEntry{}
)

type GCPApi interface {
	GetServiceAccountIAMPolicy(ctx context.Context, name string) (*iam.Policy, error)
	ListAPIKeys(ctx context.Context, name string) ([]*apikeys.V2Key, error)
	ListBackendServices(ctx context.Context, name string) ([]*compute.BackendService, error)
	ListBuckets(ctx context.Context, name string) ([]*storage.Bucket, error)
	ListBucketsIamPolicy(bucketID string) (*storage.Policy, error)
	ListCertificates(ctx context.Context, name string) ([]*compute.SslCertificate, error)
	ListCloudSQLDatabases(ctx context.Context, name string) ([]*sqladmin.DatabaseInstance, error)
	ListClustersByZone(ctx context.Context, name string, zone string) ([]*container.Cluster, error)
	ListFoldersIamPolicy(ctx context.Context, name string) (*cloudresourcemanager.Policy, error)
	ListInstances(ctx context.Context, name string) ([]*compute.Instance, error)
	ListNamespaces(ctx context.Context, name string) ([]*cloudasset.ResourceSearchResult, error)
	ListPods(ctx context.Context, name string) ([]*cloudasset.ResourceSearchResult, error)
	ListOrganizationsIamPolicy(ctx context.Context, name string) (*cloudresourcemanager.Policy, error)
	ListProjectIamPolicy(ctx context.Context, name string) (*cloudresourcemanager.Policy, error)
	ListRegions(ctx context.Context, name string) ([]*compute.Region, error)
	ListResourceGroups(ctx context.Context, rgNames []string) ([]*cloudasset.ResourceSearchResult, error)
	ListServiceAccount(ctx context.Context, name string) ([]*iam.ServiceAccount, error)
	ListServiceAccountKeys(ctx context.Context, name string) ([]*iam.ServiceAccountKey, error)
	ListServiceAccountKeyUsage(ctx context.Context, resourceGroup string, request *monitoring.QueryTimeSeriesRequest) *monitoring.ProjectsTimeSeriesQueryCall
	ListSccFindings(ctx context.Context, name string) ([]*securitycenter.Finding, error)
	ListSpannerDatabases(ctx context.Context, name string) ([]*spanner.Database, error)
	ListSslPolicies(ctx context.Context, name string) ([]*compute.SslPolicy, error)
	ListSubNetworksByRegion(ctx context.Context, name string, region string) ([]*compute.Subnetwork, error)
	ListTargetHTTPSProxies(ctx context.Context, name string) ([]*compute.TargetHttpsProxy, error)
	ListTargetSslProxies(ctx context.Context, name string) ([]*compute.TargetSslProxy, error)
	ListURLMaps(ctx context.Context, name string) ([]*compute.UrlMap, error)
	ListUsersInGroup(ctx context.Context, group string) ([]string, error)
	ListZones(ctx context.Context, name string) ([]*compute.Zone, error)
	// TODO: Get rid of the scope argument since it's not used
	SearchIamPolicy(ctx context.Context, scope string, query string) ([]*cloudasset.IamPolicySearchResult, error)
}

type GCPApiReal struct {
	// TODO: Support multiple orgID in the future
	orgID                       string
	scope                       string
	apiKeyService               *apikeys.Service
	cloudAssetService           rateLimitedCloudAssetV1Service
	cloudIdentityService        *cloudidentity.Service
	cloudResourceManagerService *cloudresourcemanager.Service
	computeService              *compute.Service
	containerService            *container.Service
	iamService                  *iam.Service
	monitoringService           *monitoring.Service
	securityCenterService       *securitycenter.Service
	spannerService              *spanner.Service
	sqlAdminService             *sqladmin.Service
	storageService              *storage.Service
}

func detailedGoogleError(e error, apiDetail string) error {
	if e == nil {
		return nil
	}
	var gErr *googleapi.Error
	if errors.As(e, &gErr) {
		gErr.Message = fmt.Sprintf("%s: %s", apiDetail, gErr.Message)
		return gErr
	}
	return e
}

func NewGCPApiReal(ctx context.Context, orgID string) (GCPApi, error) {
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
	securitycenterService, err := securitycenter.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("securitycenter.NewService: %w", err)
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
		orgID:                       orgID,
		scope:                       constants.GCPOrgIDPrefix + orgID,
		apiKeyService:               apiKeyService,
		cloudAssetService:           newRateLimitedCloudAssetInventoryV1(cloudAssetService.V1),
		cloudIdentityService:        cloudIdentityService,
		cloudResourceManagerService: cloudresourcemanagerService,
		computeService:              computeService,
		containerService:            containerService,
		iamService:                  iamService,
		monitoringService:           monitoringService,
		securityCenterService:       securitycenterService,
		spannerService:              spannerService,
		sqlAdminService:             sqladminService,
		storageService:              storageService,
	}, nil
}

func (api *GCPApiReal) GetServiceAccountIAMPolicy(ctx context.Context, name string) (policy *iam.Policy, err error) {
	policy, err = api.iamService.Projects.ServiceAccounts.GetIamPolicy(name).Context(ctx).Do()
	return policy, detailedGoogleError(err, "ServiceAccounts.GetIamPolicy")
}

func (api *GCPApiReal) ListAPIKeys(ctx context.Context, name string) (apiKeys []*apikeys.V2Key, err error) {
	err = api.apiKeyService.Projects.Locations.Keys.List(constants.ResourceWithProjectsPrefix(name)).
		Context(ctx).
		PageSize(apiKeysPageSize).
		Pages(ctx, func(vlkr *apikeys.V2ListKeysResponse) error {
			apiKeys = append(apiKeys, vlkr.Keys...)
			return nil
		})
	log.Debugf("%s fetched %d api keys", name, len(apiKeys))
	return apiKeys, detailedGoogleError(err, "ApiKey.List")
}

func (api *GCPApiReal) ListBackendServices(ctx context.Context, name string) (backendSvcs []*compute.BackendService, err error) {
	err = api.computeService.BackendServices.AggregatedList(constants.ResourceWithoutProjectsPrefix(name)).Context(ctx).Pages(ctx, func(bsal *compute.BackendServiceAggregatedList) error {
		for _, be := range bsal.Items {
			backendSvcs = append(backendSvcs, be.BackendServices...)
		}
		return nil
	})
	log.Debugf("%s fetched %d backend services", name, len(backendSvcs))
	return backendSvcs, detailedGoogleError(err, "BackendServices.AggregatedList")
}

func (api *GCPApiReal) ListBuckets(ctx context.Context, name string) (buckets []*storage.Bucket, err error) {
	err = api.storageService.Buckets.List(constants.ResourceWithoutProjectsPrefix(name)).Context(ctx).Pages(ctx, func(b *storage.Buckets) error {
		buckets = append(buckets, b.Items...)
		return nil
	})
	log.Debugf("%s fetched %d buckets", name, len(buckets))
	return buckets, detailedGoogleError(err, "Buckets.List")
}

func (api *GCPApiReal) ListBucketsIamPolicy(bucketID string) (*storage.Policy, error) {
	iamPolicies, err := api.storageService.Buckets.GetIamPolicy(bucketID).Do()
	log.Debugf("fetched bucket %s iampolicy", bucketID)
	return iamPolicies, detailedGoogleError(err, "Buckets.GetIamPolicy")
}

func (api *GCPApiReal) ListCertificates(ctx context.Context, name string) (certs []*compute.SslCertificate, err error) {
	err = api.computeService.SslCertificates.AggregatedList(constants.ResourceWithoutProjectsPrefix(name)).Context(ctx).Pages(ctx, func(scal *compute.SslCertificateAggregatedList) error {
		for _, c := range scal.Items {
			certs = append(certs, c.SslCertificates...)
		}
		return nil
	})
	log.Debugf("%s fetched %d certificates", name, len(certs))
	return certs, detailedGoogleError(err, "SslCertificates.AggregatedList")
}

func (api *GCPApiReal) ListCloudSQLDatabases(ctx context.Context, name string) (instances []*sqladmin.DatabaseInstance, err error) {
	instanceSvc := sqladmin.NewInstancesService(api.sqlAdminService)
	err = instanceSvc.List(constants.ResourceWithoutProjectsPrefix(name)).Context(ctx).Pages(ctx, func(ilr *sqladmin.InstancesListResponse) error {
		instances = append(instances, ilr.Items...)
		return nil
	})
	log.Debugf("%s fetched %d cloudSql databases", name, len(instances))
	return instances, detailedGoogleError(err, "SqlAdmin.Instances.List")
}

func (api *GCPApiReal) ListClustersByZone(ctx context.Context, name string, zone string) (clusters []*container.Cluster, err error) {
	resp, err := api.containerService.Projects.Zones.Clusters.List(constants.ResourceWithoutProjectsPrefix(name), zone).Context(ctx).Do()
	if err != nil {
		return nil, detailedGoogleError(err, "Clusters.List")
	}
	log.Debugf("%s fetched %d clusters", name, len(clusters))
	return resp.Clusters, nil
}

func (api *GCPApiReal) ListFoldersIamPolicy(ctx context.Context, name string) (*cloudresourcemanager.Policy, error) {
	resp, err := api.cloudResourceManagerService.Folders.
		GetIamPolicy(name, new(cloudresourcemanager.GetIamPolicyRequest)).
		Context(ctx).
		Do()
	log.Debugf("fetched folder %s iam policy", name)
	return resp, detailedGoogleError(err, "Folders.GetIamPolicy")
}

func (api *GCPApiReal) ListInstances(ctx context.Context, name string) (instances []*compute.Instance, err error) {
	err = api.computeService.Instances.AggregatedList(constants.ResourceWithoutProjectsPrefix(name)).Context(ctx).Pages(ctx, func(ial *compute.InstanceAggregatedList) error {
		for _, ia := range ial.Items {
			instances = append(instances, ia.Instances...)
		}
		return nil
	})
	log.Debugf("%s fetched %d instances", name, len(instances))
	return instances, detailedGoogleError(err, "Instances.AggregatedList")
}

func (api *GCPApiReal) ListNamespaces(ctx context.Context, scope string) (namespaces []*cloudasset.ResourceSearchResult, err error) {
	if scope == "" {
		// We don't list namespaces for the whole org, this is a too large amount of data.
		return nil, fmt.Errorf("ListNamespaces: name is empty")
	}
	searchAllResources, err := api.cloudAssetService.SearchAllResources(ctx, scope)
	if err != nil {
		return nil, fmt.Errorf("SearchAllResources: %w", err)
	}
	err = searchAllResources.AssetTypes(k8sNamespaceAssetType).
		PageSize(pageSize).
		Pages(ctx, func(sarr *cloudasset.SearchAllResourcesResponse) error {
			namespaces = append(namespaces, sarr.Results...)
			return nil
		})
	log.Debugf("%q fetched %d namespaces", scope, len(namespaces))
	return namespaces, detailedGoogleError(err, "ListNamespaces")
}

func (api *GCPApiReal) ListPods(ctx context.Context, scope string) (pods []*cloudasset.ResourceSearchResult, err error) {
	if scope == "" {
		// We don't list namespaces for the whole org, this is a too large amount of data.
		return nil, fmt.Errorf("ListPods: name is empty")
	}
	searchAllResources, err := api.cloudAssetService.SearchAllResources(ctx, scope)
	if err != nil {
		return nil, fmt.Errorf("SearchAllResources: %w", err)
	}
	err = searchAllResources.ReadMask("*").AssetTypes(k8sPodAssetType).Context(ctx).Pages(ctx, func(sarr *cloudasset.SearchAllResourcesResponse) error {
		for _, r := range sarr.Results {
			if r.State == "Running" {
				pods = append(pods, r)
			}
		}
		return nil
	})
	log.Debugf("%q fetched %d pods", scope, len(pods))
	return pods, detailedGoogleError(err, "ListPods")
}

func (api *GCPApiReal) ListOrganizationsIamPolicy(ctx context.Context, name string) (*cloudresourcemanager.Policy, error) {
	resp, err := api.cloudResourceManagerService.Organizations.
		GetIamPolicy(name, new(cloudresourcemanager.GetIamPolicyRequest)).
		Context(ctx).
		Do()
	log.Debugf("fetched organization %s iam policy", name)
	return resp, detailedGoogleError(err, "Organizations.GetIamPolicy")
}

func (api *GCPApiReal) ListProjectIamPolicy(ctx context.Context, name string) (*cloudresourcemanager.Policy, error) {
	resp, err := api.cloudResourceManagerService.Projects.
		GetIamPolicy(constants.ResourceWithProjectsPrefix(name), &cloudresourcemanager.GetIamPolicyRequest{}).
		Context(ctx).
		Do()
	log.Debugf("fetched project %s iam policy", name)
	return resp, detailedGoogleError(err, "Project.GetIamPolicy")
}

func (api *GCPApiReal) ListRegions(ctx context.Context, name string) (regions []*compute.Region, err error) {
	ctx, span := tracer.Start(ctx, "ListRegions")
	span.SetAttributes(
		attribute.String(constants.TraceKeyName, name),
	)
	defer span.End()
	err = api.computeService.Regions.
		List(constants.ResourceWithoutProjectsPrefix(name)).
		Context(ctx).
		Pages(ctx, func(rl *compute.RegionList) error {
			regions = append(regions, rl.Items...)
			return nil
		})
	log.Debugf("%s fetched %d regions", name, len(regions))
	return regions, detailedGoogleError(err, "Regions.List")
}

func (api *GCPApiReal) ListResourceGroups(ctx context.Context, rgNames []string) (resourceGroups []*cloudasset.ResourceSearchResult, err error) {
	searchAllResources, err := api.cloudAssetService.SearchAllResources(ctx, api.scope)
	if err != nil {
		return nil, fmt.Errorf("SearchAllResources: %w", err)
	}
	call := searchAllResources.
		AssetTypes(projectAssetType, folderAssetType, organizationAssetType).
		Context(ctx)
	rgNamesMap := make(map[string]struct{})
	for _, rgName := range rgNames {
		rgNamesMap[rgName] = struct{}{}
	}
	var query string
	if len(rgNames) == 1 {
		query = fmt.Sprintf(searchForProject, rgNames[0])
	} else {
		query = excludeSysProjectsQuery
	}
	err = call.Query(query).
		Context(ctx).
		Pages(ctx, func(sarr *cloudasset.SearchAllResourcesResponse) error {
			for _, r := range sarr.Results {
				if r.State == "ACTIVE" {
					if len(rgNames) > 0 {
						if _, ok := rgNamesMap[strings.TrimPrefix(r.Name, cloudResourceManagerPrefix)]; ok {
							resourceGroups = append(resourceGroups, r)
						}
					} else {
						resourceGroups = append(resourceGroups, r)
					}
				}
			}
			return nil
		})
	log.WithField("rg_names", rgNames).Debugf("fetched %d resourcegroups", len(resourceGroups))
	return resourceGroups, detailedGoogleError(err, "ListResourceGroups")
}

func (api *GCPApiReal) ListSccFindings(ctx context.Context, name string) (findings []*securitycenter.Finding, err error) {
	ctx, span := tracer.Start(ctx, "ListSccFindings")
	defer span.End()
	err = api.securityCenterService.Projects.Sources.Findings.List(name+"/sources/-").
		Context(ctx).
		Filter("state=\"ACTIVE\" AND NOT finding_class=\"THREAT\"").
		PageSize(sccPageSize).
		Pages(ctx, func(flr *securitycenter.ListFindingsResponse) error {
			ctx, span := tracer.Start(ctx, "ListSccFindingsPage")
			defer span.End()
			if err := rlSCC.Wait(ctx); err != nil {
				return err
			}
			for _, fr := range flr.ListFindingsResults {
				findings = append(findings, fr.Finding)
			}
			return nil
		})
	return
}

func (api *GCPApiReal) ListServiceAccount(ctx context.Context, name string) (serviceaccounts []*iam.ServiceAccount, err error) {
	err = api.iamService.Projects.ServiceAccounts.List(constants.ResourceWithProjectsPrefix(name)).Context(ctx).
		Pages(ctx, func(lsar *iam.ListServiceAccountsResponse) error {
			serviceaccounts = append(serviceaccounts, lsar.Accounts...)
			return nil
		})
	log.Debugf("%s fetched %d service accounts", name, len(serviceaccounts))
	return serviceaccounts, detailedGoogleError(err, "ServiceAccounts.List")
}

func (api *GCPApiReal) ListServiceAccountKeys(ctx context.Context, name string) (keys []*iam.ServiceAccountKey, err error) {
	resp, err := api.iamService.Projects.ServiceAccounts.Keys.
		List(constants.ResourceWithProjectsPrefix(name)).
		Context(ctx).
		Do()
	if err != nil {
		return nil, detailedGoogleError(err, "ServiceAccounts.Keys.List")
	}
	keys = append(keys, resp.Keys...)
	log.Debugf("%s fetched %d service account keys", name, len(keys))
	return keys, nil
}

func (api *GCPApiReal) ListServiceAccountKeyUsage(ctx context.Context, resourceGroup string, request *monitoring.QueryTimeSeriesRequest) *monitoring.ProjectsTimeSeriesQueryCall {
	return api.monitoringService.Projects.TimeSeries.
		Query(constants.ResourceWithProjectsPrefix(resourceGroup), request).
		Context(ctx)
}

func (api *GCPApiReal) ListSpannerDatabases(ctx context.Context, name string) (dbs []*spanner.Database, err error) {
	instanceSvc := spanner.NewProjectsInstancesService(api.spannerService)
	err = instanceSvc.List(constants.ResourceWithoutProjectsPrefix(name)).
		Context(ctx).
		Pages(ctx, func(lir *spanner.ListInstancesResponse) error {
			for _, instance := range lir.Instances {
				databaseSvc := spanner.NewProjectsInstancesDatabasesService(api.spannerService)
				return databaseSvc.List(instance.Name).Pages(ctx, func(ldr *spanner.ListDatabasesResponse) error {
					dbs = append(dbs, ldr.Databases...)
					return nil
				})
			}
			return nil
		})
	log.Debugf("%s fetched %d spanner databases", name, len(dbs))
	return dbs, detailedGoogleError(err, "Spanner.Instances.List")
}

func (api *GCPApiReal) ListSslPolicies(ctx context.Context, name string) (policies []*compute.SslPolicy, err error) {
	err = api.computeService.SslPolicies.
		AggregatedList(constants.ResourceWithoutProjectsPrefix(name)).
		Context(ctx).
		Pages(ctx, func(spal *compute.SslPoliciesAggregatedList) error {
			for _, i := range spal.Items {
				policies = append(policies, i.SslPolicies...)
			}
			return nil
		})
	log.Debugf("%s fetched %d ssl policies", name, len(policies))
	return policies, detailedGoogleError(err, "SslPolicies,AggregatedList")
}

func (api *GCPApiReal) ListSubNetworksByRegion(ctx context.Context, name string, region string) (subnetworks []*compute.Subnetwork, err error) {
	if err := rlSubnetRanges.Wait(ctx); err != nil {
		return nil, fmt.Errorf("snRangesLimiter.Wait: %w", err)
	}
	err = api.computeService.Subnetworks.
		List(constants.ResourceWithoutProjectsPrefix(name), region).
		Context(ctx).
		Pages(ctx, func(sl *compute.SubnetworkList) error {
			subnetworks = append(subnetworks, sl.Items...)
			return nil
		})
	log.Debugf("%s region %s fetched %d subnetworks", name, region, len(subnetworks))
	return subnetworks, detailedGoogleError(err, "Subnetworks.List")
}

func (api *GCPApiReal) ListTargetHTTPSProxies(ctx context.Context, name string) (proxies []*compute.TargetHttpsProxy, err error) {
	err = api.computeService.TargetHttpsProxies.
		AggregatedList(constants.ResourceWithoutProjectsPrefix(name)).
		Context(ctx).
		Pages(ctx, func(thpal *compute.TargetHttpsProxyAggregatedList) error {
			for _, p := range thpal.Items {
				proxies = append(proxies, p.TargetHttpsProxies...)
			}
			return nil
		})
	log.Debugf("%s fetched %d https proxies", name, len(proxies))
	return proxies, detailedGoogleError(err, "TargetHttpsProxies.AggregatedList")
}

func (api *GCPApiReal) ListTargetSslProxies(ctx context.Context, name string) (proxies []*compute.TargetSslProxy, err error) {
	err = api.computeService.TargetSslProxies.
		List(constants.ResourceWithoutProjectsPrefix(name)).
		Context(ctx).
		Pages(ctx, func(tspl *compute.TargetSslProxyList) error {
			proxies = append(proxies, tspl.Items...)
			return nil
		})
	log.Debugf("%s fetched %d ssl proxies", name, len(proxies))
	return proxies, detailedGoogleError(err, "TargetSslProxies.List")
}

func (api *GCPApiReal) ListURLMaps(ctx context.Context, name string) (maps []*compute.UrlMap, err error) {
	err = api.computeService.UrlMaps.
		AggregatedList(constants.ResourceWithoutProjectsPrefix(name)).
		Context(ctx).
		Pages(ctx, func(umal *compute.UrlMapsAggregatedList) error {
			for _, m := range umal.Items {
				maps = append(maps, m.UrlMaps...)
			}
			return nil
		})
	log.Debugf("%s fetched %d url maps", name, len(maps))
	return maps, detailedGoogleError(err, "UrlMaps.AggregatedList")
}

func (api *GCPApiReal) ListUsersInGroup(ctx context.Context, group string) (groupMembers []string, err error) {
	group = strings.TrimPrefix(group, "group:")
	if e, ok := listGroupCache[group]; ok && time.Since(e.Creation) < time.Second*300 {
		return e.Members, nil
	}
	groupID, err := api.cloudIdentityService.Groups.Lookup().
		Context(ctx).
		GroupKeyId(group).
		Do()
	if err != nil {
		return []string{}, fmt.Errorf("group lookup: %w", err)
	}
	err = api.cloudIdentityService.Groups.Memberships.
		List(groupID.Name).
		Context(ctx).
		PageSize(pageSize).
		Pages(ctx, func(lmr *cloudidentity.ListMembershipsResponse) error {
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
	log.Debugf("%s fetched %d group members", group, len(groupMembers))
	listGroupCache[group] = GroupCacheEntry{Creation: time.Now(), Members: groupMembers}
	return groupMembers, detailedGoogleError(err, "CloudIdentity.Groups.Memberships.List")
}

func (api *GCPApiReal) ListZones(ctx context.Context, name string) (zones []*compute.Zone, err error) {
	err = api.computeService.Zones.
		List(constants.ResourceWithoutProjectsPrefix(name)).
		Context(ctx).
		Pages(ctx, func(zl *compute.ZoneList) error {
			zones = append(zones, zl.Items...)
			return nil
		})
	log.Debugf("%s fetched %d zones", name, len(zones))
	return zones, detailedGoogleError(err, "Zones.List")
}

func (api *GCPApiReal) SearchIamPolicy(ctx context.Context, scope string, query string) (iamPolicies []*cloudasset.IamPolicySearchResult, err error) {
	attemptNumber := 1
	var waitSec float64
	start := time.Now()
	for {
		// TODO: Change scope to api.scope and get rid of the scope parameter
		searchAllIamPolicies, err := api.cloudAssetService.SearchAllIamPolicies(ctx, scope)
		if err != nil {
			return nil, fmt.Errorf("SearchAllIamPolicies: %w", err)
		}
		err = searchAllIamPolicies.
			Query(query).
			Context(ctx).
			Pages(ctx, func(resp *cloudasset.SearchAllIamPoliciesResponse) error {
				iamPolicies = append(iamPolicies, resp.Results...)
				return nil
			})
		if err == nil {
			break
		}
		if time.Since(start) > maxSecAcrossAttempts {
			return nil, fmt.Errorf("SearchIamPolicy.tooLong: after %v seconds %w", time.Since(start), detailedGoogleError(err, fmt.Sprintf("IamPolicies.SearchAll.Query %q", query)))
		}
		if attemptNumber >= maxAttemptNumber {
			return nil, fmt.Errorf("SearchIamPolicy.maxAttempt: after %v attempts %w", attemptNumber, detailedGoogleError(err, fmt.Sprintf("IamPolicies.SearchAll.Query %q", query)))
		}
		if !isErrorCodeRetryable(getErrorCode(err)) {
			return nil, fmt.Errorf("SearchIamPolicy.notRetryableCode: %w", detailedGoogleError(err, fmt.Sprintf("IamPolicies.SearchAll.Query %q", query)))
		}
		waitSec = math.Min(math.Pow(2, float64(attemptNumber))+rand.Float64(), maxSecBtwAttempts) //nolint:mnd,gosec
		time.Sleep(time.Duration(waitSec) * time.Second)
		attemptNumber++
	}
	log.Debugf("fetched %d iam policies", len(iamPolicies))
	return iamPolicies, nil
}
