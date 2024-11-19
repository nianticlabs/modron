// Package test is a collection of test utils for storage tests
package test

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	oneDay = 24 * time.Hour
)

func AreEqualResources(t *testing.T, want []*pb.Resource, got []*pb.Resource) {
	t.Helper()
	sortResources := cmp.Transformer("sort", func(in []*pb.Resource) []*pb.Resource {
		out := append([]*pb.Resource{}, in...)
		sort.SliceStable(out, func(i, j int) bool {
			return out[i].Uid < out[j].Uid
		})
		return out
	})
	if diff := cmp.Diff(want, got, protocmp.Transform(), sortResources); diff != "" {
		t.Errorf("unexpected diff (-want, +got): %v", diff)
	}
}

func AreEqualObservations(t *testing.T, want []*pb.Observation, got []*pb.Observation) {
	t.Helper()
	sortObservations := cmp.Transformer("sort", func(in []*pb.Observation) []*pb.Observation {
		out := append([]*pb.Observation{}, in...)
		sort.SliceStable(out, func(i, j int) bool {
			return out[i].Uid < out[j].Uid
		})
		return out
	})
	if diff := cmp.Diff(want, got, protocmp.Transform(), sortObservations); diff != "" {
		t.Errorf("unexpected diff (-want, +got): %v", diff)
	}
}

