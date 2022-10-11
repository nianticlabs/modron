package main

import (
	"context"
	"testing"
	"time"

	"github.com/nianticlabs/modron/nagatha/src/memstorage"
	"github.com/nianticlabs/modron/nagatha/src/pb"
	"github.com/nianticlabs/modron/nagatha/src/sendgridsender"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	sevenDaysInSeconds = int64(60 * 60 * 24 * 7)
)

func TestExceptionCrudService(t *testing.T) {
	ctx := context.Background()
	storage := memstorage.New()
	fes := &fakeEmailSender{}
	nag := newServerForTesting(storage, fes)

	want := &pb.Exception{
		CreatedOnTime:    timestamppb.New(time.Time{}),
		ValidUntilTime:   timestamppb.New(time.Time{}),
		Justification:    "test-justification",
		NotificationName: "test-notification",
		SourceSystem:     "test-system",
		UserEmail:        "test@example.com",
	}
	resp, err := nag.CreateException(ctx, &pb.CreateExceptionRequest{Exception: want})
	if err != nil {
		t.Fatalf("CreateException(ctx, %+v) unexpected error: %v", want, err)
	}
	want.Uuid = resp.Uuid
	getReq := &pb.GetExceptionRequest{Uuid: resp.Uuid}
	got, err := nag.GetException(ctx, getReq)
	if err != nil {
		t.Fatalf("GetException(ctx, %q) error: %v", resp.Uuid, err)
	}
	if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
		t.Errorf("unexpected diff (-want, +got): %v", diff)
	}

	fm, err := fieldmaskpb.New(&pb.UpdateExceptionRequest{}, "exception.notification_name")
	if err != nil {
		t.Fatalf("invalid fieldmask %+v: %v", fm, err)
	}
	updateReq := &pb.UpdateExceptionRequest{
		Exception: &pb.Exception{
			Uuid:             resp.Uuid,
			NotificationName: "test-notification-updated",
			UserEmail:        "fake_user@example.com",
		},
		UpdateMask: fm,
	}
	if _, err := nag.UpdateException(ctx, updateReq); err != nil {
		t.Fatalf("UpdateException(ctx, %+v) error: %v", updateReq, err)
	}

	want.NotificationName = "test-notification-updated"
	want.CreatedOnTime = timestamppb.New(time.Time{})
	gotAfterUpdate, err := nag.GetException(ctx, getReq)
	if err != nil {
		t.Fatalf("GetException(ctx, %+v) error: %v", getReq, err)
	}
	if diff := cmp.Diff(want, gotAfterUpdate, protocmp.Transform()); diff != "" {
		t.Errorf("unexpected diff (-want, +got): %v", diff)
	}

	deleteReq := &pb.DeleteExceptionRequest{Uuid: resp.Uuid}
	if _, err := nag.DeleteException(ctx, deleteReq); err != nil {
		t.Errorf("DeleteException(ctx, %+v) error: %v", deleteReq, err)
	}

	if _, err := nag.GetException(ctx, getReq); err == nil {
		t.Errorf("GetException(ctx, %+v) want error, got none", getReq)
	} else {
		if s, ok := status.FromError(err); !ok {
			t.Fatalf("parse status error: %v", err)
		} else {
			if s.Code() != codes.NotFound {
				t.Errorf("got code %s, want %s", s.Code(), codes.NotFound)
			}
		}
	}
}

func TestFailures(t *testing.T) {
	ctx := context.Background()
	storage := memstorage.New()
	fes := &fakeEmailSender{}
	nag := newServerForTesting(storage, fes)

	nonExistingUuid := "nonExistingUuid"
	getReq := &pb.GetExceptionRequest{Uuid: nonExistingUuid}
	if _, err := nag.GetException(ctx, getReq); err == nil {
		t.Errorf("GetException(ctx, %+v) expected error, got none", getReq)
	} else {
		s, ok := status.FromError(err)
		if !ok {
			t.Errorf("invalid status %v", err)
		}
		if s.Code() != codes.NotFound {
			t.Errorf("code got %s, want %s", s.Code(), codes.NotFound)
		}
	}

	deleteReq := &pb.DeleteExceptionRequest{Uuid: ""}
	if _, err := nag.DeleteException(ctx, deleteReq); err == nil {
		t.Errorf("DeleteException(ctx, %+v) expected error, got none", deleteReq)
	} else {
		if s, ok := status.FromError(err); !ok {
			t.Fatalf("parse status error: %v", err)
		} else {
			if s.Code() != codes.InvalidArgument {
				t.Errorf("got code %s, want %s", s.Code(), codes.InvalidArgument)
			}
		}
	}
}

