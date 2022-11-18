package memstorage

import (
	"testing"

	"github.com/nianticlabs/modron/src/storage/test"
)

func TestResourceStorage(t *testing.T) {
	test.TestStorageResource(t, New())
}

func TestObservationStorage(t *testing.T) {
	test.TestStorageObservation(t, New())
}
