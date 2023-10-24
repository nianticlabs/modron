package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
	"github.com/nianticlabs/modron/src/storage/test"

	_ "modernc.org/sqlite"
)

const (
	sqliteTestDSN        = "file:test.db?cache=private&mode=memory"
	testObservationTable = "observations"
	testOperationTable   = "operations"
	testResourceTable    = "resources"
)

var sorted = cmp.Transformer("sort", func(in []*pb.Resource) []*pb.Resource {
	out := append([]*pb.Resource{}, in...)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Uid < out[j].Uid
	})
	return out
})

func newTestDB(t *testing.T) model.Storage {
	t.Helper()
	db, err := sql.Open("sqlite", sqliteTestDSN)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	createDB(t, db, "schema.sql")
	cfg := Config{testResourceTable, testObservationTable, testOperationTable, 1}
	sqlClient, err := New(db, cfg)
	if err != nil {
		t.Fatalf("New(%+v) unexpected error: %v", cfg, err)
	}
	return sqlClient
}

func TestResourceStorageSql(t *testing.T) {
	ctx := context.Background()
	sqlClient := newTestDB(t)
	oldCollection := uuid.NewString()
	collectionID := uuid.NewString()
	resourceGroup := "projects/project-id"
	otherResourceGroup := "projects/other-resource-group"
	deletedResourceGroup := "project/deleted"
	allResources := []*pb.Resource{
		{
			Uid:               uuid.NewString(),
			CollectionUid:     oldCollection,
			Timestamp:         timestamppb.New(time.Now().Add(-time.Hour * 24 * 8)),
			DisplayName:       "test-resource-old-collection",
			Name:              "test-old-resource",
			Parent:            "projects/test-parent",
			ResourceGroupName: resourceGroup,
			Type:              &pb.Resource_Bucket{},
		},
		{
			Uid:               uuid.NewString(),
			CollectionUid:     collectionID,
			Timestamp:         timestamppb.Now(),
			DisplayName:       "test-resource",
			Name:              "test-resource",
			Parent:            "projects/test-parent",
			ResourceGroupName: resourceGroup,
			Type:              &pb.Resource_Bucket{},
		},
		{
			Uid:               uuid.NewString(),
			CollectionUid:     collectionID,
			Timestamp:         timestamppb.Now(),
			DisplayName:       "test-resource",
			Name:              "test-resource",
			Parent:            "projects/test-other-parent",
			ResourceGroupName: otherResourceGroup,
			Type:              &pb.Resource_Certificate{},
		},
		{
			Uid:               uuid.NewString(),
			CollectionUid:     oldCollection,
			Timestamp:         timestamppb.New(time.Now().Add(-time.Hour * 24 * 8)),
			DisplayName:       "test-resource-old-collection",
			Name:              "test-old-resource",
			Parent:            "projects/test-deleted-parent",
			ResourceGroupName: deletedResourceGroup,
			Type:              &pb.Resource_Bucket{},
		},
	}

	ops := []model.Operation{
		{
			ID:            oldCollection,
			ResourceGroup: resourceGroup,
			OpsType:       "collection",
			StatusTime:    time.Now().Add(-time.Hour * 24 * 8),
			Status:        model.OperationStarted,
		},
		{
			ID:            oldCollection,
			ResourceGroup: resourceGroup,
			OpsType:       "collection",
			StatusTime:    time.Now().Add(-time.Hour * 24 * 8),
			Status:        model.OperationCompleted,
		},
		{
			ID:            oldCollection,
			ResourceGroup: deletedResourceGroup,
			OpsType:       "collection",
			StatusTime:    time.Now().Add(-time.Hour * 24 * 8),
			Status:        model.OperationStarted,
		},
		{
			ID:            oldCollection,
			ResourceGroup: deletedResourceGroup,
			OpsType:       "collection",
			StatusTime:    time.Now().Add(-time.Hour * 24 * 8),
			Status:        model.OperationCompleted,
		},
		{
			ID:            collectionID,
			ResourceGroup: resourceGroup,
			OpsType:       "collection",
			StatusTime:    time.Now().Add(-time.Second * 60),
			Status:        model.OperationStarted,
		},
		{
			ID:            collectionID,
			ResourceGroup: resourceGroup,
			OpsType:       "collection",
			StatusTime:    time.Now().Add(-time.Second * 10),
			Status:        model.OperationCompleted,
		},
		{
			ID:            collectionID,
			ResourceGroup: otherResourceGroup,
			OpsType:       "collection",
			StatusTime:    time.Now().Add(-time.Second * 60),
			Status:        model.OperationStarted,
		},
		{
			ID:            collectionID,
			ResourceGroup: otherResourceGroup,
			OpsType:       "collection",
			StatusTime:    time.Now().Add(-time.Second * 10),
			Status:        model.OperationCompleted,
		},
	}

	err := sqlClient.AddOperationLog(ctx, ops)
	if err != nil {
		t.Fatalf("AddOperationLog(%+v) unexpected error: %v", ops, err)
	}

	got, err := sqlClient.BatchCreateResources(ctx, allResources)
	if err != nil {
		t.Fatalf("BatchCreateResources(%+v) unexpected error: %v", allResources, err)
	}

	if diff := cmp.Diff(allResources, got, protocmp.Transform(), sorted); diff != "" {
		t.Errorf("unexpected diff (-want, +got): %v", diff)
	}

	gotRead, err := sqlClient.ListResources(ctx, model.StorageFilter{ResourceGroupNames: []string{resourceGroup}})
	if err != nil {
		t.Errorf("ListResources() unexpected error: %v", err)
	}

	newResources := []*pb.Resource{allResources[1]}
	if diff := cmp.Diff(newResources, gotRead, protocmp.Transform(), sorted); diff != "" {
		t.Errorf("unexpected diff (-want, +got): %v", diff)
	}

	sevenDaysAgo := time.Now().Add(-time.Hour * 24 * 7)
	gotReadFiltered, err := sqlClient.ListResources(ctx, model.StorageFilter{
		ResourceTypes: []string{common.ResourceBucket},
		StartTime:     sevenDaysAgo,
		TimeOffset:    time.Since(sevenDaysAgo),
	})
	if err != nil {
		t.Errorf("ListResources() unexpected error: %v", err)
	}
	if diff := cmp.Diff(newResources, gotReadFiltered, protocmp.Transform(), sorted); diff != "" {
		t.Errorf("unexpected diff (-want, +got): %v", diff)
	}

	gotReadFilteredEmpty, err := sqlClient.ListResources(ctx, model.StorageFilter{ResourceTypes: []string{common.ResourceDatabase}})
	if err != nil {
		t.Errorf("ListResources() unexpected error: %v", err)
	}
	if len(gotReadFilteredEmpty) > 0 {
		t.Errorf("len(gotReadFilteredEmpty) is %d, should be 0", len(gotReadFilteredEmpty))
	}

	gotReadNoOutdated, err := sqlClient.ListResources(ctx, model.StorageFilter{
		StartTime:  sevenDaysAgo,
		TimeOffset: time.Since(sevenDaysAgo),
	})
	if err != nil {
		t.Errorf("ListResources() unexpected error: %v", err)
	}

	allNewerResources := []*pb.Resource{allResources[1], allResources[2]}
	if diff := cmp.Diff(allNewerResources, gotReadNoOutdated, protocmp.Transform(), sorted); diff != "" {
		t.Errorf("unexpected diff (-want, +got): %v", diff)
	}
}

