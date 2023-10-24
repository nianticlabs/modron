package rules

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/types/known/structpb"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

func TestCheckDetectTooOldMinTlsVersion(t *testing.T) {
	resources := []*pb.Resource{
		{
			Name:              testProjectName,
			Parent:            "",
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{},
			},
		},
		{
			Name:              "lb-good-tls-policy",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_LoadBalancer{
				LoadBalancer: &pb.LoadBalancer{
					Type: pb.LoadBalancer_EXTERNAL,
					SslPolicy: &pb.SslPolicy{
						MinTlsVersion: pb.SslPolicy_TLS_1_2,
						Profile:       pb.SslPolicy_MODERN,
						Name:          "Great SSL Policy",
					},
				},
			},
		},
		{
			Name:              "lb-gcp-default",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_LoadBalancer{
				LoadBalancer: &pb.LoadBalancer{
					Type: pb.LoadBalancer_EXTERNAL,
					SslPolicy: &pb.SslPolicy{
						MinTlsVersion: pb.SslPolicy_TLS_1_0,
						Profile:       pb.SslPolicy_COMPATIBLE,
						Name:          "GCP Default",
					},
				},
			},
		},
		{
			Name:              "lb-gcp-default-internal",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_LoadBalancer{
				LoadBalancer: &pb.LoadBalancer{
					Type: pb.LoadBalancer_INTERNAL,
					SslPolicy: &pb.SslPolicy{
						MinTlsVersion: pb.SslPolicy_TLS_1_0,
						Profile:       pb.SslPolicy_COMPATIBLE,
						Name:          "GCP Default",
					},
				},
			},
		},
	}

	want := []*pb.Observation{
		{
			Name: LbMinTlsVersionTooOldRuleName,
			Resource: &pb.Resource{
				Name: "lb-gcp-default",
			},
			ExpectedValue: structpb.NewStringValue(protoToVersionMap[pb.SslPolicy_TLS_1_2]),
			ObservedValue: structpb.NewStringValue(protoToVersionMap[pb.SslPolicy_TLS_1_0]),
		},
	}
	got := TestRuleRun(t, resources, []model.Rule{NewLbMinTlsVersionTooOldRule()})

	// Check that the observations are correct.
	if diff := cmp.Diff(want, got, cmp.Comparer(observationComparer), cmpopts.SortSlices(observationsSorter)); diff != "" {
		t.Errorf("CheckRules unexpected diff (-want, +got): %v", diff)
	}
}
