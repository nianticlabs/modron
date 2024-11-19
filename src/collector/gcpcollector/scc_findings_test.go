package gcpcollector

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"google.golang.org/api/securitycenter/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"

	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/risk"
	"github.com/nianticlabs/modron/src/storage/memstorage"
)

var (
	vuln1 = &securitycenter.Finding{
		CanonicalName: "projects/1234/sources/5678/locations/global/findings/0000",
		Name:          "projects/1234/sources/5678/locations/global/findings/0000",
		Category:      "GKE_RUNTIME_OS_VULNERABILITY",
		ResourceName:  "//container.googleapis.com/projects/modron0test/locations/us-west1/clusters/my-cluster",
		Severity:      "SEVERITY_UNSPECIFIED",
		State:         "ACTIVE",
		CreateTime:    "2024-05-19T13:20:02.112Z",
		EventTime:     "2024-07-04T15:47:51.202Z",
		Vulnerability: &securitycenter.Vulnerability{
			Cve: &securitycenter.Cve{
				Cvssv3: &securitycenter.Cvssv3{
					BaseScore:             5.5,
					AttackComplexity:      "ATTACK_COMPLEXITY_LOW",
					AttackVector:          "ATTACK_VECTOR_LOCAL",
					AvailabilityImpact:    "IMPACT_HIGH",
					ConfidentialityImpact: "IMPACT_NONE",
					IntegrityImpact:       "IMPACT_NONE",
					PrivilegesRequired:    "PRIVILEGES_REQUIRED_LOW",
					Scope:                 "SCOPE_UNCHANGED",
					UserInteraction:       "USER_INTERACTION_NONE",
				},
				ExploitationActivity: "NO_KNOWN",
				Id:                   "CVE-2024-35863",
				Impact:               "LOW",
				ObservedInTheWild:    false,
				References:           nil,
				UpstreamFixAvailable: true,
				ZeroDay:              false,
			},
			OffendingPackage: &securitycenter.Package{
				CpeUri:         "cpe:/o:debian:debian_linux:12",
				PackageName:    "linux",
				PackageType:    "OS",
				PackageVersion: "6.1.76-1",
			},
			FixedPackage: &securitycenter.Package{
				CpeUri:         "cpe:/o:debian:debian_linux:12",
				PackageName:    "linux",
				PackageType:    "OS",
				PackageVersion: "6.1.85-1",
			},
			SecurityBulletin: nil,
		},
		Description: "In the Linux kernel, the following vulnerability has been resolved:  smb: client: fix potential UAF in is_valid_oplock_break()  Skip sessions that are being teared down (status == SES_EXITING) to avoid UAF.",
		NextSteps:   "Use the following resources to help you mitigate CVE-2024-35863.\n\n**More information about CVE-2024-35863**\n* https://security-tracker.debian.org/tracker/CVE-2024-35863\n* https://access.redhat.com/security/cve/CVE-2024-35863\n* https://www.suse.com/security/cve/CVE-2024-35863\n* http://people.ubuntu.com/~ubuntu-security/cve/CVE-2024-35863\n\n**Fixed location**\ncpe:/o:debian:debian_linux:12\n\n**Fixed package**\nlinux\n\n**Fixed version**\n6.1.85-1\n",
		Kubernetes: &securitycenter.Kubernetes{
			Objects: []*securitycenter.Object{
				{
					Kind: "Deployment",
					Ns:   "default",
					Name: "my-deployment",
					Containers: []*securitycenter.Container{
						{
							Name:    "us-west1-docker.pkg.dev/project-id/my-image:latest",
							ImageId: "us-west1-docker.pkg.dev/project-id/my-image:latest",
						},
					},
				},
			},
		},
	}
)

func TestFindingToObservation(t *testing.T) {
	got := FindingToObservation(vuln1, "projects/modron-test")

	want := &pb.Observation{
		ExpectedValue: structpb.NewStringValue("linux 6.1.85-1"),
		ObservedValue: structpb.NewStringValue("linux 6.1.76-1"),
		Uid:           "projects/1234/sources/5678/locations/global/findings/0000",
		Name:          "GKE_RUNTIME_OS_VULNERABILITY",
		Remediation: &pb.Remediation{
			Description:    "CVE-2024-35863 in linux 6.1.76-1 (us-west1-docker.pkg.dev/project-id/my-image:latest): update to linux 6.1.85-1\n\nIn the Linux kernel, the following vulnerability has been resolved:  smb: client: fix potential UAF in is_valid_oplock_break()  Skip sessions that are being teared down (status == SES_EXITING) to avoid UAF.",
			Recommendation: "Use the following resources to help you mitigate CVE-2024-35863.\n\n**More information about CVE-2024-35863**\n* https://security-tracker.debian.org/tracker/CVE-2024-35863\n* https://access.redhat.com/security/cve/CVE-2024-35863\n* https://www.suse.com/security/cve/CVE-2024-35863\n* http://people.ubuntu.com/~ubuntu-security/cve/CVE-2024-35863\n\n**Fixed location**\ncpe:/o:debian:debian_linux:12\n\n**Fixed package**\nlinux\n\n**Fixed version**\n6.1.85-1\n",
		},
		ResourceRef: &pb.ResourceRef{
			GroupName:     "projects/modron-test",
			ExternalId:    proto.String("//container.googleapis.com/projects/modron0test/locations/us-west1/clusters/my-cluster/k8s/namespaces/default/apps/deployments/my-deployment"),
			CloudPlatform: pb.CloudPlatform_GCP,
		},
		ExternalId: proto.String("//securitycenter.googleapis.com/projects/1234/sources/5678/locations/global/findings/0000"),
		Source:     pb.Observation_SOURCE_SCC,
	}

	if diff := cmp.Diff(want, got, protocmp.Transform(), protocmp.IgnoreFields(&pb.Observation{}, "timestamp", "uid")); diff != "" {
		t.Errorf("FindingToObservation() mismatch (-want +got):\n%s", diff)
	}
}

