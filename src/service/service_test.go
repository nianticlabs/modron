package service_test

import (
	"context"
	"encoding/json"
	"regexp"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/nianticlabs/modron/src/acl/fakeacl"
	"github.com/nianticlabs/modron/src/collector/gcpcollector"
	"github.com/nianticlabs/modron/src/engine"
	"github.com/nianticlabs/modron/src/engine/rules"
	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/risk"
	"github.com/nianticlabs/modron/src/service"
	"github.com/nianticlabs/modron/src/statemanager/reqdepstatemanager"
	"github.com/nianticlabs/modron/src/storage/memstorage"
)

const (
	bucketPublicRemediationDesc  = "Bucket [\"bucket-public\"](https://console.cloud.google.com/storage/browser/bucket-public) is publicly accessible"
	bucketPublicRemediationRecom = "Unless strictly needed, restrict the IAM policy of bucket [\"bucket-public\"](https://console.cloud.google.com/storage/browser/bucket-public) to prevent unconditional access by anyone. For more details, see [here](https://cloud.google.com/storage/docs/using-public-access-prevention)"

	bucketPublicAllUsersRemediationDesc         = "Bucket [\"bucket-public-allusers\"](https://console.cloud.google.com/storage/browser/bucket-public-allusers) is publicly accessible"
	bucketPublicAllUsersRemediationRecom string = "Unless strictly needed, restrict the IAM policy of bucket [\"bucket-public-allusers\"](https://console.cloud.google.com/storage/browser/bucket-public-allusers) to prevent unconditional access by anyone. For more details, see [here](https://cloud.google.com/storage/docs/using-public-access-prevention)"

	sqlRemediationDesc  = "To lower your attack surface, Cloud SQL databases should not have public IPs. Private IPs provide improved network security and lower latency for your application."
	sqlRemediationRecom = "Go to https://console.cloud.google.com/sql/instances/xyz/connections?project=project-id and click the \"Networking\" tab. Uncheck the \"Public IP\" checkbox and click \"SAVE\". If your instance is not configured to use a private IP, you will first have to enable private IP by following the instructions here: https://cloud.google.com/sql/docs/mysql/configure-private-ip#existing-private-instance"
)

var impactMap = map[string]pb.Impact{
	"prod":       pb.Impact_IMPACT_HIGH,
	"pre-prod":   pb.Impact_IMPACT_MEDIUM,
	"dev":        pb.Impact_IMPACT_LOW,
	"playground": pb.Impact_IMPACT_LOW,
}

func getService(ctx context.Context, t *testing.T) (*service.Modron, *mockNotifier) {
	t.Helper()
	st := memstorage.New()
	engineRules := []model.Rule{
		rules.NewBucketIsPublicRule(),
	}
	notifier := newMockNotifier()
	sm, err := reqdepstatemanager.New()
	tagConfig := risk.TagConfig{
		ImpactMap:    impactMap,
		Environment:  "111111111111/environment",
		EmployeeData: "111111111111/employee_data",
		CustomerData: "111111111111/customer_data",
	}
	if err != nil {
		t.Fatalf("reqdepstatemanager.New: %v", err)
	}
	e, err := engine.New(st, engineRules, map[string]json.RawMessage{}, nil, tagConfig)
	if err != nil {
		t.Fatalf("engine.New: %v", err)
	}
	svc, err := service.New(
		fakeacl.New(),
		10*time.Second,
		gcpcollector.NewFake(ctx, st, tagConfig),
		24*time.Hour,
		notifier,
		"example.com",
		e,
		"https://modron.example.com",
		sm,
		st,
		nil,
		regexp.MustCompile("(.*)_(.*?)_(.*?)$"),
		"$1@$2.$3",
	)
	if err != nil {
		t.Fatalf("service.New: %v", err)
	}
	return svc, notifier
}

