package nagatha

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/durationpb"
	"github.com/nianticlabs/modron/src/model"
)

const (
	sourceSystem = "modron"
	clientID     = "143415353591-bsmr7ii98a2493kts699289n2ommqi07.apps.googleusercontent.com"
)

func New(ctx context.Context, addr string) (model.NotificationService, error) {
	c, err := NewNagathaClient(addr)
	if err != nil {
		return nil, err
	}
	return &Service{
		client: c,
	}, nil
}

type Service struct {
	model.NotificationService
	client *NagathaClient
}

func (svc *Service) CreateNotification(ctx context.Context, notification model.Notification) (model.Notification, error) {
	if notification.Name == "" {
		return model.Notification{}, fmt.Errorf("name can't be empty")
	}
	err := svc.client.CreateNotification(ctx, &Notification{
		SourceSystem: sourceSystem,
		Name:         notification.Name,
		UserEmail:    notification.Recipient,
		Content:      notification.Content,
		Interval:     durationpb.New(notification.Interval),
	})
	if err != nil {
		return model.Notification{}, err
	}
	return model.Notification{}, nil
}

func (svc *Service) GetException(ctx context.Context, uuid string) (model.Exception, error) {
	ex, err := svc.client.GetException(ctx, uuid)
	if err != nil {
		return model.Exception{}, err
	}
	return exceptionModelFromNagathaProto(ex), nil
}

func (svc *Service) CreateException(ctx context.Context, exception model.Exception) (model.Exception, error) {
	return exception, svc.client.CreateException(ctx, exceptionNagathaProtoFromModel(exception))
}

func (svc *Service) UpdateException(ctx context.Context, exception model.Exception) (model.Exception, error) {
	return exception, svc.client.UpdateException(ctx, exceptionNagathaProtoFromModel(exception), nil)
}

func (svc *Service) DeleteException(ctx context.Context, id string) error {
	return svc.client.DeleteException(ctx, id)
}

func (svc *Service) ListExceptions(ctx context.Context, userEmail string, pageSize int32, pageToken string) ([]model.Exception, error) {
	exceptions := make([]model.Exception, 0)
	resp, err := svc.client.ListExceptions(ctx, userEmail, pageSize, pageToken)
	if err != nil {
		return nil, err
	}
	for _, e := range resp.Exceptions {
		exceptions = append(exceptions, exceptionModelFromNagathaProto(e))
	}
	return exceptions, nil
}
