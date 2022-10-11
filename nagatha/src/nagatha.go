// Binary nagatha provides a way to notify people repeatidly until conditions are met.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nianticlabs/modron/nagatha/src/model"
	"github.com/nianticlabs/modron/nagatha/src/pb"
	"github.com/nianticlabs/modron/nagatha/src/sendgridsender"

	"github.com/golang/glog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	emailSenderAddressEnvVar  = "EMAIL_SENDER_ADDRESS"
	exceptionTableIDEnvVar    = "EXCEPTION_TABLE_ID"
	gcpProjectIDEnvVar        = "GCP_PROJECT_ID"
	notificationTableIDEnvVar = "NOTIFICATION_TABLE_ID"
	portEnvVar                = "PORT"
	sendGridAPIKeyEnvVar      = "SENDGRID_API_KEY" //#nosec G101
	testRecipientEnvVar       = "TEST_NOTIFICATION_RECIPIENT"
	emailSubject              = "🚨 Niantic Action Items 🚨"
)

var (
	emailSenderAddress  = os.Getenv(emailSenderAddressEnvVar)
	exceptionTableID    = os.Getenv(exceptionTableIDEnvVar)
	notificationTableID = os.Getenv(notificationTableIDEnvVar)
	projectID           = os.Getenv(gcpProjectIDEnvVar)
	sendgridApiKey      = os.Getenv(sendGridAPIKeyEnvVar)

	port       int32
	oneDay     = time.Duration(time.Hour * 24)
	ninetyDays = time.Duration(time.Hour * 24 * 90)
)

type nagathaService struct {
	storage     model.Storage
	emailSender model.EmailSender

	// Required
	pb.UnimplementedNagathaServer
}

func (nag *nagathaService) CreateNotification(ctx context.Context, req *pb.CreateNotificationRequest) (*pb.Notification, error) {
	if req.Notification == nil {
		return nil, status.Errorf(codes.InvalidArgument, "notification must be provided, got %q", req.Notification)
	}
	if req.Notification.Uuid != "" {
		return nil, status.Errorf(codes.InvalidArgument, "uuid can't be set, got %q", req.Notification.Uuid)
	}
	if req.Notification.Recipient == "" {
		return nil, status.Errorf(codes.InvalidArgument, "recipient can't be empty")
	}
	if req.Notification.SourceSystem == "" {
		return nil, status.Errorf(codes.InvalidArgument, "source_system can't be empty")
	}
	if req.Notification.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name cannot be empty")
	}
	if req.Notification.CreatedOn != nil {
		return nil, status.Errorf(codes.InvalidArgument, "created on is set automatically by the backend and can't be set by the client")
	}
	if req.Notification.SentOn != nil {
		return nil, status.Errorf(codes.InvalidArgument, "sent on is set automatically by the backend and can't be set by the client")
	}
	if req.Notification.Interval.AsDuration() < oneDay {
		return nil, status.Errorf(codes.InvalidArgument, "notification interval can't be shorter than %q, got %q", oneDay, req.Notification.Interval.AsDuration())
	}
	existingNotifications, err := nag.storage.ListNotifications(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "existing notifications: %v", err)
	}
	exceptions, err := nag.storage.ListExceptions(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list exceptions: %v", err)
	}
	if model.NotificationFromProto(req.Notification).HasMatchingException(exceptions) {
		return nil, status.Error(codes.FailedPrecondition, "an exception exist for this notification, creation aborted")
	}
	for _, n := range existingNotifications {
		if n.Name == req.Notification.Name && n.Recipient == req.Notification.Recipient {
			if n.SentOn.IsZero() {
				return nil, status.Errorf(codes.AlreadyExists, "an unsent notification for that user already exists")
			}
			if n.SentOn.Add(n.Interval).After(time.Now()) {
				return nil, status.Errorf(codes.InvalidArgument, "last notification sent on %s, wait until %s to send the next one", n.SentOn.Format(time.RFC3339), n.SentOn.Add(n.Interval).Format(time.RFC3339))
			}
		}
	}
	req.Notification.CreatedOn = timestamppb.Now()
	n, err := nag.storage.CreateNotification(ctx, model.NotificationFromProto(req.Notification))
	if err != nil {
		return nil, err
	}
	return n.ToProto(), nil
}

func (nag *nagathaService) GetException(ctx context.Context, req *pb.GetExceptionRequest) (*pb.Exception, error) {
	exp, err := nag.storage.GetException(ctx, req.Uuid)
	if err != nil {
		return nil, status.Error(codes.NotFound, "error")
	}
	return exp.ToProto(), nil
}

func (nag *nagathaService) CreateException(ctx context.Context, req *pb.CreateExceptionRequest) (*pb.Exception, error) {
	if req.Exception.Uuid != "" {
		return nil, status.Errorf(codes.InvalidArgument, "uuid can't be set, got %q", req.Exception.Uuid)
	}
	if req.Exception.Justification == "" {
		return nil, status.Error(codes.InvalidArgument, "justification cannot be empty")
	}
	if req.Exception.SourceSystem == "" {
		return nil, status.Error(codes.InvalidArgument, "source system cannot be empty")
	}
	if req.Exception.UserEmail == "" {
		return nil, status.Error(codes.InvalidArgument, "user email cannot be empty")
	}
	maxValidityTime := time.Now().Add(ninetyDays)
	if req.Exception.ValidUntilTime.AsTime().After(maxValidityTime) {
		return nil, status.Errorf(codes.InvalidArgument, "validity cannot be set beyond %s", maxValidityTime.Format(time.RFC3339))
	}
	exp, err := nag.storage.CreateException(ctx, model.ExceptionFromProto(req.Exception))
	if err != nil {
		return nil, err
	}
	return exp.ToProto(), nil
}

