package nagatha

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/golang/glog"
	"google.golang.org/api/idtoken"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpcMetadata "google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/durationpb"
	"github.com/nianticlabs/modron/src/model"
)

const (
	sourceSystem = "modron"
)

func New(ctx context.Context, addr string) (model.NotificationService, error) {
	var opts []grpc.DialOption
	cp, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("cert pool: %v", err)
	}
	opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		RootCAs:            cp,
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS13,
	})))
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("dialing %s: %v", addr, err)
	}
	return &Service{
		client: NewNagathaClient(conn),
	}, nil
}

type Service struct {
	model.NotificationService
	client NagathaClient
}

func (svc *Service) CreateNotification(ctx context.Context, notification model.Notification) (model.Notification, error) {
	if notification.Name == "" {
		return model.Notification{}, fmt.Errorf("name can't be empty")
	}
	_, err := svc.client.CreateNotification(addAuthenticationToCtx(ctx), &CreateNotificationRequest{
		Notification: &Notification{
			SourceSystem: sourceSystem,
			Name:         notification.Name,
			Recipient:    notification.Recipient,
			Content:      notification.Content,
			Interval:     durationpb.New(notification.Interval),
		},
	})
	if err != nil {
		return model.Notification{}, err
	}
	return model.Notification{}, nil
}

func (svc *Service) GetException(ctx context.Context, uuid string) (model.Exception, error) {
	ex, err := svc.client.GetException(addAuthenticationToCtx(ctx), &GetExceptionRequest{Uuid: uuid})
	if err != nil {
		return model.Exception{}, err
	}
	return exceptionModelFromNagathaProto(ex), nil
}

func (svc *Service) CreateException(ctx context.Context, exception model.Exception) (model.Exception, error) {
	ex, err := svc.client.CreateException(addAuthenticationToCtx(ctx), &CreateExceptionRequest{Exception: exceptionNagathaProtoFromModel(exception)})
	if err != nil {
		return model.Exception{}, err
	}
	return exceptionModelFromNagathaProto(ex), nil
}

func (svc *Service) UpdateException(ctx context.Context, exception model.Exception) (model.Exception, error) {
	ex, err := svc.client.UpdateException(addAuthenticationToCtx(ctx), &UpdateExceptionRequest{
		Exception: exceptionNagathaProtoFromModel(exception),
	})
	return exceptionModelFromNagathaProto(ex), err
}

func (svc *Service) DeleteException(ctx context.Context, id string) error {
	_, err := svc.client.DeleteException(addAuthenticationToCtx(ctx), &DeleteExceptionRequest{Uuid: id})
	return err
}

func (svc *Service) ListExceptions(ctx context.Context, userEmail string, pageSize int32, pageToken string) ([]model.Exception, error) {
	exceptions := make([]model.Exception, 0)
	resp, err := svc.client.ListExceptions(addAuthenticationToCtx(ctx), &ListExceptionsRequest{UserEmail: userEmail, PageSize: pageSize, PageToken: pageToken})
	if err != nil {
		return nil, err
	}
	for _, e := range resp.Exceptions {
		exceptions = append(exceptions, exceptionModelFromNagathaProto(e))
	}
	return exceptions, nil
}

func addAuthenticationToCtx(ctx context.Context) context.Context {
	// Create an identity token.
	// With a global TokenSource tokens would be reused and auto-refreshed at need.
	// A given TokenSource is specific to the audience.
	tokenSource, err := idtoken.NewTokenSource(ctx, "143415353591-bsmr7ii98a2493kts699289n2ommqi07.apps.googleusercontent.com")
	if err != nil {
		glog.Warningf("idtoken.NewTokenSource: %v", err)
	}
	token, err := tokenSource.Token()
	if err != nil {
		glog.Warningf("TokenSource.Token: %v", err)
	}

	// Add token to gRPC Request.
	return grpcMetadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token.AccessToken)
}
