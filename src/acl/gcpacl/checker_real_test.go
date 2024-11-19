//go:build integration

package gcpacl

import (
	"context"
	"os"
	"testing"

	"github.com/nianticlabs/modron/src/collector/gcpcollector"
	"github.com/nianticlabs/modron/src/risk"
	"github.com/nianticlabs/modron/src/storage/memstorage"
)

func TestCheckerReal(t *testing.T) {
	orgID := os.Getenv("ORG_ID")
	orgSuffix := os.Getenv("ORG_SUFFIX")
	if orgID == "" || orgSuffix == "" {
		t.Fatalf("ORG_ID and ORG_SUFFIX are required, orgID=%q, orgSuffix=%q", orgID, orgSuffix)
	}

	ctx := context.Background()
	storage := memstorage.New()
	gcpCollector, err := gcpcollector.New(ctx, storage, orgID, orgSuffix, []string{}, risk.TagConfig{}, []string{})
	if err != nil {
		t.Error(err)
	}

	checker, err := New(ctx, gcpCollector, Config{})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := checker.ListResourceGroupNamesOwned(ctx); err != nil {
		t.Error(err)
	}
}
