package gcpacl

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/nianticlabs/modron/src/collector/testcollector"
	"github.com/nianticlabs/modron/src/model"

	"github.com/google/go-cmp/cmp"
)

func clearCache(t *testing.T) {
	t.Helper()
	if err := os.Remove(localACLCacheFile); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("cannot delete cache: %v", err)
		}
	}
}

func TestCache(t *testing.T) {
	clearCache(t)
	defer clearCache(t)

	checker := GcpChecker{
		cfg: Config{PersistentCache: true, PersistentCacheTimeout: time.Second * 10},
	}
	collector := testcollector.TestCollector{}
	checker.collector = &collector
	fsACLCache, err := checker.getLocalACLCache()
	if err != nil {
		t.Fatalf("getLocalACLCache: %v", err)
	}
	if fsACLCache != nil {
		t.Fatalf("getLocalACLCache should be empty")
	}
	ctx := context.Background()
	aclStoreTime := time.Now()
	if err := checker.loadACLCache(ctx); err != nil {
		t.Fatalf("loadACLCache: %v", err)
	}
	if checker.aclCache == nil {
		t.Fatalf("checker.aclCache should be initialized")
	}

	expectedCache := model.ACLCache{
		"*": {
			"projects/modron-test":  {},
			"projects/super-secret": {},
		},
		"user@example.com": {
			"projects/modron-test": {},
		},
	}
	if !cmp.Equal(expectedCache, checker.aclCache) {
		t.Errorf("aclCache mismatch: %v", cmp.Diff(expectedCache, checker.aclCache))
	}

	fsACLCache, err = checker.getLocalACLCache()
	if err != nil {
		t.Fatalf("getLocalACLCache: %v", err)
	}
	if fsACLCache == nil {
		t.Fatalf("getLocalACLCache should be initialized, but it's nil")
	}

	if diff := cmp.Diff(checker.aclCache, fsACLCache.Content); diff != "" {
		t.Errorf("aclCache mismatch (-want +got):\n%s", diff)
	}
	if !aclStoreTime.Before(fsACLCache.LastUpdate) {
		t.Errorf("the filesystem cache should have been updated, its last update was %v", fsACLCache.LastUpdate)
	}
	if aclStoreTime.Add(time.Second * 10).Before(fsACLCache.LastUpdate) {
		t.Fatalf("the filesystem cache should not be older than 10 seconds from the moment we started (lastUpdate: %v)", fsACLCache.LastUpdate)
	}
}
