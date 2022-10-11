package gcpcollector

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/pb"
)

const (
	serviceAccountPrefix = "serviceAccount:"
	userPrefix           = "user:"
	groupPrefix          = "group:"

	projectResourcePrefix = "//cloudresourcemanager.googleapis.com/projects/"
)

var (
	// Syntax https://cloud.google.com/asset-inventory/docs/query-syntax
	projectOwnersQuery = fmt.Sprintf("resource:%s*", projectResourcePrefix)
)

func (collector *GCPCollector) GetResourceGroup(ctx context.Context, collectId string, name string) (*pb.Resource, error) {
	resp, err := collector.api.ListProjectIamPolicy(name)
	if err != nil {
		return nil, err
	}
	permissions := []*pb.Permission{}
	for _, binding := range resp.Bindings {
		for i := range binding.Members {
			binding.Members[i] = strings.TrimPrefix(binding.Members[i], serviceAccountPrefix)
			binding.Members[i] = strings.TrimPrefix(binding.Members[i], userPrefix)
		}
		permissions = append(permissions, &pb.Permission{
			Role:       strings.TrimPrefix(binding.Role, constants.GCPRolePrefix),
			Principals: binding.Members,
		})
	}
	resourceGroup := &pb.Resource{
		Uid:               collector.getNewUid(),
		ResourceGroupName: name,
		CollectionUid:     collectId,
		Timestamp:         timestamppb.Now(),
		Name:              formatResourceName(name, name),
		IamPolicy: &pb.IamPolicy{
			Resource:    nil,
			Permissions: permissions,
		},
		Type: &pb.Resource_ResourceGroup{},
	}
	return resourceGroup, nil
}

func (collector *GCPCollector) ListResourceGroupNames(ctx context.Context) ([]string, error) {
	projects, err := collector.api.ListAllResourceGroups(ctx)
	if err != nil {
		return nil, err
	}
	projectIds := []string{}
	for _, rg := range projects {
		projectIds = append(projectIds, rg.ProjectId)
	}
	return projectIds, nil
}

func (collector *GCPCollector) ListResourceGroupAdmins(ctx context.Context) (map[string]map[string]struct{}, error) {
	resp, err := collector.api.SearchIamPolicy(ctx, orgId, projectOwnersQuery)
	if err != nil {
		return nil, err
	}
	projectAdmins := map[string]map[string]struct{}{}
	projectAdmins["*"] = map[string]struct{}{}
	for _, res := range resp {
		if res.Policy == nil {
			continue
		}
		pID := strings.TrimPrefix(res.Resource, projectResourcePrefix)
		for _, binding := range res.Policy.Bindings {
			if _, ok := constants.AdminRoles[strings.ToLower(strings.TrimPrefix(binding.Role, constants.GCPRolePrefix))]; !ok {
				continue
			}
			for _, u := range binding.Members {
				users := []string{u}
				if strings.HasPrefix(u, groupPrefix) {
					users = []string{}
				}
				for _, user := range users {
					user = strings.TrimPrefix(user, userPrefix)
					if strings.HasSuffix(user, orgSuffix) {
						if _, ok := projectAdmins[user]; !ok {
							projectAdmins[user] = make(map[string]struct{})
						}
						projectAdmins[user][pID] = struct{}{}
						projectAdmins["*"][pID] = struct{}{}
					}
				}
			}
		}
	}
	return projectAdmins, nil
}
