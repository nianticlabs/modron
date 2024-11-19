package gcpcollector

import (
	"context"
	"flag"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/risk"
	"github.com/nianticlabs/modron/src/storage/memstorage"
)

var (
	collectorTestProjectID string
	projectListFile        string
)

func init() {
	flag.StringVar(&collectorTestProjectID, "projectId", testProjectID, "GCP project Id")
	flag.StringVar(&projectListFile, "projectIdList", "resourceGroupList.txt", "GCP project Id list")
}

const (
	testProjectID = "projects/modron-test"
	collectID     = "collectID-1"
)

func TestResourceGroupResources(t *testing.T) {
	ctx := context.Background()
	storage := memstorage.New()
	gcpCollector := NewFake(ctx, storage, risk.TagConfig{})

	resourceGroup, err := gcpCollector.GetResourceGroupWithIamPolicy(ctx, collectID, testProjectID)
	if err != nil {
		t.Fatalf("No resourceGroup found: %v", err)
	}
	if resourceGroup.CollectionUid != collectID {
		t.Errorf("wrong collectUid, want %v got %v", collectID, resourceGroup.CollectionUid)
	}

	resourcesCollected, errArr := gcpCollector.ListResourceGroupResources(ctx, collectID, resourceGroup.Name)
	for _, err := range errArr {
		t.Errorf("%v", err)
	}

	for _, r := range resourcesCollected {
		if r.CollectionUid != collectID {
			t.Errorf("wrong collectUid, want %v got %v", collectID, r.CollectionUid)
		}
	}

	wantResourcesCollected := 28 // TODO: Create a better test for this functionality
	if len(resourcesCollected) != wantResourcesCollected {
		t.Errorf("resources collected: got %d, want %d", len(resourcesCollected), wantResourcesCollected)
	}
}

func TestResourceGroup(t *testing.T) {
	ctx := context.Background()
	storage := memstorage.New()
	gcpCollector := NewFake(ctx, storage, risk.TagConfig{})
	resourceGroup, err := gcpCollector.GetResourceGroupWithIamPolicy(ctx, collectID, testProjectID)
	if err != nil {
		t.Fatalf("No resourceGroup found: %v", err)
	}

	if resourceGroup.Name != testProjectID {
		t.Errorf("wrong resourceGroup Name: %v", resourceGroup.Name)
	}
	if len(resourceGroup.IamPolicy.Permissions) != 6 {
		t.Errorf("iam policy count: got %d, want %d", len(resourceGroup.IamPolicy.Permissions), 5)
	}
	if resourceGroup.CollectionUid != collectID {
		t.Errorf("wrong collectUid, want %v got %v", collectID, resourceGroup.CollectionUid)
	}
}

func modronTestResource(name string) *pb.Resource {
	return &pb.Resource{
		Name:              name,
		Parent:            "projects/modron-test",
		ResourceGroupName: "projects/modron-test",
	}
}

