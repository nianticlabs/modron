package e2e

import pb "github.com/nianticlabs/modron/src/proto/generated"

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
