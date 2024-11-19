package model

import (
	"golang.org/x/net/context"

	pb "github.com/nianticlabs/modron/src/proto/generated"
)

type Collector interface {
	CollectAndStoreAll(ctx context.Context, collectID string, resourceGroupNames []string, preCollectedRgs []*pb.Resource) error

	GetResourceGroupWithIamPolicy(ctx context.Context, collectID string, rgName string) (*pb.Resource, error)
	ListResourceGroups(ctx context.Context, rgNames []string) ([]*pb.Resource, error)
	ListResourceGroupsWithIamPolicies(ctx context.Context, rgNames []string) ([]*pb.Resource, error)
	ListResourceGroupNames(ctx context.Context) ([]string, error)
	ListResourceGroupAdmins(ctx context.Context) (ACLCache, error)
	ListResourceGroupResources(ctx context.Context, collectID string, rgName string) ([]*pb.Resource, []error)
	ListResourceGroupObservations(ctx context.Context, collectID string, rgName string) ([]*pb.Observation, []error)
}
