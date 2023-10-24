package e2e

import "github.com/nianticlabs/modron/src/pb"

func New() pb.NotificationServiceServer {
	return newFakeServer()
}

type FakeNotificationService struct {
	// Required
	pb.UnimplementedNotificationServiceServer
}

func newFakeServer() pb.NotificationServiceServer {
	return FakeNotificationService{}
}
