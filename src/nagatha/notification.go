package nagatha

import (
	"time"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/proto/generated/nagatha"
)

func NotificationFromObservation(contact string, interval time.Duration, obs *pb.Observation) model.Notification {
	return model.Notification{
		SourceSystem: "modron",
		Name:         obs.Name,
		Recipient:    contact,
		Content:      formatNotificationContent(obs),
		Interval:     interval,
	}
}

func notificationFromProto(p *nagatha.Notification) model.Notification {
	return model.Notification{
		UUID:         p.Uuid,
		SourceSystem: p.SourceSystem,
		Name:         p.Name,
		Recipient:    p.Recipient,
		Content:      p.Content,
		CreatedOn:    p.CreatedOn.AsTime(),
		SentOn:       p.SentOn.AsTime(),
		Interval:     p.Interval.AsDuration(),
	}
}

func formatNotificationContent(obs *pb.Observation) string {
	var out string
	out += obs.Remediation.Description + "\n\n"
	out += obs.Remediation.Recommendation + "  \n  \n"
	return out
}
