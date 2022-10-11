// Package storag provides a storage backend
package test

import (
	"context"
	"testing"
	"time"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

// this function checks if the two ResourceEntry objects are equal
func areEqualResources(t *testing.T, have *pb.Resource, want *pb.Resource) {
	t.Helper()
	if diff := cmp.Diff(want, have, protocmp.Transform()); diff != "" {
		t.Errorf("unexpected diff (-want, +got): %v", diff)
	}
}

func areEqualObservations(t *testing.T, have *pb.Observation, want *pb.Observation) {
	t.Helper()
	if diff := cmp.Diff(want, have, protocmp.Transform()); diff != "" {
		t.Errorf("unexpected diff (-want, +got): %v", diff)
	}
}

// TODO: don't check length of list, but compare the actual returned arrays
func TestStorageResource(t *testing.T, storage model.Storage) {
	ctx := context.Background()
	testResourceName := "testname"
	testResource := &pb.Resource{
		Name:              testResourceName,
		Parent:            "ParentResource",
		ResourceGroupName: "projectID1",
		Type: &pb.Resource_ApiKey{
			ApiKey: &pb.APIKey{
				Scopes: []string{"TEST1"},
			},
		},
	}

	testResourceName2 := "testname2"
	testResource2 := &pb.Resource{
		Name:              testResourceName2,
		Parent:            "ParentResource",
		ResourceGroupName: "projectID2",
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
		ResourceGroupNames: &[]string{"projectID1"},
	}
	resourceGroup2AndNameFilter := model.StorageFilter{
		ResourceNames:      &[]string{testResourceName2},
		ResourceGroupNames: &[]string{"projectID2"},
	}

	// should not error with empty storage
	allResources, err := storage.ListResources(ctx, model.StorageFilter{})
	if err != nil {
		t.Errorf("ListResources(ctx, filter) failed with: %v", err)
	}
	if len(allResources) != 0 {
		t.Errorf("len(allresources) got %d, want %d", 0, len(allResources))
	}

	// add and get first resource and see if they are equal
	rHave, err := storage.BatchCreateResources(ctx, []*pb.Resource{testResource})
	if err != nil {
		t.Fatalf("BatchCreateResources(ctx, %v+) failed with: %v", testResource, err)
	}
	rWant, err := storage.ListResources(ctx, model.StorageFilter{})
	if err != nil {
		t.Fatalf("ListResources(ctx, %v) failed with: %v", testResourceName, err)
	}
	areEqualResources(t, rHave[0], rWant[0])

	// add second resource
	if _, err = storage.BatchCreateResources(ctx, []*pb.Resource{testResource2}); err != nil {
		t.Fatalf("AddResource(ctx, %v+) failed with: %v", testResource, err)
	}

	// check if both elements are there
	allResources, err = storage.ListResources(ctx, model.StorageFilter{})
	if err != nil {
		t.Errorf("ListResources(ctx, filter) failed with: %v", err)
	}

	if len(allResources) != 2 {
		t.Errorf("len(allresources) got %d, want %d", 2, len(allResources))
	}

	// only get one element
	allResources, err = storage.ListResources(ctx, limitOneFilter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", limitOneFilter, err)
	}
	if len(allResources) != 1 {
		t.Errorf("len(allresources) got %d, want %d", 1, len(allResources))
	}

	// only get a specific resourceEntry based on name
	allResources, err = storage.ListResources(ctx, resourceName2Filter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", resourceName2Filter, err)
	}
	if len(allResources) != 1 {
		t.Errorf("len(allresources) got %d, want %d", 1, len(allResources))
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
		t.Errorf("len(allresources) got %d, want %d", 1, len(allResources))
	}

	// get the first resourceName but no limit
	allResources, err = storage.ListResources(ctx, resourceNameFilter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", resourceNameFilter, err)
	}
	if len(allResources) != 2 {
		t.Errorf("len(allresources) got %d, want %d", 2, len(allResources))
	}

	// filter non-existing resourceType
	allResources, err = storage.ListResources(ctx, resourceTypeFilter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", resourceTypeFilter, err)
	}
	if len(allResources) != 0 {
		t.Errorf("len(allresources) got %d, want %d", 0, len(allResources))
	}

	// filter by resource group name
	allResources, err = storage.ListResources(ctx, resourceGroup1Filter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", resourceGroup1Filter, err)
	}
	if len(allResources) != 2 {
		t.Errorf("len(allresources) got %d, want %d", 2, len(allResources))
	}

	allResources, err = storage.ListResources(ctx, resourceGroup2AndNameFilter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", resourceGroup2AndNameFilter, err)
	}
	if len(allResources) != 1 {
		t.Errorf("len(allresources) got %d, want %d", 1, len(allResources))
	}
}

func TestStorageObservation(t *testing.T, storage model.Storage) {
	ctx := context.Background()
	testResourceName := "testname"
	testResourceGroupName1 := "projectID1"
	testResource := &pb.Resource{
		Name:              testResourceName,
		Parent:            "ParentResource",
		ResourceGroupName: testResourceGroupName1,
		Type: &pb.Resource_ApiKey{
			ApiKey: &pb.APIKey{
				Scopes: []string{"TEST1"},
			},
		},
	}

	testResourceName2 := "testname2"
	testResourceGroupName2 := "projectID2"
	testResource2 := &pb.Resource{
		Name:              testResourceName2,
		Parent:            "ParentResource",
		ResourceGroupName: testResourceGroupName2,
		Type: &pb.Resource_ApiKey{
			ApiKey: &pb.APIKey{
				Scopes: []string{"TEST2"},
			},
		},
	}

	testObservation1 := &pb.Observation{
		Uid:      common.GetUUID(3),
		Resource: testResource,
		Name:     "testObservation1",
	}

	testObservation2 := &pb.Observation{
		Uid:      common.GetUUID(3),
		Resource: testResource2,
		Name:     "testObservation2",
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
	allObservations, err := storage.ListObservations(ctx, model.StorageFilter{})
	if err != nil {
		t.Errorf("ListObservations(ctx, filter) failed with: %v", err)
	}
	if len(allObservations) != 0 {
		t.Errorf("len(allObservation) got %d, want %d", len(allObservations), 0)
	}

	// add and get first observation and see if they are equal
	rHave, err := storage.BatchCreateObservations(ctx, []*pb.Observation{testObservation1})
	if err != nil {
		t.Fatalf("BatchCreateObservations(ctx, %v+) failed with: %v", testObservation1, err)
	}
	rWant, err := storage.ListObservations(ctx, model.StorageFilter{})
	if err != nil {
		t.Fatalf("ListObservations(ctx, %v) failed with: %v", testResourceName, err)
	}
	areEqualObservations(t, rHave[0], rWant[0])

	// add second observation
	if _, err = storage.BatchCreateObservations(ctx, []*pb.Observation{testObservation2}); err != nil {
		t.Fatalf("AddResource(ctx, %v+) failed with: %v", testResource, err)
	}

	// check if both elements are there
	allObservations, err = storage.ListObservations(ctx, model.StorageFilter{})
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

func TestBQStorageResource(t *testing.T, storage model.Storage) {
	if testing.Short() {
		t.Skip("skipping test in short mode: needs GCP credentails")
	}

	ctx := context.Background()
	testResourceName := "testname"
	testResource := &pb.Resource{
		Uid:               common.GetUUID(3),
		Name:              testResourceName,
		Parent:            "ParentResource",
		ResourceGroupName: "projectID1",
		Type: &pb.Resource_ApiKey{
			ApiKey: &pb.APIKey{
				Scopes: []string{"TEST1"},
			},
		},
	}

	testResourceName2 := "testname2"
	testResource2 := &pb.Resource{
		Uid:               common.GetUUID(3),
		Name:              testResourceName2,
		Parent:            "ParentResource",
		ResourceGroupName: "projectID1",
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

	tstart := time.Now()
	toffset, _ := time.ParseDuration("-0.5m")
	resourceTimeFilter := model.StorageFilter{
		StartTime:  &tstart,
		TimeOffset: &toffset,
	}

	// should not error with empty storage
	allResources, err := storage.ListResources(ctx, model.StorageFilter{})
	if err != nil {
		t.Errorf("ListResources(ctx, filter) failed with: %v", err)
	}
	if len(allResources) != 0 {
		t.Errorf("len(allresources) got %d, want %d", 0, len(allResources))
	}

	// add and get first resource and see if they are equal
	rHave, err := storage.BatchCreateResources(ctx, []*pb.Resource{testResource})
	if err != nil {
		t.Fatalf("BatchCreateResources(ctx, %v+) failed with: %v", testResource, err)
	}
	rWant, err := storage.ListResources(ctx, model.StorageFilter{})
	if err != nil {
		t.Fatalf("ListResources(ctx, %v) failed with: %v", testResourceName, err)
	}
	areEqualResources(t, rHave[0], rWant[0])

	// add second resource
	if _, err = storage.BatchCreateResources(ctx, []*pb.Resource{testResource2}); err != nil {
		t.Fatalf("AddResource(ctx, %v+) failed with: %v", testResource, err)
	}

	// check if both elements are there
	allResources, err = storage.ListResources(ctx, model.StorageFilter{})
	if err != nil {
		t.Errorf("ListResources(ctx, filter) failed with: %v", err)
	}

	if len(allResources) != 2 {
		t.Errorf("len(allresources) got %d, want %d", 2, len(allResources))
	}

	// only get one element
	allResources, err = storage.ListResources(ctx, limitOneFilter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", limitOneFilter, err)
	}
	if len(allResources) != 1 {
		t.Errorf("len(allresources) got %d, want %d", 1, len(allResources))
	}

	// only get a specific resourceEntry based on name
	allResources, err = storage.ListResources(ctx, resourceName2Filter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", resourceName2Filter, err)
	}
	if len(allResources) != 1 {
		t.Errorf("len(allresources) got %d, want %d", 1, len(allResources))
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
		t.Errorf("len(allresources) got %d, want %d", 1, len(allResources))
	}

	// get the first resourceName but no limit
	allResources, err = storage.ListResources(ctx, resourceNameFilter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", resourceNameFilter, err)
	}
	if len(allResources) != 2 {
		t.Errorf("len(allresources) got %d, want %d", 2, len(allResources))
	}

	// filter non-existing resourceType
	allResources, err = storage.ListResources(ctx, resourceTypeFilter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", resourceTypeFilter, err)
	}
	if len(allResources) != 0 {
		t.Errorf("len(allresources) got %d, want %d", 0, len(allResources))
	}

	// filter timeframe
	allResources, err = storage.ListResources(ctx, resourceTimeFilter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", resourceTimeFilter, err)
	}
	if len(allResources) != 3 {
		t.Errorf("len(allresources) got %d, want %d", 3, len(allResources))
	}
}

func TestBQStorageObservation(t *testing.T, storage model.Storage) {
	if testing.Short() {
		t.Skip("skipping test in short mode: needs GCP credentails")
	}
	ctx := context.Background()
	testResourceName := "testname"
	testResourceGroupName1 := "projectID1"
	testResource := &pb.Resource{
		Name:              testResourceName,
		Parent:            "ParentResource",
		ResourceGroupName: testResourceGroupName1,
		Type: &pb.Resource_ApiKey{
			ApiKey: &pb.APIKey{
				Scopes: []string{"TEST1"},
			},
		},
	}

	testResourceName2 := "testname2"
	testResourceGroupName2 := "projectID2"
	testResource2 := &pb.Resource{

		Name:              testResourceName2,
		Parent:            "ParentResource",
		ResourceGroupName: testResourceGroupName2,
		Type: &pb.Resource_ApiKey{
			ApiKey: &pb.APIKey{
				Scopes: []string{"TEST2"},
			},
		},
	}

	testObservation1 := &pb.Observation{
		Uid:      common.GetUUID(3),
		Resource: testResource,
		Name:     "testObservation1",
	}

	testObservation2 := &pb.Observation{
		Uid:      common.GetUUID(3),
		Resource: testResource2,
		Name:     "testObservation2",
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
	allObservations, err := storage.ListObservations(ctx, model.StorageFilter{})
	if err != nil {
		t.Errorf("ListObservations(ctx, filter) failed with: %v", err)
	}
	if len(allObservations) != 0 {
		t.Errorf("len(allObservation) got %d, want %d", len(allObservations), 0)
	}

	// add and get first observation and see if they are equal
	rHave, err := storage.BatchCreateObservations(ctx, []*pb.Observation{testObservation1})
	if err != nil {
		t.Fatalf("BatchCreateObservations(ctx, %v+) failed with: %v", testObservation1, err)
	}
	rWant, err := storage.ListObservations(ctx, model.StorageFilter{})
	if err != nil {
		t.Fatalf("ListObservations(ctx, %v) failed with: %v", testResourceName, err)
	}
	areEqualObservations(t, rHave[0], rWant[0])

	// add second observation

	if _, err = storage.BatchCreateObservations(ctx, []*pb.Observation{testObservation2}); err != nil {
		t.Fatalf("AddResource(ctx, %v+) failed with: %v", testResource, err)
	}

	// check if both elements are there
	allObservations, err = storage.ListObservations(ctx, model.StorageFilter{})
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
		t.Errorf("len(allresources) got %d, want %d", 1, len(allObservations))
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
