package rules

import (
	"testing"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestIAPDisabledRule(t *testing.T) {
	iapDisabledResource := &pb.Resource{
		Name:              "load-balancer-with-iap-disabled",
		ResourceGroupName: testProjectName,
		Type: &pb.Resource_LoadBalancer{
			LoadBalancer: &pb.LoadBalancer{
				Type: pb.LoadBalancer_EXTERNAL,
				Iap: &pb.IAP{
					Enabled: false,
				},
			},
		},
	}
	iapUnspecifiedResource := &pb.Resource{
		Name:              "load-balancer-with-iap-unspecified",
		ResourceGroupName: testProjectName,
		Type: &pb.Resource_LoadBalancer{
			LoadBalancer: &pb.LoadBalancer{
				Type: pb.LoadBalancer_EXTERNAL,
			},
		},
	}

	resources := []*pb.Resource{
		{
			Name:              "load-balancer-with-iap-enabled",
			ResourceGroupName: testProjectName,
			Type: &pb.Resource_LoadBalancer{
				LoadBalancer: &pb.LoadBalancer{
					Type: pb.LoadBalancer_EXTERNAL,
					Iap: &pb.IAP{
						Enabled: true,
					},
				},
			},
		},
		iapDisabledResource,
		iapUnspecifiedResource,
	}

	want := []*pb.Observation{
		{
			Name:          IAPDisabledRuleName,
			ResourceRef:   utils.GetResourceRef(iapDisabledResource),
			ExpectedValue: structpb.NewBoolValue(true),
			ObservedValue: structpb.NewBoolValue(false),
			Remediation: &pb.Remediation{
				Description:    "IAP is disabled on Load Balancer [\"load-balancer-with-iap-disabled\"](https://console.cloud.google.com/net-services/loadbalancing/list/loadBalancers?project=project-0) which exposes internal resources on the internet",
				Recommendation: "Enable IAP on Load Balancer [\"load-balancer-with-iap-disabled\"](https://console.cloud.google.com/net-services/loadbalancing/list/loadBalancers?project=project-0) to secure the access to internal resources and prevent unauthorized access.",
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		},
		{
			Name:          IAPDisabledRuleName,
			ResourceRef:   utils.GetResourceRef(iapUnspecifiedResource),
			ExpectedValue: structpb.NewBoolValue(true),
			ObservedValue: structpb.NewBoolValue(false),
			Remediation: &pb.Remediation{
				Description:    "IAP is disabled on Load Balancer [\"load-balancer-with-iap-unspecified\"](https://console.cloud.google.com/net-services/loadbalancing/list/loadBalancers?project=project-0) which exposes internal resources on the internet",
				Recommendation: "Enable IAP on Load Balancer [\"load-balancer-with-iap-unspecified\"](https://console.cloud.google.com/net-services/loadbalancing/list/loadBalancers?project=project-0) to secure the access to internal resources and prevent unauthorized access.",
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		},
	}

	TestRuleRun(t, resources, []model.Rule{NewIAPDisabledRule()}, want)
}
