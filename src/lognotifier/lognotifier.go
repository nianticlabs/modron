package lognotifier

import (
	"context"
	"errors"

	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"

	"github.com/sirupsen/logrus"
)

func New() model.NotificationService {
	return &LogNotifier{}
}

var log = logrus.StandardLogger().WithField(constants.LogKeyPkg, "lognotifier")

type LogNotifier struct {
	exceptions []model.Exception
}

func (ln *LogNotifier) BatchCreateNotifications(ctx context.Context, notifications []model.Notification) ([]model.Notification, error) {
	var resultNotifications []model.Notification
	var errArr []error
	for _, v := range notifications {
		notif, err := ln.CreateNotification(ctx, v)
		if err != nil {
			errArr = append(errArr, err)
			continue
		}
		resultNotifications = append(resultNotifications, notif)
	}
	return resultNotifications, errors.Join(errArr...)
}

func (ln *LogNotifier) CreateNotification(_ context.Context, notification model.Notification) (model.Notification, error) {
	log.Infof("create notification: %+v", notification)
	return notification, nil
}

func (ln *LogNotifier) GetException(_ context.Context, uuid string) (model.Exception, error) {
	log.Infof("get exception called with %q", uuid)
	return model.Exception{UUID: uuid}, nil
}

func (ln *LogNotifier) CreateException(_ context.Context, exception model.Exception) (model.Exception, error) {
	log.Infof("create exception %+v", exception)
	ln.exceptions = append(ln.exceptions, exception)
	return exception, nil
}

func (ln *LogNotifier) UpdateException(_ context.Context, exception model.Exception) (model.Exception, error) {
	log.Infof("update exception: %+v", exception)
	return exception, nil
}

func (ln *LogNotifier) DeleteException(_ context.Context, id string) error {
	log.Infof("delete exception %q", id)
	return nil
}

func (ln *LogNotifier) ListExceptions(_ context.Context, userEmail string, _ int32, _ string) ([]model.Exception, error) {
	log.Infof("list exceptions for user %q", userEmail)
	return ln.exceptions, nil
}
