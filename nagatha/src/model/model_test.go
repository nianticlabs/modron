package model

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestConversion(t *testing.T) {
	exp := Exception{
		Uuid:             "testuuid",
		UserEmail:        "test@example.com",
		NotificationName: "test-name",
		Justification:    "This is a very long justification for testing.",
		ValidUntil:       time.Time{},
	}
	if diff := cmp.Diff(exp, ExceptionFromProto(exp.ToProto())); diff != "" {
		t.Errorf("convert unexpected diff (-want, +got): %v", diff)
	}
}
