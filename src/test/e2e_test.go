package e2e

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/pb"
)

func init() {
	flag.StringVar(&projectListFile, "projectIdList", "resourceGroupList.txt", "GCP project Id list")
}

const (
	notificationPortEnvVar = "FAKE_NOTIFICATION_SERVICE_PORT"
	serverAddrEnvVar       = "BACKEND_ADDRESS"
	fakeServerAddrEnvVar   = "FAKE_BACKEND_ADDRESS"
)

var (
	serverAddr     = os.Getenv(serverAddrEnvVar)
	fakeServerAddr = os.Getenv(fakeServerAddrEnvVar)
)

var projectListFile string

func runFakeNotificationService(t *testing.T, port int64) {
	t.Helper()
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		t.Fatalf("cannot listen on port %d: %v", port, err)
	}
	srvCert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		panic(fmt.Sprintln("load certificate: ", err))
	}
	grpcServer := grpc.NewServer(grpc.Creds(credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{srvCert},
		ClientAuth:   tls.NoClientCert,
	})))
	svc := New()
	pb.RegisterNotificationServiceServer(grpcServer, svc)
	fmt.Printf("starting fake notification service on port %d\n", port)
	if err := grpcServer.Serve(lis); err != nil {
		t.Fatalf("grpcServer serve: %v", err)
	}
}

func testModronE2e(t *testing.T, addr string, resourceGroupNames []string, want map[string][]*structpb.Value) {
	flag.Parse()
	ctx := context.Background()
	go func() {
		runFakeNotificationService(t, extractNotificationServicePortFromEnvironment(t))
	}()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		t.Fatalf("fail to dial %q: %v", addr, err)
	}
	defer conn.Close()
	client := pb.NewModronServiceClient(conn)

	doCollectAndScan(ctx, t, client, resourceGroupNames)
	listObsResponse, err := client.ListObservations(ctx, &pb.ListObservationsRequest{
		ResourceGroupNames: resourceGroupNames,
	})
	if err != nil {
		if s, ok := status.FromError(err); !ok {
			t.Fatalf("invalid error: %v", err)
		} else {
			t.Fatalf("ListObs unexpected error with status code %s: %s", s.Code(), s.Message())
		}
	}

	allObs := []*pb.Observation{}
	for _, el := range listObsResponse.ResourceGroupsObservations {
		for _, rules := range el.RulesObservations {
			allObs = append(allObs, rules.Observations...)
		}
	}
	if len(allObs) < 1 {
		t.Fatalf("no observations returned")
	}
	got := map[string][]*structpb.Value{}
	for _, ob := range allObs {
		if _, ok := got[ob.Resource.Name]; !ok {
			got[ob.Resource.Name] = []*structpb.Value{}
		}
		temp := got[ob.Resource.Name]
		got[ob.Resource.Name] = append(temp, ob.ExpectedValue)
	}
	for k, v := range got {
		if diff := cmp.Diff(want[k], v, protocmp.Transform()); diff != "" {
			t.Errorf("Resource %v has unexpected observations (-want, +got): %v", k, diff)
		}
	}

	// TODO extract this to its own test
	manualObservation := &pb.Observation{
		Resource:      allObs[0].GetResource(),
		Name:          "test-observation",
		ObservedValue: structpb.NewStringValue("test observation"),
		Remediation: &pb.Remediation{
			Description:    "test observation",
			Recommendation: "test observation, no recommendation",
		},
	}
	cmpOpts := cmp.Options{
		protocmp.Transform(),
		cmpopts.EquateApproxTime(time.Second),
		// Approximate comparison of timestamppb timestamps.
		// https://github.com/golang/protobuf/issues/1347
		cmp.FilterPath(
			func(p cmp.Path) bool {
				if p.Last().Type() == reflect.TypeOf(protocmp.Message{}) {
					a, b := p.Last().Values()
					return msgIsTimestamp(a) && msgIsTimestamp(b)
				}
				return false
			},
			cmp.Transformer("timestamppb", func(t protocmp.Message) time.Time {
				if t["seconds"] == nil {
					return time.Time{}
				}
				return time.Unix(t["seconds"].(int64), 0).UTC()
			}),
		),
	}
	manualObservation.Timestamp = timestamppb.Now()
	gotManualObs, err := client.CreateObservation(ctx, &pb.CreateObservationRequest{Observation: manualObservation})
	if err != nil {
		t.Errorf("CreateObservation(ctx, %+v) unexpected error %v", manualObservation, err)
	} else {
		if diff := cmp.Diff(manualObservation, gotManualObs, cmpOpts); diff != "" {
			t.Errorf("CreateObservation(ctx, %+v) diff (-want, +got): %v", manualObservation, diff)
		}
	}

	manualObservation.Resource = &pb.Resource{Name: "non existing"}
	_, err = client.CreateObservation(ctx, &pb.CreateObservationRequest{Observation: manualObservation})
	if err == nil {
		t.Errorf("CreateObservation(ctx, %+v) wanted error, got nil", manualObservation)
	}
	if s, ok := status.FromError(err); !ok {
		t.Fatalf("invalid error: %v", err)
	} else {
		if s.Code() != codes.FailedPrecondition {
			t.Errorf("CreateObservation(ctx, %+v) unexpected error code got %s, want %s", manualObservation, s.Code(), s.Message())
		}
	}
}

