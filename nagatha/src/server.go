package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/nianticlabs/modron/nagatha/src/bqstorage"
	"github.com/nianticlabs/modron/nagatha/src/model"
	"github.com/nianticlabs/modron/nagatha/src/pb"
	"github.com/nianticlabs/modron/nagatha/src/sendgridsender"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	grpcPort   = 64232
	portEnvVar = "PORT"
)

var (
	port int32
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
		glog.Errorf("%s must be set, got empty", emailSenderAddressEnvVar)
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
	// Handle SIGINT (for Ctrl+C) and SIGTERM (for Cloud Run) signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-c
		glog.Infof("received signal: %+v", sig)
		cancel()
	}()
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
		if err != nil {
			glog.Errorf("failed to listen: %v", err)
			os.Exit(1)
		}
		var opts []grpc.ServerOption
		grpcServer := grpc.NewServer(opts...) // nosemgrep: go.grpc.security.grpc-server-insecure-connection.grpc-server-insecure-connection
		glog.Infof("gRPC server starting on port %d", grpcPort)
		srv, err := newServer(ctx)
		if err != nil {
			glog.Errorf("server creation: %v", err)
			os.Exit(3)
		}
		go func() {
			err := srv.NotificationTriggerListener(ctx)
			if err != nil {
				glog.Errorf("notification trigger: %v", err)
			}
		}()
		pb.RegisterNagathaServer(grpcServer, srv)
		if err := grpcServer.Serve(lis); err != nil {
			glog.Errorf("error while listening: %v", err)
			os.Exit(3)
		}
	}()
	go func() {
		mux := runtime.NewServeMux()
		localGrpcService := fmt.Sprintf("localhost:%d", grpcPort)
		glog.Infof("waiting for gRPC backend to start on %s", localGrpcService)
		if err := waitForService(localGrpcService); err != nil {
			glog.Errorf("wait for gRPC: %v", err)
			os.Exit(4)
		}

		opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

		err := pb.RegisterNagathaHandlerFromEndpoint(ctx, mux, localGrpcService, opts)
		if err != nil {
			glog.Errorf("failed to start HTTP gateway: %v", err)
			os.Exit(5)
		}
		localHttpService := fmt.Sprintf(":%d", port)
		glog.Infof("Starting gRPC-Gateway on http://0.0.0.0%s", localHttpService)
		glog.Fatal(http.ListenAndServe(localHttpService, mux))
	}()
	<-ctx.Done()
	glog.Infof("server stopped")
}

func waitForService(addr string) (err error) {
	timeout := 3 * time.Second
	for i := 0; i < 10; i++ {
		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err != nil {
			continue
		}
		if conn != nil {
			break
		}
		time.Sleep(time.Second)
	}
	return nil
}