func TestBatchObservationInsert(t *testing.T) {
	ctx := context.Background()
	sqlClient := newTestDB(t)

	oldScan := uuid.NewString()
	scanID := uuid.NewString()
	resourceGroup := "projects/project-id"
	allObservations := []*pb.Observation{
		{
			Uid:       uuid.NewString(),
			ScanUid:   oldScan,
			Timestamp: timestamppb.New(time.Now().Add(-time.Hour * 24 * 8)),
			Name:      "test-observation-old-scan",
			Resource: &pb.Resource{
				Uid:               uuid.NewString(),
				ResourceGroupName: resourceGroup,
			},
		},
		{
			Uid:       uuid.NewString(),
			ScanUid:   scanID,
			Timestamp: timestamppb.Now(),
			Name:      "test-observation",
			Resource: &pb.Resource{
				Uid:               uuid.NewString(),
				ResourceGroupName: resourceGroup,
			},
		},
	}

	ops := []model.Operation{
		{
			ID:            oldScan,
			ResourceGroup: resourceGroup,
			OpsType:       "scan",
			StatusTime:    time.Now().Add(-time.Hour * 24 * 8),
			Status:        model.OperationStarted,
		},
		{
			ID:            oldScan,
			ResourceGroup: resourceGroup,
			OpsType:       "scan",
			StatusTime:    time.Now().Add(-time.Hour * 24 * 8),
			Status:        model.OperationCompleted,
		},
		{
			ID:            scanID,
			ResourceGroup: resourceGroup,
			OpsType:       "scan",
			StatusTime:    time.Now(),
			Status:        model.OperationStarted,
		},
		{
			ID:            scanID,
			ResourceGroup: resourceGroup,
			OpsType:       "scan",
			StatusTime:    time.Now(),
			Status:        model.OperationCompleted,
		},
	}

	err := sqlClient.AddOperationLog(ctx, ops)
	if err != nil {
		t.Errorf("AddOperationLog(%+v) unexpected error: %v", ops, err)
	}

	got, err := sqlClient.BatchCreateObservations(ctx, allObservations)
	if err != nil {
		t.Fatalf("BatchCreateObservations(%+v) unexpected error: %v", allObservations, err)
	}

	if diff := cmp.Diff(allObservations, got, protocmp.Transform(), sorted); diff != "" {
		t.Errorf("unexpected diff (-want, +got): %v", diff)
	}

	gotRead, err := sqlClient.ListObservations(ctx, model.StorageFilter{})
	if err != nil {
		t.Errorf("ListObservations() unexpected error: %v", err)
	}

	want := []*pb.Observation{allObservations[1]}
	if diff := cmp.Diff(want, gotRead, protocmp.Transform()); diff != "" {
		t.Errorf("unexpected diff (-want, +got): %v", diff)
	}
}

