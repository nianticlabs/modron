package reqdepstatemanager

import (
	"testing"

	"github.com/nianticlabs/modron/src/pb"
)

func TestSimpleSateManager(t *testing.T) {

	stateManager, err := New()
	if err != nil {
		t.Fatal(err)
	}

	collectId1 := "collect-id-1"
	collectId2 := "collect-id-2"
	scanId1 := "scan-id-1"
	scanId2 := "scan-id-2"

	if state := stateManager.GetCollectState(collectId1); state != pb.RequestStatus_UNKNOWN {
		t.Errorf("GetCollectState(%s) got %s, want %s", collectId1, state, pb.RequestStatus_UNKNOWN)
	}
	if state := stateManager.GetScanState(scanId1); state != pb.RequestStatus_UNKNOWN {
		t.Errorf("GetScanState(%s) got %s, want %s", scanId1, state, pb.RequestStatus_UNKNOWN)
	}

	resourceGroups := []string{"p1", "p2", "p3"}
	collecting := stateManager.AddCollect(collectId1, resourceGroups)
	if len(collecting) != 3 {
		t.Errorf("AddCollect(%v): got len %d, want %d", resourceGroups, len(collecting), 3)
	}

	scanning := stateManager.AddScan(scanId1, resourceGroups)
	if len(scanning) != 3 {
		t.Errorf("AddScan(%v): got len %d, want %d", resourceGroups, len(scanning), 3)
	}

	if state := stateManager.GetCollectState(collectId1); state != pb.RequestStatus_RUNNING {
		t.Errorf("GetCollectState(%s): got %s, want %s", collectId1, state, pb.RequestStatus_RUNNING)
	}
	if state := stateManager.GetCollectState(collectId2); state != pb.RequestStatus_UNKNOWN {
		t.Errorf("GetCollectState(%s): got %s, want %s", collectId2, state, pb.RequestStatus_UNKNOWN)
	}

	if state := stateManager.GetScanState(scanId1); state != pb.RequestStatus_RUNNING {
		t.Errorf("GetScanState(%s): got %s, want %s", scanId1, state, pb.RequestStatus_RUNNING)
	}
	if state := stateManager.GetScanState(scanId2); state != pb.RequestStatus_UNKNOWN {
		t.Errorf("GetScanState(%s): got %s, want %s", scanId2, state, pb.RequestStatus_UNKNOWN)
	}

	stateManager.EndCollect(collectId2, resourceGroups)
	stateManager.EndScan(scanId2, resourceGroups)

	stateManager.EndCollect(collectId1, resourceGroups)
	stateManager.EndScan(scanId1, resourceGroups)

	if state := stateManager.GetCollectState(collectId1); state != pb.RequestStatus_DONE {
		t.Errorf("GetCollectState(%s) got %s, want %s", collectId1, state, pb.RequestStatus_DONE)
	}

	if state := stateManager.GetScanState(scanId1); state != pb.RequestStatus_DONE {
		t.Errorf("GetScanState(%s): got %s, want %s", scanId1, state, pb.RequestStatus_DONE)
	}
}