func TestNotifyUser(t *testing.T) {
	ctx := context.Background()
	storage := memstorage.New()
	fes := &fakeEmailSender{}
	nag := newServerForTesting(storage, fes)

	notifyReq := &pb.NotifyUserRequest{
		UserEmail: "test-user@test.com",
		Title:     "test-title",
		Content:   "test-content",
	}
	if _, err := nag.NotifyUser(ctx, notifyReq); err != nil {
		t.Errorf("NotifyUser(ctx, %+v) unexpected error: %v", notifyReq, err)
	}
	if fes.Called != 1 {
		t.Errorf("emailSender called %d times, want 1", fes.Called)
	}
	wantCall := []string{emailSenderAddress, "test-user@test.com", "test-title", "test-content"}
	if diff := cmp.Diff(wantCall, fes.Calls[0]); diff != "" {
		t.Errorf("emailSender calls diff (-want, +got): %v", diff)
	}
}

func TestNotifyUserFailures(t *testing.T) {
	ctx := context.Background()
	storage := memstorage.New()
	fes := &fakeEmailSender{}
	nag := newServerForTesting(storage, fes)

	for _, test := range []struct {
		desc        string
		req         *pb.NotifyUserRequest
		wantCode    codes.Code
		wantMessage string
	}{
		{
			desc: "no email fails",
			req: &pb.NotifyUserRequest{
				Title:   "test-title",
				Content: "test-content",
			},
			wantCode:    codes.InvalidArgument,
			wantMessage: "user email can't be empty, got \"\"",
		},
		{
			desc: "no title fails",
			req: &pb.NotifyUserRequest{
				UserEmail: "test-useremail",
				Content:   "test-content",
			},
			wantCode:    codes.InvalidArgument,
			wantMessage: "title can't be empty, got \"\"",
		},
		{
			desc: "no content fails",
			req: &pb.NotifyUserRequest{
				UserEmail: "test-useremail",
				Title:     "test-title",
			},
			wantCode:    codes.InvalidArgument,
			wantMessage: "content can't be empty, got \"\"",
		},
		{
			desc: "wrong email fails",
			req: &pb.NotifyUserRequest{
				UserEmail: "emailwithnoatsign",
				Title:     "test-title",
				Content:   "test-content",
			},
			wantCode:    codes.InvalidArgument,
			wantMessage: "user email must contain an '@', got \"emailwithnoatsign\"",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			if _, err := nag.NotifyUser(ctx, test.req); err == nil {
				t.Fatalf("NotifyUser(ctx, %+v) want error %q, got none", test.req, test.wantMessage)
			} else {
				s, ok := status.FromError(err)
				if !ok {
					t.Fatalf("want status error")
				}
				if s.Code() != test.wantCode {
					t.Errorf("NotifyUser(ctx, %+v) status got %s, want %s", test.req, s.Code(), test.wantCode)
				}
				if s.Message() != test.wantMessage {
					t.Errorf("NotifyUser(ctx, %+v) message got %s, want %s", test.req, s.Message(), test.wantMessage)
				}
			}

		})
	}
}

func TestNotificationCreationAndSend(t *testing.T) {
	ctx := context.Background()
	storage := memstorage.New()
	fes := &fakeEmailSender{}
	nag := newServerForTesting(storage, fes)

	req := &pb.CreateNotificationRequest{
		Notification: &pb.Notification{
			Name:         "test-name",
			Recipient:    "test-recipient@example.com",
			SourceSystem: "test",
			Content:      "notification content",
			Interval:     &durationpb.Duration{Seconds: sevenDaysInSeconds},
		},
	}
	if _, err := nag.CreateNotification(ctx, req); err != nil {
		t.Fatalf("CreateNotification(ctx, %+v) unexpected error: %v", req, err)
	}

	reqSecond := &pb.CreateNotificationRequest{
		Notification: &pb.Notification{
			Name:         "test-name",
			Recipient:    "test-recipient@example.com",
			SourceSystem: "test",
			Content:      "notification content",
			Interval:     &durationpb.Duration{Seconds: sevenDaysInSeconds},
		},
	}
	if _, err := nag.CreateNotification(ctx, reqSecond); err == nil {
		t.Errorf("CreateNotification(ctx, %+v) wanted error, got nil", reqSecond)
	} else {
		if s, ok := status.FromError(err); !ok {
			t.Fatalf("invalid status %s", s)
		} else {
			if s.Code() != codes.AlreadyExists {
				t.Errorf("status got %s, want %s: err: %v", s.Code(), codes.AlreadyExists, err)
			}
		}
	}

	sentReq := &pb.NotifyAllRequest{}
	if _, err := nag.NotifyAll(ctx, sentReq); err != nil {
		t.Fatalf("NotifyAll(ctx, %+v) unexpected error: %v", sentReq, err)
	}

	if _, err := nag.CreateNotification(ctx, reqSecond); err == nil {
		t.Errorf("CreateNotification(ctx, %+v) wanted error, got nil", reqSecond)
	} else {
		if s, ok := status.FromError(err); !ok {
			t.Fatalf("invalid status %s", s)
		} else {
			if s.Code() != codes.InvalidArgument {
				t.Errorf("status got %s, want %s: err: %v", s.Code(), codes.InvalidArgument, err)
			}
		}
	}
}

