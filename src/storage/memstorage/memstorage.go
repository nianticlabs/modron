// Package memstorage provides a storage backend that runs locally in memory. It is supposed to be used primarily for API testing.
package memstorage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

type MemStorage struct {
	resources        sync.Map
	observations     sync.Map
	operations       sync.Map
	mostRecentScanID sync.Map
}

func New() model.Storage {
	return &MemStorage{
		resources:        sync.Map{},
		observations:     sync.Map{},
		operations:       sync.Map{},
		mostRecentScanID: sync.Map{},
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
		if o.Resource == nil {
			glog.Warningf("can't store observation with no attached resource: %+v", o)
			continue
		}
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
			if filteredV, err := mem.filterObs(v.([]*pb.Observation), filter); err != nil {
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
			if filteredV, err := mem.filterObs(ob.([]*pb.Observation), filter); err != nil {
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
		// Here we assume that operation are always added in chronological order.
		if o.OpsType == "scan" && o.Status == model.OperationCompleted {
			mem.mostRecentScanID.Store(o.ResourceGroup, o.ID)
		}
	}
	return nil
}

func (mem *MemStorage) FlushOpsLog(ctx context.Context) error {
	return nil
}

func (mem *MemStorage) filterObs(obs []*pb.Observation, filter model.StorageFilter) ([]*pb.Observation, error) {
	res := []*pb.Observation{}

	if len(obs) == 0 {
		return res, nil
	}

	// Here we assume that the observations are sorted by date.
	for i := len(obs) - 1; i >= 0; i-- {
		// TODO: This will fail if we insert observation without a corresponding scan.
		mostRecentScanID, ok := mem.mostRecentScanID.Load(obs[i].Resource.ResourceGroupName)
		if !ok {
			glog.Warningf("no scan found, but observation exist: %v", obs[i])
			continue
		}
		appendResource := true
		if obs[i].ScanUid != mostRecentScanID {
			continue
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
				start, end := extractStartAndEndTimes(filter)
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

	// Here we assume that the resources are sorted by date.
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
			if filter.StartTime != nil && filter.TimeOffset != nil {
				timeStamp := resources[i].GetTimestamp().AsTime()
				start, end := extractStartAndEndTimes(filter)
				if !timeStamp.After(start) || !timeStamp.Before(end) {
					appendResource = false
				}
			} else {
				return nil, fmt.Errorf("StartTime and TimeOffset must both be set")
			}
		}
		if appendResource {
			res = append(res, resources[i])
		}
	}
	return res, nil
}

func extractStartAndEndTimes(filter model.StorageFilter) (start time.Time, end time.Time) {
	startTimeF, offsetTimeF := filter.StartTime, filter.StartTime.Add(*filter.TimeOffset)
	if startTimeF.Before(offsetTimeF) {
		start = *startTimeF
		end = offsetTimeF
	} else {
		end = *startTimeF
		start = offsetTimeF
	}
	return start, end
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