func TestService_CollectAndScan(t *testing.T) {
	logrus.StandardLogger().SetLevel(logrus.DebugLevel)
	logrus.StandardLogger().SetFormatter(&logrus.TextFormatter{ForceColors: true})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	svc, notifier := getService(ctx, t)
	res, err := svc.CollectAndScan(ctx, &pb.CollectAndScanRequest{
		ResourceGroupNames: []string{"projects/modron-test"},
	})
	if err != nil {
		t.Fatalf("CollectAndScan: %v", err)
	}

	scanID := res.ScanId
	if scanID == "" {
		t.Fatalf("empty scan ID")
	}
	collectID := res.CollectId
	if collectID == "" {
		t.Fatalf("empty collect ID")
	}

	// Wait for the scan to complete
	tick := time.NewTicker(1 * time.Second)
	for {
		done := false
		select {
		case <-ctx.Done():
			t.Fatalf("timeout waiting for scan to complete")
		case <-tick.C:
			status, err := svc.GetStatusCollectAndScan(ctx, &pb.GetStatusCollectAndScanRequest{
				CollectId: collectID,
				ScanId:    scanID,
			})
			if err != nil {
				t.Fatalf("GetStatusCollectAndScan: %v", err)
			}
			if status.CollectStatus == pb.RequestStatus_DONE && status.ScanStatus == pb.RequestStatus_DONE {
				t.Log("Collect and scan completed")
				done = true
				break
			} else {
				t.Logf("collect=%s, scan=%s", status.CollectStatus.String(), status.ScanStatus.String())
			}
		}
		if done {
			break
		}
	}

	// Done
	obs, err := svc.ListObservations(ctx, &pb.ListObservationsRequest{})
	if err != nil {
		t.Fatalf("ListObservations: %v", err)
	}
	got := obs.ResourceGroupsObservations
	want := []*pb.ResourceGroupObservationsPair{
		{
			ResourceGroupName: "projects/modron-test",
			RulesObservations: []*pb.RuleObservationPair{
				{
					Rule: "BUCKET_IS_PUBLIC",
					Observations: []*pb.Observation{
						{
							Name: "BUCKET_IS_PUBLIC",
							Remediation: &pb.Remediation{
								Description:    bucketPublicRemediationDesc,
								Recommendation: bucketPublicRemediationRecom,
							},
							ObservedValue: structpb.NewStringValue("PUBLIC"),
							ExpectedValue: structpb.NewStringValue("PRIVATE"),
							ResourceRef: &pb.ResourceRef{
								GroupName:     "projects/modron-test",
								ExternalId:    proto.String("bucket-public"),
								CloudPlatform: pb.CloudPlatform_GCP,
							},
							Source:       pb.Observation_SOURCE_MODRON,
							Severity:     pb.Severity_SEVERITY_MEDIUM,
							Impact:       pb.Impact_IMPACT_HIGH,
							ImpactReason: "environment=prod",
							RiskScore:    pb.Severity_SEVERITY_HIGH,
							ScanUid:      proto.String(scanID),
						},
						{
							Name: "BUCKET_IS_PUBLIC",
							Remediation: &pb.Remediation{
								Description:    bucketPublicAllUsersRemediationDesc,
								Recommendation: bucketPublicAllUsersRemediationRecom,
							},
							ResourceRef: &pb.ResourceRef{
								GroupName:     "projects/modron-test",
								ExternalId:    proto.String("bucket-public-allusers"),
								CloudPlatform: pb.CloudPlatform_GCP,
							},
							ScanUid:       proto.String(scanID),
							ObservedValue: structpb.NewStringValue("PUBLIC"),
							ExpectedValue: structpb.NewStringValue("PRIVATE"),
							Source:        pb.Observation_SOURCE_MODRON,
							Severity:      pb.Severity_SEVERITY_MEDIUM,
							Impact:        pb.Impact_IMPACT_HIGH,
							ImpactReason:  "environment=prod",
							RiskScore:     pb.Severity_SEVERITY_HIGH,
						},
					},
				},
				{
					Rule: "SQL_PUBLIC_IP",
					Observations: []*pb.Observation{
						{
							Name: "SQL_PUBLIC_IP",
							ResourceRef: &pb.ResourceRef{
								GroupName:     "projects/modron-test",
								ExternalId:    proto.String("//cloudsql.googleapis.com/projects/project-id/instances/xyz"),
								CloudPlatform: pb.CloudPlatform_GCP,
							},
							Remediation: &pb.Remediation{
								Description:    sqlRemediationDesc,
								Recommendation: sqlRemediationRecom,
							},
							CollectionId: proto.String(collectID),
							Category:     pb.Observation_CATEGORY_MISCONFIGURATION,
							ExternalId:   proto.String("//securitycenter.googleapis.com/projects/12345/sources/123/findings/48230f1978594ffb9d09a3cb1fe5e1b3"),
							Source:       pb.Observation_SOURCE_SCC,
							Impact:       pb.Impact_IMPACT_HIGH,
							ImpactReason: "environment=prod",
							Severity:     pb.Severity_SEVERITY_MEDIUM,
							RiskScore:    pb.Severity_SEVERITY_HIGH,
						},
					},
				},
			},
		},
	}

	if diff := cmp.Diff(want, got,
		protocmp.Transform(),
		protocmp.IgnoreFields(&pb.Observation{}, "uid", "timestamp"),
		protocmp.IgnoreFields(&pb.ResourceRef{}, "uid"),
	); diff != "" {
		t.Fatalf("ListObservations: diff (-want +got):\n%s", diff)
	}

	user1 := "user-1@example.com"
	user2 := "user-2@example.com"
	owner1 := "owner1@example.com"
	owner2 := "owner2@example.com"

	notificationContent := func(desc, rec string) string {
		return desc + "\n\n" + rec + "  \n  \n"
	}
	wantNotif := []model.Notification{
		{
			SourceSystem: "modron",
			Name:         "BUCKET_IS_PUBLIC",
			Recipient:    owner1,
			Content:      notificationContent(bucketPublicRemediationDesc, bucketPublicRemediationRecom),
			Interval:     24 * time.Hour,
		},
		{
			SourceSystem: "modron",
			Name:         "BUCKET_IS_PUBLIC",
			Recipient:    owner2,
			Content:      notificationContent(bucketPublicRemediationDesc, bucketPublicRemediationRecom),
			Interval:     24 * time.Hour,
		},
		{
			SourceSystem: "modron",
			Name:         "BUCKET_IS_PUBLIC",
			Recipient:    user1,
			Content:      notificationContent(bucketPublicRemediationDesc, bucketPublicRemediationRecom),
			Interval:     24 * time.Hour,
		},
		{
			SourceSystem: "modron",
			Name:         "BUCKET_IS_PUBLIC",
			Recipient:    user2,
			Content:      notificationContent(bucketPublicRemediationDesc, bucketPublicRemediationRecom),
			Interval:     24 * time.Hour,
		},
		{
			SourceSystem: "modron",
			Name:         "BUCKET_IS_PUBLIC",
			Recipient:    owner1,
			Content:      notificationContent(bucketPublicAllUsersRemediationDesc, bucketPublicAllUsersRemediationRecom),
			Interval:     24 * time.Hour,
		},
		{
			SourceSystem: "modron",
			Name:         "BUCKET_IS_PUBLIC",
			Recipient:    owner2,
			Content:      notificationContent(bucketPublicAllUsersRemediationDesc, bucketPublicAllUsersRemediationRecom),
			Interval:     24 * time.Hour,
		},
		{
			SourceSystem: "modron",
			Name:         "BUCKET_IS_PUBLIC",
			Recipient:    user1,
			Content:      notificationContent(bucketPublicAllUsersRemediationDesc, bucketPublicAllUsersRemediationRecom),
			Interval:     24 * time.Hour,
		},
		{
			SourceSystem: "modron",
			Name:         "BUCKET_IS_PUBLIC",
			Recipient:    user2,
			Content:      notificationContent(bucketPublicAllUsersRemediationDesc, bucketPublicAllUsersRemediationRecom),
			Interval:     24 * time.Hour,
		},
		{
			SourceSystem: "modron",
			Name:         "SQL_PUBLIC_IP",
			Recipient:    owner1,
			Content:      notificationContent(sqlRemediationDesc, sqlRemediationRecom),
			Interval:     24 * time.Hour,
		},
		{
			SourceSystem: "modron",
			Name:         "SQL_PUBLIC_IP",
			Recipient:    owner2,
			Content:      notificationContent(sqlRemediationDesc, sqlRemediationRecom),
			Interval:     24 * time.Hour,
		},
		{
			SourceSystem: "modron",
			Name:         "SQL_PUBLIC_IP",
			Recipient:    user1,
			Content:      notificationContent(sqlRemediationDesc, sqlRemediationRecom),
			Interval:     24 * time.Hour,
		},
		{
			SourceSystem: "modron",
			Name:         "SQL_PUBLIC_IP",
			Recipient:    user2,
			Content:      notificationContent(sqlRemediationDesc, sqlRemediationRecom),
			Interval:     24 * time.Hour,
		},
	}

	gotNotif := notifier.getNotifications()
	sort.Sort(sortNotifications(gotNotif))
	if diff := cmp.Diff(wantNotif, gotNotif); diff != "" {
		t.Fatalf("notifications diff (-want +got):\n%s", diff)
	}
}

