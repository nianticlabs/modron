package gcpcollector

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/api/cloudasset/v1"
	"google.golang.org/api/cloudresourcemanager/v3"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
)

const (
	projectResourcePrefix      = "//cloudresourcemanager.googleapis.com/projects/"
	folderResourcePrefix       = "//cloudresourcemanager.googleapis.com/folders/"
	organizationResourcePrefix = "//cloudresourcemanager.googleapis.com/organizations/"
	cloudResourceManagerPrefix = "//cloudresourcemanager.googleapis.com/"
	welcomePage                = "https://console.cloud.google.com/welcome"
)

var (
	// Syntax https://cloud.google.com/asset-inventory/docs/query-syntax
	projectOwnersQuery       = fmt.Sprintf("resource:%s*", projectResourcePrefix)
	foldersOwnersQuery       = fmt.Sprintf("resource:%s*", folderResourcePrefix)
	organizationsOwnersQuery = fmt.Sprintf("resource:%s*", organizationResourcePrefix)
)

func (collector *GCPCollector) GetResourceGroupWithIamPolicy(ctx context.Context, collectID string, rgName string) (res *pb.Resource, err error) {
	ctx, span := tracer.Start(ctx, "GetResourceGroup")
	defer span.End()

	rgs, err := collector.ListResourceGroupsWithIamPolicies(ctx, []string{rgName})
	if err != nil {
		return nil, err
	}
	if len(rgs) != 1 {
		return nil, fmt.Errorf("found %d resource groups for %s, expected 1", len(rgs), rgName)
	}
	rgs[0].CollectionUid = collectID
	return rgs[0], nil
}

func (collector *GCPCollector) ListResourceGroupsWithIamPolicies(ctx context.Context, rgNames []string) ([]*pb.Resource, error) {
	ctx, span := tracer.Start(ctx, "ListResourceGroupsWithIamPolicies")
	defer span.End()
	rgs, err := collector.ListResourceGroups(ctx, rgNames)
	if err != nil {
		return nil, fmt.Errorf("ListResourceGroups: %w", err)
	}
	for i, rg := range rgs {
		var resp *cloudresourcemanager.Policy
		switch {
		case strings.HasPrefix(rg.Name, constants.GCPFolderIDPrefix):
			resp, err = collector.api.ListFoldersIamPolicy(ctx, rg.Name)
		case strings.HasPrefix(rg.Name, constants.GCPOrgIDPrefix):
			resp, err = collector.api.ListOrganizationsIamPolicy(ctx, rg.Name)
		default: // Default to project
			resp, err = collector.api.ListProjectIamPolicy(ctx, rg.Name)
		}
		if err != nil {
			return nil, fmt.Errorf("cannot get IAM policies for resource group %q: %w", rg.Name, err)
		}
		var permissions []*pb.Permission
		for _, binding := range resp.Bindings {
			permissions = append(permissions, &pb.Permission{
				Role:       constants.ToRole(binding.Role).String(),
				Principals: binding.Members,
			})
		}
		rgs[i].IamPolicy = &pb.IamPolicy{
			Permissions: permissions,
		}
	}
	return rgs, nil
}

func (collector *GCPCollector) ListResourceGroups(ctx context.Context, rgNames []string) (rgs []*pb.Resource, err error) {
	ctx, span := tracer.Start(ctx, "ListResourceGroups")
	defer span.End()
	resourceGroups, err := collector.api.ListResourceGroups(ctx, rgNames)
	if err != nil {
		return nil, err
	}
	for _, rg := range resourceGroups {
		var ancestors []string
		ancestors = append(ancestors, rg.Folders...)
		if rg.Organization != "" {
			ancestors = append(ancestors, rg.Organization)
		}
		if rg.State != "ACTIVE" {
			log.Warnf("resource group %s is not active", rg.Name)
			continue
		}
		rgName := strings.TrimPrefix(rg.Name, cloudResourceManagerPrefix)
		res := &pb.Resource{
			Uid:               common.GetUUID(uuidGenRetries),
			ResourceGroupName: rgName,
			Name:              rgName,
			DisplayName:       rg.DisplayName,
			Parent:            strings.TrimPrefix(rg.ParentFullResourceName, cloudResourceManagerPrefix),
			Link:              getResourceGroupLink(rg.Name),
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{
					Identifier: strings.TrimPrefix(rg.Name, cloudResourceManagerPrefix),
					Name:       rg.DisplayName,
				},
			},
			Labels:    rg.Labels,
			Tags:      toKV(rg.Tags),
			Ancestors: ancestors,
		}
		rgs = append(rgs, res)
	}
	return rgs, nil
}

func toKV(tags []*cloudasset.Tag) map[string]string {
	kv := make(map[string]string)
	for _, tag := range tags {
		kv[tag.TagKey] = tag.TagValue
	}
	return kv
}

