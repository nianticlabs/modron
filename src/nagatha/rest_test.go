package nagatha_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/h2non/gock"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/nagatha"
)

const testEndpoint = "https://nagatha.localhost"

func mockNagathaCreateNotification() func() {
	gock.New(testEndpoint).
		Post("/v2/notifications").
		MatchHeader("Authorization", "Bearer hunter2").
		Reply(200).
		JSON(map[string]any{"uuid": "notification-uuid"})
	return gock.Off
}

func mockNagathaCreateBatchNotifications() func() {
	gock.New(testEndpoint).
		Post("/v2/notifications:batchCreate").
		MatchHeader("Authorization", "Bearer hunter2").
		Reply(200).
		JSON(map[string]any{
			"name": "operations/my-operation-1",
			"done": false,
		})
	gock.New(testEndpoint).
		Get("/v2/operations/my-operation-1").
		MatchHeader("Authorization", "Bearer hunter2").
		Reply(200).
		JSON(map[string]any{
			"name": "my-operation-1",
			"done": false,
		})
	gock.New(testEndpoint).
		Get("/v2/operations/my-operation-1").
		MatchHeader("Authorization", "Bearer hunter2").
		Reply(200).
		JSON(map[string]any{
			"name": "my-operation-1",
			"done": false,
		})
	gock.New(testEndpoint).
		Get("/v2/operations/my-operation-1").
		MatchHeader("Authorization", "Bearer hunter2").
		Reply(200).
		JSON(map[string]any{
			"name": "operations/my-operation-1",
			"done": true,
			"response": map[string]any{
				"@type": "com.nianticlabs.nagatha.BatchCreateNotificationsResponse",
				"notifications": []map[string]any{
					{
						"uuid": "notification-uuid-1",
					},
					{
						"uuid": "notification-uuid-2",
					},
				},
			},
		})
	return gock.Off
}

func getClient(t *testing.T) model.NotificationService {
	t.Helper()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "hunter2"})
	c, err := nagatha.New(testEndpoint, "modron.localhost", ts)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}
func TestClient_CreateNagathaNotification_WithToken(t *testing.T) {
	ctx := context.Background()
	off := mockNagathaCreateNotification()
	defer off()

	c := getClient(t)
	got, err := c.CreateNotification(ctx, model.Notification{
		SourceSystem: "modron",
		Name:         "this is a test",
		Recipient:    "user@example.com",
		Content:      "notification content",
		CreatedOn:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Interval:     24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("CreateNotification: %v", err)
	}
	want := model.Notification{
		UUID:      "notification-uuid",
		CreatedOn: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		SentOn:    time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("notification mismatch (-want, +got): %s", diff)
	}
}

func TestClient_BatchCreateNotifications(t *testing.T) {
	ctx := context.Background()
	off := mockNagathaCreateBatchNotifications()
	defer off()
	logrus.StandardLogger().SetLevel(logrus.DebugLevel)

	c := getClient(t)
	notifications := []model.Notification{
		{
			SourceSystem: "modron",
			Name:         "this is a test",
			Recipient:    "user@example.com",
			Content:      "notification content",
			CreatedOn:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Interval:     24 * time.Hour,
		},
		{
			SourceSystem: "modron",
			Name:         "this is another test",
			Recipient:    "user@example.com",
			Content:      "notification content",
			CreatedOn:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Interval:     24 * time.Hour,
		},
	}
	got, err := c.BatchCreateNotifications(ctx, notifications)
	if err != nil {
		t.Fatalf("BatchCreateNotifications: %v", err)
	}

	want := []model.Notification{
		{
			UUID:      "notification-uuid-1",
			SentOn:    time.UnixMilli(0),
			CreatedOn: time.UnixMilli(0),
		},
		{
			UUID:      "notification-uuid-2",
			SentOn:    time.UnixMilli(0),
			CreatedOn: time.UnixMilli(0),
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("notifications mismatch (-want, +got):\n%s", diff)
	}
}

func TestClient_CreateNagathaNotification_WithoutToken(t *testing.T) {
	if _, err := nagatha.New(testEndpoint, "modron.localhost", nil); err == nil {
		t.Fatalf("nagatha.New: got nil, want error")
	}
}
