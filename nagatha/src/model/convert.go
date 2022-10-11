package model

import (
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/nianticlabs/modron/nagatha/src/pb"
)

func (n Notification) ToProto() *pb.Notification {
	return &pb.Notification{
		Uuid:         n.Uuid,
		SourceSystem: n.SourceSystem,
		Name:         n.Name,
		Recipient:    n.Recipient,
		Content:      n.Content,
		SentOn:       timestamppb.New(n.SentOn),
		Interval:     durationpb.New(n.Interval),
	}
}

func NotificationFromProto(p *pb.Notification) Notification {
	n := Notification{
		Uuid:         p.Uuid,
		SourceSystem: p.SourceSystem,
		Name:         p.Name,
		Recipient:    p.Recipient,
		Content:      p.Content,
	}
	if p.CreatedOn != nil {
		n.CreatedOn = p.CreatedOn.AsTime()
	}
	if p.SentOn != nil {
		n.SentOn = p.SentOn.AsTime()
	}
	if p.Interval != nil {
		n.Interval = p.Interval.AsDuration()
	}
	return n
}

func (e Exception) ToProto() *pb.Exception {
	return &pb.Exception{
		Uuid:             e.Uuid,
		SourceSystem:     e.SourceSystem,
		UserEmail:        e.UserEmail,
		NotificationName: e.NotificationName,
		Justification:    e.Justification,
		CreatedOnTime:    timestamppb.New(e.CreatedOn),
		ValidUntilTime:   timestamppb.New(e.ValidUntil),
	}
}

func ExceptionFromProto(p *pb.Exception) Exception {
	e := Exception{
		Uuid:             p.Uuid,
		SourceSystem:     p.SourceSystem,
		UserEmail:        p.UserEmail,
		NotificationName: p.NotificationName,
		Justification:    p.Justification,
	}
	if p.CreatedOnTime != nil {
		e.CreatedOn = p.CreatedOnTime.AsTime()
	}
	if p.ValidUntilTime != nil {
		e.ValidUntil = p.ValidUntilTime.AsTime()
	}
	return e
}