func TestOperationLog(t *testing.T) {
	// We can't use newTestDB as we need direct access to the database.
	db, err := sql.Open("sqlite", sqliteTestDSN)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	createDB(t, db, "schema.sql")
	cfg := Config{testResourceTable, testObservationTable, testOperationTable, 1}
	sqlClient, err := New(db, cfg)
	if err != nil {
		t.Fatalf("New(%+v) unexpected error: %v", cfg, err)
	}
	opsId := uuid.NewString()
	allOps := []model.Operation{
		{
			ID:            opsId,
			ResourceGroup: "test-resource-group",
			OpsType:       "scan",
			StatusTime:    time.Now(),
			Status:        model.OperationStarted,
		},
		{
			ID:            opsId,
			ResourceGroup: "test-resource-group",
			OpsType:       "scan",
			StatusTime:    time.Now().Add(time.Second * 60),
			Status:        model.OperationCompleted,
		},
	}

	if err := sqlClient.AddOperationLog(context.Background(), allOps); err != nil {
		t.Fatalf("AddOperationLog(%+v) unexpected error: %v", allOps, err)
	}
	q := fmt.Sprintf("SELECT * FROM %s", testOperationTable)
	rows, err := db.Query(q)
	if err != nil {
		t.Errorf("query(%s) unexpected error: %v", q, err)
	}
	defer rows.Close()
	type DBOp struct {
		ID            string
		ResourceGroup string
		OpsType       string
		StartTime     *time.Time
		EndTime       *time.Time
		Status        string
		Reason        string
	}
	want := []DBOp{
		{
			ID:            allOps[1].ID,
			ResourceGroup: allOps[1].ResourceGroup,
			OpsType:       allOps[1].OpsType,
			StartTime:     &allOps[0].StatusTime,
			EndTime:       &allOps[1].StatusTime,
			Status:        allOps[1].Status.String(),
			Reason:        "",
		},
	}
	got := []DBOp{}
	for rows.Next() {
		op := DBOp{}
		if err := rows.Scan(&op.ID, &op.ResourceGroup, &op.OpsType, &op.StartTime, &op.EndTime, &op.Status, &op.Reason); err != nil {
			t.Errorf("invalid operation: %v", err)
		}
		got = append(got, op)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected diff (-want, +got): %v", diff)
	}
}

