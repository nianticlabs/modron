package nagatha

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nianticlabs/modron/src/constants"
	modronmetric "github.com/nianticlabs/modron/src/metric"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/proto/generated/nagatha"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"golang.org/x/oauth2"
	"google.golang.org/protobuf/types/known/durationpb"
)

var meter = otel.Meter("github.com/nianticlabs/modron/src/nagatha")
var tracer = otel.Tracer("github.com/nianticlabs/modron/src/nagatha")
var log = logrus.StandardLogger().WithField(constants.LogKeyPkg, "nagatha")

const (
	sourceSystem = "modron"
	batchSize    = 1000
)

func New(addr, modronURL string, tokenSource oauth2.TokenSource) (model.NotificationService, error) {
	c, err := newNagathaClient(addr, tokenSource)
	if err != nil {
		return nil, err
	}
	s := &Service{
		client:    c,
		modronURL: modronURL,
	}
	err = s.initMetrics()
	return s, err
}

type Service struct {
	client    *Client
	modronURL string
	metrics   metrics
}

type metrics struct {
	ClientReqDuration    metric.Float64Histogram
	NotificationDuration metric.Float64Histogram
}

func (svc *Service) CreateNotification(ctx context.Context, notification model.Notification) (model.Notification, error) {
	ctx, span := tracer.Start(ctx, "CreateNotification")
	defer span.End()
	start := time.Now()
	if notification.Name == "" {
		return model.Notification{}, fmt.Errorf("name can't be empty")
	}
	notif, err := svc.client.CreateNotification(ctx, &nagatha.Notification{
		SourceSystem: sourceSystem,
		Name:         notification.Name,
		Recipient:    notification.Recipient,
		Content:      notification.Content,
		Interval:     durationpb.New(notification.Interval),
	})
	status := modronmetric.StatusSuccess
	if err != nil {
		status = modronmetric.StatusError
	}
	svc.metrics.NotificationDuration.
		Record(ctx, time.Since(start).Seconds(),
			metric.WithAttributes(
				attribute.String(modronmetric.KeyStatus, status),
				attribute.String(modronmetric.KeyRecipient, notification.Recipient),
			),
		)
	return notif, err
}

func (svc *Service) BatchCreateNotifications(ctx context.Context, notifications []model.Notification) ([]model.Notification, error) {
	ctx, span := tracer.Start(ctx, "BatchCreateNotifications")
	defer span.End()
	start := time.Now()
	var resultNotifications []model.Notification
	var errArr []error
	status := "success"
	for i := 0; i < len(notifications); i += batchSize {
		end := i + batchSize
		if end > len(notifications) {
			end = len(notifications)
		}
		notificationsProto := make([]*nagatha.Notification, 0, end-i)
		for _, n := range notifications[i:end] {
			if n.Recipient == "" {
				log.Warnf("recipient empty for notification %s", n.Name)
				continue
			}
			notificationsProto = append(notificationsProto, &nagatha.Notification{
				SourceSystem: sourceSystem,
				Name:         n.Name,
				Recipient:    n.Recipient,
				Content:      n.Content,
				Interval:     durationpb.New(n.Interval),
			})
		}
		notif, err := svc.client.BatchCreateNotifications(ctx, notificationsProto)
		if err != nil {
			status = "error"
			log.Warnf("error creating notifications: %v", err)
			errArr = append(errArr, err)
			continue
		}
		resultNotifications = append(resultNotifications, notif...)
		log.Infof("%d notifications remaining", len(notifications)-end)
	}
	svc.metrics.NotificationDuration.
		Record(ctx, time.Since(start).Seconds(),
			metric.WithAttributes(
				attribute.String(modronmetric.KeyStatus, status),
				attribute.Int(modronmetric.KeyCount, len(notifications)),
			),
		)
	return resultNotifications, errors.Join(errArr...)
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

func (svc *Service) initMetrics() error {
	notificationDurationHistogram, err := meter.Float64Histogram(constants.MetricsPrefix + "notifications_sent_total")
	if err != nil {
		return err
	}
	clientReqDurationHist, err := meter.Float64Histogram(
		constants.MetricsPrefix+"client_requests_duration",
		metric.WithDescription("Duration of client requests in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}
	svc.metrics = metrics{
		NotificationDuration: notificationDurationHistogram,
		ClientReqDuration:    clientReqDurationHist,
	}
	return nil
}
