package gcpcollector

import (
	"fmt"
	"os"
	"strings"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/pb"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/api/cloudasset/v1"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	projectResourcePrefix      = "//cloudresourcemanager.googleapis.com/projects/"
	folderResourcePrefix       = "//cloudresourcemanager.googleapis.com/folders/"
	organizationResourcePrefix = "//cloudresourcemanager.googleapis.com/organizations/"
	cloudResourceManagerPrefix = "//cloudresourcemanager.googleapis.com/"
	projectLinkPrefix          = "https://console.cloud.google.com/home/dashboard?project=%s"
)

var (
	// Syntax https://cloud.google.com/asset-inventory/docs/query-syntax
	projectOwnersQuery       = fmt.Sprintf("resource:%s*", projectResourcePrefix)
	foldersOwnersQuery       = fmt.Sprintf("resource:%s*", folderResourcePrefix)
	organizationsOwnersQuery = fmt.Sprintf("resource:%s*", organizationResourcePrefix)
)

func (collector *GCPCollector) GetResourceGroup(ctx context.Context, collectId string, name string) (*pb.Resource, error) {
	var resp *cloudresourcemanager.Policy
	var err error
	switch {
	case strings.HasPrefix(name, constants.GCPFolderIdPrefix):
		resp, err = collector.api.ListFoldersIamPolicy(name)
	case strings.HasPrefix(name, constants.GCPOrgIdPrefix):
		resp, err = collector.api.ListOrganizationsIamPolicy(name)
	default: // Default to project
		resp, err = collector.api.ListProjectIamPolicy(name)
	}
	if err != nil {
		return nil, err
	}
	permissions := []*pb.Permission{}
	for _, binding := range resp.Bindings {
		for i := range binding.Members {
			binding.Members[i] = strings.TrimPrefix(binding.Members[i], constants.GCPAccountGroupPrefix)
			binding.Members[i] = strings.TrimPrefix(binding.Members[i], constants.GCPServiceAccountPrefix)
			binding.Members[i] = strings.TrimPrefix(binding.Members[i], constants.GCPUserAccountPrefix)
		}
		permissions = append(permissions, &pb.Permission{
			Role:       strings.TrimPrefix(binding.Role, constants.GCPRolePrefix),
			Principals: binding.Members,
		})
	}
	rgs, err := collector.ListResourceGroups(ctx, name)
	if err != nil {
		return nil, err
	}
	if len(rgs) != 1 {
		return nil, fmt.Errorf("found %d resource groups for %s, expected 1", len(rgs), name)
	}
	rgs[0].Timestamp = timestamppb.Now()
	rgs[0].CollectionUid = collectId
	rgs[0].IamPolicy = &pb.IamPolicy{
		Resource:    nil,
		Permissions: permissions,
	}
	return rgs[0], nil
}

func (collector *GCPCollector) ListResourceGroups(ctx context.Context, name string) ([]*pb.Resource, error) {
	resourceGroups, err := collector.api.ListResourceGroups(ctx, name)
	if err != nil {
		return nil, err
	}
	rgs := []*pb.Resource{}
	for _, rg := range resourceGroups {
		rgs = append(rgs, &pb.Resource{
			Uid:               common.GetUUID(3),
			ResourceGroupName: strings.TrimPrefix(rg.Name, cloudResourceManagerPrefix),
			Name:              strings.TrimPrefix(rg.Name, cloudResourceManagerPrefix),
			Parent:            rg.ParentFullResourceName,
			// TODO: This only works for project, yet other resource groups in GCP do not have much to see on the UI side.
			Link: fmt.Sprintf(projectLinkPrefix, strings.TrimPrefix(rg.Name, projectResourcePrefix)),
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{
					Identifier: rg.Project,
				},
			},
		})
	}
	return rgs, nil
}

func (collector *GCPCollector) ListResourceGroupNames(ctx context.Context) ([]string, error) {
	resourceGroups, err := collector.ListResourceGroups(ctx, "")
	if err != nil {
		return nil, err
	}
	rgNames := []string{}
	for _, rg := range resourceGroups {
		rgNames = append(rgNames, strings.TrimPrefix(rg.Name, cloudResourceManagerPrefix))
	}
	glog.V(5).Infof("found %d names: %+v", len(rgNames), rgNames)
	return rgNames, nil
}

func (collector *GCPCollector) listResourceGroupAdmins(
	ctx context.Context,
	resourceGroupAdmins map[string]map[string]struct{},
	iamPolicyResult []*cloudasset.IamPolicySearchResult,
) map[string]map[string]struct{} {
	for _, res := range iamPolicyResult {
		if res.Policy == nil {
			continue
		}
		resourceID := strings.TrimPrefix(res.Resource, "//cloudresourcemanager.googleapis.com/")
		for _, binding := range res.Policy.Bindings {
			if _, ok := constants.AdminRoles[strings.ToLower(strings.TrimPrefix(binding.Role, constants.GCPRolePrefix))]; !ok {
				continue
			}
			for _, u := range binding.Members {
				users := []string{u}
				var err error
				if strings.HasPrefix(u, constants.GCPAccountGroupPrefix) && strings.HasSuffix(u, os.Getenv(constants.OrgSuffixEnvVar)) {
					users, err = collector.api.ListUsersInGroup(ctx, u)
					if err != nil {
						glog.Warningf("cannot list users in group %q: %v", u, err)
						continue
					}
				}
				for _, user := range users {
					user = strings.TrimPrefix(user, constants.GCPUserAccountPrefix)
					if strings.HasSuffix(user, orgSuffix) {
						if _, ok := resourceGroupAdmins[user]; !ok {
							resourceGroupAdmins[user] = make(map[string]struct{})
						}
						resourceGroupAdmins[user][resourceID] = struct{}{}
						resourceGroupAdmins["*"][resourceID] = struct{}{}
					}
				}
			}
		}
	}
	return resourceGroupAdmins
}

func (collector *GCPCollector) ListResourceGroupAdmins(ctx context.Context) (map[string]map[string]struct{}, error) {
	resp, err := collector.api.SearchIamPolicy(ctx, orgId, projectOwnersQuery)
	if err != nil {
		return nil, err
	}
	resourceGroupAdmins := map[string]map[string]struct{}{}
	resourceGroupAdmins["*"] = map[string]struct{}{}
	resourceGroupAdmins = collector.listResourceGroupAdmins(ctx, resourceGroupAdmins, resp)

	resp, err = collector.api.SearchIamPolicy(ctx, orgId, foldersOwnersQuery)
	if err != nil {
		return nil, err
	}
	resourceGroupAdmins = collector.listResourceGroupAdmins(ctx, resourceGroupAdmins, resp)

	resp, err = collector.api.SearchIamPolicy(ctx, orgId, organizationsOwnersQuery)
	if err != nil {
		return nil, err
	}
	resourceGroupAdmins = collector.listResourceGroupAdmins(ctx, resourceGroupAdmins, resp)

	return resourceGroupAdmins, nil
}