func TestListSccFindings(t *testing.T) {
	otelMeter := otel.GetMeterProvider()
	var mr *metric.ManualReader
	if otelMeter == nil {
		mr = metric.NewManualReader()
		otel.SetMeterProvider(
			metric.NewMeterProvider(
				metric.WithReader(mr),
			),
		)
	}

	ctx := context.Background()
	storage := memstorage.New()
	gcpCollector := NewFake(ctx, storage, risk.TagConfig{})
	got, err := gcpCollector.ListSccFindings(ctx, "projects/project-id")
	if err != nil {
		t.Fatalf("ListSccFindings: %v", err)
	}

	want := []*pb.Observation{
		{
			Name: "SQL_PUBLIC_IP",
			Remediation: &pb.Remediation{
				Description:    "To lower your attack surface, Cloud SQL databases should not have public IPs. Private IPs provide improved network security and lower latency for your application.",
				Recommendation: "Go to https://console.cloud.google.com/sql/instances/xyz/connections?project=project-id and click the \"Networking\" tab. Uncheck the \"Public IP\" checkbox and click \"SAVE\". If your instance is not configured to use a private IP, you will first have to enable private IP by following the instructions here: https://cloud.google.com/sql/docs/mysql/configure-private-ip#existing-private-instance",
			},
			ExternalId: proto.String("//securitycenter.googleapis.com/projects/12345/sources/123/findings/48230f1978594ffb9d09a3cb1fe5e1b3"),
			Source:     pb.Observation_SOURCE_SCC,
			Severity:   pb.Severity_SEVERITY_MEDIUM,
			Category:   pb.Observation_CATEGORY_MISCONFIGURATION,
			ResourceRef: &pb.ResourceRef{
				ExternalId:    proto.String("//cloudsql.googleapis.com/projects/project-id/instances/xyz"),
				CloudPlatform: pb.CloudPlatform_GCP,
				GroupName:     "projects/project-id",
			},
		},
	}
	if diff := cmp.Diff(want, got, protocmp.Transform(), protocmp.IgnoreFields(&pb.Observation{}, "timestamp", "uid")); diff != "" {
		t.Errorf("ListSccFindings() mismatch (-want +got):\n%s", diff)
	}

	if mr != nil {
		resourceMetrics := metricdata.ResourceMetrics{}
		if err := mr.Collect(context.Background(), &resourceMetrics); err != nil {
			t.Fatalf("Collect: %v", err)
		}

		metricScope := "github.com/nianticlabs/modron/src/collector/gcpcollector"
		gotMetrics := getMetrics(resourceMetrics, metricScope)
		if gotMetrics == nil {
			t.Fatalf("no metrics found for scope %q", metricScope)
		}

		wantMetrics := []metricdata.Metrics{
			{
				Name:        "modron_scc_collected_observations",
				Description: "Number of collected observations from SCC",
				Data: metricdata.Sum[int64]{
					DataPoints: []metricdata.DataPoint[int64]{
						{
							Attributes: attribute.NewSet(
								attribute.String("category", "SQL_PUBLIC_IP"),
								attribute.String("severity", "MEDIUM"),
							),
							Value: 1,
						},
					},
				},
			},
		}

		if diff := cmp.Diff(wantMetrics, gotMetrics, metricsCmpOpts()...); diff != "" {
			t.Errorf("ResourceMetrics mismatch (-want +got):\n%s", diff)
		}
	} else {
		t.Log("The test was run in parallel mode, thus we'll not check the result of the metrics (partial!)")
	}

}

// Adapted from https://github.com/temporalio/temporal/blob/8752a90c5b15851d81c7aff98e0fe94a6cb57d13/common/metrics/otel_metrics_handler_test.go#L200-L218
func metricsCmpOpts() []cmp.Option {
	return []cmp.Option{
		cmp.Comparer(func(e1, e2 metricdata.Extrema[int64]) bool {
			v1, ok1 := e1.Value()
			v2, ok2 := e2.Value()
			return ok1 && ok2 && v1 == v2
		}),
		cmp.Comparer(func(a1, a2 attribute.Set) bool {
			return a1.Equals(&a2)
		}),
		cmpopts.SortSlices(func(x, y metricdata.Metrics) bool {
			return x.Name < y.Name
		}),
		cmpopts.IgnoreFields(metricdata.DataPoint[int64]{}, "StartTime", "Time"),
		cmpopts.IgnoreFields(metricdata.DataPoint[float64]{}, "StartTime", "Time"),
		cmpopts.IgnoreFields(metricdata.Sum[int64]{}, "Temporality", "IsMonotonic"),
		cmpopts.IgnoreFields(metricdata.HistogramDataPoint[int64]{}, "StartTime", "Time", "Bounds"),
	}
}

func getMetrics(resourceMetrics metricdata.ResourceMetrics, s string) []metricdata.Metrics {
	for _, v := range resourceMetrics.ScopeMetrics {
		if v.Scope.Name == s {
			return v.Metrics
		}
	}
	return nil
}
