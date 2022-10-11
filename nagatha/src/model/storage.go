// Package storage has a storage interface for exceptions
package model

import (
	"context"
)

type Storage interface {
	CreateNotification(ctx context.Context, notification Notification) (Notification, error)
	ListNotificationsToSend(ctx context.Context) ([]Notification, error)
	ListNotifications(ctx context.Context) ([]Notification, error)
	NotificationSent(ctx context.Context, notification Notification) error
	GetException(ctx context.Context, uuid string) (Exception, error)
	CreateException(ctx context.Context, exception Exception) (Exception, error)
	EditException(ctx context.Context, exception Exception) (Exception, error)
	DeleteException(ctx context.Context, uuid string) error
	ListExceptions(ctx context.Context) ([]Exception, error)
}