func TestCollectAndStore(t *testing.T) {
	logrus.StandardLogger().SetLevel(logrus.DebugLevel)
	logrus.StandardLogger().SetFormatter(&logrus.TextFormatter{ForceColors: true})
	ctx := context.Background()
	storage := memstorage.New()
	gcpCollector := NewFake(ctx, storage, risk.TagConfig{})
	limitFilter := model.StorageFilter{
		Limit: 100,
	}

	for _, testResourceID := range []string{"organizations/1111", testProjectID} {
		err := gcpCollector.collectAndStoreAllInRg(ctx, collectID, testResourceID, nil)
		if err != nil {
			t.Errorf("collectAndStoreResources(ctx, %s, %s): %v", collectID, testResourceID, err)
		}
	}

	if err := storage.FlushOpsLog(ctx); err != nil {
		t.Errorf("flush ops log: %v", err)
	}

	got, err := storage.ListResources(ctx, limitFilter)
	if err != nil {
		t.Errorf("error storing resources: %v", err)
	}

	for _, r := range got {
		if r.CollectionUid != collectID {
			t.Errorf("wrong collectUid, want %v got %v", collectID, r.CollectionUid)
		}
	}

	want := []*pb.Resource{
		{
			Name:              "//container.googleapis.com/projects/modron-test/locations/us-central1/clusters/modron-test-cluster/k8s/namespaces/modron-test-namespace",
			Parent:            "//container.googleapis.com/projects/modron-test/locations/us-central1/clusters/modron-test-cluster",
			ResourceGroupName: "projects/modron-test",
		},
		{
			Name:              "//container.googleapis.com/projects/modron-test/locations/us-central1/clusters/modron-test-cluster/k8s/namespaces/modron-test-namespace-other",
			Parent:            "//container.googleapis.com/projects/modron-test/locations/us-central1/clusters/modron-test-cluster",
			ResourceGroupName: "projects/modron-test",
		},
		modronTestResource("api-key-unrestricted-0"),
		modronTestResource("api-key-unrestricted-1"),
		modronTestResource("api-key-with-overbroad-scope-1"),
		modronTestResource("api-key-without-overbroad-scope"),
		modronTestResource("backend-svc-external-modern"),
		modronTestResource("backend-svc-external-no-modern"),
		modronTestResource("backend-svc-iap"),
		modronTestResource("backend-svc-internal"),
		modronTestResource("backend-svc-no-iap"),
		{
			Name:              "bucket-2",
			Parent:            "projects/modron-test",
			ResourceGroupName: "projects/modron-test",
			IamPolicy: &pb.IamPolicy{
				Permissions: []*pb.Permission{
					{
						Role: "storage.objectViewer",
						Principals: []string{
							"serviceAccount:account-1@modron-test.iam.gserviceaccount.com",
						},
					},
					{
						Role: "storage.objectViewer",
						Principals: []string{
							"serviceAccount:account-2@modron-test.iam.gserviceaccount.com",
						},
					},
				},
			},
		},
		{
			Name:              "bucket-accessible-from-other-project",
			Parent:            "projects/modron-test",
			ResourceGroupName: "projects/modron-test",
			IamPolicy: &pb.IamPolicy{
				Permissions: []*pb.Permission{
					{
						Role: "storage.legacyBucketOwner",
						Principals: []string{
							"serviceAccount:account-3@modron-other-test.iam.gserviceaccount.com",
						},
					},
				},
			},
		},
		{
			Name:              "bucket-public",
			Parent:            "projects/modron-test",
			ResourceGroupName: "projects/modron-test",
			IamPolicy: &pb.IamPolicy{
				Permissions: []*pb.Permission{
					{
						Role: "storage.objectViewer",
						Principals: []string{
							"allAuthenticatedUsers",
						},
					},
				},
			},
		},
		{
			Name:              "bucket-public-allusers",
			Parent:            "projects/modron-test",
			ResourceGroupName: "projects/modron-test",
			IamPolicy: &pb.IamPolicy{
				Permissions: []*pb.Permission{
					{
						Role: "storage.objectViewer",
						Principals: []string{
							"allUsers",
						},
					},
				},
			},
		},
		modronTestResource("cloudsql-report-not-enforcing-tls"),
		modronTestResource("cloudsql-test-db-ok"),
		modronTestResource("cloudsql-test-db-public-and-authorized-networks"),
		modronTestResource("cloudsql-test-db-public-and-no-authorized-networks"),
		// Groups belong to another parent resource - they shouldn't show up here as they have not been scanned
		// as part of the collection phase for the specified resourcegroups (they're one level up)
		modronTestResource("instance-1"),
		{
			Name:              "modron-pod-test-name-1",
			ResourceGroupName: "projects/modron-test",
			Parent:            "//container.googleapis.com/projects/modron-test/locations/us-central1/clusters/modron-test-cluster/k8s/namespaces/modron-test-namespace",
		},
		{
			Name:              "modron-pod-test-name-2",
			ResourceGroupName: "projects/modron-test",
			Parent:            "//container.googleapis.com/projects/modron-test/locations/us-central1/clusters/modron-test-cluster/k8s/namespaces/modron-test-namespace",
		},
		{
			Name:              "organizations/1111",
			ResourceGroupName: "organizations/1111",
			Link:              "https://console.cloud.google.com/welcome?organizationId=1111",
			IamPolicy: &pb.IamPolicy{
				Resource: nil,
				Permissions: []*pb.Permission{
					{
						Role: "owner",
						Principals: []string{
							"user:account-1@example.com",
							"user:account-2@example.com",
						},
					},
					{
						Role: "test2",
						Principals: []string{
							"account-2@example.com",
						},
					},
					{
						Role: "iam.serviceAccountAdmin",
						Principals: []string{
							"account-1@example.com",
						},
					},
					{
						Role: "dataflow.admin",
						Principals: []string{
							"account-1@example.com",
						},
					},
					{
						Role: "viewer",
						Principals: []string{
							"account-2@example.com",
						},
					},
				},
			},
		},
		{
			Name:              "projects/modron-test",
			ResourceGroupName: "projects/modron-test",
			Link:              "https://console.cloud.google.com/welcome?project=modron-test",
			Parent:            "folders/234",
			Ancestors: []string{
				"folders/234", "folders/123", "organizations/1111",
			},
			IamPolicy: &pb.IamPolicy{
				Permissions: []*pb.Permission{
					{
						Role:       "owner",
						Principals: []string{"user:owner1@example.com", "user:owner2@example.com"},
					},
					{
						Role:       "test2",
						Principals: []string{"serviceAccount:account-2@modron-test.iam.gserviceaccount.com"},
					},
					/*
					   `{role:"iam.serviceAccountAdmin", principals:["account-1@modron-test.iam.gserviceaccount.com"]}`,
					   `{role:"dataflow.admin", principals:["account-3@modron-other-test.iam.gserviceaccount.com"]}`,
					   `{role:"iam.serviceAccountAdmin", principals:["account-3@modron-other-test.iam.gserviceaccount.com"]}`,
					   `{role:"viewer", principals:["account-2@modron-test.iam.gserviceaccount.com"]}`,
					*/
					{
						Role: "iam.serviceAccountAdmin",
						Principals: []string{
							"serviceAccount:account-1@modron-test.iam.gserviceaccount.com",
						},
					},
					{
						Role: "dataflow.admin",
						Principals: []string{
							"serviceAccount:account-3@modron-other-test.iam.gserviceaccount.com",
						},
					},
					{
						Role: "iam.serviceAccountAdmin",
						Principals: []string{
							"serviceAccount:account-3@modron-other-test.iam.gserviceaccount.com",
						},
					},
					{
						Role: "viewer",
						Principals: []string{
							"serviceAccount:account-2@modron-test.iam.gserviceaccount.com",
						},
					},
				},
			},
			Labels: map[string]string{
				"contact1": "user-1_example_com",
				"contact2": "user-2_example_com",
			},
		},
		modronTestResource("psc-network-should-not-be-reported"),
		modronTestResource("spanner-test-db-1"),
		modronTestResource("subnetwork-no-private-access-should-be-reported"),
		modronTestResource("subnetwork-private-access-should-not-be-reported"),
		{
			Name:              "user:account-1@modron-test",
			Parent:            "projects/modron-test",
			ResourceGroupName: "projects/modron-test",
			IamPolicy: &pb.IamPolicy{
				Permissions: []*pb.Permission{
					{
						Role:       "iam.serviceAccountUser",
						Principals: []string{"user:user-1@example.com"},
					},
				},
			},
		},
		{
			Name:              "user:account-2@modron-test",
			Parent:            "projects/modron-test",
			ResourceGroupName: "projects/modron-test",
			IamPolicy: &pb.IamPolicy{
				Permissions: []*pb.Permission{
					{
						Role:       "iam.serviceAccountUser",
						Principals: []string{"user:user-1@example.com"},
					},
				},
			},
		},
	}

	if diff := cmp.Diff(want, got, protocmp.Transform(),
		protocmp.IgnoreOneofs(&pb.Resource{}, "type"),
		protocmp.IgnoreFields(&pb.Resource{}, "uid", "collection_uid", "timestamp"),
	); diff != "" {
		t.Errorf("resources collected: -want +got\n%s", diff)
	}
}

