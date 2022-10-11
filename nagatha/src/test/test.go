package test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/nianticlabs/modron/nagatha/src/model"
)

func TestStorage(t *testing.T, storage model.Storage) {
	t.Helper()
	ctx := context.Background()

	want := model.Exception{
		ValidUntil: time.Time{},
	}
	resp, err := storage.CreateException(ctx, want)
	if err != nil {
		t.Fatalf("CreateException(ctx, %+v) unexpected error: %v", want, err)
	}

	want.Uuid = resp.Uuid
	got, err := storage.GetException(ctx, resp.Uuid)
	if err != nil {
		t.Fatalf("GetException(ctx, %q) error: %v", resp.Uuid, err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("GetException(ctx, %q) unexpected diff (-want, +got): %v", resp.Uuid, diff)
	}

	updateReq := model.Exception{
		Uuid:             resp.Uuid,
		NotificationName: "test-notification",
		ValidUntil:       time.Time{},
	}
	if _, err := storage.EditException(ctx, updateReq); err != nil {
		t.Fatalf("EditException(ctx, %+v) error: %v", updateReq, err)
	}

	want.NotificationName = "test-notification"
	if gotAfterUpdate, err := storage.GetException(ctx, resp.Uuid); err != nil {
		t.Fatalf("GetException(ctx, %q) error: %v", resp.Uuid, err)
	} else if diff := cmp.Diff(want, gotAfterUpdate); diff != "" {
		t.Errorf("GetException(ctx, %q) unexpected diff (-want, +got): %v", resp.Uuid, diff)
	}

	if err := storage.DeleteException(ctx, resp.Uuid); err != nil {
		t.Errorf("DeleteException(ctx, %q) error: %v", resp.Uuid, err)
	}

	if _, err := storage.GetException(ctx, resp.Uuid); err == nil {
		t.Fatalf("GetException(ctx, %q) wanted error, got none", resp.Uuid)
	}
}
