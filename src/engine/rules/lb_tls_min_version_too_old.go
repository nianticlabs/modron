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

const lbMinTLSVersionTooOldRule = "LOAD_BALANCER_MIN_TLS_VERSION_TOO_OLD"

var (
	minTLSVersion     = pb.SslPolicy_TLS_1_2
	protoToVersionMap = map[pb.SslPolicy_MinTlsVersion]string{
		pb.SslPolicy_TLS_1_0: "TLS 1.0",
		pb.SslPolicy_TLS_1_1: "TLS 1.1",
		pb.SslPolicy_TLS_1_2: "TLS 1.2",
		pb.SslPolicy_TLS_1_3: "TLS 1.3",
	}
)

type LbMinTLSVersionTooOldRule struct {
	info model.RuleInfo
}

func init() {
	AddRule(NewLbMinTLSVersionTooOldRule())
}

func NewLbMinTLSVersionTooOldRule() model.Rule {
	return &LbMinTLSVersionTooOldRule{
		info: model.RuleInfo{
			Name: lbMinTLSVersionTooOldRule,
			AcceptedResourceTypes: []proto.Message{
				&pb.LoadBalancer{},
			},
		},
	}
}

func (r *LbMinTLSVersionTooOldRule) Check(_ context.Context, _ model.Engine, rsrc *pb.Resource) (obs []*pb.Observation, errs []error) {
	lb := rsrc.GetLoadBalancer()
	sslPolicy := lb.SslPolicy

	if sslPolicy == nil {
		return nil, []error{fmt.Errorf("SSL policy is nil")}
	}

	if lb.Type == pb.LoadBalancer_EXTERNAL && sslPolicy.MinTlsVersion < minTLSVersion {
		obs = append(obs, &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			ResourceRef:   utils.GetResourceRef(rsrc),
			Name:          r.Info().Name,
			ExpectedValue: structpb.NewStringValue(protoToVersionMap[minTLSVersion]),
			ObservedValue: structpb.NewStringValue(protoToVersionMap[sslPolicy.MinTlsVersion]),
			Remediation: &pb.Remediation{
				Description: fmt.Sprintf(
					"The load balancer [%q](https://console.cloud.google.com/net-services/loadbalancing/list/loadBalancers?project=%s) allows connections over an outdated TLS protocol version. Outdated protocol versions use ciphers which can be attacked by dedicated threat actors",
					getGcpReadableResourceName(rsrc.Name),
					constants.ResourceWithoutProjectsPrefix(rsrc.ResourceGroupName),
				),
				Recommendation: fmt.Sprintf(
					"Configure an SSL policy for the load balancer [%q](https://console.cloud.google.com/net-services/loadbalancing/list/loadBalancers?project=%s) that has a Minimum TLS version of %s and uses e.g. the \"MODERN\" or \"RESTRICTED\" configuration",
					getGcpReadableResourceName(rsrc.Name),
					constants.ResourceWithoutProjectsPrefix(rsrc.ResourceGroupName),
					protoToVersionMap[minTLSVersion],
				),
			},
			Severity: pb.Severity_SEVERITY_HIGH,
		})
	}

	return obs, errs
}

func (r *LbMinTLSVersionTooOldRule) Info() *model.RuleInfo {
	return &r.info
}
