package sendgridsender

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/nianticlabs/modron/nagatha/src/model"
)

type Config struct {
	ApiKey        string
	TestRecipient string
}

type SendGridSender struct {
	cfg Config
}

func NewSendGridSender(cfg Config) model.EmailSender {
	return &SendGridSender{cfg: cfg}
}

func (es *SendGridSender) SendEmail(ctx context.Context, sender, subject, content string, recipients []string) error {
	if es.cfg.ApiKey == "" {
		return fmt.Errorf("API key is empty, can't send email")
	}
	if es.cfg.TestRecipient != "" {
		glog.Infof("Debug mode enabled, all emails going to %s", es.cfg.TestRecipient)
		recipients = []string{es.cfg.TestRecipient}
	}
	errs := []string{}
	for _, recipient := range recipients {
		message := mail.NewSingleEmail(
			mail.NewEmail("", sender),
			subject,
			mail.NewEmail("", recipient),
			content,
			content,
		)
		client := sendgrid.NewSendClient(es.cfg.ApiKey)
		response, err := client.Send(message)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%v", err))
			continue
		}
		if response.StatusCode != 202 {
			errs = append(errs, fmt.Sprintf("sendgrid %v: %+v", response.StatusCode, response.Body))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("sendgrid: %v", errs)
	}
	return nil
}
