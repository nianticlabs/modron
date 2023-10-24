package rules

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const LbMinTlsVersionTooOldRuleName = "LOAD_BALANCER_MIN_TLS_VERSION_TOO_OLD"

var (
	minTlsVersion     = pb.SslPolicy_TLS_1_2
	protoToVersionMap = map[pb.SslPolicy_MinTlsVersion]string{
		pb.SslPolicy_TLS_1_0: "TLS 1.0",
		pb.SslPolicy_TLS_1_1: "TLS 1.1",
		pb.SslPolicy_TLS_1_2: "TLS 1.2",
		pb.SslPolicy_TLS_1_3: "TLS 1.3",
	}
)

type LbMinTlsVersionTooOldRule struct {
	info model.RuleInfo
}

func init() {
	AddRule(NewLbMinTlsVersionTooOldRule())
}

func NewLbMinTlsVersionTooOldRule() model.Rule {
	return &LbMinTlsVersionTooOldRule{
		info: model.RuleInfo{
			Name: LbMinTlsVersionTooOldRuleName,
			AcceptedResourceTypes: []string{
				common.ResourceLoadBalancer,
			},
		},
	}
}

func (r *LbMinTlsVersionTooOldRule) Check(ctx context.Context, rsrc *pb.Resource) (obs []*pb.Observation, errs []error) {
	lb := rsrc.GetLoadBalancer()

	sslPolicy := lb.SslPolicy

	if lb.Type == pb.LoadBalancer_EXTERNAL && sslPolicy.MinTlsVersion < minTlsVersion {
		obs = append(obs, &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			Resource:      rsrc,
			Name:          r.Info().Name,
			ExpectedValue: structpb.NewStringValue(protoToVersionMap[minTlsVersion]),
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
					protoToVersionMap[minTlsVersion],
				),
			},
		})
	}

	return obs, errs
}

func (r *LbMinTlsVersionTooOldRule) Info() *model.RuleInfo {
	return &r.info
}
