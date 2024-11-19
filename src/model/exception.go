package model

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/nianticlabs/modron/src/proto/generated"
)

type Exception struct {
	UUID             string    `json:"uuid,omitempty"`
	SourceSystem     string    `json:"sourceSystem,omitempty"`
	UserEmail        string    `json:"userEmail,omitempty"`
	NotificationName string    `json:"notification_name,omitempty"`
	Justification    string    `json:"justification,omitempty"`
	CreatedOn        time.Time `json:"createdOn,omitempty"`
	ValidUntil       time.Time `json:"validUntil,omitempty"`
}

type Notification struct {
	UUID         string        `json:"uuid,omitempty"`
	SourceSystem string        `json:"sourceSystem,omitempty"`
	Name         string        `json:"name,omitempty"`
	Recipient    string        `json:"recipient,omitempty"`
	Content      string        `json:"content,omitempty"`
	CreatedOn    time.Time     `json:"created_on,omitempty"`
	SentOn       time.Time     `json:"sentOn,omitempty"`
	Interval     time.Duration `json:"interval,omitempty"`
}

func (e *Exception) ToProto() *pb.NotificationException {
	return &pb.NotificationException{
		Uuid:             e.UUID,
		SourceSystem:     e.SourceSystem,
		UserEmail:        e.UserEmail,
		NotificationName: e.NotificationName,
		Justification:    e.Justification,
		CreatedOnTime:    timestamppb.New(e.CreatedOn),
		ValidUntilTime:   timestamppb.New(e.ValidUntil),
	}
}

func ExceptionFromProto(p *pb.NotificationException) Exception {
	return Exception{
		UUID:             p.Uuid,
		SourceSystem:     p.SourceSystem,
		UserEmail:        p.UserEmail,
		NotificationName: p.NotificationName,
		Justification:    p.Justification,
		CreatedOn:        p.CreatedOnTime.AsTime(),
		ValidUntil:       p.ValidUntilTime.AsTime(),
	}
}

type NotificationService interface {
	BatchCreateNotifications(ctx context.Context, notifications []Notification) ([]Notification, error)
	CreateNotification(ctx context.Context, notification Notification) (Notification, error)

	GetException(ctx context.Context, uuid string) (Exception, error)
	CreateException(ctx context.Context, exception Exception) (Exception, error)
	UpdateException(ctx context.Context, exception Exception) (Exception, error)
	DeleteException(ctx context.Context, id string) error
	ListExceptions(ctx context.Context, userEmail string, pageSize int32, pageToken string) ([]Exception, error)
}
