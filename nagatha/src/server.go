package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/nianticlabs/modron/nagatha/src/bqstorage"
	"github.com/nianticlabs/modron/nagatha/src/model"
	"github.com/nianticlabs/modron/nagatha/src/pb"
	"github.com/nianticlabs/modron/nagatha/src/sendgridsender"

	"github.com/golang/glog"
	"google.golang.org/grpc"
)

func newServer(ctx context.Context) (*nagathaService, error) {
	bq, err := bqstorage.New(ctx, bqstorage.Config{ProjectID: projectID, NotificationTableID: notificationTableID, ExceptionTableID: exceptionTableID})
	if err != nil {
		return nil, err
	}
	return newServerForTesting(
		bq,
		sendgridsender.NewSendGridSender(sendgridsender.Config{ApiKey: sendgridApiKey, TestRecipient: os.Getenv(testRecipientEnvVar)}),
	), nil
}

func newServerForTesting(storage model.Storage, emailSender model.EmailSender) *nagathaService {
	return &nagathaService{
		storage:     storage,
		emailSender: emailSender,
	}
}

func validateEnvironment() {
	hasError := false
	var err error
	portStr := os.Getenv(portEnvVar)
	parsedPort, err := strconv.ParseInt(portStr, 10, 32)
	if err != nil {
		glog.Errorf("%s contains an invalid port number %s: %v", portEnvVar, portStr, err)
		hasError = true
	}
	port = int32(parsedPort)
	if projectID == "" {
		glog.Errorf("%s is empty", gcpProjectIDEnvVar)
		hasError = true
	}
	if exceptionTableID == "" {
		glog.Errorf("%s is empty", exceptionTableIDEnvVar)
		hasError = true
	}
	if notificationTableID == "" {
		glog.Errorf("%s is empty", notificationTableIDEnvVar)
		hasError = true
	}
	if emailSenderAddress == "" {
		glog.Errorf("%q must be set, got empty", emailSenderAddressEnvVar)
		hasError = true
	}
	if hasError {
		glog.Errorf("Fix all errors and restart. Stopping.")
		os.Exit(1)
	}
}

func main() {
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	validateEnvironment()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		glog.Errorf("failed to listen: %v", err)
		os.Exit(1)
	}
	// Handle SIGINT (for Ctrl+C) and SIGTERM (for Cloud Run) signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-c
		glog.Infof("received signal: %+v", sig)
		cancel()
	}()
	go func() {
		var opts []grpc.ServerOption
		grpcServer := grpc.NewServer(opts...)
		srv, err := newServer(ctx)
		if err != nil {
			glog.Errorf("server: %v", err)
			os.Exit(3)
		}
		pb.RegisterNagathaServer(grpcServer, srv)
		glog.Infof("starting nagatha on port %d", port)
		if err := grpcServer.Serve(lis); err != nil {
			glog.Errorf("error while listening: %v", err)
			os.Exit(2)
		}
	}()
	<-ctx.Done()
	glog.Infof("server stopped")
}
