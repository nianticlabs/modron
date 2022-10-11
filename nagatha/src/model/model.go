package model

import (
	"context"
	"time"
)

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

func (n Notification) HasMatchingException(exceptions []Exception) bool {
	for _, e := range exceptions {
		if n.Recipient == e.UserEmail && n.SourceSystem == e.SourceSystem {
			if e.NotificationName == "" || n.Name == e.NotificationName {
				return true
			}
		}
	}
	return false
}

type Exception struct {
	Uuid             string
	SourceSystem     string
	UserEmail        string
	NotificationName string
	Justification    string
	CreatedOn        time.Time
	ValidUntil       time.Time
}

type EmailSender interface {
	SendEmail(ctx context.Context, sender, object, content string, recipients []string) error
}
