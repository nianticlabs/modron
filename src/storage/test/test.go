// Package storage provides a storage backend
package test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"google.golang.org/protobuf/testing/protocmp"
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

func AreEqualObservations(t *testing.T, got []*pb.Observation, want []*pb.Observation) {
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
	testResourceName := fmt.Sprintf("test-%s", uuid.NewString())
	testResourceName2 := fmt.Sprintf("test2-%s", uuid.NewString())
	parentResourceName := fmt.Sprintf("test-parent-%s", uuid.NewString())
	resourceGroupName1 := fmt.Sprintf("projectID1-%s", uuid.NewString())
	testResource := &pb.Resource{
		Name:              testResourceName,
		Parent:            parentResourceName,
		ResourceGroupName: resourceGroupName1,
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
		Type: &pb.Resource_ApiKey{
			ApiKey: &pb.APIKey{
				Scopes: []string{"TEST2"},
			},
		},
	}

	// Filters
	limit := 1
	limitOneFilter := model.StorageFilter{
		Limit: &limit,
	}

	resourceName2Filter := model.StorageFilter{
		ResourceNames: &[]string{testResourceName2},
	}

	resourceNameLimitFilter := model.StorageFilter{
		ResourceNames: &[]string{testResourceName},
		Limit:         &limit,
	}

	resourceNameFilter := model.StorageFilter{
		ResourceNames: &[]string{testResourceName},
	}

	testType := 31337
	resourceTypeFilter := model.StorageFilter{
		ResourceTypes: &[]int{testType},
	}

	resourceGroup1Filter := model.StorageFilter{
		ResourceGroupNames: &[]string{resourceGroupName1},
	}
	resourceGroup2AndNameFilter := model.StorageFilter{
		ResourceNames:      &[]string{testResourceName2},
		ResourceGroupNames: &[]string{resourceGroupName2},
	}

	// should not error with empty storage
	allResources, err := storage.ListResources(ctx, model.StorageFilter{
		ResourceGroupNames: &[]string{resourceGroupName1, resourceGroupName2},
	})
	if err != nil {
		t.Errorf("ListResources(ctx, filter) failed with: %v", err)
	}
	if len(allResources) != 0 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 0)
	}

	// add and get first resource and see if they are equal
	rGot, err := storage.BatchCreateResources(ctx, []*pb.Resource{testResource})
	if err != nil {
		t.Fatalf("BatchCreateResources(ctx, %v+) failed with: %v", testResource, err)
	}

	rWant, err := storage.ListResources(ctx, model.StorageFilter{
		ResourceGroupNames: &[]string{resourceGroupName1, resourceGroupName2},
	})
	if err != nil {
		t.Fatalf("ListResources(ctx, %v) failed with: %v", testResourceName, err)
	}
	AreEqualResources(t, rWant, rGot)

	// add second resource
	if _, err = storage.BatchCreateResources(ctx, []*pb.Resource{testResource2}); err != nil {
		t.Fatalf("AddResource(ctx, %v+) failed with: %v", testResource, err)
	}

	// check if both elements are there
	allResources, err = storage.ListResources(ctx, model.StorageFilter{
		ResourceGroupNames: &[]string{resourceGroupName1, resourceGroupName2},
	})
	if err != nil {
		t.Errorf("ListResources(ctx, filter) failed with: %v", err)
	}
	if len(allResources) != 2 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 2)
	}

	// only get one element
	allResources, err = storage.ListResources(ctx, limitOneFilter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", limitOneFilter, err)
	}
	if len(allResources) != 1 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 1)
	}

	// only get a specific resourceEntry based on name
	allResources, err = storage.ListResources(ctx, resourceName2Filter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", resourceName2Filter, err)
	}
	if len(allResources) != 1 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 1)
	}

	// add first resource again
	if _, err = storage.BatchCreateResources(ctx, []*pb.Resource{testResource}); err != nil {
		t.Fatalf("BatchCreateResources(ctx, %v+) failed with: %v", testResource, err)
	}

	// get the first resourceName but limit to 1
	allResources, err = storage.ListResources(ctx, resourceNameLimitFilter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", resourceNameLimitFilter, err)
	}
	if len(allResources) != 1 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 1)
	}

	// get the first resourceName but no limit
	allResources, err = storage.ListResources(ctx, resourceNameFilter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", resourceNameFilter, err)
	}
	if len(allResources) != 2 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 2)
	}

	// filter non-existing resourceType
	allResources, err = storage.ListResources(ctx, resourceTypeFilter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", resourceTypeFilter, err)
	}
	if len(allResources) != 0 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 0)
	}

	// filter by resource group name
	allResources, err = storage.ListResources(ctx, resourceGroup1Filter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", resourceGroup1Filter, err)
	}
	if len(allResources) != 2 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 2)
	}

	allResources, err = storage.ListResources(ctx, resourceGroup2AndNameFilter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", resourceGroup2AndNameFilter, err)
	}
	if len(allResources) != 1 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 1)
	}
}

