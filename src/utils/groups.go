package utils

import (
	"golang.org/x/exp/maps"

	pb "github.com/nianticlabs/modron/src/proto/generated"
)

func GroupsFromResources(resources []*pb.Resource) (allGroups []string) {
	resourceGroups := map[string]struct{}{}
	for _, r := range resources {
		resourceGroups[r.ResourceGroupName] = struct{}{}
	}
	return maps.Keys(resourceGroups)
}