func TestDepSateManager(t *testing.T) {

	stateManager, err := New()
	if err != nil {
		t.Fatal(err)
	}

	collectId1 := "collect-id-1"
	collectId2 := "collect-id-2"
	scanId1 := "scan-id-1"
	scanId2 := "scan-id-2"
	resourceGroups := []string{"p1", "p2", "p3"}

	if state := stateManager.GetCollectState(collectId1); state != pb.RequestStatus_UNKNOWN {
		t.Errorf("GetCollectState(%s) got %s, want %s", collectId1, state, pb.RequestStatus_UNKNOWN)
	}
	if state := stateManager.GetScanState(scanId1); state != pb.RequestStatus_UNKNOWN {
		t.Errorf("GetScanState(%s) got %s, want %s", scanId1, state, pb.RequestStatus_UNKNOWN)
	}

	if collecting := stateManager.AddCollect(collectId1, resourceGroups); len(collecting) != 3 {
		t.Errorf("AddCollect(%v): got len %d, want %d", resourceGroups, len(collecting), 3)
	}

	overlappingResourceGroups := []string{"p0", "p1", "p2"}
	if collecting := stateManager.AddCollect("collect-id-3", overlappingResourceGroups); len(collecting) != 1 {
		t.Errorf("AddCollect(%s): got len %d, want %d", collectId1, len(collecting), 1)
	}

	if scanning := stateManager.AddScan(scanId1, resourceGroups); len(scanning) != 3 {
		t.Errorf("AddScan(%s): got len %d, want %d", scanId1, len(scanning), 1)
	}
	scanId3 := "scan-id-3"
	if scanning := stateManager.AddScan(scanId3, []string{"p1", "p2"}); len(scanning) != 0 {
		t.Errorf("AddScan(%s): got len %d, want %d", scanId3, len(scanning), 0)
	}

	if state := stateManager.GetCollectState(collectId1); state != pb.RequestStatus_RUNNING {
		t.Errorf("GetCollectState(%s): got %s, want %s", collectId1, state, pb.RequestStatus_RUNNING)
	}
	collectId3 := "collect-id-3"
	if state := stateManager.GetCollectState(collectId3); state != pb.RequestStatus_RUNNING {
		t.Errorf("GetCollectState(%s): got %s, want %s", collectId3, state, pb.RequestStatus_RUNNING)
	}
	if state := stateManager.GetCollectState(collectId2); state != pb.RequestStatus_UNKNOWN {
		t.Errorf("GetCollectState(%s): got %s, want %s", collectId2, state, pb.RequestStatus_UNKNOWN)
	}

	if state := stateManager.GetScanState(scanId1); state != pb.RequestStatus_RUNNING {
		t.Errorf("GetScanState(%s): got %s, want %s", scanId1, state, pb.RequestStatus_RUNNING)
	}
	if state := stateManager.GetScanState(scanId3); state != pb.RequestStatus_RUNNING {
		t.Errorf("GetScanState(%s): got %s, want %s", scanId3, state, pb.RequestStatus_RUNNING)
	}
	if state := stateManager.GetScanState(scanId2); state != pb.RequestStatus_UNKNOWN {
		t.Errorf("GetScanState(%s): got %s, want %s", scanId2, state, pb.RequestStatus_UNKNOWN)
	}

	stateManager.EndCollect(collectId2, resourceGroups)
	stateManager.EndScan(scanId2, resourceGroups)

	stateManager.EndCollect(collectId1, resourceGroups)
	stateManager.EndScan(scanId1, resourceGroups)

	if state := stateManager.GetCollectState(collectId1); state != pb.RequestStatus_DONE {
		t.Errorf("GetCollectState(%s): got %s, want %s", collectId1, state, pb.RequestStatus_DONE)
	}
	if state := stateManager.GetCollectState(collectId2); state != pb.RequestStatus_UNKNOWN {
		t.Errorf("GetCollectState(%s): got %s, want %s", collectId2, state, pb.RequestStatus_UNKNOWN)
	}

	if state := stateManager.GetScanState(scanId1); state != pb.RequestStatus_DONE {
		t.Errorf("GetScanState(%s): got %s, want %s", scanId1, state, pb.RequestStatus_DONE)
	}
	if state := stateManager.GetScanState(scanId3); state != pb.RequestStatus_DONE {
		t.Errorf("GetScanState(%s): got %s, want %s", scanId3, state, pb.RequestStatus_DONE)
	}
	if state := stateManager.GetScanState(scanId2); state != pb.RequestStatus_UNKNOWN {
		t.Errorf("GetScanState(%s): got %s, want %s", scanId2, state, pb.RequestStatus_UNKNOWN)
	}

	if state := stateManager.GetCollectState(collectId3); state != pb.RequestStatus_RUNNING {
		t.Errorf("GetCollectState(%s): got %s, want %s", collectId3, state, pb.RequestStatus_RUNNING)
	}
	stateManager.EndCollect(collectId3, overlappingResourceGroups)
	if state := stateManager.GetCollectState(collectId3); state != pb.RequestStatus_DONE {
		t.Errorf("GetCollectState(%s): got %s, want %s", collectId3, state, pb.RequestStatus_DONE)
	}
}
