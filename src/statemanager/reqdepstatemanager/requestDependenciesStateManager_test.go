package reqdepstatemanager

import (
	"testing"

	pb "github.com/nianticlabs/modron/src/proto/generated"
)

func TestSimpleStateManager(t *testing.T) {
	stateManager, err := New()
	if err != nil {
		t.Fatal(err)
	}

	collectID1 := "collect-id-1"
	collectID2 := "collect-id-2"
	scanID1 := "scan-id-1"
	scanID2 := "scan-id-2"

	if state := stateManager.GetCollectState(collectID1); state != pb.RequestStatus_UNKNOWN {
		t.Errorf("GetCollectState(%s) got %s, want %s", collectID1, state, pb.RequestStatus_UNKNOWN)
	}
	if state := stateManager.GetScanState(scanID1); state != pb.RequestStatus_UNKNOWN {
		t.Errorf("GetScanState(%s) got %s, want %s", scanID1, state, pb.RequestStatus_UNKNOWN)
	}

	resourceGroups := []string{"projects/p1", "projects/p2", "projects/p3"}
	collecting := stateManager.AddCollect(collectID1, resourceGroups)
	if len(collecting) != 3 {
		t.Errorf("AddCollect(%v): got len %d, want %d", resourceGroups, len(collecting), 3)
	}

	scanning := stateManager.AddScan(scanID1, resourceGroups)
	if len(scanning) != 3 {
		t.Errorf("AddScan(%v): got len %d, want %d", resourceGroups, len(scanning), 3)
	}

	if state := stateManager.GetCollectState(collectID1); state != pb.RequestStatus_RUNNING {
		t.Errorf("GetCollectState(%s): got %s, want %s", collectID1, state, pb.RequestStatus_RUNNING)
	}
	if state := stateManager.GetCollectState(collectID2); state != pb.RequestStatus_UNKNOWN {
		t.Errorf("GetCollectState(%s): got %s, want %s", collectID2, state, pb.RequestStatus_UNKNOWN)
	}

	if state := stateManager.GetScanState(scanID1); state != pb.RequestStatus_RUNNING {
		t.Errorf("GetScanState(%s): got %s, want %s", scanID1, state, pb.RequestStatus_RUNNING)
	}
	if state := stateManager.GetScanState(scanID2); state != pb.RequestStatus_UNKNOWN {
		t.Errorf("GetScanState(%s): got %s, want %s", scanID2, state, pb.RequestStatus_UNKNOWN)
	}

	stateManager.EndCollect(collectID2, resourceGroups)
	stateManager.EndScan(scanID2, resourceGroups)

	stateManager.EndCollect(collectID1, resourceGroups)
	stateManager.EndScan(scanID1, resourceGroups)

	if state := stateManager.GetCollectState(collectID1); state != pb.RequestStatus_DONE {
		t.Errorf("GetCollectState(%s) got %s, want %s", collectID1, state, pb.RequestStatus_DONE)
	}

	if state := stateManager.GetScanState(scanID1); state != pb.RequestStatus_DONE {
		t.Errorf("GetScanState(%s): got %s, want %s", scanID1, state, pb.RequestStatus_DONE)
	}
}

