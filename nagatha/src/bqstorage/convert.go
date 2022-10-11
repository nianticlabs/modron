package bqstorage

import (
	"time"

	"github.com/nianticlabs/modron/nagatha/src/model"
)

type BqException struct {
	Uuid             string
	SourceSystem     string
	UserEmail        string
	NotificationName string
	Justification    string
	CreatedOn        time.Time
	ValidUntil       time.Time
}

func (e BqException) BqEntryToException() model.Exception {
	return model.Exception{
		Uuid:             e.Uuid,
		SourceSystem:     e.SourceSystem,
		UserEmail:        e.UserEmail,
		NotificationName: e.NotificationName,
		Justification:    e.Justification,
		CreatedOn:        e.CreatedOn,
		ValidUntil:       e.ValidUntil,
	}
}

func ExceptionToBqEntry(e model.Exception) BqException {
	return BqException{
		Uuid:             e.Uuid,
		SourceSystem:     e.SourceSystem,
		UserEmail:        e.UserEmail,
		NotificationName: e.NotificationName,
		Justification:    e.Justification,
		CreatedOn:        e.CreatedOn,
		ValidUntil:       e.ValidUntil,
	}
}

type BqNotification struct {
	Uuid                     string
	SourceSystem             string
	Name                     string
	Recipient                string
	Content                  string
	CreatedOn                time.Time
	SentOn                   time.Time
	IntervalBetweenReminders string
}

func (n BqNotification) BqEntryToNotification() (model.Notification, error) {
	notif := model.Notification{
		Uuid:         n.Uuid,
		SourceSystem: n.SourceSystem,
		Name:         n.Name,
		Recipient:    n.Recipient,
		Content:      n.Content,
	}
	if !n.CreatedOn.IsZero() {
		notif.CreatedOn = n.CreatedOn
	} else {
		notif.CreatedOn = time.Time{}
	}
	if !n.SentOn.IsZero() {
		notif.SentOn = n.SentOn
	} else {
		notif.SentOn = time.Time{}
	}
	if n.IntervalBetweenReminders != "" {
		d, err := time.ParseDuration(n.IntervalBetweenReminders)
		if err != nil {
			return model.Notification{}, err
		}
		notif.Interval = d
	}
	return notif, nil
}

func NotificationToBqEntry(n model.Notification) BqNotification {
	bqn := BqNotification{
		Uuid:                     n.Uuid,
		SourceSystem:             n.SourceSystem,
		Name:                     n.Name,
		Recipient:                n.Recipient,
		Content:                  n.Content,
		IntervalBetweenReminders: n.Interval.String(),
	}
	if !n.CreatedOn.IsZero() {
		bqn.CreatedOn = n.CreatedOn
	} else {
		bqn.CreatedOn = time.Time{}
	}
	if !n.SentOn.IsZero() {
		bqn.SentOn = n.SentOn
	} else {
		bqn.SentOn = time.Time{}
	}
	return bqn
}