type sortNotifications []model.Notification

func (s sortNotifications) Len() int {
	return len(s)
}

func (s sortNotifications) Less(i, j int) bool {
	if s[i].Name < s[j].Name {
		return true
	} else if s[i].Name > s[j].Name {
		return false
	}
	if s[i].Content < s[j].Content {
		return true
	} else if s[i].Content > s[j].Content {
		return false
	}
	return s[i].Recipient < s[j].Recipient
}

func (s sortNotifications) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

var _ sort.Interface = (*sortNotifications)(nil)

func TestCrossEnvRule(t *testing.T) {
	ctx := context.Background()
	svc, _ := getService(ctx, t)
	rgNames := []string{"projects/modron-test"}
	svc.RuleEngine, _ = engine.New(svc.Storage, []model.Rule{
		rules.NewCrossEnvironmentPermissionsRule(),
	},
		map[string]json.RawMessage{},
		nil,
		risk.TagConfig{
			ImpactMap:    impactMap,
			Environment:  "111111111111/environment",
			EmployeeData: "111111111111/employee_data",
			CustomerData: "111111111111/customer_data",
		})
	csRes, err := svc.CollectAndScan(ctx, &pb.CollectAndScanRequest{
		ResourceGroupNames: rgNames,
	})
	if err != nil {
		t.Fatalf("CollectAndScan: %v", err)
	}
	collectID := csRes.CollectId
	scanID := csRes.ScanId
	for {
		status, err := svc.GetStatusCollectAndScan(ctx, &pb.GetStatusCollectAndScanRequest{
			CollectId: collectID,
			ScanId:    scanID,
		})
		if err != nil {
			t.Fatalf("GetStatusCollectAndScan: %v", err)
		}
		if status.CollectStatus == pb.RequestStatus_DONE && status.ScanStatus == pb.RequestStatus_DONE {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Logf("Scan done")

	observations, err := svc.ListObservations(ctx, &pb.ListObservationsRequest{
		ResourceGroupNames: rgNames,
	})
	if err != nil {
		t.Fatalf("ListObservations: %v", err)
	}

	want := []*pb.ResourceGroupObservationsPair{
		{
			ResourceGroupName: "projects/modron-test",
			RulesObservations: []*pb.RuleObservationPair{
				{
					Rule: "CROSS_ENVIRONMENT_PERMISSIONS",
					Observations: []*pb.Observation{
						{
							Name:          "CROSS_ENVIRONMENT_PERMISSIONS",
							Category:      pb.Observation_CATEGORY_MISCONFIGURATION,
							ObservedValue: structpb.NewStringValue(""),
							ExpectedValue: structpb.NewStringValue("prod"),
							Remediation: &pb.Remediation{
								Description:    "account-3@modron-other-test.iam.gserviceaccount.com is in a different environment than the resource \"bucket-accessible-from-other-project\"",
								Recommendation: "Revoke the access of \"account-3@modron-other-test.iam.gserviceaccount.com\" to the resource \"bucket-accessible-from-other-project\"",
							},
							ResourceRef: &pb.ResourceRef{
								GroupName:     "projects/modron-test",
								CloudPlatform: pb.CloudPlatform_GCP,
								ExternalId:    proto.String("bucket-accessible-from-other-project"),
							},
							Source:       pb.Observation_SOURCE_MODRON,
							Severity:     pb.Severity_SEVERITY_HIGH,
							Impact:       pb.Impact_IMPACT_HIGH,
							RiskScore:    pb.Severity_SEVERITY_CRITICAL,
							ImpactReason: "environment=prod",
						},
						{
							Name:          "CROSS_ENVIRONMENT_PERMISSIONS",
							Category:      pb.Observation_CATEGORY_MISCONFIGURATION,
							ObservedValue: structpb.NewStringValue(""),
							ExpectedValue: structpb.NewStringValue("prod"),
							Remediation: &pb.Remediation{
								Description:    "account-3@modron-other-test.iam.gserviceaccount.com is in a different environment than the resource \"projects/modron-test\"",
								Recommendation: "Revoke the access of \"account-3@modron-other-test.iam.gserviceaccount.com\" to the resource \"projects/modron-test\"",
							},
							ResourceRef: &pb.ResourceRef{
								GroupName:     "projects/modron-test",
								CloudPlatform: pb.CloudPlatform_GCP,
								ExternalId:    proto.String("projects/modron-test"),
							},
							Source:       pb.Observation_SOURCE_MODRON,
							Severity:     pb.Severity_SEVERITY_HIGH,
							Impact:       pb.Impact_IMPACT_HIGH,
							RiskScore:    pb.Severity_SEVERITY_CRITICAL,
							ImpactReason: "environment=prod",
						},
					},
				},
				{
					Rule: "SQL_PUBLIC_IP",
					Observations: []*pb.Observation{
						{
							Name: "SQL_PUBLIC_IP",
							ResourceRef: &pb.ResourceRef{
								GroupName:     "projects/modron-test",
								ExternalId:    proto.String("//cloudsql.googleapis.com/projects/project-id/instances/xyz"),
								CloudPlatform: pb.CloudPlatform_GCP,
							},
							Remediation: &pb.Remediation{
								Description:    "To lower your attack surface, Cloud SQL databases should not have public IPs. Private IPs provide improved network security and lower latency for your application.",
								Recommendation: "Go to https://console.cloud.google.com/sql/instances/xyz/connections?project=project-id and click the \"Networking\" tab. Uncheck the \"Public IP\" checkbox and click \"SAVE\". If your instance is not configured to use a private IP, you will first have to enable private IP by following the instructions here: https://cloud.google.com/sql/docs/mysql/configure-private-ip#existing-private-instance",
							},
							CollectionId: proto.String(collectID),
							Category:     pb.Observation_CATEGORY_MISCONFIGURATION,
							ExternalId:   proto.String("//securitycenter.googleapis.com/projects/12345/sources/123/findings/48230f1978594ffb9d09a3cb1fe5e1b3"),
							Source:       pb.Observation_SOURCE_SCC,
							Impact:       pb.Impact_IMPACT_HIGH,
							ImpactReason: "environment=prod",
							Severity:     pb.Severity_SEVERITY_MEDIUM,
							RiskScore:    pb.Severity_SEVERITY_HIGH,
						},
					},
				},
			},
		},
	}
	if diff := cmp.Diff(
		want,
		observations.ResourceGroupsObservations,
		protocmp.Transform(),
		protocmp.IgnoreFields(&pb.Observation{}, "uid", "timestamp", "scan_uid"),
		protocmp.IgnoreFields(&pb.ResourceRef{}, "uid"),
	); diff != "" {
		t.Fatalf("ListObservations: diff (-want +got):\n%s", diff)
	}
}

func TestLabelToEmail(t *testing.T) {
	s, _ := getService(context.Background(), t)
	tests := []struct {
		labelContent string
		want         string
	}{
		{
			labelContent: "user_example_com",
			want:         "user@example.com",
		},
		{
			labelContent: "first_last_example_com",
			want:         "first.last@example.com",
		},
		{
			labelContent: "first_second_third_example_tokyo",
			want:         "first.second.third@example.tokyo",
		},
		{
			labelContent: "user.test_example_com",
			want:         "user.test@example.com",
		},
		{
			labelContent: "test",
			want:         "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.labelContent, func(t *testing.T) {
			got := s.LabelToEmail(tt.labelContent)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("LabelToEmail: diff (-want +got):\n%s", diff)
			}
		})
	}
}
