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
			AcceptedResourceTypes: []proto.Message{
				&pb.LoadBalancer{},
			},
		},
	}
}

func (r *LbUserManagedCertRule) Check(_ context.Context, _ model.Engine, rsrc *pb.Resource) (obs []*pb.Observation, errs []error) {
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
				ResourceRef:   utils.GetResourceRef(rsrc),
				Name:          r.Info().Name,
				ExpectedValue: structpb.NewStringValue(pb.Certificate_MANAGED.String()),
				ObservedValue: structpb.NewStringValue(cert.Type.String()),
				Remediation: &pb.Remediation{
					Description: fmt.Sprintf(
						"Load balancer [%q](https://console.cloud.google.com/net-services/loadbalancing/list/loadBalancers?project=%s) has user-managed certificate issued by %q for the domain %q",
						getGcpReadableResourceName(rsrc.Name),
						constants.ResourceWithoutProjectsPrefix(rsrc.ResourceGroupName),
						cert.Issuer,
						cert.DomainName,
					),
					Recommendation: fmt.Sprintf(
						"Configure a platform-managed certificate for load balancer [%q](https://console.cloud.google.com/net-services/loadbalancing/list/loadBalancers?project=%s) to ensure lower management overhead, better security and prevent outages caused by certificate expiry",
						getGcpReadableResourceName(rsrc.Name),
						constants.ResourceWithoutProjectsPrefix(rsrc.ResourceGroupName),
					),
				},
				Severity: pb.Severity_SEVERITY_INFO,
			}
			obs = append(obs, ob)
		}
	}

	return
}

func (r *LbUserManagedCertRule) Info() *model.RuleInfo {
	return &r.info
}
