package model

import (
	"golang.org/x/net/context"
	"github.com/nianticlabs/modron/src/pb"
)

type Collector interface {
	CollectAndStoreAllResourceGroupResources(ctx context.Context, collectId string, resourceGroupNames []string) []error
	CollectAndStoreResources(ctx context.Context, collectId string, resourecGroupId string) []error
	GetResourceGroup(ctx context.Context, collectId string, resourecGroupId string) (*pb.Resource, error)
	ListResourceGroupAdmins(ctx context.Context) (map[string]map[string]struct{}, error)
	ListResourceGroupResources(ctx context.Context, collectId string, resourecGroup *pb.Resource) ([]*pb.Resource, []error)
}
