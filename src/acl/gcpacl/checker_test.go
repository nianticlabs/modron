package gcpacl

import (
	"context"
	"os"
	"testing"

	"github.com/nianticlabs/modron/src/collector/gcpcollector"
	"github.com/nianticlabs/modron/src/risk"
	"github.com/nianticlabs/modron/src/storage/memstorage"

	"google.golang.org/grpc/metadata"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestInvalidNoToken(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode: need GCP credentials")
	}

	ctx := context.Background()
	storage := memstorage.New()
	gcpCollector := gcpcollector.NewFake(ctx, storage, risk.TagConfig{})

	checker, err := New(ctx, gcpCollector, Config{})
	if err != nil {
		t.Error(err)
	}

	if _, err = checker.GetValidatedUser(ctx); err == nil {
		t.Error("expected error: the context does not have a tokenid but the checker authenticated a user")
	}
}

func TestInvalidParseToken(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(
		map[string]string{"Authorization": "Bearer xyz.abc.123"}))
	storage := memstorage.New()
	gcpCollector := gcpcollector.NewFake(ctx, storage, risk.TagConfig{})

	checker, err := New(ctx, gcpCollector, Config{})
	if err != nil {
		t.Error(err)
	}

	if _, err = checker.GetValidatedUser(ctx); err == nil {
		t.Error("expected error: checker parsed a jwt tokenId that is invalid")
	}
}
