package service_test

import (
	"context"
	"errors"
	"sync"

	"github.com/nianticlabs/modron/src/model"
)

type mockNotifier struct {
	notificationLock sync.Mutex
	notifications    []model.Notification
}

func (m *mockNotifier) BatchCreateNotifications(ctx context.Context, notifications []model.Notification) ([]model.Notification, error) {
	var createdNotifications []model.Notification
	var errArr []error
	for _, n := range notifications {
		created, err := m.CreateNotification(ctx, n)
		if err != nil {
			errArr = append(errArr, err)
			continue
		}
		createdNotifications = append(createdNotifications, created)
	}
	return createdNotifications, errors.Join(errArr...)
}

func (m *mockNotifier) CreateNotification(_ context.Context, notification model.Notification) (model.Notification, error) {
	m.notificationLock.Lock()
	defer m.notificationLock.Unlock()
	m.notifications = append(m.notifications, notification)
	return notification, nil
}

func (m *mockNotifier) GetException(context.Context, string) (model.Exception, error) {
	panic("implement me")
}

func (m *mockNotifier) CreateException(context.Context, model.Exception) (model.Exception, error) {
	panic("implement me")
}

func (m *mockNotifier) UpdateException(context.Context, model.Exception) (model.Exception, error) {
	panic("implement me")
}

func (m *mockNotifier) DeleteException(context.Context, string) error {
	panic("implement me")
}

func (m *mockNotifier) ListExceptions(context.Context, string, int32, string) ([]model.Exception, error) {
	panic("implement me")
}

var _ model.NotificationService = (*mockNotifier)(nil)

func newMockNotifier() *mockNotifier {
	return &mockNotifier{}
}

func (m *mockNotifier) getNotifications() []model.Notification {
	// Clone the notifications
	m.notificationLock.Lock()
	defer m.notificationLock.Unlock()
	notifications := make([]model.Notification, len(m.notifications))
	copy(notifications, m.notifications)
	return notifications
}