func TestPurgeIncompleteOperations(t *testing.T) {
	ctx := context.Background()
	// We can't use newTestDB as we need direct access to the database.
	db, err := sql.Open("sqlite", sqliteTestDSN)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	createDB(t, db, "schema.sql")
	cfg := Config{testResourceTable, testObservationTable, testOperationTable, 1}
	sqlClient, err := New(db, cfg)
	if err != nil {
		t.Fatalf("New(%+v) unexpected error: %v", cfg, err)
	}
	allOps := []model.Operation{
		{
			ID:            uuid.NewString(),
			ResourceGroup: "test-resource-group",
			OpsType:       "scan",
			StatusTime:    time.Now(),
			Status:        model.OperationStarted,
		},
	}

	if err := sqlClient.AddOperationLog(ctx, allOps); err != nil {
		t.Fatalf("AddOperationLog(%+v) unexpected error: %v", allOps, err)
	}

	if err := sqlClient.PurgeIncompleteOperations(ctx); err != nil {
		t.Errorf("PurgeIncompleteOperations() unexpected error: %v", err)
	}

	q := fmt.Sprintf("SELECT operationID, resourceGroupName, opsType, startTime, endTime, status FROM %s", testOperationTable)
	rows, err := db.Query(q)
	if err != nil {
		t.Errorf("query(%s) unexpected error: %v", q, err)
	}
	defer rows.Close()
	type DBOp struct {
		ID            string
		ResourceGroup string
		OpsType       string
		StartTime     *time.Time
		EndTime       *time.Time
		Status        string
	}
	want := []DBOp{
		{
			ID:            allOps[0].ID,
			ResourceGroup: allOps[0].ResourceGroup,
			OpsType:       allOps[0].OpsType,
			StartTime:     &allOps[0].StatusTime,
			EndTime:       &allOps[0].StatusTime,
			Status:        model.OperationFailed.String(),
		},
	}
	got := []DBOp{}
	for rows.Next() {
		op := DBOp{}
		if err := rows.Scan(&op.ID, &op.ResourceGroup, &op.OpsType, &op.StartTime, &op.EndTime, &op.Status); err != nil {
			t.Errorf("invalid operation: %v", err)
		}
		got = append(got, op)
	}
	if diff := cmp.Diff(want, got, cmpopts.EquateApproxTime(time.Second)); diff != "" {
		t.Errorf("unexpected diff (-want, +got): %v", diff)
	}
}

func createDB(t *testing.T, db *sql.DB, sqlScriptFile string) {
	t.Helper()
	if _, err := os.Stat(sqlScriptFile); errors.Is(err, os.ErrNotExist) {
		t.Fatalf("%s does not exist", sqlScriptFile)
	}
	content, err := os.ReadFile(sqlScriptFile)
	if err != nil {
		t.Fatalf("ReadFile(%s) unexpected error: %v", sqlScriptFile, err)
	}
	_, err = db.Exec(string(content))
	if err != nil {
		t.Fatalf("db.Exec(%s) unexpected error: %v", string(content), err)
	}
}

func TestStorageResource(t *testing.T) {
	test.TestStorageResource(t, newTestDB(t))
}

func TestStorageObservation(t *testing.T) {
	test.TestStorageObservation(t, newTestDB(t))
}
