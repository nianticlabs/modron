package rules

import (
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
)

func TestCheckDetectExpiringCertificate(t *testing.T) {
	in10Days := timestamppb.New(time.Now().Add(10 * 24 * time.Hour))
	resources := []*pb.Resource{
		{
			Name:              "lb-cert-expiring-in-10-days",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_LoadBalancer{
				LoadBalancer: &pb.LoadBalancer{
					Type: pb.LoadBalancer_EXTERNAL,
					Certificates: []*pb.Certificate{
						{
							ExpirationDate: in10Days,
							Type:           pb.Certificate_IMPORTED,
						},
					},
					SslPolicy: &pb.SslPolicy{
						MinTlsVersion: pb.SslPolicy_TLS_1_2,
						Profile:       pb.SslPolicy_MODERN,
						Name:          "Great SSL Policy",
					},
				},
			},
		},
		{
			Name:              "lb-cert-expiring-in-10-days-managed-no-report",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_LoadBalancer{
				LoadBalancer: &pb.LoadBalancer{
					Type: pb.LoadBalancer_EXTERNAL,
					Certificates: []*pb.Certificate{
						{
							ExpirationDate: in10Days,
							Type:           pb.Certificate_MANAGED,
						},
					},
					SslPolicy: &pb.SslPolicy{
						MinTlsVersion: pb.SslPolicy_TLS_1_2,
						Profile:       pb.SslPolicy_MODERN,
						Name:          "Great SSL Policy",
					},
				},
			},
		},
		{
			Name:              "lb-cert-expiring-in-10-days-internal-no-report",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_LoadBalancer{
				LoadBalancer: &pb.LoadBalancer{
					Type: pb.LoadBalancer_INTERNAL,
					Certificates: []*pb.Certificate{
						{
							ExpirationDate: in10Days,
							Type:           pb.Certificate_IMPORTED,
						},
					},
					SslPolicy: &pb.SslPolicy{
						MinTlsVersion: pb.SslPolicy_TLS_1_2,
						Profile:       pb.SslPolicy_MODERN,
						Name:          "Great SSL Policy",
					},
				},
			},
		},
	}

	want := []*pb.Observation{
		{
			Name: LoadBalancerTLSCertExpiringSoonRuleName,
			ResourceRef: &pb.ResourceRef{
				Uid:           proto.String("uuid-0"),
				ExternalId:    proto.String("lb-cert-expiring-in-10-days"),
				GroupName:     testProjectName,
				CloudPlatform: pb.CloudPlatform_GCP,
			},
			Remediation: &pb.Remediation{
				Description:    "The TLS certificate for load balancer [\"lb-cert-expiring-in-10-days\"](https://console.cloud.google.com/net-services/loadbalancing/list/loadBalancers?project=project-0) will expire in 9 days. Certificates should be renewed before expiry to avoid service disruptions.",
				Recommendation: "Renew the TLS certificate for load balancer [\"lb-cert-expiring-in-10-days\"](https://console.cloud.google.com/net-services/loadbalancing/list/loadBalancers?project=project-0). go/renew-certificate has information on how to proceed.",
			},
			ExpectedValue: structpb.NewStringValue("More than 90 days"),
			ObservedValue: structpb.NewStringValue("9 days"),
			Severity:      pb.Severity_SEVERITY_HIGH,
		},
	}

	TestRuleRun(t, resources, []model.Rule{NewLoadBalancerTLSCertExpiringSoonRule()}, want)
}