func TestStorageObservation(t *testing.T, storage model.Storage) {
	ctx := context.Background()
	parentResourceName := fmt.Sprintf("test-parent-%s", uuid.NewString())
	testResourceName := fmt.Sprintf("testName-%s", uuid.NewString())
	testResourceGroupName1 := fmt.Sprintf("projectID1-%s", uuid.NewString())
	scanUID := uuid.NewString()
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

	testResourceName2 := fmt.Sprintf("testName2-%s", uuid.NewString())
	testResourceGroupName2 := fmt.Sprintf("projectID2-%s", uuid.NewString())
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
		Uid:      uuid.NewString(),
		Resource: testResource,
		Name:     "testObservation1",
		ScanUid:  scanUID,
	}

	testObservation2 := &pb.Observation{
		Uid:      uuid.NewString(),
		Resource: testResource2,
		Name:     "testObservation2",
		ScanUid:  scanUID,
	}

	// Filters
	limit := 1
	limitOneFilter := model.StorageFilter{
		Limit: &limit,
	}

	resourceName2Filter := model.StorageFilter{
		ResourceGroupNames: &[]string{testResourceGroupName2},
	}

	// should not error with empty storage
	allObservations, err := storage.ListObservations(ctx, model.StorageFilter{
		ResourceGroupNames: &[]string{testResourceGroupName1, testResourceGroupName2},
	})
	if err != nil {
		t.Errorf("ListObservations(ctx, filter) failed with: %v", err)
	}
	if len(allObservations) != 0 {
		t.Errorf("len(allObservation) got %d, want %d", len(allObservations), 0)
	}

	// Add fist observation.
	rGot, err := storage.BatchCreateObservations(ctx, []*pb.Observation{testObservation1})
	if err != nil {
		t.Fatalf("BatchCreateObservations(ctx, %v+) failed with: %v", testObservation1, err)
	}
	if err := storage.AddOperationLog(ctx, []model.Operation{{
		ID:            testObservation1.ScanUid,
		ResourceGroup: testObservation1.Resource.ResourceGroupName,
		OpsType:       "scan",
		Status:        model.OperationCompleted,
	}}); err != nil {
		t.Fatalf("AddOperationLog() unexpected error: %v", err)
	}
	rWant, err := storage.ListObservations(ctx, model.StorageFilter{
		ResourceGroupNames: &[]string{testResourceGroupName1, testResourceGroupName2},
	})
	if err != nil {
		t.Fatalf("ListObservations(ctx, %v) failed with: %v", testResourceName, err)
	}
	AreEqualObservations(t, rWant, rGot)

	// add second observation
	if _, err = storage.BatchCreateObservations(ctx, []*pb.Observation{testObservation2}); err != nil {
		t.Fatalf("AddResource(ctx, %v+) failed with: %v", testResource, err)
	}
	if err := storage.AddOperationLog(ctx, []model.Operation{{
		ID:            testObservation2.ScanUid,
		ResourceGroup: testObservation2.Resource.ResourceGroupName,
		OpsType:       "scan",
		Status:        model.OperationCompleted,
	}}); err != nil {
		t.Fatalf("AddOperationLog() unexpected error: %v", err)
	}
	// check if both elements are there
	allObservations, err = storage.ListObservations(ctx, model.StorageFilter{
		ResourceGroupNames: &[]string{testResourceGroupName1, testResourceGroupName2},
	})
	if err != nil {
		t.Errorf("ListObservations(ctx, filter) failed with: %v", err)
	}

	if len(allObservations) != 2 {
		t.Errorf("len(allObservation) got %d, want %d", len(allObservations), 2)
	}

	// only get one element
	allObservations, err = storage.ListObservations(ctx, limitOneFilter)
	if err != nil {
		t.Errorf("ListObservations(ctx, %v) failed with: %v", limitOneFilter, err)
	}
	if len(allObservations) != 1 {
		t.Errorf("len(allObservation) got %d, want %d", len(allObservations), 1)
	}

	// only get a resource group based on name
	allObservations, err = storage.ListObservations(ctx, resourceName2Filter)
	if err != nil {
		t.Errorf("ListObservations(ctx, %v) failed with: %v", resourceName2Filter, err)
	}
	if len(allObservations) != 1 {
		t.Errorf("len(allObservation) got %d, want %d", len(allObservations), 1)
	}

	// add second observation again
	if _, err = storage.BatchCreateObservations(ctx, []*pb.Observation{testObservation2}); err != nil {
		t.Fatalf("AddResource(ctx, %v+) failed with: %v", testResource, err)
	}

	// only get entries based on resource group name (two this time)
	allObservations, err = storage.ListObservations(ctx, resourceName2Filter)
	if err != nil {
		t.Errorf("ListObservations(ctx, %v) failed with: %v", resourceName2Filter, err)
	}
	if len(allObservations) != 2 {
		t.Errorf("len(allObservation) got %d, want %d", len(allObservations), 2)
	}
}