func msgIsTimestamp(x reflect.Value) bool {
	if !x.IsValid() || x.IsZero() || x.IsNil() {
		return false
	}
	return x.Interface().(protocmp.Message).Descriptor().FullName() == "google.protobuf.Timestamp"
}

func doCollectAndScan(ctx context.Context, t *testing.T, client pb.ModronServiceClient, resourceGroupNames []string) {
	res, err := client.CollectAndScan(ctx,
		&pb.CollectAndScanRequest{
			ResourceGroupNames: resourceGroupNames,
		},
	)
	if err != nil {
		if s, ok := status.FromError(err); !ok {
			t.Fatalf("invalid error: %v", err)
		} else if s.Code() == codes.Unavailable {
			t.Fatalf("backend is not available: %s, %s", s.Code(), s.Message())
		} else {
			t.Fatalf("CollectAndScan(ctx, req): %s message: %s", s.Code(), s.Message())
		}
	}

	for {
		time.Sleep(time.Second)
		resp, err := client.GetStatusCollectAndScan(ctx, &pb.GetStatusCollectAndScanRequest{
			CollectId: res.CollectId,
			ScanId:    res.ScanId,
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		fmt.Printf("status: Collect: %s, Scan: %s\n", resp.CollectStatus, resp.ScanStatus)
		if resp.CollectStatus == pb.RequestStatus_DONE && resp.ScanStatus == pb.RequestStatus_DONE {
			break
		}
	}
}

func TestModronE2e(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping test TestModronE2e against address %v", serverAddr)
	}

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/app/secrets/key.json")
	defer os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")

	content, err := os.ReadFile(projectListFile)
	if err != nil {
		t.Errorf("error with projectID list file: %v", err)
	}
	projectIDs := strings.Split(string(content), "\n")

	// TODO: Fill expected values for real test.
	want := map[string][]*structpb.Value{}
	testModronE2e(t, serverAddr, projectIDs, want)
}

func TestModronE2eFake(t *testing.T) {
	want := map[string][]*structpb.Value{
		"account-1":                                          {structpb.NewStringValue(""), structpb.NewStringValue("")},
		"bucket-public":                                      {structpb.NewStringValue("PRIVATE")},
		"bucket-public-allusers":                             {structpb.NewStringValue("PRIVATE")},
		"bucket-accessible-from-other-project":               {structpb.NewStringValue("")},
		"api-key-unrestricted-0[]":                           {structpb.NewStringValue("restricted")},
		"api-key-unrestricted-1[]":                           {structpb.NewStringValue("restricted")},
		"api-key-with-overbroad-scope-1[]":                   {structpb.NewStringValue(""), structpb.NewStringValue("")},
		"backend-svc-2[0]":                                   {structpb.NewNumberValue(float64(pb.Certificate_MANAGED))},
		"backend-svc-3[0]":                                   {structpb.NewNumberValue(float64(pb.Certificate_MANAGED))},
		"backend-svc-5[0]":                                   {structpb.NewStringValue("TLS 1.2")},
		"subnetwork-no-private-access-should-be-reported[0]": {structpb.NewStringValue("enabled")},
		"cloudsql-report-not-enforcing-tls":                  {structpb.NewBoolValue(true)},
		"cloudsql-test-db-public-and-no-authorized-networks": {structpb.NewStringValue("AUTHORIZED_NETWORKS_SET")},
		"instance-1[0]":                                      {structpb.NewStringValue("empty")},
		"projects/modron-test":                               {structpb.NewStringValue(""), structpb.NewStringValue("")},
	}
	testModronE2e(t, fakeServerAddr, []string{"projects/modron-test"}, want)
}

func extractNotificationServicePortFromEnvironment(t *testing.T) int64 {
	t.Helper()
	p, err := strconv.ParseInt(os.Getenv(notificationPortEnvVar), 10, 64)
	if err != nil {
		t.Fatalf("parse %s as int: %v", os.Getenv(notificationPortEnvVar), err)
	}
	return p
}