func msgIsTimestamp(x reflect.Value) bool {
	if !x.IsValid() || x.IsZero() || x.IsNil() {
		return false
	}
	return x.Interface().(protocmp.Message).Descriptor().FullName() == "google.protobuf.Timestamp"
}

func TestCollectAndStoreObservations(t *testing.T) {
	logrus.StandardLogger().SetLevel(logrus.DebugLevel)
	logrus.StandardLogger().SetFormatter(&logrus.TextFormatter{ForceColors: true})
	ctx := context.Background()
	storage := memstorage.New()
	gcpCollector := NewFake(ctx, storage, risk.TagConfig{})
	collectID := uuid.NewString()

	if err := gcpCollector.CollectAndStoreAll(ctx, collectID, []string{testProjectID}, nil); err != nil {
		t.Fatalf("collectAndStoreObservations: %v", err)
	}
	if err := storage.AddOperationLog(ctx, []*pb.Operation{
		{
			Id:            collectID,
			ResourceGroup: testProjectID,
			Type:          "scan",
			Status:        pb.Operation_COMPLETED,
		},
	}); err != nil {
		t.Fatalf("add operation log: %v", err)
	}
	if err := storage.FlushOpsLog(ctx); err != nil {
		t.Errorf("flush ops log: %v", err)
	}
	got, err := storage.ListObservations(ctx, model.StorageFilter{Limit: 100})
	if err != nil {
		t.Errorf("error storing observations: %v", err)
	}

	want := []*pb.Observation{
		{
			Name:         "SQL_PUBLIC_IP",
			CollectionId: proto.String(collectID),
			Timestamp:    timestamppb.Now(),
			Remediation: &pb.Remediation{
				Description:    "To lower your attack surface, Cloud SQL databases should not have public IPs. Private IPs provide improved network security and lower latency for your application.",
				Recommendation: "Go to https://console.cloud.google.com/sql/instances/xyz/connections?project=project-id and click the \"Networking\" tab. Uncheck the \"Public IP\" checkbox and click \"SAVE\". If your instance is not configured to use a private IP, you will first have to enable private IP by following the instructions here: https://cloud.google.com/sql/docs/mysql/configure-private-ip#existing-private-instance",
			},
			ResourceRef: &pb.ResourceRef{
				CloudPlatform: pb.CloudPlatform_GCP,
				ExternalId:    proto.String("//cloudsql.googleapis.com/projects/project-id/instances/xyz"),
				GroupName:     testProjectID,
			},
			ExternalId: proto.String("//securitycenter.googleapis.com/projects/12345/sources/123/findings/48230f1978594ffb9d09a3cb1fe5e1b3"),
			Source:     pb.Observation_SOURCE_SCC,
			Severity:   pb.Severity_SEVERITY_MEDIUM,

			// We have no information about the folders here, so the impact is MEDIUM and the risk score is equal
			// to the severity.
			Impact:    pb.Impact_IMPACT_MEDIUM,
			RiskScore: pb.Severity_SEVERITY_MEDIUM,
			Category:  pb.Observation_CATEGORY_MISCONFIGURATION,
		},
	}
	if diff := cmp.Diff(want, got, protocmp.Transform(),
		protocmp.IgnoreFields(&pb.Observation{}, "uid"),
		cmpopts.EquateApproxTime(10*time.Second),
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
	); diff != "" {
		t.Errorf("observations collected: -want +got\n%s", diff)
	}
}

func TestResourceGroupRegex(t *testing.T) {
	validRGNames := []string{
		"projects/google.com:xyz",
		"organizations/111111111111",
		"folders/111111111111",
		"folders/11111111111",
		"projects/hello-world",
	}
	invalidRGNames := []string{
		"projects/!",
		"organizations/example",
		"folders/test",
	}
	for _, v := range validRGNames {
		t.Run(v, func(t *testing.T) {
			if !validGcpResourceGroupRegex.MatchString(v) {
				t.Errorf("expected %s to be valid", v)
			}
		})
	}

	for _, v := range invalidRGNames {
		t.Run(v, func(t *testing.T) {
			if validGcpResourceGroupRegex.MatchString(v) {
				t.Errorf("expected %s to be invalid", v)
			}
		})
	}
}
