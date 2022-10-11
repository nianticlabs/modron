// Package memstorage implements an inmemory storage for testing purposes.
package memstorage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nianticlabs/modron/nagatha/src/model"
)

type MemStorage struct {
	exceptions    []model.Exception
	notifications []model.Notification
}

func New() model.Storage {
	return &MemStorage{}
}

func (mem *MemStorage) CreateNotification(ctx context.Context, notification model.Notification) (model.Notification, error) {
	notification.Uuid = uuid.NewString()
	mem.notifications = append(mem.notifications, notification)
	return notification, nil
}

func (mem *MemStorage) GetException(ctx context.Context, uuid string) (model.Exception, error) {
	for _, e := range mem.exceptions {
		if e.Uuid == uuid {
			return e, nil
		}
	}
	return model.Exception{}, fmt.Errorf("%s not found", uuid)
}

func (mem *MemStorage) CreateException(ctx context.Context, exception model.Exception) (model.Exception, error) {
	exception.Uuid = uuid.NewString()
	mem.exceptions = append(mem.exceptions, exception)
	return exception, nil
}

func (mem *MemStorage) EditException(ctx context.Context, exception model.Exception) (model.Exception, error) {
	for i, e := range mem.exceptions {
		if e.Uuid == exception.Uuid {
			mem.exceptions[i].Uuid = exception.Uuid
			mem.exceptions[i].SourceSystem = exception.SourceSystem
			mem.exceptions[i].UserEmail = exception.UserEmail
			mem.exceptions[i].NotificationName = exception.NotificationName
			mem.exceptions[i].Justification = exception.Justification
			mem.exceptions[i].CreatedOn = exception.CreatedOn
			mem.exceptions[i].ValidUntil = exception.ValidUntil
			return exception, nil
		}
	}
	return model.Exception{}, fmt.Errorf("%s not found", exception.Uuid)
}

func (mem *MemStorage) DeleteException(ctx context.Context, uuid string) error {
	for i, e := range mem.exceptions {
		if e.Uuid == uuid {
			mem.exceptions = append(mem.exceptions[:i], mem.exceptions[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("%s not found", uuid)
}

func (mem *MemStorage) ListExceptions(ctx context.Context) ([]model.Exception, error) {
	return mem.exceptions, nil
}

func (mem *MemStorage) ListNotificationsToSend(ctx context.Context) ([]model.Notification, error) {
	notifs := make([]model.Notification, 0)
	for _, n := range mem.notifications {
		if !n.SentOn.IsZero() {
			continue
		}
		sendNotification := true
		for _, e := range mem.exceptions {
			if n.Name == e.NotificationName && n.Recipient == e.UserEmail && n.SourceSystem == e.SourceSystem {
				sendNotification = false
			}
		}
		if sendNotification {
			notifs = append(notifs, n)
		}
	}
	return notifs, nil
}

func (mem *MemStorage) ListNotifications(ctx context.Context) ([]model.Notification, error) {
	return mem.notifications, nil
}

func (mem *MemStorage) NotificationSent(ctx context.Context, notification model.Notification) error {
	for i, n := range mem.notifications {
		if n.Uuid == notification.Uuid {
			mem.notifications[i].SentOn = time.Now()
			return nil
		}
	}
	return fmt.Errorf("could not find notification %q", notification)
}
