package reqdepstatemanager

import (
	"sync"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
)

type RequestStateManager struct {
	scanIDs                  sync.Map
	scanIDsDependencies      map[string]map[string]struct{}
	collectIDs               sync.Map
	collectIDsDependencies   map[string]map[string]struct{}
	resourceGroupsScanning   map[string]string
	resourceGroupsCollecting map[string]string
}

func New() (model.StateManager, error) {
	return &RequestStateManager{
		scanIDs:                  sync.Map{},
		scanIDsDependencies:      map[string]map[string]struct{}{},
		collectIDs:               sync.Map{},
		collectIDsDependencies:   map[string]map[string]struct{}{},
		resourceGroupsScanning:   map[string]string{},
		resourceGroupsCollecting: map[string]string{},
	}, nil
}

func (manager *RequestStateManager) GetCollectState(collectID string) pb.RequestStatus {
	status := pb.RequestStatus_UNKNOWN
	if v, ok := manager.collectIDs.Load(collectID); ok {
		status = v.(pb.RequestStatus)
	}
	if status == pb.RequestStatus_CANCELLED ||
		status == pb.RequestStatus_UNKNOWN {
		return status
	}
	if status == pb.RequestStatus_ALREADY_RUNNING {
		manager.collectIDs.Store(collectID, pb.RequestStatus_DONE)
		status = pb.RequestStatus_DONE
	}
	if mapDep, ok := manager.collectIDsDependencies[collectID]; ok {
		for dep := range mapDep {
			state := manager.GetCollectState(dep)
			if state == pb.RequestStatus_UNKNOWN ||
				state == pb.RequestStatus_CANCELLED {
				manager.collectIDs.Store(collectID, pb.RequestStatus_CANCELLED)
			} else if state == pb.RequestStatus_RUNNING {
				return pb.RequestStatus_RUNNING
			}
		}
	}
	return status
}

func (manager *RequestStateManager) GetScanState(scanID string) pb.RequestStatus {
	status := pb.RequestStatus_UNKNOWN
	if v, ok := manager.scanIDs.Load(scanID); ok {
		status = v.(pb.RequestStatus)
	}
	if status == pb.RequestStatus_CANCELLED ||
		status == pb.RequestStatus_UNKNOWN {
		return status
	}
	if status == pb.RequestStatus_ALREADY_RUNNING {
		manager.scanIDs.Store(scanID, pb.RequestStatus_DONE)
		status = pb.RequestStatus_DONE
	}
	if mapDep, ok := manager.scanIDsDependencies[scanID]; ok {
		for dep := range mapDep {
			state := manager.GetScanState(dep)
			if state == pb.RequestStatus_UNKNOWN ||
				state == pb.RequestStatus_CANCELLED {
				manager.scanIDs.Store(scanID, pb.RequestStatus_CANCELLED)
			} else if state == pb.RequestStatus_RUNNING {
				return pb.RequestStatus_RUNNING
			}
		}
	}
	return status
}

func (manager *RequestStateManager) AddScan(scanID string, resourceGroupNames []string) []string {
	manager.scanIDs.Store(scanID, pb.RequestStatus_RUNNING)
	filteredRG := []string{}
	for _, rs := range resourceGroupNames {
		scan, ok := manager.resourceGroupsScanning[rs]
		if !ok {
			manager.resourceGroupsScanning[rs] = scanID
			filteredRG = append(filteredRG, rs)
		} else {
			if _, ok := manager.scanIDsDependencies[scanID]; !ok {
				manager.scanIDsDependencies[scanID] = map[string]struct{}{}
			}
			manager.scanIDsDependencies[scanID][scan] = struct{}{}
		}
	}
	if len(filteredRG) < 1 {
		manager.scanIDs.Store(scanID, pb.RequestStatus_ALREADY_RUNNING)
	}
	return filteredRG
}

func (manager *RequestStateManager) EndScan(scanID string, resourceGroupNames []string) {
	if _, ok := manager.scanIDs.Load(scanID); ok {
		for _, rs := range resourceGroupNames {
			delete(manager.resourceGroupsScanning, rs)
		}
		manager.scanIDs.Store(scanID, pb.RequestStatus_DONE)
	}
}

func (manager *RequestStateManager) AddCollect(collectID string, resourceGroupNames []string) []string {
	manager.collectIDs.Store(collectID, pb.RequestStatus_RUNNING)
	filteredRG := []string{}
	for _, rs := range resourceGroupNames {
		collect, ok := manager.resourceGroupsCollecting[rs]
		if !ok {
			manager.resourceGroupsCollecting[rs] = collectID
			filteredRG = append(filteredRG, rs)
		} else {
			if _, ok := manager.collectIDsDependencies[collectID]; !ok {
				manager.collectIDsDependencies[collectID] = map[string]struct{}{}
			}
			manager.collectIDsDependencies[collectID][collect] = struct{}{}
		}
	}
	if len(filteredRG) < 1 {
		manager.collectIDs.Store(collectID, pb.RequestStatus_ALREADY_RUNNING)
	}
	return filteredRG
}

func (manager *RequestStateManager) EndCollect(collectID string, resourceGroupNames []string) {
	if _, ok := manager.collectIDs.Load(collectID); ok {
		for _, rs := range resourceGroupNames {
			delete(manager.resourceGroupsCollecting, rs)
		}
		manager.collectIDs.Store(collectID, pb.RequestStatus_DONE)
	}
}
