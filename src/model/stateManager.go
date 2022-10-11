package model

import (
	"github.com/nianticlabs/modron/src/pb"
)

type StateManager interface {
	GetCollectState(collectId string) pb.RequestStatus
	GetScanState(scanId string) pb.RequestStatus

	AddScan(scanId string, resourceGroupNames []string) []string
	EndScan(scanId string, resourceGroupNames []string)

	AddCollect(collectId string, resourceGroupNames []string) []string
	EndCollect(collectId string, resourceGroupNames []string)
}