func (nag *nagathaService) UpdateException(ctx context.Context, req *pb.UpdateExceptionRequest) (*pb.Exception, error) {
	excp, err := nag.storage.GetException(ctx, req.Exception.Uuid)
	if err != nil {
		s, ok := status.FromError(err)
		if !ok {
			return nil, fmt.Errorf("invalid status error: %v", err)
		}
		if s.Code() == codes.NotFound {
			return nil, fmt.Errorf("create exception before updating")
		}
		return nil, fmt.Errorf("update error: %v", s)
	}
	if !req.UpdateMask.IsValid(req) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid field mask")
	}
	newExpProto := excp.ToProto()
	for _, path := range req.GetUpdateMask().GetPaths() {
		switch path {
		case "exception.notification_name":
			newExpProto.NotificationName = req.Exception.GetNotificationName()
		case "exception.valid_until":
			newExpProto.ValidUntilTime = req.Exception.GetValidUntilTime()
		case "exception.user_email":
			newExpProto.UserEmail = req.Exception.GetUserEmail()
		case "exception.justification":
			newExpProto.Justification = req.Exception.GetJustification()
		case "exception.source_system":
			newExpProto.SourceSystem = req.Exception.GetSourceSystem()
		case "exception.uuid":
			return nil, status.Errorf(codes.InvalidArgument, "uuid cannot be updated")
		}
	}
	newExp, err := nag.storage.EditException(ctx, model.ExceptionFromProto(newExpProto))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update errror: %v", err)
	}
	return newExp.ToProto(), nil
}

func (nag *nagathaService) DeleteException(ctx context.Context, req *pb.DeleteExceptionRequest) (*emptypb.Empty, error) {
	if req.Uuid == "" {
		return nil, status.Error(codes.InvalidArgument, "uuid is required, was empty")
	}
	if err := nag.storage.DeleteException(ctx, req.Uuid); err != nil {
		return nil, status.Errorf(codes.Internal, "delete exception: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (nag *nagathaService) ListExceptions(ctx context.Context, req *pb.ListExceptionsRequest) (*pb.ListExceptionsResponse, error) {
	expList, err := nag.storage.ListExceptions(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list: %v", err)
	}
	ret := &pb.ListExceptionsResponse{}
	for _, e := range expList {
		ret.Exceptions = append(ret.Exceptions, e.ToProto())
	}
	return ret, nil
}

func (nag *nagathaService) NotifyUser(ctx context.Context, req *pb.NotifyUserRequest) (*pb.NotifyUserResponse, error) {
	// TODO(lds): Group notifications for the same user and send one email only with a summary of all.
	if req.UserEmail == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user email can't be empty, got %q", req.UserEmail)
	}
	if req.Title == "" {
		return nil, status.Errorf(codes.InvalidArgument, "title can't be empty, got %q", req.Title)
	}
	if req.Content == "" {
		return nil, status.Errorf(codes.InvalidArgument, "content can't be empty, got %q", req.Content)
	}
	if !strings.Contains(req.UserEmail, "@") {
		return nil, status.Errorf(codes.InvalidArgument, "user email must contain an '@', got %q", req.UserEmail)
	}
	if err := nag.emailSender.SendEmail(ctx, emailSenderAddress, req.Title, req.Content, []string{req.UserEmail}); err != nil {
		return nil, status.Errorf(codes.Internal, "send email: %v", err)
	}
	glog.Infof("notified user %s", req.UserEmail)
	return &pb.NotifyUserResponse{}, nil
}

func (nag *nagathaService) NotifyAll(ctx context.Context, req *pb.NotifyAllRequest) (*pb.NotifyAllResponse, error) {
	exceptions, err := nag.storage.ListExceptions(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list exceptions: %v", err)
	}
	notifications, err := nag.storage.ListNotificationsToSend(ctx)
	if err != nil {
		return &pb.NotifyAllResponse{HasCompleted: false}, status.Errorf(codes.Internal, "list notification: %v", err)
	}
	glog.Infof("%d notifications to send", len(notifications))
	notificationsSent := 0
	notificationsByUser := make(map[string][]model.Notification)
	for _, n := range notifications {
		if !n.HasMatchingException(exceptions) {
			if _, ok := notificationsByUser[n.Recipient]; !ok {
				notificationsByUser[n.Recipient] = []model.Notification{}
			}
			notificationsByUser[n.Recipient] = append(notificationsByUser[n.Recipient], n)
		}
	}
	for user, notif := range notificationsByUser {
		allMessages := []sendgridsender.EmailNotificationsContent{}
		for _, n := range notif {
			allMessages = append(allMessages, sendgridsender.EmailNotificationsContent{
				Topic:   n.Name,
				Message: n.Content,
			})
		}
		content, err := sendgridsender.Template(emailSubject, allMessages)
		if err != nil {
			fmt.Printf("email format error: %v\n", err)
			for _, n := range notif {
				content += n.Content + "\n"
			}
		}
		if _, err := nag.NotifyUser(ctx, &pb.NotifyUserRequest{
			UserEmail: user,
			Title:     emailSubject,
			Content:   content,
		}); err != nil {
			glog.Errorf("%s notification failed: %v", user, err)
		} else {
			for _, n := range notif {
				if err := nag.storage.NotificationSent(ctx, n); err != nil {
					glog.Errorf("could not set %v as sent: %v", n, err)
				}
			}
			notificationsSent += len(notif)
		}
	}
	glog.Infof("sent %d notifications", notificationsSent)
	// TODO(lds): Add a list of notification errors to the response.
	return &pb.NotifyAllResponse{HasCompleted: true}, nil
}
