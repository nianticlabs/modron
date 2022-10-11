package memstorage

import (
	"testing"

	"github.com/nianticlabs/modron/src/storage/test"
)

func TestMemStorage(t *testing.T) {
	test.TestStorageResource(t, New())
	test.TestStorageObservation(t, New())
	//test.TestStorageResourceFailure(t, New())
}
