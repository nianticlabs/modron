package utils

import (
	"fmt"

	pb "github.com/nianticlabs/modron/src/proto/generated"
)

func ComputeRgHierarchy(resources []*pb.Resource) (map[string]*pb.RecursiveResource, error) {
	resourceMap := make(map[string]*pb.RecursiveResource)
	for _, r := range resources {
		if r.GetResourceGroup() == nil {
			continue
		}
		recRes, err := ToRecursiveResource(r)
		if err != nil {
			return nil, err
		}
		resourceMap[r.Name] = recRes
	}

	for _, r := range resources {
		if r.Parent == "" {
			var err error
			resourceMap[""], err = ToRecursiveResource(r)
			if err != nil {
				return nil, err
			}
			continue
		}
		parent, ok := resourceMap[r.Parent]
		if !ok {
			log.Warnf("parent %q not found", r.Parent)
			if r.ResourceGroupName == "" {
				log.Errorf("resource %q has no parent and no resource group", r.Name)
				continue
			}
			if r.ResourceGroupName == r.Name {
				log.Errorf("resource %q is its own parent", r.Name)
				continue
			}
			parent, ok = resourceMap[r.ResourceGroupName]
			if !ok {
				log.Errorf("resource group %q not found, is %q orphan?", r.ResourceGroupName, r.Name)
				continue
			}
		}
		recRes, err := ToRecursiveResource(r)
		if err != nil {
			return nil, fmt.Errorf("toRecursiveResource: %w", err)
		}
		parent.Children = append(parent.Children, recRes)
		resourceMap[r.Parent] = parent
	}
	return resourceMap, nil
}

func ToRecursiveResource(r *pb.Resource) (*pb.RecursiveResource, error) {
	t, err := TypeFromResource(r)
	if err != nil {
		return nil, fmt.Errorf("typeFromResourceAsString: %w", err)
	}
	return &pb.RecursiveResource{
		Uuid:        r.Uid,
		Name:        r.Name,
		DisplayName: r.DisplayName,
		Parent:      r.Parent,
		Type:        t,
		Labels:      r.Labels,
		Tags:        r.Tags,
	}, nil
}
