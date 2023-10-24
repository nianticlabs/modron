// Package storage has a storage interface for exceptions
package model

import (
	"context"
	"time"
)

type Storage interface {
	CreateException(ctx context.Context, exception Exception) (Exception, error)
	CreateNotification(ctx context.Context, notification Notification) (Notification, error)
	DeleteException(ctx context.Context, uuid string) error
	EditException(ctx context.Context, exception Exception) (Exception, error)
	GetException(ctx context.Context, uuid string) (Exception, error)
	LastSendDate(ctx context.Context) (time.Time, error)
	ListExceptions(ctx context.Context, userEmail string) ([]Exception, error)
	ListNotifications(ctx context.Context) ([]Notification, error)
	ListNotificationsToSend(ctx context.Context) ([]Notification, error)
	NotificationSent(ctx context.Context, notification Notification) error
}
