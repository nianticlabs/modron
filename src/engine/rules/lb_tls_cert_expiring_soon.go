package rules

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const LoadBalancerTLSCertExpiringSoonRuleName = "LOAD_BALANCER_TLS_CERTIFICATE_EXPIRING_SOON"

const (
	hoursInADay       = 24
	oneDay            = hoursInADay * time.Hour
	certExpiryWarning = 90 * oneDay // 90 days
)

type LoadBalancerTLSCertExpiringSoonRule struct {
	info model.RuleInfo
}

func init() {
	AddRule(NewLoadBalancerTLSCertExpiringSoonRule())
}

func NewLoadBalancerTLSCertExpiringSoonRule() model.Rule {
	return &LoadBalancerTLSCertExpiringSoonRule{
		info: model.RuleInfo{
			Name: LoadBalancerTLSCertExpiringSoonRuleName,
			AcceptedResourceTypes: []proto.Message{
				&pb.LoadBalancer{},
			},
		},
	}
}

func (r *LoadBalancerTLSCertExpiringSoonRule) Check(_ context.Context, _ model.Engine, rsrc *pb.Resource) (obs []*pb.Observation, errs []error) {
	lb := rsrc.GetLoadBalancer()

	if lb.Type == pb.LoadBalancer_EXTERNAL && lb.SslPolicy != nil {
		for _, cert := range lb.Certificates {
			daysUntilExpiry := int(time.Until(cert.ExpirationDate.AsTime()).Hours() / hoursInADay)
			if cert.Type == pb.Certificate_IMPORTED {
				if time.Until(cert.ExpirationDate.AsTime()) < certExpiryWarning {
					obs = append(obs, &pb.Observation{
						Uid:           uuid.NewString(),
						Timestamp:     timestamppb.Now(),
						ResourceRef:   utils.GetResourceRef(rsrc),
						Name:          r.Info().Name,
						ExpectedValue: structpb.NewStringValue(fmt.Sprintf("More than %d days", int(certExpiryWarning.Hours()/hoursInADay))),
						ObservedValue: structpb.NewStringValue(fmt.Sprintf("%d days", daysUntilExpiry)),
						Remediation: &pb.Remediation{
							Description: fmt.Sprintf(
								"The TLS certificate for load balancer [%q](https://console.cloud.google.com/net-services/loadbalancing/list/loadBalancers?project=%s) will expire in %d days. Certificates should be renewed before expiry to avoid service disruptions.",
								getGcpReadableResourceName(rsrc.Name),
								constants.ResourceWithoutProjectsPrefix(rsrc.ResourceGroupName),
								daysUntilExpiry,
							),
							Recommendation: fmt.Sprintf(
								"Renew the TLS certificate for load balancer [%q](https://console.cloud.google.com/net-services/loadbalancing/list/loadBalancers?project=%s). go/renew-certificate has information on how to proceed.",
								getGcpReadableResourceName(rsrc.Name),
								constants.ResourceWithoutProjectsPrefix(rsrc.ResourceGroupName),
							),
						},
						Severity: pb.Severity_SEVERITY_HIGH,
					})
				}
			}
		}
	}

	return obs, errs
}

func (r *LoadBalancerTLSCertExpiringSoonRule) Info() *model.RuleInfo {
	return &r.info
}
