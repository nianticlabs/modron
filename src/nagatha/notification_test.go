package nagatha_test

import (
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/nagatha"
	pb "github.com/nianticlabs/modron/src/proto/generated"
)

func TestNotificationFromObservation(t *testing.T) {
	pbObs := pb.Observation{
		Uid:       "47fc94f3-b6e9-4ae5-b719-c8dd2157744b",
		ScanUid:   proto.String("2f86b47d-a386-4f4e-88a1-786f2a572ad8"),
		Timestamp: timestamppb.New(time.Now()),
		ResourceRef: &pb.ResourceRef{
			Uid:       proto.String("79bc4bd2-a454-4837-8a09-0d769cb36d0f"),
			GroupName: "projects/some-project",
		},
		Name: "DATABASE_ALLOWS_UNENCRYPTED_CONNECTIONS",
		Remediation: &pb.Remediation{
			Description:    "Database example-psql allows for unencrypted connections.",
			Recommendation: "Enable the require SSL setting in the database settings to allow only encrypted connections to example-psql.",
		},
	}

	want := model.Notification{
		SourceSystem: "modron",
		Name:         "DATABASE_ALLOWS_UNENCRYPTED_CONNECTIONS",
		Content:      "Database example-psql allows for unencrypted connections.\n\nEnable the require SSL setting in the database settings to allow only encrypted connections to example-psql.  \n  \n",
		Recipient:    "test@example.com",
		Interval:     24 * time.Hour,
	}
	got := nagatha.NotificationFromObservation("test@example.com", 24*time.Hour, &pbObs)
	if diff := cmp.Diff(&got, &want); diff != "" {
		t.Errorf("NotificationFromObservation() mismatch (-got +want):\n%s", diff)
	}
}
