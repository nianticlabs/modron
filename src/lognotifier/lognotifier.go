package lognotifier

import (
	"context"

	"github.com/golang/glog"
	"github.com/nianticlabs/modron/src/model"
)

func New() model.NotificationService {
	return &LogNotifier{}
}

type LogNotifier struct {
	exceptions []model.Exception
}

func (ln *LogNotifier) CreateNotification(ctx context.Context, notification model.Notification) (model.Notification, error) {
	glog.Infof("create notification: %+v", notification)
	return notification, nil
}

func (ln *LogNotifier) GetException(ctx context.Context, uuid string) (model.Exception, error) {
	glog.Infof("get exception called with %q", uuid)
	return model.Exception{Uuid: uuid}, nil
}
func (ln *LogNotifier) CreateException(ctx context.Context, exception model.Exception) (model.Exception, error) {
	glog.Infof("create exception %+v", exception)
	ln.exceptions = append(ln.exceptions, exception)
	return exception, nil
}
func (ln *LogNotifier) UpdateException(ctx context.Context, exception model.Exception) (model.Exception, error) {
	glog.Infof("update exception: %+v", exception)
	return exception, nil
}
func (ln *LogNotifier) DeleteException(ctx context.Context, id string) error {
	glog.Infof("delete exception %q", id)
	return nil
}
func (ln *LogNotifier) ListExceptions(ctx context.Context, userEmail string, pageSize int32, pageToken string) ([]model.Exception, error) {
	glog.Infof("list exceptions for user %q", userEmail)
	return ln.exceptions, nil
}