// TODO: don't check length of list, but compare the actual returned arrays
func StorageResource(t *testing.T, storage model.Storage) {
	ctx := context.Background()
	collectionID := uuid.NewString()
	testResourceName := fmt.Sprintf("test-%s", uuid.NewString())
	testResourceName2 := fmt.Sprintf("test2-%s", uuid.NewString())
	parentResourceName := fmt.Sprintf("test-parent-%s", uuid.NewString())
	resourceGroupName1 := fmt.Sprintf("projectID1-%s", uuid.NewString())
	testResource := &pb.Resource{
		Name:              testResourceName,
		Parent:            parentResourceName,
		ResourceGroupName: resourceGroupName1,
		CollectionUid:     collectionID,
		Timestamp:         timestamppb.New(time.Now().Add(-oneDay)),
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
		CollectionUid:     collectionID,
		Timestamp:         timestamppb.New(time.Now().Add(-oneDay)),
		Type: &pb.Resource_ApiKey{
			ApiKey: &pb.APIKey{
				Scopes: []string{"TEST2"},
			},
		},
	}

	now := time.Now()
	testOps := []*pb.Operation{
		{
			Id:            collectionID,
			ResourceGroup: resourceGroupName1,
			Type:          "collection",
			StatusTime:    timestamppb.New(now),
			Status:        pb.Operation_STARTED,
		},
		{
			Id:            collectionID,
			ResourceGroup: resourceGroupName2,
			Type:          "collection",
			StatusTime:    timestamppb.New(now),
			Status:        pb.Operation_STARTED,
		},
		{
			Id:            collectionID,
			ResourceGroup: resourceGroupName1,
			Type:          "collection",
			StatusTime:    timestamppb.New(now),
			Status:        pb.Operation_COMPLETED,
		},
		{
			Id:            collectionID,
			ResourceGroup: resourceGroupName2,
			Type:          "collection",
			StatusTime:    timestamppb.New(now),
			Status:        pb.Operation_COMPLETED,
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
	if len(allResources) != 2 { //nolint:mnd
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 2) //nolint:mnd
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
	if len(allResources) != 2 { //nolint:mnd
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 2) //nolint:mnd
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
	if len(allResources) != 2 { //nolint:mnd
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 2) //nolint:mnd
	}

	allResources, err = storage.ListResources(ctx, resourceGroup2AndNameFilter)
	if err != nil {
		t.Errorf("ListResources(ctx, %v) error: %v", resourceGroup2AndNameFilter, err)
	}
	if len(allResources) != 1 {
		t.Errorf("len(allResources) got %d, want %d", len(allResources), 1)
	}
}

func StorageObservation(t *testing.T, storage model.Storage) {
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
		Uid:         "observation1",
		ResourceRef: utils.GetResourceRef(testResource),
		Name:        "testObservation1",
		ScanUid:     utils.RefOrNull(firstScanUID),
		Timestamp:   timestamppb.Now(),
		Source:      pb.Observation_SOURCE_MODRON,
		Severity:    pb.Severity_SEVERITY_HIGH,
		RiskScore:   pb.Severity_SEVERITY_HIGH,
		Impact:      pb.Impact_IMPACT_MEDIUM,
	}

	testObservation2 := &pb.Observation{
		Uid:         "observation2",
		ResourceRef: utils.GetResourceRef(testResource2),
		Name:        "testObservation2",
		ScanUid:     utils.RefOrNull(firstScanUID),
		Timestamp:   timestamppb.Now(),
		Source:      pb.Observation_SOURCE_MODRON,
		Severity:    pb.Severity_SEVERITY_HIGH,
		RiskScore:   pb.Severity_SEVERITY_HIGH,
		Impact:      pb.Impact_IMPACT_MEDIUM,
	}

	// Filters

	resourceName2Filter := model.StorageFilter{
		ResourceGroupNames: []string{testResourceGroupName2},
	}

	testOps := []*pb.Operation{
		{
			Id:            firstScanUID,
			ResourceGroup: testResourceGroupName1,
			Type:          "scan",
			StatusTime:    timestamppb.New(firstScanTime),
			Status:        pb.Operation_STARTED,
		},
		{
			Id:            firstScanUID,
			ResourceGroup: testResourceGroupName2,
			Type:          "scan",
			StatusTime:    timestamppb.New(firstScanTime),
			Status:        pb.Operation_STARTED,
		},
		{
			Id:            firstScanUID,
			ResourceGroup: testResourceGroupName1,
			Type:          "scan",
			StatusTime:    timestamppb.New(firstScanTime),
			Status:        pb.Operation_COMPLETED,
		},
		{
			Id:            firstScanUID,
			ResourceGroup: testResourceGroupName2,
			Type:          "scan",
			StatusTime:    timestamppb.New(firstScanTime),
			Status:        pb.Operation_COMPLETED,
		},
	}

	addOps(ctx, t, storage, testOps)

	_, err := storage.BatchCreateResources(ctx, []*pb.Resource{testResource, testResource2})
	if err != nil {
		t.Fatalf("BatchCreateResources: %v", err)
	}

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
	secondScanOps := []*pb.Operation{
		{
			Id:            secondScanUID,
			ResourceGroup: testResourceGroupName1,
			Type:          "scan",
			StatusTime:    timestamppb.New(secondScanTime),
			Status:        pb.Operation_STARTED,
		},
		{
			Id:            secondScanUID,
			ResourceGroup: testResourceGroupName1,
			Type:          "scan",
			StatusTime:    timestamppb.New(secondScanTime),
			Status:        pb.Operation_COMPLETED,
		},
	}
	addOps(ctx, t, storage, secondScanOps)

	testObservationSecondScan := &pb.Observation{
		Uid:         "observation3",
		ResourceRef: utils.GetResourceRef(testResource),
		Name:        "testObservationSecondScan",
		ScanUid:     utils.RefOrNull(secondScanUID),
		Timestamp:   timestamppb.Now(),
		Source:      pb.Observation_SOURCE_MODRON,
		Severity:    pb.Severity_SEVERITY_LOW,
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

func StorageListObservations2(t *testing.T, storage model.Storage) {
	ctx := context.Background()
	scanUUID := uuid.NewString()
	collectionUUID := uuid.NewString()
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := startTime.Add(time.Minute)

	addOps(ctx, t, storage, []*pb.Operation{
		{
			Id:            collectionUUID,
			ResourceGroup: "projects/test-1",
			// IMPORTANT: We use "collect" here because the collection of observations from SCC
			// is part of the "collection" step - thus we need to make sure that the storage is aware of the fact
			// that observations can either come from a "collection" or a "scan" (and they should be merged).
			Type:       "collection",
			StatusTime: timestamppb.New(startTime),
			Status:     pb.Operation_STARTED,
			Reason:     "",
		},
		{
			Id:            scanUUID,
			ResourceGroup: "projects/test-1",
			// IMPORTANT: We use "scan" here because we pretend that we've run a Modron scan that generated
			// Modron observations, and these appear only after a "scan" operation.
			Type:       "scan",
			StatusTime: timestamppb.New(startTime),
			Status:     pb.Operation_STARTED,
			Reason:     "",
		},
	})
	flushOps(ctx, t, storage)

	// Create a resource
	resourceUUID := uuid.NewString()
	rsrc := &pb.Resource{
		Uid:  resourceUUID,
		Name: "custom-resource",
		Type: &pb.Resource_ApiKey{ApiKey: &pb.APIKey{Scopes: []string{"this", "is", "an", "example"}}},
	}
	_, err := storage.BatchCreateResources(ctx, []*pb.Resource{rsrc})
	if err != nil {
		t.Fatalf("BatchCreateResources: %v", err)
	}

	// Create an observation
	scanTs := startTime.Add(10 * time.Second) //nolint:mnd
	originalModronObs := &pb.Observation{
		Uid:           uuid.NewString(),
		ScanUid:       utils.RefOrNull(scanUUID),
		Timestamp:     timestamppb.New(scanTs),
		Name:          "MY_CUSTOM_OBSERVATION",
		ExpectedValue: structpb.NewStringValue("expected"),
		ObservedValue: structpb.NewStringValue("observed"),
		Remediation: &pb.Remediation{
			Description:    "Desc",
			Recommendation: "Recommendation",
		},
		ResourceRef: &pb.ResourceRef{
			Uid:           &resourceUUID,
			GroupName:     "projects/test-1",
			ExternalId:    utils.RefOrNull("//cloud.google.com/example"),
			CloudPlatform: pb.CloudPlatform_GCP,
		},

		Source:    pb.Observation_SOURCE_MODRON,
		Severity:  pb.Severity_SEVERITY_INFO,
		Impact:    pb.Impact_IMPACT_MEDIUM,
		RiskScore: pb.Severity_SEVERITY_INFO,
	}

	unknownSeverityObs := &pb.Observation{
		Uid:           uuid.NewString(),
		ScanUid:       utils.RefOrNull(scanUUID),
		Timestamp:     timestamppb.New(scanTs),
		Name:          "UNKNOWN_SEVERITY_OBSERVATIONS_WILL_NEVER_SHOW_UP",
		ExpectedValue: structpb.NewStringValue("expected"),
		ObservedValue: structpb.NewStringValue("observed"),
		Remediation: &pb.Remediation{
			Description:    "Desc",
			Recommendation: "Recommendation",
		},
		ResourceRef: &pb.ResourceRef{
			Uid:           &resourceUUID,
			GroupName:     "projects/test-1",
			ExternalId:    utils.RefOrNull("//cloud.google.com/example"),
			CloudPlatform: pb.CloudPlatform_GCP,
		},

		Source:    pb.Observation_SOURCE_MODRON,
		Severity:  pb.Severity_SEVERITY_UNKNOWN,
		Impact:    pb.Impact_IMPACT_HIGH,
		RiskScore: pb.Severity_SEVERITY_UNKNOWN,
	}

	sccObs := &pb.Observation{
		Uid:           uuid.NewString(),
		CollectionId:  utils.RefOrNull(collectionUUID),
		Timestamp:     timestamppb.New(scanTs),
		Name:          "GCP_STORAGE_BUCKET_READABLE",
		ExpectedValue: nil,
		ObservedValue: nil,
		Remediation: &pb.Remediation{
			Description:    "A world readable GCP Storage bucket was discovered which may contain potentially sensitive data.",
			Recommendation: "Investigate whether this bucket should be readable, and if not, adjust the permissions.",
		},
		ResourceRef: &pb.ResourceRef{
			Uid:           nil,
			GroupName:     "projects/test-1",
			ExternalId:    utils.RefOrNull("//storage.googleapis.com/test-dev-example-public"),
			CloudPlatform: pb.CloudPlatform_GCP,
		},
		ExternalId: utils.RefOrNull("//securitycenter.googleapis.com/projects/12345/sources/123/findings/42000000"),
		Source:     pb.Observation_SOURCE_SCC,
		Category:   pb.Observation_CATEGORY_MISCONFIGURATION,
		Severity:   pb.Severity_SEVERITY_LOW,
		Impact:     pb.Impact_IMPACT_HIGH,
		RiskScore:  pb.Severity_SEVERITY_MEDIUM,
	}

	listObs := []*pb.Observation{sccObs, originalModronObs, unknownSeverityObs}
	_, err = storage.BatchCreateObservations(ctx, listObs)
	if err != nil {
		t.Fatalf("BatchCreateObservations(ctx, %v) error: %v", listObs, err)
	}
	addOps(ctx, t, storage, []*pb.Operation{
		{
			Id:            collectionUUID,
			ResourceGroup: "projects/test-1",
			Type:          "collection",
			StatusTime:    timestamppb.New(endTime),
			Status:        pb.Operation_COMPLETED,
			Reason:        "",
		},
		{
			Id:            scanUUID,
			ResourceGroup: "projects/test-1",
			Type:          "scan",
			StatusTime:    timestamppb.New(endTime),
			Status:        pb.Operation_COMPLETED,
			Reason:        "",
		},
	})
	flushOps(ctx, t, storage)

	// Sorted by severity
	want := []*pb.Observation{sccObs, originalModronObs}
	got, err := storage.ListObservations(ctx, model.StorageFilter{})
	if err != nil {
		t.Errorf("ListObservations(ctx, %v) error: %v", model.StorageFilter{}, err)
	}

	if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
		t.Errorf("unexpected diff (-want, +got): %v", diff)
	}
}

func flushOps(ctx context.Context, t *testing.T, storage model.Storage) {
	if err := storage.FlushOpsLog(ctx); err != nil {
		t.Fatalf("Flushops: %v", err)
	}
}

func addOps(ctx context.Context, t *testing.T, storage model.Storage, ops []*pb.Operation) {
	if err := storage.AddOperationLog(ctx, ops); err != nil {
		t.Fatalf("AddOperation unexpected error: %v", err)
	}
	if err := storage.FlushOpsLog(ctx); err != nil {
		t.Fatalf("Flushops: %v", err)
	}
}
