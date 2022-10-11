package memstorage

import (
	"testing"

	"github.com/nianticlabs/modron/nagatha/src/test"
)

func TestStorage(t *testing.T) {
	test.TestStorage(t, New())
}