func TestDepSateManager(t *testing.T) {
	stateManager, err := New()
	if err != nil {
		t.Fatal(err)
	}

	collectID1 := "collect-id-1"
	collectID2 := "collect-id-2"
	collectID3 := "collect-id-3"
	scanID1 := "scan-id-1"
	scanID2 := "scan-id-2"
	scanID3 := "scan-id-3"
	resourceGroups := []string{"projects/p1", "projects/p2", "projects/p3"}

	if collecting := stateManager.AddCollect(collectID1, resourceGroups); len(collecting) != 3 {
		t.Errorf("AddCollect(%v): got len %d, want %d", resourceGroups, len(collecting), 3)
	}

	overlappingResourceGroups := []string{"projects/p0", "projects/p1", "projects/p2"}
	if collecting := stateManager.AddCollect("collect-id-3", overlappingResourceGroups); len(collecting) != 1 {
		t.Errorf("AddCollect(%s): got len %d, want %d", collectID1, len(collecting), 1)
	}

	if scanning := stateManager.AddScan(scanID1, resourceGroups); len(scanning) != 3 {
		t.Errorf("AddScan(%s): got len %d, want %d", scanID1, len(scanning), 1)
	}

	if scanning := stateManager.AddScan(scanID3, []string{"projects/p1", "projects/p2"}); len(scanning) != 0 {
		t.Errorf("AddScan(%s): got len %d, want %d", scanID3, len(scanning), 0)
	}

	if state := stateManager.GetCollectState(collectID1); state != pb.RequestStatus_RUNNING {
		t.Errorf("GetCollectState(%s): got %s, want %s", collectID1, state, pb.RequestStatus_RUNNING)
	}
	if state := stateManager.GetCollectState(collectID2); state != pb.RequestStatus_UNKNOWN {
		t.Errorf("GetCollectState(%s): got %s, want %s", collectID2, state, pb.RequestStatus_UNKNOWN)
	}
	if state := stateManager.GetCollectState(collectID3); state != pb.RequestStatus_RUNNING {
		t.Errorf("GetCollectState(%s): got %s, want %s", collectID3, state, pb.RequestStatus_RUNNING)
	}

	if state := stateManager.GetScanState(scanID1); state != pb.RequestStatus_RUNNING {
		t.Errorf("GetScanState(%s): got %s, want %s", scanID1, state, pb.RequestStatus_RUNNING)
	}
	if state := stateManager.GetScanState(scanID2); state != pb.RequestStatus_UNKNOWN {
		t.Errorf("GetScanState(%s): got %s, want %s", scanID2, state, pb.RequestStatus_UNKNOWN)
	}
	if state := stateManager.GetScanState(scanID3); state != pb.RequestStatus_RUNNING {
		t.Errorf("GetScanState(%s): got %s, want %s", scanID3, state, pb.RequestStatus_RUNNING)
	}

	stateManager.EndCollect(collectID1, resourceGroups)
	stateManager.EndCollect(collectID2, resourceGroups)
	stateManager.EndScan(scanID1, resourceGroups)
	stateManager.EndScan(scanID2, resourceGroups)

	if state := stateManager.GetCollectState(collectID1); state != pb.RequestStatus_DONE {
		t.Errorf("GetCollectState(%s): got %s, want %s", collectID1, state, pb.RequestStatus_DONE)
	}
	if state := stateManager.GetCollectState(collectID2); state != pb.RequestStatus_UNKNOWN {
		t.Errorf("GetCollectState(%s): got %s, want %s", collectID2, state, pb.RequestStatus_UNKNOWN)
	}

	if state := stateManager.GetScanState(scanID1); state != pb.RequestStatus_DONE {
		t.Errorf("GetScanState(%s): got %s, want %s", scanID1, state, pb.RequestStatus_DONE)
	}
	if state := stateManager.GetScanState(scanID2); state != pb.RequestStatus_UNKNOWN {
		t.Errorf("GetScanState(%s): got %s, want %s", scanID2, state, pb.RequestStatus_UNKNOWN)
	}
	if state := stateManager.GetScanState(scanID3); state != pb.RequestStatus_DONE {
		t.Errorf("GetScanState(%s): got %s, want %s", scanID3, state, pb.RequestStatus_DONE)
	}

	if state := stateManager.GetCollectState(collectID3); state != pb.RequestStatus_RUNNING {
		t.Errorf("GetCollectState(%s): got %s, want %s", collectID3, state, pb.RequestStatus_RUNNING)
	}
	stateManager.EndCollect(collectID3, overlappingResourceGroups)
	if state := stateManager.GetCollectState(collectID3); state != pb.RequestStatus_DONE {
		t.Errorf("GetCollectState(%s): got %s, want %s", collectID3, state, pb.RequestStatus_DONE)
	}
}
