package rules

import (
	"testing"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"
)

func TestCheckDetectTooOldMinTlsVersion(t *testing.T) {
	lbGcpDefault := &pb.Resource{
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
	}

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
		lbGcpDefault,
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
			Name:          lbMinTLSVersionTooOldRule,
			ResourceRef:   utils.GetResourceRef(lbGcpDefault),
			ExpectedValue: structpb.NewStringValue(protoToVersionMap[pb.SslPolicy_TLS_1_2]),
			ObservedValue: structpb.NewStringValue(protoToVersionMap[pb.SslPolicy_TLS_1_0]),
			Remediation: &pb.Remediation{
				Description:    "The load balancer [\"lb-gcp-default\"](https://console.cloud.google.com/net-services/loadbalancing/list/loadBalancers?project=project-0) allows connections over an outdated TLS protocol version. Outdated protocol versions use ciphers which can be attacked by dedicated threat actors",
				Recommendation: "Configure an SSL policy for the load balancer [\"lb-gcp-default\"](https://console.cloud.google.com/net-services/loadbalancing/list/loadBalancers?project=project-0) that has a Minimum TLS version of TLS 1.2 and uses e.g. the \"MODERN\" or \"RESTRICTED\" configuration",
			},
			Severity: pb.Severity_SEVERITY_HIGH,
		},
	}

	TestRuleRun(t, resources, []model.Rule{NewLbMinTLSVersionTooOldRule()}, want)
}
