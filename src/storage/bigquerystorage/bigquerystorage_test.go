package bigquerystorage

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
	"github.com/nianticlabs/modron/src/storage/test"
)

const (
	datasetIdEnvVar    = "DATASET_ID"
	gcpProjectIdEnvVar = "GCP_PROJECT_ID"

	observationTableIdEnvVar = "OBSERVATION_TABLE_ID"
	operationTableIdEnvVar   = "OPERATION_TABLE_ID"
	resourceTableIdEnvVar    = "RESOURCE_TABLE_ID"
)

var (
	requiredEnvVars = []string{datasetIdEnvVar, gcpProjectIdEnvVar, observationTableIdEnvVar, operationTableIdEnvVar, resourceTableIdEnvVar}
	storage         model.Storage
)

func setup() error {
	for _, envVar := range requiredEnvVars {
		if env := os.Getenv(envVar); env == "" {
			return fmt.Errorf("environment variable %q is not set", envVar)
		}
	}
	ctx := context.Background()
	var err error
	storage, err = New(
		ctx,
		os.Getenv(gcpProjectIdEnvVar),
		os.Getenv(datasetIdEnvVar),
		os.Getenv(resourceTableIdEnvVar),
		os.Getenv(observationTableIdEnvVar),
		os.Getenv(operationTableIdEnvVar),
	)
	if err != nil {
		return fmt.Errorf("BigQueryStorage.New unexpected error: %v", err)
	}
	return nil
}

func TestBQStorageResource(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode: needs GCP credentials")
	}
	if err := setup(); err != nil {
		t.Fatalf("setup error: %v", err)
	}
	ctx := context.Background()
	testStartTime := time.Now()
	testResourceName := fmt.Sprintf("testName-%s", uuid.NewString())
	parentResourceName := fmt.Sprintf("test-parent-%s", uuid.NewString())
	resourceGroupName := fmt.Sprintf("projectID-%s", uuid.NewString())
	testResource := &pb.Resource{
		Uid:               uuid.NewString(),
		Name:              testResourceName,
		Parent:            parentResourceName,
		ResourceGroupName: resourceGroupName,
		Timestamp:         timestamppb.New(testStartTime),
		Type: &pb.Resource_ApiKey{
			ApiKey: &pb.APIKey{
				Scopes: []string{"TEST1"},
			},
		},
	}

	testResourceName2 := fmt.Sprintf("testName2-%s", uuid.NewString())
	testResource2 := &pb.Resource{
		Uid:               uuid.NewString(),
		Name:              testResourceName2,
		Parent:            parentResourceName,
		ResourceGroupName: resourceGroupName,
		Timestamp:         timestamppb.New(testStartTime),
		Type: &pb.Resource_ApiKey{
			ApiKey: &pb.APIKey{
				Scopes: []string{"TEST2"},
			},
		},
	}

	// should not error with empty storage
	_, err := storage.ListResources(ctx, model.StorageFilter{
		ResourceNames: &[]string{testResourceName, testResourceName2},
	})
	if err != nil {
		t.Errorf("ListResources(ctx, filter) failed with: %v", err)
	}

	// add and get first resource and see if they are equal
	rGot, err := storage.BatchCreateResources(ctx, []*pb.Resource{testResource})
	if err != nil {
		t.Fatalf("BatchCreateResources(ctx, %v+) failed with: %v", testResource, err)
	}
	want := testResource
	test.AreEqualResources(t, []*pb.Resource{want}, rGot)

	// add second resource
	if _, err = storage.BatchCreateResources(ctx, []*pb.Resource{testResource2}); err != nil {
		t.Fatalf("AddResource(ctx, %v+) failed with: %v", testResource, err)
	}

	// check if both elements are there
	allResources, err := storage.ListResources(ctx, model.StorageFilter{
		ResourceGroupNames: &[]string{resourceGroupName},
	})
	if err != nil {
		t.Errorf("ListResources(ctx, filter) failed with: %v", err)
	}
	if len(allResources) != 2 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 2)
	}

	// only get one element
	limit := 1
	limitOneFilter := model.StorageFilter{
		Limit: &limit,
	}
	allResources, err = storage.ListResources(ctx, limitOneFilter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", limitOneFilter, err)
	}
	if len(allResources) != 1 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 1)
	}

	// only get a specific resourceEntry based on name
	allResources, err = storage.ListResources(ctx, model.StorageFilter{
		ResourceNames: &[]string{testResourceName},
	})
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", model.StorageFilter{
			ResourceNames: &[]string{testResourceName},
		}, err)
	}
	if len(allResources) != 1 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 1)
	}

	// add first resource again
	if _, err = storage.BatchCreateResources(ctx, []*pb.Resource{testResource}); err != nil {
		t.Fatalf("BatchCreateResources(ctx, %v+) failed with: %v", testResource, err)
	}

	// get the first resourceName but limit to 1
	resourceNameLimitFilter := model.StorageFilter{
		ResourceNames: &[]string{testResourceName},
		Limit:         &limit,
	}
	allResources, err = storage.ListResources(ctx, resourceNameLimitFilter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", resourceNameLimitFilter, err)
	}
	if len(allResources) != 1 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 1)
	}

	// get the first resourceName but no limit
	resourceNameFilter := model.StorageFilter{
		ResourceNames: &[]string{testResourceName},
	}
	allResources, err = storage.ListResources(ctx, resourceNameFilter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", resourceNameFilter, err)
	}
	if len(allResources) != 2 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 2)
	}

	// filter non-existing resourceType
	testType := 31337
	resourceTypeFilter := model.StorageFilter{
		ResourceTypes: &[]int{testType},
	}
	allResources, err = storage.ListResources(ctx, resourceTypeFilter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", resourceTypeFilter, err)
	}
	if len(allResources) != 0 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 0)
	}

	// filter time frame
	tStart := time.Now()
	tOffset, _ := time.ParseDuration("-2s")
	resourceTimeFilter := model.StorageFilter{
		StartTime:  &tStart,
		TimeOffset: &tOffset,
	}
	allResources, err = storage.ListResources(ctx, resourceTimeFilter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) failed with: %v", resourceTimeFilter, err)
	}
	if len(allResources) != 3 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 3)
	}
}

