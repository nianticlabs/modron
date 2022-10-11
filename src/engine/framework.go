package engine

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

type Storage struct {
	model.Storage
}

var storage *Storage

const (
	PrincipalServiceAccount        = "serviceAccount"
	PrincipalUser                  = "user"
	PrincipalGroup                 = "group"
	PrincipalAllUsers              = "allUsers"
	PrincipalAllAuthenticatedUsers = "allAuthenticatedUsers"
	PrincipalDomain                = "domain"
)

func GetResource(ctx context.Context, resourceName string) (*pb.Resource, error) {
	limit := 1
	filter := model.StorageFilter{
		Limit:         &limit,
		ResourceNames: &[]string{resourceName},
	}
	res, err := storage.Storage.ListResources(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("resource %q could not be fetched: %w", resourceName, err)
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("resource %q does not exist", resourceName)
	}
	return res[0], nil
}

// TODO: Add SelfLink and HumanReadableName field to Protobuf and move this logic to the collector.
func GetGcpReadableResourceName(resourceName string) string {
	if !(strings.Contains(resourceName, "[") && strings.Contains(resourceName, "]")) {
		return resourceName
	}
	m := regexp.MustCompile(`^.*\[|\].*$`)
	return m.ReplaceAllLiteralString(resourceName, "")
}

func GetAccountRoles(perm *pb.Permission, account string) ([]string, error) {
	roles := []string{}
	for _, principal := range perm.Principals {
		if principal != account {
			continue
		}

		roles = append(roles, perm.Role)
	}

	return roles, nil
}
