// Package storage provides a storage backend
package test

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// this function checks if the two ResourceEntry objects are equal
func AreEqualResources(t *testing.T, want []*pb.Resource, got []*pb.Resource) {
	t.Helper()
	sort := cmp.Transformer("sort", func(in []*pb.Resource) []*pb.Resource {
		out := append([]*pb.Resource{}, in...)
		sort.SliceStable(out, func(i, j int) bool {
			return out[i].Uid < out[j].Uid
		})
		return out
	})
	if diff := cmp.Diff(want, got, protocmp.Transform(), sort); diff != "" {
		t.Errorf("unexpected diff (-want, +got): %v", diff)
	}
}

func AreEqualObservations(t *testing.T, want []*pb.Observation, got []*pb.Observation) {
	t.Helper()
	sort := cmp.Transformer("sort", func(in []*pb.Observation) []*pb.Observation {
		out := append([]*pb.Observation{}, in...)
		sort.SliceStable(out, func(i, j int) bool {
			return out[i].Uid < out[j].Uid
		})
		return out
	})
	if diff := cmp.Diff(want, got, protocmp.Transform(), sort); diff != "" {
		t.Errorf("unexpected diff (-want, +got): %v", diff)
	}
}

// TODO: don't check length of list, but compare the actual returned arrays
func TestStorageResource(t *testing.T, storage model.Storage) {
	ctx := context.Background()
	collectionId := uuid.NewString()
	testResourceName := fmt.Sprintf("test-%s", uuid.NewString())
	testResourceName2 := fmt.Sprintf("test2-%s", uuid.NewString())
	parentResourceName := fmt.Sprintf("test-parent-%s", uuid.NewString())
	resourceGroupName1 := fmt.Sprintf("projectID1-%s", uuid.NewString())
	testResource := &pb.Resource{
		Name:              testResourceName,
		Parent:            parentResourceName,
		ResourceGroupName: resourceGroupName1,
		CollectionUid:     collectionId,
		Timestamp:         timestamppb.New(time.Now().Add(-time.Hour * 24)),
		Type: &pb.Resource_ApiKey{
			ApiKey: &pb.APIKey{
				Scopes: []string{"TEST1"},
			},
		},
	}

	resourceGroupName2 := fmt.Sprintf("projectID2-%s", uuid.NewString())
	testResource2 := &pb.Resource{
		Name:              testResourceName2,
		Parent:            parentResourceName,
		ResourceGroupName: resourceGroupName2,
		CollectionUid:     collectionId,
		Timestamp:         timestamppb.New(time.Now().Add(-time.Hour * 24)),
		Type: &pb.Resource_ApiKey{
			ApiKey: &pb.APIKey{
				Scopes: []string{"TEST2"},
			},
		},
	}

	testOps := []model.Operation{
		{
			ID:            collectionId,
			ResourceGroup: resourceGroupName1,
			OpsType:       "collection",
			StatusTime:    time.Now(),
			Status:        model.OperationStarted,
		},
		{
			ID:            collectionId,
			ResourceGroup: resourceGroupName2,
			OpsType:       "collection",
			StatusTime:    time.Now(),
			Status:        model.OperationStarted,
		},
		{
			ID:            collectionId,
			ResourceGroup: resourceGroupName1,
			OpsType:       "collection",
			StatusTime:    time.Now().Add(time.Second * 60),
			Status:        model.OperationCompleted,
		},
		{
			ID:            collectionId,
			ResourceGroup: resourceGroupName2,
			OpsType:       "collection",
			StatusTime:    time.Now().Add(time.Second * 60),
			Status:        model.OperationCompleted,
		},
	}

	addOps(ctx, t, storage, testOps)

	resourceName2Filter := model.StorageFilter{
		ResourceNames: []string{testResourceName2},
	}

	resourceNameLimitFilter := model.StorageFilter{
		ResourceNames: []string{testResourceName},
		Limit:         1,
	}

	resourceNameFilter := model.StorageFilter{
		ResourceNames: []string{testResourceName},
	}

	resourceTypeFilter := model.StorageFilter{
		ResourceTypes: []string{"DATABASE"},
	}

	resourceGroup1Filter := model.StorageFilter{
		ResourceGroupNames: []string{resourceGroupName1},
	}
	resourceGroup2AndNameFilter := model.StorageFilter{
		ResourceNames:      []string{testResourceName2},
		ResourceGroupNames: []string{resourceGroupName2},
	}

	// should not error with empty storage
	allResources, err := storage.ListResources(ctx, model.StorageFilter{
		ResourceGroupNames: []string{resourceGroupName1, resourceGroupName2},
	})
	if err != nil {
		t.Errorf("ListResources(ctx, filter) error: %v", err)
	}
	if len(allResources) != 0 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 0)
	}

	// add and get first resource and see if they are equal
	rWant, err := storage.BatchCreateResources(ctx, []*pb.Resource{testResource})
	if err != nil {
		t.Fatalf("BatchCreateResources(ctx, %v+) error: %v", testResource, err)
	}

	rGot, err := storage.ListResources(ctx, model.StorageFilter{
		ResourceGroupNames: []string{resourceGroupName1, resourceGroupName2},
	})
	if err != nil {
		t.Fatalf("ListResources(ctx, %v) error: %v", testResourceName, err)
	}
	AreEqualResources(t, rWant, rGot)

	// add second resource
	if _, err = storage.BatchCreateResources(ctx, []*pb.Resource{testResource2}); err != nil {
		t.Fatalf("AddResource(ctx, %v+) error: %v", testResource, err)
	}

	// check if both elements are there
	allResources, err = storage.ListResources(ctx, model.StorageFilter{
		ResourceGroupNames: []string{resourceGroupName1, resourceGroupName2},
	})
	if err != nil {
		t.Errorf("ListResources(ctx, filter) error: %v", err)
	}
	if len(allResources) != 2 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 2)
	}

	// only get one element
	allResources, err = storage.ListResources(ctx, model.StorageFilter{Limit: 1})
	if err != nil {
		t.Errorf("ListResources(ctx, %v) error: %v", model.StorageFilter{Limit: 1}, err)
	}
	if len(allResources) != 1 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 1)
	}

	// only get a specific resourceEntry based on name
	allResources, err = storage.ListResources(ctx, resourceName2Filter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) error: %v", resourceName2Filter, err)
	}
	if len(allResources) != 1 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 1)
	}

	// add a second resource
	testResource.Uid = uuid.NewString()
	if _, err = storage.BatchCreateResources(ctx, []*pb.Resource{testResource}); err != nil {
		t.Fatalf("BatchCreateResources(ctx, %v+) error: %v", testResource, err)
	}

	// get the first resourceName but limit to 1
	allResources, err = storage.ListResources(ctx, resourceNameLimitFilter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) error: %v", resourceNameLimitFilter, err)
	}
	if len(allResources) != 1 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 1)
	}

	// get the first resourceName but no limit
	allResources, err = storage.ListResources(ctx, resourceNameFilter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) error: %v", resourceNameFilter, err)
	}
	if len(allResources) != 2 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 2)
	}

	// filter non-existing resourceType
	allResources, err = storage.ListResources(ctx, resourceTypeFilter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) error: %v", resourceTypeFilter, err)
	}
	if len(allResources) != 0 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 0)
	}

	// filter by resource group name
	allResources, err = storage.ListResources(ctx, resourceGroup1Filter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) error: %v", resourceGroup1Filter, err)
	}
	if len(allResources) != 2 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 2)
	}

	allResources, err = storage.ListResources(ctx, resourceGroup2AndNameFilter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) error: %v", resourceGroup2AndNameFilter, err)
	}
	if len(allResources) != 1 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 1)
	}
}

