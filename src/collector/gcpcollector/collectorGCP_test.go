package gcpcollector

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/storage/memstorage"
)

var (
	collectorTestProjectId string
	projectListFile        string
)

func init() {
	flag.StringVar(&collectorTestProjectId, "projectId", "modron-test-project", "GCP project Id")
	flag.StringVar(&projectListFile, "projectIdList", "resourceGroupList.txt", "GCP project Id list")
}

const (
	testProjectID = "modron-test"
	collectId     = "collectId-1"
)

func TestResourceGroupResources(t *testing.T) {
	ctx := context.Background()
	storage := memstorage.New()
	gcpCollector := NewFake(ctx, storage)

	resourceGroup, err := gcpCollector.GetResourceGroup(ctx, collectId, testProjectID)
	if err != nil {
		t.Fatalf("No resourceGroup found: %v", err)
	}

	resourcesCollected, errors := gcpCollector.ListResourceGroupResources(ctx, collectId, resourceGroup)
	for _, err := range errors {
		t.Errorf("%v", err)
	}

	for _, r := range resourcesCollected {
		if r.CollectionUid != collectId {
			t.Errorf("wrong collectUid, want %v got %v", collectId, r.CollectionUid)
		}
	}

	wantResourcesCollected := 21
	if len(resourcesCollected) != wantResourcesCollected {
		t.Errorf("resources collected: got %d, want %d", len(resourcesCollected), wantResourcesCollected)
	}
}

func TestResourceGroup(t *testing.T) {
	ctx := context.Background()
	storage := memstorage.New()
	gcpCollector := NewFake(ctx, storage)
	resourceGroup, err := gcpCollector.GetResourceGroup(ctx, collectId, testProjectID)
	if err != nil {
		t.Fatalf("No resourceGroup found: %v", err)
	}

	if resourceGroup.Name != formatResourceName(testProjectID, testProjectID) {
		t.Errorf("wrong resourceGroup Name: %v", resourceGroup.Name)
	}
	if resourceGroup.Parent != "" {
		t.Errorf("wrong resourceGroup Parent: %v", resourceGroup.Name)
	}
	if len(resourceGroup.IamPolicy.Permissions) != 5 {
		t.Errorf("iam policy count: got %d, want %d", len(resourceGroup.IamPolicy.Permissions), 5)
	}
	if resourceGroup.CollectionUid != collectId {
		t.Errorf("wrong collectUid, want %v got %v", collectId, resourceGroup.CollectionUid)
	}
}

func TestCollectAndStore(t *testing.T) {
	ctx := context.Background()
	storage := memstorage.New()
	gcpCollector := NewFake(ctx, storage)

	limit := 100
	limitFilter := model.StorageFilter{
		Limit: &limit,
	}

	testProjectID := "testResourceGroupName1"
	errors := gcpCollector.CollectAndStoreResources(ctx, collectId, testProjectID)

	if len(errors) != 0 {
		for _, e := range errors {
			t.Errorf("error storing resources: %v", e)
		}
	}

	res, err := storage.ListResources(ctx, limitFilter)

	for _, r := range res {
		if r.CollectionUid != collectId {
			t.Errorf("wrong collectUid, want %v got %v", collectId, r.CollectionUid)
		}
	}

	if err != nil {
		t.Errorf("error storing resources: %v", err)
	}

	wantResourcesStored := 22
	if len(res) != wantResourcesStored {
		t.Errorf("stored resources: got %d, want %d", len(res), wantResourcesStored)
	}

}

func TestOnRealGCPProject(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode: long test and needed GCP credentials")
	}

	limit := 100000
	limitFilter := model.StorageFilter{
		Limit: &limit,
	}

	// Initialize Collector
	ctx := context.Background()
	storage := memstorage.New()
	gcpCollector, err := New(ctx, storage)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	errors := gcpCollector.CollectAndStoreResources(ctx, collectId, collectorTestProjectId)
	if len(errors) > 0 {
		t.Errorf("CollectAndStoreResources(): %v", errors)
	}
	resourcesCollected, err := storage.ListResources(ctx, limitFilter)
	if err != nil {
		t.Errorf("storage.ListResources: %v", err)
	}
	fmt.Printf("Collected %v resources\n", len(resourcesCollected))
	fmt.Printf("Resources: %v\n", resourcesCollected)
}

//go test -v -run ^TestOnRealGCPProjectAll$ github.com/nianticlabs/modron/src/collector/gcpcollector -timeout 99999s
func TestOnRealGCPProjectAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode: long test and needed GCP credentials")
	}

	limit := 100000
	limitFilter := model.StorageFilter{
		Limit: &limit,
	}

	// Initialize Collector
	ctx := context.Background()
	storage := memstorage.New()
	gcpCollector, err := New(ctx, storage)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	content, err := os.ReadFile(projectListFile)
	if err != nil {
		t.Errorf("error with projectID list file: %v", err)
	}
	projectIDs := strings.Split(string(content), "\n")

	errors := gcpCollector.CollectAndStoreAllResourceGroupResources(ctx, collectId, projectIDs)
	if len(errors) > 0 {
		t.Errorf("CollectAndStoreResources(): %v", errors)
	}

	resources, err := storage.ListResources(ctx, limitFilter)
	if err != nil {
		t.Errorf("storage.ListResources: %v", err)
	}

	fmt.Printf("Collected %v resources\n", len(resources))
}