func getResourceGroupLink(resourceName string) string {
	parts := strings.SplitN(resourceName, "/", 2) //nolint:mnd
	if len(parts) != 2 {                          //nolint:mnd
		return ""
	}
	switch parts[0] {
	case "projects":
		return welcomePage + "?project=" + parts[1]
	case "folders":
		return welcomePage + "?folder=" + parts[1]
	case "organizations":
		return welcomePage + "?organizationId=" + parts[1]
	}
	return ""
}

func (collector *GCPCollector) ListResourceGroupNames(ctx context.Context) (rgNames []string, err error) {
	ctx, span := tracer.Start(ctx, "ListResourceGroupNames")
	defer span.End()
	resourceGroups, err := collector.ListResourceGroups(ctx, nil)
	if err != nil {
		return nil, err
	}
	for _, rg := range resourceGroups {
		rgNames = append(rgNames, strings.TrimPrefix(rg.Name, cloudResourceManagerPrefix))
	}
	log.Infof("found %d names", len(rgNames))
	return rgNames, nil
}

func (collector *GCPCollector) listResourceGroupAdmins(
	ctx context.Context,
	resourceGroupAdmins map[string]map[string]struct{},
	iamPolicyResult []*cloudasset.IamPolicySearchResult,
) model.ACLCache {
	ctx, span := tracer.Start(ctx, "listResourceGroupAdmins")
	defer span.End()
	for _, res := range iamPolicyResult {
		if res.Policy == nil {
			continue
		}
		resourceID := strings.TrimPrefix(res.Resource, cloudResourceManagerPrefix)
		if strings.HasPrefix(resourceID, constants.GCPProjectsNamePrefix+constants.GCPSysProjectPrefix) {
			log.Debugf("skipping system project %q", resourceID)
			continue
		}
		if strings.HasPrefix(resourceID, constants.GCPFolderIDPrefix) {
			// TODO: remove this once we support the hierarchical structure of GCP
			log.Debugf("skipping folder %q", resourceID)
			continue
		}
		// Allow admins to see all the resources, regardless on whether they are owned by someone or not.
		resourceGroupAdmins["*"][resourceID] = struct{}{}
		for _, binding := range res.Policy.Bindings {
			theRole := constants.Role(strings.TrimPrefix(binding.Role, constants.GCPRolePrefix))
			_, hasAdminRole := constants.AdminRoles[theRole]
			_, hasAdditionalAdminRole := collector.additionalAdminRolesMap[theRole]
			if !hasAdminRole && !hasAdditionalAdminRole {
				continue
			}

			for _, u := range binding.Members {
				users := []string{u}
				var err error
				if strings.HasPrefix(u, constants.GCPAccountGroupPrefix) && strings.HasSuffix(u, collector.orgSuffix) {
					users, err = collector.api.ListUsersInGroup(ctx, u)
					if err != nil {
						log.Warnf("cannot list users in group %q: %v", u, err)
						continue
					}
				}
				for _, user := range users {
					user = strings.TrimPrefix(user, constants.GCPUserAccountPrefix)
					if strings.HasSuffix(user, collector.orgSuffix) {
						if _, ok := resourceGroupAdmins[user]; !ok {
							resourceGroupAdmins[user] = make(map[string]struct{})
						}
						resourceGroupAdmins[user][resourceID] = struct{}{}
					}
				}
			}
		}
	}
	return resourceGroupAdmins
}

func (collector *GCPCollector) ListResourceGroupAdmins(ctx context.Context) (model.ACLCache, error) {
	ctx, span := tracer.Start(ctx, "ListResourceGroupAdmins")
	defer span.End()
	scope := constants.GCPOrgIDPrefix + collector.orgID

	resp, err := collector.api.SearchIamPolicy(ctx, scope, projectOwnersQuery)
	if err != nil {
		return nil, err
	}
	resourceGroupAdmins := model.ACLCache{}
	resourceGroupAdmins["*"] = map[string]struct{}{}
	resourceGroupAdmins = collector.listResourceGroupAdmins(ctx, resourceGroupAdmins, resp)

	resp, err = collector.api.SearchIamPolicy(ctx, scope, foldersOwnersQuery)
	if err != nil {
		return nil, err
	}
	resourceGroupAdmins = collector.listResourceGroupAdmins(ctx, resourceGroupAdmins, resp)

	resp, err = collector.api.SearchIamPolicy(ctx, scope, organizationsOwnersQuery)
	if err != nil {
		return nil, err
	}
	resourceGroupAdmins = collector.listResourceGroupAdmins(ctx, resourceGroupAdmins, resp)

	return resourceGroupAdmins, nil
}