func TestNotificationCreationValidation(t *testing.T) {
	ctx := context.Background()
	storage := memstorage.New()
	fes := &fakeEmailSender{}
	nag := newServerForTesting(storage, fes)

	for _, test := range []struct {
		desc        string
		req         *pb.CreateNotificationRequest
		wantMessage string
	}{
		{
			desc: "no empty notification",
			req: &pb.CreateNotificationRequest{
				Notification: nil,
			},
			wantMessage: "notification must be provided, got \"<nil>\"",
		},
		{
			desc: "no empty recipient",
			req: &pb.CreateNotificationRequest{
				Notification: &pb.Notification{
					Name:         "test-name",
					SourceSystem: "test",
					Content:      "notification content",
					Interval:     &durationpb.Duration{Seconds: sevenDaysInSeconds},
				},
			},
			wantMessage: "recipient can't be empty",
		},
		{
			desc: "no more than one notification a day",
			req: &pb.CreateNotificationRequest{
				Notification: &pb.Notification{
					Name:         "test-name",
					Recipient:    "test-recipient",
					SourceSystem: "test",
					Content:      "notification content",
					Interval:     &durationpb.Duration{Seconds: 1},
				},
			},
			wantMessage: "notification interval can't be shorter than \"24h0m0s\", got \"1s\"",
		},
		{
			desc: "no empty name",
			req: &pb.CreateNotificationRequest{
				Notification: &pb.Notification{
					Recipient:    "test-recipient",
					SourceSystem: "test",
					Content:      "notification content",
					Interval:     &durationpb.Duration{Seconds: sevenDaysInSeconds},
				},
			},
			wantMessage: "name cannot be empty",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			if _, err := nag.CreateNotification(ctx, test.req); err == nil {
				t.Fatalf("NotifyUser(ctx, %+v) want error %q, got none", test.req, test.wantMessage)
			} else {
				s, ok := status.FromError(err)
				if !ok {
					t.Fatalf("want status error")
				}
				if s.Code() != codes.InvalidArgument {
					t.Errorf("NotifyUser(ctx, %+v) status got %q, want %q", test.req, s.Code(), codes.InvalidArgument)
				}
				if s.Message() != test.wantMessage {
					t.Errorf("NotifyUser(ctx, %+v) message got %q, want %q", test.req, s.Message(), test.wantMessage)
				}
			}

		})
	}
}

func TestNotificationCreationFailsIfAnExceptionExists(t *testing.T) {
	ctx := context.Background()
	storage := memstorage.New()
	fes := &fakeEmailSender{}
	nag := newServerForTesting(storage, fes)

	want := &pb.Exception{
		CreatedOnTime:    timestamppb.New(time.Time{}),
		ValidUntilTime:   timestamppb.New(time.Time{}),
		Justification:    "test-justification",
		NotificationName: "test-notification",
		SourceSystem:     "test-system",
		UserEmail:        "test@example.com",
	}
	_, err := nag.CreateException(ctx, &pb.CreateExceptionRequest{Exception: want})
	if err != nil {
		t.Fatalf("CreateException(ctx, %+v) unexpected error: %v", want, err)
	}

	req := &pb.CreateNotificationRequest{
		Notification: &pb.Notification{
			Name:         "test-notification",
			Recipient:    "test@example.com",
			SourceSystem: "test-system",
			Content:      "notification content",
			Interval:     &durationpb.Duration{Seconds: sevenDaysInSeconds},
		},
	}
	if _, err := nag.CreateNotification(ctx, req); err == nil {
		t.Fatalf("CreateNotification(ctx, %+v) wanted error, got nil", req)
	} else {
		s, ok := status.FromError(err)
		if !ok {
			t.Fatalf("want status error")
		}
		if s.Code() != codes.FailedPrecondition {
			t.Errorf("CreateNotification(ctx, %+v) status got %q, want %q", req, s.Code(), codes.FailedPrecondition)
		}
	}
}

