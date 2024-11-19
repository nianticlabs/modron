package model

import (
	pb "github.com/nianticlabs/modron/src/proto/generated"
)

type StateManager interface {
	GetCollectState(collectID string) pb.RequestStatus
	GetScanState(scanID string) pb.RequestStatus

	AddScan(scanID string, resourceGroupNames []string) []string
	EndScan(scanID string, resourceGroupNames []string)

	AddCollect(collectID string, resourceGroupNames []string) []string
	EndCollect(collectID string, resourceGroupNames []string)
}
