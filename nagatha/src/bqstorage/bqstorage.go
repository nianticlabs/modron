package bqstorage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"github.com/nianticlabs/modron/nagatha/src/model"
)

const (
	bigQueryZero = "0001-01-01 00:00:00 UTC"
)

type BigQueryStorage struct {
	client *bigquery.Client
	cfg    Config
}

type Config struct {
	NotificationTableID string
	ExceptionTableID    string
	ProjectID           string
}

func New(ctx context.Context, cfg Config) (model.Storage, error) {
	c, err := bigquery.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		return nil, err
	}
	return &BigQueryStorage{client: c, cfg: cfg}, nil
}

func (bqs *BigQueryStorage) CreateNotification(ctx context.Context, notification model.Notification) (model.Notification, error) {
	notification.Uuid = uuid.NewString()
	v := strings.Split(bqs.cfg.NotificationTableID, ".")
	ins := bqs.client.Dataset(v[0]).Table(v[1]).Inserter()
	if err := ins.Put(ctx, NotificationToBqEntry(notification)); err != nil {
		return model.Notification{}, fmt.Errorf("insert notification: %v", err)
	}
	return notification, nil
}

func (bqs *BigQueryStorage) GetException(ctx context.Context, uuid string) (model.Exception, error) {
	query := bqs.client.Query(`SELECT * FROM ` + bqs.cfg.ExceptionTableID + ` WHERE uuid = @uuid`)
	query.Parameters = []bigquery.QueryParameter{
		{
			Name:  "uuid",
			Value: uuid,
		},
	}
	it, err := bqs.runQuery(ctx, query)
	if err != nil {
		return model.Exception{}, fmt.Errorf("get: %v; query: %v parameters: %v", err, query, query.Parameters)
	}
	exceptions, err := exceptionsFromAnswer(it)
	if err != nil {
		return model.Exception{}, fmt.Errorf("get: %v", err)
	}
	if len(exceptions) != 1 {
		return model.Exception{}, fmt.Errorf("get: more than one exception with id %q", uuid)
	}
	return exceptions[0], nil
}

func (bqs *BigQueryStorage) CreateException(ctx context.Context, exception model.Exception) (model.Exception, error) {
	v := strings.Split(bqs.cfg.ExceptionTableID, ".")
	ins := bqs.client.Dataset(v[0]).Table(v[1]).Inserter()
	exception.Uuid = uuid.NewString()
	if err := ins.Put(ctx, ExceptionToBqEntry(exception)); err != nil {
		return model.Exception{}, fmt.Errorf("insert exception: %v", err)
	}
	return exception, nil
}
func (bqs *BigQueryStorage) EditException(ctx context.Context, exception model.Exception) (model.Exception, error) {
	return model.Exception{}, fmt.Errorf("unimplemented")
}
func (bqs *BigQueryStorage) DeleteException(ctx context.Context, uuid string) error {
	return fmt.Errorf("unimplemented")
}
func (bqs *BigQueryStorage) ListExceptions(ctx context.Context, user string) ([]model.Exception, error) {
	// TODO(lds): Add pagination here if the list becomes too large.
	querySql := `SELECT * FROM ` + bqs.cfg.ExceptionTableID
	if user != "" {
		querySql += ` WHERE userEmail = @userEmail`
	}
	query := bqs.client.Query(querySql)
	query.Parameters = []bigquery.QueryParameter{
		{
			Name:  "userEmail",
			Value: user,
		},
	}
	it, err := bqs.runQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list exceptions: %q", err)
	}
	exceptions, err := exceptionsFromAnswer(it)
	if err != nil {
		return nil, fmt.Errorf("list: %v", err)
	}
	return exceptions, nil
}

func (bqs *BigQueryStorage) ListNotifications(ctx context.Context) ([]model.Notification, error) {
	query := bqs.client.Query(
		`SELECT uuid, sourceSystem, name, content, recipient, intervalBetweenReminders, createdOn , countif(sentOn != "` + bigQueryZero + `") as sent
		FROM  ` + bqs.cfg.NotificationTableID + `
		GROUP BY uuid, sourceSystem, name, content, recipient, createdOn, intervalBetweenReminders`)
	it, err := bqs.runQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %q", err)
	}
	notifications, err := notificationsFromAnswer(it)
	if err != nil {
		return nil, fmt.Errorf("list: %v", err)
	}
	return notifications, nil
}

func (bqs *BigQueryStorage) ListNotificationsToSend(ctx context.Context) ([]model.Notification, error) {
	query := bqs.client.Query(
		`SELECT * FROM (
			SELECT uuid, sourceSystem, name, content, recipient, intervalBetweenReminders, createdOn , countif(sentOn != "` + time.Time{}.Format(time.RFC3339) + `") as sent
			FROM ` + bqs.cfg.NotificationTableID + `
			GROUP BY uuid, sourceSystem, name,content, recipient, createdOn, intervalBetweenReminders)
		  WHERE sent = 0`)
	it, err := bqs.runQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %q", err)
	}
	notifications, err := notificationsFromAnswer(it)
	if err != nil {
		return nil, fmt.Errorf("list: %v", err)
	}
	return notifications, nil
}

func (bqs *BigQueryStorage) LastSendDate(ctx context.Context) (time.Time, error) {
	query := bqs.client.Query(
		`SELECT * FROM ` + bqs.cfg.NotificationTableID + ` ORDER BY sentOn DESC LIMIT 1`)

	it, err := bqs.runQuery(ctx, query)
	if err != nil {
		return time.Time{}, fmt.Errorf("list notifications: %q", err)
	}
	notifications, err := notificationsFromAnswer(it)
	if err != nil {
		return time.Time{}, fmt.Errorf("last send date: %v", err)
	}
	if len(notifications) != 1 {
		return time.Time{}, fmt.Errorf("last send date expected len(notifications) to be 1, is %d", len(notifications))
	}
	return notifications[0].SentOn, nil
}

func (bqs *BigQueryStorage) NotificationSent(ctx context.Context, notification model.Notification) error {
	notification.SentOn = time.Now()
	v := strings.Split(bqs.cfg.NotificationTableID, ".")
	ins := bqs.client.Dataset(v[0]).Table(v[1]).Inserter()
	if err := ins.Put(ctx, NotificationToBqEntry(notification)); err != nil {
		return fmt.Errorf("insert notification sent: %v", err)
	}
	return nil
}

func (bqs *BigQueryStorage) runQuery(ctx context.Context, q *bigquery.Query) (*bigquery.RowIterator, error) {
	j, err := q.Run(ctx)
	if err != nil {
		return nil, err
	}
	s, err := j.Wait(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	it, err := j.Read(ctx)
	if err != nil {
		return nil, err
	}
	return it, nil
}

func exceptionsFromAnswer(it *bigquery.RowIterator) ([]model.Exception, error) {
	exceptions := []model.Exception{}
	for {
		var row BqException
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		exceptions = append(exceptions, row.BqEntryToException())
	}
	return exceptions, nil
}

func notificationsFromAnswer(it *bigquery.RowIterator) ([]model.Notification, error) {
	notifications := []model.Notification{}
	for {
		var row BqNotification
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		n, err := row.BqEntryToNotification()
		if err != nil {
			glog.Warningf("notification conversion: %v", err)
			continue
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}
