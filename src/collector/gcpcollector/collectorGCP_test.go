package gcpcollector

import (
	"context"
	"flag"
	"testing"

	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/storage/memstorage"
)

var (
	collectorTestProjectId string
	projectListFile        string
)

func init() {
	flag.StringVar(&collectorTestProjectId, "projectId", testProjectID, "GCP project Id")
	flag.StringVar(&projectListFile, "projectIdList", "resourceGroupList.txt", "GCP project Id list")
}

const (
	testProjectID = "projects/modron-test"
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

	wantResourcesCollected := 26 // TODO: Create a better test for this functionality
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
	if len(resourceGroup.IamPolicy.Permissions) != 6 {
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

	limitFilter := model.StorageFilter{
		Limit: 100,
	}

	for _, testResourceID := range []string{"organizations/1111", testProjectID} {
		errors := gcpCollector.CollectAndStoreResources(ctx, collectId, testResourceID)

		if len(errors) != 0 {
			for _, e := range errors {
				t.Errorf("CollectAndStoreResources(ctx, %s, %s): %v", collectId, testResourceID, e)
			}
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

	wantResourcesStored := 30 // TODO: Create a better test for this functionality
	if len(res) != wantResourcesStored {
		t.Errorf("stored resources: got %d, want %d", len(res), wantResourcesStored)
	}
}
