package rules

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/engine"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const LbUserManagedCertRuleName = "LOAD_BALANCER_USER_MANAGED_CERTIFICATE"

type LbUserManagedCertRule struct {
	info model.RuleInfo
}

func init() {
	AddRule(NewLbUserManagedCertRule())
}

func NewLbUserManagedCertRule() model.Rule {
	return &LbUserManagedCertRule{
		info: model.RuleInfo{
			Name: LbUserManagedCertRuleName,
			AcceptedResourceTypes: []string{
				common.ResourceLoadBalancer,
			},
		},
	}
}

func (r *LbUserManagedCertRule) Check(ctx context.Context, rsrc *pb.Resource) (obs []*pb.Observation, errs []error) {
	lb := rsrc.GetLoadBalancer()

	for _, cert := range lb.Certificates {
		if cert.Type == pb.Certificate_UNKNOWN {
			errs = append(
				errs,
				fmt.Errorf("certificate issued by %q for the domain %q is of unknown type", cert.Issuer, cert.DomainName),
			)
			continue
		}
		if cert.Type == pb.Certificate_IMPORTED {
			ob := &pb.Observation{
				Uid:           uuid.NewString(),
				Timestamp:     timestamppb.Now(),
				Resource:      rsrc,
				Name:          r.Info().Name,
				ExpectedValue: structpb.NewNumberValue(float64(pb.Certificate_MANAGED)),
				ObservedValue: structpb.NewNumberValue(float64(cert.Type)),
				Remediation: &pb.Remediation{
					Description: fmt.Sprintf(
						"Load balancer [%q](https://console.cloud.google.com/net-services/loadbalancing/list/loadBalancers?project=%s) has user-managed certificate issued by %q for the domain %q",
						engine.GetGcpReadableResourceName(rsrc.Name),
						rsrc.ResourceGroupName,
						cert.Issuer,
						cert.DomainName,
					),
					Recommendation: fmt.Sprintf(
						"Configure a platform-managed certificate for load balancer [%q](https://console.cloud.google.com/net-services/loadbalancing/list/loadBalancers?project=%s) to ensure lower management overhead, better security and prevent outages caused by certificate expiry",
						engine.GetGcpReadableResourceName(rsrc.Name),
						rsrc.ResourceGroupName,
					),
				},
			}
			obs = append(obs, ob)
		}
	}

	return
}

func (r *LbUserManagedCertRule) Info() *model.RuleInfo {
	return &r.info
}