func TestBQStorageObservation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode: needs GCP credentials")
	}
	if err := setup(); err != nil {
		t.Fatalf("setup error: %v", err)
	}
	ctx := context.Background()
	testResourceName := fmt.Sprintf("testName-%s", uuid.NewString())
	testResourceGroupName1 := fmt.Sprintf("projectID1-%s", uuid.NewString())
	parentResourceName := fmt.Sprintf("test-parent-%s", uuid.NewString())
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
	}

	testObservation2 := &pb.Observation{
		Uid:      uuid.NewString(),
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

	// should not error with empty response
	allObservations, err := storage.ListObservations(ctx, model.StorageFilter{
		ResourceGroupNames: &[]string{testResourceGroupName1, testResourceGroupName2},
	})
	if err != nil {
		t.Errorf("ListObservations(ctx, filter) failed with: %v", err)
	}
	if len(allObservations) != 0 {
		t.Errorf("len(allObservation) got %d, want %d", len(allObservations), 0)
	}

	// add and get first observation and see if they are equal
	rGot, err := storage.BatchCreateObservations(ctx, []*pb.Observation{testObservation1})
	if err != nil {
		t.Fatalf("BatchCreateObservations(ctx, %v+) failed with: %v", testObservation1, err)
	}
	want := testObservation1
	test.AreEqualObservations(t, []*pb.Observation{want}, rGot)

	// add second observation

	if _, err = storage.BatchCreateObservations(ctx, []*pb.Observation{testObservation2}); err != nil {
		t.Fatalf("AddResource(ctx, %v+) failed with: %v", testResource, err)
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
		t.Errorf("len(allResources) got %d, want %d", len(allObservations), 1)
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
