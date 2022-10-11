package model

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/nianticlabs/modron/src/pb"
)

type Exception struct {
	Uuid             string
	SourceSystem     string
	UserEmail        string
	NotificationName string
	Justification    string
	CreatedOn        time.Time
	ValidUntil       time.Time
}

type Notification struct {
	Uuid         string
	SourceSystem string
	Name         string
	Recipient    string
	Content      string
	CreatedOn    time.Time
	SentOn       time.Time
	Interval     time.Duration
}

func (e *Exception) ToProto() *pb.NotificationException {
	return &pb.NotificationException{
		Uuid:             e.Uuid,
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
		Uuid:             p.Uuid,
		SourceSystem:     p.SourceSystem,
		UserEmail:        p.UserEmail,
		NotificationName: p.NotificationName,
		Justification:    p.Justification,
		CreatedOn:        p.CreatedOnTime.AsTime(),
		ValidUntil:       p.ValidUntilTime.AsTime(),
	}
}

type NotificationService interface {
	CreateNotification(ctx context.Context, notification Notification) (Notification, error)

	GetException(ctx context.Context, uuid string) (Exception, error)
	CreateException(ctx context.Context, exception Exception) (Exception, error)
	UpdateException(ctx context.Context, exception Exception) (Exception, error)
	DeleteException(ctx context.Context, id string) error
	ListExceptions(ctx context.Context, userEmail string, pageSize int32, pageToken string) ([]Exception, error)
}
