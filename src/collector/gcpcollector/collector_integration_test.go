//go:build integration

package gcpcollector

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/risk"
	"github.com/nianticlabs/modron/src/storage/memstorage"
)

func getCollector(ctx context.Context, t *testing.T) (model.Collector, model.Storage) {
	storage := memstorage.New()
	logrus.StandardLogger().SetLevel(logrus.DebugLevel)
	logrus.StandardLogger().SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})

	orgID := os.Getenv("ORG_ID")
	if orgID == "" {
		t.Fatalf("ORG_ID is empty")
	}
	orgSuffix := os.Getenv("ORG_SUFFIX")
	if orgSuffix == "" {
		t.Fatalf("ORG_SUFFIX is empty")
	}
	tagConfig := risk.TagConfig{}
	coll, err := New(ctx, storage, orgID, orgSuffix, []string{}, tagConfig, []string{})
	if err != nil {
		t.Fatalf("failed to create collector: %v", err)
	}
	return coll, storage
}

func TestGetResourceGroups(t *testing.T) {
	ctx := context.Background()
	coll, _ := getCollector(ctx, t)
	rgs, err := coll.ListResourceGroups(ctx, nil)
	if err != nil {
		t.Fatalf("failed to list resource groups: %v", err)
	}
	if len(rgs) == 0 {
		t.Fatalf("no resource groups found")
	}
	t.Logf("rg=%+v", rgs)
}

func TestGetSpecificResourceGroups(t *testing.T) {
	ctx := context.Background()
	coll, _ := getCollector(ctx, t)
	rgs, err := coll.ListResourceGroups(ctx, []string{
		"projects/modron-dev",
		"projects/modron",
	})
	if err != nil {
		t.Fatalf("failed to list resource groups: %v", err)
	}
	if len(rgs) != 2 {
		t.Fatalf("expected 2 resource groups, got %d", len(rgs))
	}
	t.Logf("rg=%+v", rgs)
}

func TestCollect(t *testing.T) {
	ctx := context.Background()
	coll, storage := getCollector(ctx, t)
	collectID := uuid.NewString()
	err := coll.CollectAndStoreAll(ctx, collectID, []string{"projects/modron-dev"}, []*pb.Resource{})
	if err != nil {
		t.Fatalf("failed to collect: %v", err)
	}

	resources, err := storage.GetChildrenOfResource(
		ctx, collectID, "", proto.String("ResourceGroup"),
	)
	if err != nil {
		t.Fatalf("GetChildrenOfResource: %v", err)
	}
	encoded, err := json.Marshal(resources)
	if err != nil {
		t.Fatalf("failed to marshal resources: %v", err)
	}

	t.Logf("resources=%s", encoded)
}