func TestNotificationCreationFailsIfAnExceptionNoNameExists(t *testing.T) {
	ctx := context.Background()
	storage := memstorage.New()
	fes := &fakeEmailSender{}
	nag := newServerForTesting(storage, fes)

	want := &pb.Exception{
		CreatedOnTime:  timestamppb.New(time.Time{}),
		ValidUntilTime: timestamppb.New(time.Time{}),
		Justification:  "test-justification",
		SourceSystem:   "test-system",
		UserEmail:      "test@example.com",
	}
	_, err := nag.CreateException(ctx, &pb.CreateExceptionRequest{Exception: want})
	if err != nil {
		t.Fatalf("CreateException(ctx, %+v) unexpected error: %v", want, err)
	}

	req := &pb.CreateNotificationRequest{
		Notification: &pb.Notification{
			Name:         "test-notification",
			Recipient:    "test@example.com",
			SourceSystem: "test-system",
			Content:      "notification content",
			Interval:     &durationpb.Duration{Seconds: sevenDaysInSeconds},
		},
	}
	if _, err := nag.CreateNotification(ctx, req); err == nil {
		t.Fatalf("CreateNotification(ctx, %+v) wanted error, got nil", req)
	} else {
		s, ok := status.FromError(err)
		if !ok {
			t.Fatalf("want status error")
		}
		if s.Code() != codes.FailedPrecondition {
			t.Errorf("CreateNotification(ctx, %+v) status got %q, want %q", req, s.Code(), codes.FailedPrecondition)
		}
	}
}

func TestNotifyAll(t *testing.T) {
	ctx := context.Background()
	storage := memstorage.New()
	fes := &fakeEmailSender{}
	nag := newServerForTesting(storage, fes)
	testRecipient := "test@example.com"
	exceptionSourceSystem := "source-sytem-with-exception"

	req := &pb.CreateNotificationRequest{
		Notification: &pb.Notification{
			Name:         "notification-to-send",
			Recipient:    testRecipient,
			SourceSystem: "test",
			Content:      "notification content",
			Interval:     &durationpb.Duration{Seconds: sevenDaysInSeconds},
		},
	}
	_, err := nag.CreateNotification(ctx, req)
	if err != nil {
		t.Fatalf("NotifyUser(ctx, %+v) unexpected error: %v", req, err)
	}

	req = &pb.CreateNotificationRequest{
		Notification: &pb.Notification{
			Name:         "notification-with-exception",
			Recipient:    testRecipient,
			SourceSystem: exceptionSourceSystem,
			Content:      "should not be in content",
			Interval:     &durationpb.Duration{Seconds: sevenDaysInSeconds},
		},
	}
	_, err = nag.CreateNotification(ctx, req)
	if err != nil {
		t.Fatalf("NotifyUser(ctx, %+v) unexpected error: %v", req, err)
	}

	want := &pb.Exception{
		NotificationName: "notification-with-exception",
		UserEmail:        testRecipient,
		SourceSystem:     exceptionSourceSystem,
		Justification:    "test-exception",
		CreatedOnTime:    timestamppb.New(time.Time{}),
		ValidUntilTime:   timestamppb.New(time.Time{}),
	}
	_, err = nag.CreateException(ctx, &pb.CreateExceptionRequest{Exception: want})
	if err != nil {
		t.Fatalf("CreateException(ctx, %+v) unexpected error: %v", want, err)
	}

	notAllReq := &pb.NotifyAllRequest{}
	gotNotAll, err := nag.NotifyAll(ctx, notAllReq)
	if err != nil {
		t.Fatalf("NotifyAll(ctx, %+v) unexpected error: %v", notAllReq, err)
	}
	wantNotAll := &pb.NotifyAllResponse{HasCompleted: true}
	if diff := cmp.Diff(wantNotAll, gotNotAll, protocmp.Transform()); diff != "" {
		t.Errorf("NofifyAll(ctx, %v) unexpected diff (-want, +got): %v", notAllReq, diff)
	}

	content, err := sendgridsender.Template(emailSubject, []sendgridsender.EmailNotificationsContent{
		{
			Topic:   "notification-to-send",
			Message: "notification content",
		},
	})
	if err != nil {
		t.Fatalf("templating: %v", err)
	}
	wantEmails := [][]string{
		{
			"",
			"test@example.com",
			emailSubject,
			content,
		},
	}
	if diff := cmp.Diff(wantEmails, fes.Calls); diff != "" {
		t.Errorf("Emails sent unexpected diff (-want, +got): %v", diff)
	}
}

type fakeEmailSender struct {
	Called int64
	Calls  [][]string
}

func (es *fakeEmailSender) SendEmail(ctx context.Context, sender, object, content string, recipients []string) error {
	for _, r := range recipients {
		es.Called++
		es.Calls = append(es.Calls, []string{sender, r, object, content})
	}
	return nil
}
