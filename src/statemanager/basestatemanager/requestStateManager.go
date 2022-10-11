package basestatemanager

import (
	"sync"

	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

type RequestStateManager struct {
	scanIds    sync.Map
	collectIds sync.Map
}

func New() (model.StateManager, error) {
	return &RequestStateManager{
		scanIds:    sync.Map{},
		collectIds: sync.Map{},
	}, nil
}

func (manager *RequestStateManager) GetCollectState(collectId string) pb.RequestStatus {
	status := pb.RequestStatus_UNKNOWN
	if v, ok := manager.collectIds.Load(collectId); ok {
		status = v.(pb.RequestStatus)
	}
	return status
}

func (manager *RequestStateManager) GetScanState(scanId string) pb.RequestStatus {
	status := pb.RequestStatus_UNKNOWN
	if v, ok := manager.scanIds.Load(scanId); ok {
		status = v.(pb.RequestStatus)
	}
	return status
}

func (manager *RequestStateManager) AddScan(scanId string, resourceGroupNames []string) []string {
	manager.scanIds.Store(scanId, pb.RequestStatus_RUNNING)
	return resourceGroupNames
}

func (manager *RequestStateManager) EndScan(scanId string, resourceGroupNames []string) {
	if _, ok := manager.scanIds.Load(scanId); ok {
		manager.scanIds.Store(scanId, pb.RequestStatus_DONE)
	}
}

func (manager *RequestStateManager) AddCollect(collectId string, resourceGroupNames []string) []string {
	manager.collectIds.Store(collectId, pb.RequestStatus_RUNNING)
	return resourceGroupNames
}

func (manager *RequestStateManager) EndCollect(collectId string, resourceGroupNames []string) {
	if _, ok := manager.collectIds.Load(collectId); ok {
		manager.collectIds.Store(collectId, pb.RequestStatus_DONE)
	}
}