func TestStorageObservation(t *testing.T, storage model.Storage) {
	ctx := context.Background()
	parentResourceName := "test-parent"
	testResourceName := "testResourceName"
	testResourceGroupName1 := "projectID1"
	firstScanUID := "firstscanUID"
	secondScanUID := "secondscanUID"
	firstScanTime := time.Now().Add(-60 * time.Second)
	secondScanTime := time.Now()

	testResource := &pb.Resource{
		Name:              testResourceName,
		Parent:            parentResourceName,
		ResourceGroupName: testResourceGroupName1,
		Type: &pb.Resource_ApiKey{
			ApiKey: &pb.APIKey{
				Scopes: []string{"TEST1"},
			},
		},
	}

	testResourceName2 := "testName2"
	testResourceGroupName2 := "projectID2"
	testResource2 := &pb.Resource{
		Name:              testResourceName2,
		Parent:            parentResourceName,
		ResourceGroupName: testResourceGroupName2,
		Type: &pb.Resource_ApiKey{
			ApiKey: &pb.APIKey{
				Scopes: []string{"TEST2"},
			},
		},
	}

	testObservation1 := &pb.Observation{
		Uid:       "observation1",
		Resource:  testResource,
		Name:      "testObservation1",
		ScanUid:   firstScanUID,
		Timestamp: timestamppb.Now(),
	}

	testObservation2 := &pb.Observation{
		Uid:       "observation2",
		Resource:  testResource2,
		Name:      "testObservation2",
		ScanUid:   firstScanUID,
		Timestamp: timestamppb.Now(),
	}

	// Filters

	resourceName2Filter := model.StorageFilter{
		ResourceGroupNames: []string{testResourceGroupName2},
	}

	testOps := []model.Operation{
		{
			ID:            firstScanUID,
			ResourceGroup: testResourceGroupName1,
			OpsType:       "scan",
			StatusTime:    firstScanTime,
			Status:        model.OperationStarted,
		},
		{
			ID:            firstScanUID,
			ResourceGroup: testResourceGroupName2,
			OpsType:       "scan",
			StatusTime:    firstScanTime,
			Status:        model.OperationStarted,
		},
		{
			ID:            firstScanUID,
			ResourceGroup: testResourceGroupName1,
			OpsType:       "scan",
			StatusTime:    firstScanTime,
			Status:        model.OperationCompleted,
		},
		{
			ID:            firstScanUID,
			ResourceGroup: testResourceGroupName2,
			OpsType:       "scan",
			StatusTime:    firstScanTime,
			Status:        model.OperationCompleted,
		},
	}

	addOps(ctx, t, storage, testOps)

	// should not error with empty storage
	allObservations, err := storage.ListObservations(ctx, model.StorageFilter{
		ResourceGroupNames: []string{testResourceGroupName1, testResourceGroupName2},
	})
	if err != nil {
		t.Errorf("ListObservations(ctx, filter) error: %v", err)
	}
	if len(allObservations) != 0 {
		t.Errorf("len(allObservation) got %d, want %d", len(allObservations), 0)
	}

	// Add fist observation.
	rWant, err := storage.BatchCreateObservations(ctx, []*pb.Observation{testObservation1})
	if err != nil {
		t.Fatalf("BatchCreateObservations(ctx, %v+) error: %v", testObservation1, err)
	}
	rGot, err := storage.ListObservations(ctx, model.StorageFilter{
		ResourceGroupNames: []string{testResourceGroupName1, testResourceGroupName2},
	})
	if err != nil {
		t.Fatalf("ListObservations(ctx, %v) error: %v", testResourceName, err)
	}
	AreEqualObservations(t, rWant, rGot)

	// add second observation
	if _, err = storage.BatchCreateObservations(ctx, []*pb.Observation{testObservation2}); err != nil {
		t.Fatalf("AddResource(ctx, %v+) error: %v", testResource, err)
	}
	// check if both elements are there
	allObservations, err = storage.ListObservations(ctx, model.StorageFilter{
		ResourceGroupNames: []string{testResourceGroupName1, testResourceGroupName2},
	})
	if err != nil {
		t.Errorf("ListObservations(ctx, filter) error: %v", err)
	}
	AreEqualObservations(t, []*pb.Observation{testObservation1, testObservation2}, allObservations)

	// only get one element
	allObservations, err = storage.ListObservations(ctx, model.StorageFilter{Limit: 1})
	if err != nil {
		t.Errorf("ListObservations(ctx, %v) error: %v", model.StorageFilter{Limit: 1}, err)
	}
	if len(allObservations) != 1 {
		t.Errorf("len(allObservation) got %d, want %d", len(allObservations), 1)
	}

	// only get a resource group based on name
	allObservations, err = storage.ListObservations(ctx, resourceName2Filter)
	if err != nil {
		t.Errorf("ListObservations(ctx, %v) error: %v", resourceName2Filter, err)
	}
	AreEqualObservations(t, []*pb.Observation{testObservation2}, allObservations)

	// Run a second scan
	secondScanOps := []model.Operation{
		{
			ID:            secondScanUID,
			ResourceGroup: testResourceGroupName1,
			OpsType:       "scan",
			StatusTime:    secondScanTime,
			Status:        model.OperationStarted,
		},
		{
			ID:            secondScanUID,
			ResourceGroup: testResourceGroupName1,
			OpsType:       "scan",
			StatusTime:    secondScanTime,
			Status:        model.OperationCompleted,
		},
	}
	addOps(ctx, t, storage, secondScanOps)

	testObservationSecondScan := &pb.Observation{
		Uid:       "observation3",
		Resource:  testResource,
		Name:      "testObservationSecondScan",
		ScanUid:   secondScanUID,
		Timestamp: timestamppb.Now(),
	}
	_, err = storage.BatchCreateObservations(ctx, []*pb.Observation{testObservationSecondScan})
	if err != nil {
		t.Fatalf("BatchCreateObservations(ctx, %v+) error: %v", testObservationSecondScan, err)
	}

	wantObs := []*pb.Observation{

		testObservationSecondScan,
		testObservation2,
	}
	gotObs, err := storage.ListObservations(ctx, model.StorageFilter{})
	if err != nil {
		t.Errorf("ListObservations(ctx, %v) error: %v", model.StorageFilter{}, err)
	}
	AreEqualObservations(t, wantObs, gotObs)
}

func addOps(ctx context.Context, t *testing.T, storage model.Storage, ops []model.Operation) {
	if err := storage.AddOperationLog(context.Background(), ops); err != nil {
		t.Fatalf("AddOperation unexpected error: %v", err)
	}
	if err := storage.FlushOpsLog(ctx); err != nil {
		t.Fatalf("Flushops: %v", err)
	}
}
