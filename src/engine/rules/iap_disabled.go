package rules

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const IAPDisabledRuleName = "IAP_DISABLED"

type IAPDisabledRule struct {
	info model.RuleInfo
}

func init() {
	AddRule(NewIAPDisabledRule())
}

func NewIAPDisabledRule() model.Rule {
	return &IAPDisabledRule{
		info: model.RuleInfo{
			Name: IAPDisabledRuleName,
			AcceptedResourceTypes: []proto.Message{
				&pb.LoadBalancer{},
			},
		},
	}
}

func (r *IAPDisabledRule) Check(_ context.Context, _ model.Engine, rsrc *pb.Resource) (obs []*pb.Observation, errs []error) {
	lb := rsrc.GetLoadBalancer()

	// TODO: Add port and name validation to have more accurate detection.
	if !lb.GetIap().GetEnabled() {
		ob := &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			ResourceRef:   utils.GetResourceRef(rsrc),
			Name:          r.Info().Name,
			ExpectedValue: structpb.NewBoolValue(true),
			ObservedValue: structpb.NewBoolValue(false),
			Remediation: &pb.Remediation{
				Description: fmt.Sprintf(
					"IAP is disabled on Load Balancer [%q](https://console.cloud.google.com/net-services/loadbalancing/list/loadBalancers?project=%s) which exposes internal resources on the internet",
					getGcpReadableResourceName(rsrc.Name),
					constants.ResourceWithoutProjectsPrefix(rsrc.ResourceGroupName),
				),
				Recommendation: fmt.Sprintf(
					"Enable IAP on Load Balancer [%q](https://console.cloud.google.com/net-services/loadbalancing/list/loadBalancers?project=%s) to secure the access to internal resources and prevent unauthorized access.",
					getGcpReadableResourceName(rsrc.Name),
					constants.ResourceWithoutProjectsPrefix(rsrc.ResourceGroupName),
				),
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		}
		obs = append(obs, ob)
	}

	return
}

func (r *IAPDisabledRule) Info() *model.RuleInfo {
	return &r.info
}
