package reqdepstatemanager

import (
	"sync"

	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

type RequestStateManager struct {
	scanIds                  sync.Map
	scanIdsDependencies      map[string]map[string]struct{}
	collectIds               sync.Map
	collectIdsDependencies   map[string]map[string]struct{}
	resourceGroupsScanning   map[string]string
	resourceGroupsCollecting map[string]string
}

func New() (model.StateManager, error) {
	return &RequestStateManager{
		scanIds:                  sync.Map{},
		scanIdsDependencies:      map[string]map[string]struct{}{},
		collectIds:               sync.Map{},
		collectIdsDependencies:   map[string]map[string]struct{}{},
		resourceGroupsScanning:   map[string]string{},
		resourceGroupsCollecting: map[string]string{},
	}, nil
}

func (manager *RequestStateManager) GetCollectState(collectId string) pb.RequestStatus {
	status := pb.RequestStatus_UNKNOWN
	if v, ok := manager.collectIds.Load(collectId); ok {
		status = v.(pb.RequestStatus)
	}
	if status == pb.RequestStatus_CANCELLED ||
		status == pb.RequestStatus_UNKNOWN {
		return status
	}
	if status == pb.RequestStatus_ALREADY_RUNNING {
		manager.collectIds.Store(collectId, pb.RequestStatus_DONE)
		status = pb.RequestStatus_DONE
	}
	if mapDep, ok := manager.collectIdsDependencies[collectId]; ok {
		for dep := range mapDep {
			state := manager.GetCollectState(dep)
			if state == pb.RequestStatus_UNKNOWN ||
				state == pb.RequestStatus_CANCELLED {
				manager.collectIds.Store(collectId, pb.RequestStatus_CANCELLED)
			} else if state == pb.RequestStatus_RUNNING {
				return pb.RequestStatus_RUNNING
			}
		}
	}
	return status
}

func (manager *RequestStateManager) GetScanState(scanId string) pb.RequestStatus {
	status := pb.RequestStatus_UNKNOWN
	if v, ok := manager.scanIds.Load(scanId); ok {
		status = v.(pb.RequestStatus)
	}
	if status == pb.RequestStatus_CANCELLED ||
		status == pb.RequestStatus_UNKNOWN {
		return status
	}
	if status == pb.RequestStatus_ALREADY_RUNNING {
		manager.scanIds.Store(scanId, pb.RequestStatus_DONE)
		status = pb.RequestStatus_DONE
	}
	if mapDep, ok := manager.scanIdsDependencies[scanId]; ok {
		for dep := range mapDep {
			state := manager.GetScanState(dep)
			if state == pb.RequestStatus_UNKNOWN ||
				state == pb.RequestStatus_CANCELLED {
				manager.scanIds.Store(scanId, pb.RequestStatus_CANCELLED)
			} else if state == pb.RequestStatus_RUNNING {
				return pb.RequestStatus_RUNNING
			}
		}
	}
	return status
}

func (manager *RequestStateManager) AddScan(scanId string, resourceGroupNames []string) []string {
	manager.scanIds.Store(scanId, pb.RequestStatus_RUNNING)
	filteredRG := []string{}
	for _, rs := range resourceGroupNames {
		scan, ok := manager.resourceGroupsScanning[rs]
		if !ok {
			manager.resourceGroupsScanning[rs] = scanId
			filteredRG = append(filteredRG, rs)
		} else {
			if _, ok := manager.scanIdsDependencies[scanId]; !ok {
				manager.scanIdsDependencies[scanId] = map[string]struct{}{}
			}
			manager.scanIdsDependencies[scanId][scan] = struct{}{}
		}
	}
	if len(filteredRG) < 1 {
		manager.scanIds.Store(scanId, pb.RequestStatus_ALREADY_RUNNING)
	}
	return filteredRG
}

func (manager *RequestStateManager) EndScan(scanId string, resourceGroupNames []string) {
	if _, ok := manager.scanIds.Load(scanId); ok {
		for _, rs := range resourceGroupNames {
			delete(manager.resourceGroupsScanning, rs)
		}
		manager.scanIds.Store(scanId, pb.RequestStatus_DONE)
	}
}

func (manager *RequestStateManager) AddCollect(collectId string, resourceGroupNames []string) []string {
	manager.collectIds.Store(collectId, pb.RequestStatus_RUNNING)
	filteredRG := []string{}
	for _, rs := range resourceGroupNames {
		collect, ok := manager.resourceGroupsCollecting[rs]
		if !ok {
			manager.resourceGroupsCollecting[rs] = collectId
			filteredRG = append(filteredRG, rs)
		} else {
			if _, ok := manager.collectIdsDependencies[collectId]; !ok {
				manager.collectIdsDependencies[collectId] = map[string]struct{}{}
			}
			manager.collectIdsDependencies[collectId][collect] = struct{}{}
		}
	}
	if len(filteredRG) < 1 {
		manager.collectIds.Store(collectId, pb.RequestStatus_ALREADY_RUNNING)
	}
	return filteredRG
}

func (manager *RequestStateManager) EndCollect(collectId string, resourceGroupNames []string) {
	if _, ok := manager.collectIds.Load(collectId); ok {
		for _, rs := range resourceGroupNames {
			delete(manager.resourceGroupsCollecting, rs)
		}
		manager.collectIds.Store(collectId, pb.RequestStatus_DONE)
	}
}
