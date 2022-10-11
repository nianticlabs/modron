// Package memstorage provides a storage backend that runs locally in memory. It is supposed to be used primarily for API testing.
package memstorage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

type MemStorage struct {
	resources    sync.Map
	observations sync.Map
	operations   sync.Map
}

func New() model.Storage {
	return &MemStorage{
		resources:    sync.Map{},
		observations: sync.Map{},
		operations:   sync.Map{},
	}
}

func (mem *MemStorage) BatchCreateResources(ctx context.Context, resources []*pb.Resource) ([]*pb.Resource, error) {
	for _, resource := range resources {
		existingRes, ok := mem.resources.Load(resource.ResourceGroupName)
		if !ok {
			existingRes = []*pb.Resource{}
		}
		mem.resources.Store(resource.ResourceGroupName, append(existingRes.([]*pb.Resource), resource))
	}
	return resources, nil
}

func (mem *MemStorage) BatchCreateObservations(ctx context.Context, observations []*pb.Observation) ([]*pb.Observation, error) {
	for _, o := range observations {
		existingObs, ok := mem.observations.Load(o.Resource.ResourceGroupName)
		if !ok {
			existingObs = []*pb.Observation{}
		}
		mem.observations.Store(o.Resource.ResourceGroupName, append(existingObs.([]*pb.Observation), o))
	}
	return observations, nil
}

func (mem *MemStorage) ListResources(ctx context.Context, filter model.StorageFilter) ([]*pb.Resource, error) {
	// Resource group filter
	latestRes := map[string][]*pb.Resource{}
	if filter.ResourceGroupNames == nil {
		mem.resources.Range(func(k, v any) bool {
			if filteredV, err := filterRes(v.([]*pb.Resource), filter); err != nil {
				return false
			} else {
				latestRes[k.(string)] = filteredV
			}
			return true
		})
	} else {
		for _, n := range *filter.ResourceGroupNames {
			res, ok := mem.resources.Load(n)
			if !ok {
				continue
			}
			if filteredV, err := filterRes(res.([]*pb.Resource), filter); err != nil {
				return nil, err
			} else {
				latestRes[n] = filteredV
			}
		}
	}
	result := flatValues(latestRes)
	if filter.Limit != nil {
		if len(result) > *filter.Limit {
			return result[:*filter.Limit], nil
		}
	}
	return result, nil
}

func (mem *MemStorage) ListObservations(ctx context.Context, filter model.StorageFilter) ([]*pb.Observation, error) {
	// Resource group filter
	latestObs := map[string][]*pb.Observation{}
	if filter.ResourceGroupNames == nil {
		mem.observations.Range(func(k, v any) bool {
			if filteredV, err := filterObs(v.([]*pb.Observation), filter); err != nil {
				return false
			} else {
				latestObs[k.(string)] = filteredV
			}
			return true
		})
	} else {
		for _, n := range *filter.ResourceGroupNames {
			ob, ok := mem.observations.Load(n)
			if !ok {
				continue
			}
			if filteredV, err := filterObs(ob.([]*pb.Observation), filter); err != nil {
				return nil, err
			} else {
				latestObs[n] = filteredV
			}
		}
	}

	result := flatValues(latestObs)
	if filter.Limit != nil {
		if len(result) > *filter.Limit {
			fmt.Println(*filter.Limit)
			return result[:*filter.Limit], nil
		}
	}
	return result, nil
}

func (mem *MemStorage) AddOperationLog(ctx context.Context, ops []model.Operation) error {
	for _, o := range ops {
		mem.operations.Store(o.ID, o)
	}
	return nil
}

func (mem *MemStorage) FlushOpsLog(ctx context.Context) error {
	return nil
}

func filterObs(obs []*pb.Observation, filter model.StorageFilter) ([]*pb.Observation, error) {
	res := []*pb.Observation{}

	if len(obs) == 0 {
		return res, nil
	}

	mostRecentScanID := obs[len(obs)-1].ScanUid
	for i := len(obs) - 1; i >= 0; i-- {
		appendResource := true
		if mostRecentScanID != obs[i].ScanUid {
			break
		}
		if filter.ResourceTypes != nil {
			t, err := common.TypeFromResource(obs[i].Resource)
			if err != nil {
				return nil, err
			}
			if _, ok := toSet(*filter.ResourceTypes)[t]; !ok {
				appendResource = false
			}
		}
		if filter.ResourceIDs != nil {
			if _, ok := toSet(*filter.ResourceIDs)[obs[i].Resource.Uid]; !ok {
				appendResource = false
			}
		}
		if filter.ResourceNames != nil {
			if _, ok := toSet(*filter.ResourceNames)[obs[i].Resource.Name]; !ok {
				appendResource = false
			}
		}
		if filter.ParentNames != nil {
			if _, ok := toSet(*filter.ParentNames)[obs[i].Resource.Parent]; !ok {
				appendResource = false
			}
		}
		if filter.StartTime != nil || filter.TimeOffset != nil {
			if filter.StartTime != nil && filter.TimeOffset != nil {
				timeStamp := obs[i].Timestamp.AsTime()
				startTimeF, offsetTimeF := *filter.StartTime, (*filter.StartTime).Add(*filter.TimeOffset)
				var start, end time.Time
				if startTimeF.Before(offsetTimeF) {
					start = startTimeF
					end = offsetTimeF
				} else {
					end = startTimeF
					start = offsetTimeF
				}
				if !timeStamp.After(start) || !timeStamp.Before(end) {
					appendResource = false
				}
			} else {
				return nil, fmt.Errorf("StartTime and TimeOffset must both be set")
			}
		}
		if appendResource {
			res = append(res, obs[i])
		}
	}
	return res, nil
}

func filterRes(resources []*pb.Resource, filter model.StorageFilter) ([]*pb.Resource, error) {
	res := []*pb.Resource{}
	if len(resources) == 0 {
		return res, nil
	}

	mostRecentCollectID := resources[len(resources)-1].CollectionUid
	for i := len(resources) - 1; i >= 0; i-- {
		appendResource := true
		if mostRecentCollectID != resources[i].CollectionUid {
			break
		}
		if filter.ResourceTypes != nil {
			t, err := common.TypeFromResource(resources[i])
			if err != nil {
				return nil, err
			}
			if _, ok := toSet(*filter.ResourceTypes)[t]; !ok {
				appendResource = false
			}
		}
		if filter.ResourceIDs != nil {
			if _, ok := toSet(*filter.ResourceIDs)[resources[i].Uid]; !ok {
				appendResource = false
			}
		}
		if filter.ResourceNames != nil {
			if _, ok := toSet(*filter.ResourceNames)[resources[i].Name]; !ok {
				appendResource = false
			}
		}
		if filter.ParentNames != nil {
			if _, ok := toSet(*filter.ParentNames)[resources[i].Parent]; !ok {
				appendResource = false
			}
		}
		if filter.StartTime != nil || filter.TimeOffset != nil {
			return nil, fmt.Errorf("resources dont have a timestamp , you cannot filter by timestamp")
		}
		if appendResource {
			res = append(res, resources[i])
		}
	}
	return res, nil
}

func flatValues[T interface{}](m map[string][]T) []T {
	res := []T{}
	for _, v := range m {
		res = append(res, v...)
	}
	return res
}

func toSet[T comparable](arr []T) map[T]struct{} {
	res := map[T]struct{}{}
	for _, e := range arr {
		res[e] = struct{}{}
	}
	return res
}
