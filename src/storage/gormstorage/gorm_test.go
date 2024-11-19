package gormstorage

import (
	"testing"

	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/storage/test"
	storageutils "github.com/nianticlabs/modron/src/storage/utils"
)

func newTestDb(t *testing.T) model.Storage {
	st, err := NewSQLite(Config{
		BatchSize:     100,
		LogAllQueries: true,
	}, storageutils.GetSqliteMemoryDbPath())
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	return st
}

func TestStorageResource(t *testing.T) {
	test.StorageResource(t, newTestDb(t))
}

func TestStorageObservation(t *testing.T) {
	test.StorageObservation(t, newTestDb(t))
}

func TestStorageListObservationsActive(t *testing.T) {
	test.StorageListObservations2(t, newTestDb(t))
}
