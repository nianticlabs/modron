package memstorage

import (
	"testing"

	"github.com/nianticlabs/modron/src/storage/test"
)

func TestStorageResource(t *testing.T) {
	test.TestStorageResource(t, New())
}

func TestStorageObservation(t *testing.T) {
	test.TestStorageObservation(t, New())
}
