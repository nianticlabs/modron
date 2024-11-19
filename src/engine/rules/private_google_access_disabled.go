package rules

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"
)

const PrivateGoogleAccessDisabled = "PRIVATE_GOOGLE_ACCESS_DISABLED"

type PrivateGoogleAccessDisabledRule struct {
	info model.RuleInfo
}

func init() {
	AddRule(NewPrivateGoogleAccessDisabledRule())
}

func NewPrivateGoogleAccessDisabledRule() model.Rule {
	return &PrivateGoogleAccessDisabledRule{
		info: model.RuleInfo{
			Name: PrivateGoogleAccessDisabled,
			AcceptedResourceTypes: []proto.Message{
				&pb.Network{},
			},
		},
	}
}

func (r *PrivateGoogleAccessDisabledRule) Check(_ context.Context, _ model.Engine, rsrc *pb.Resource) ([]*pb.Observation, []error) {
	net := rsrc.GetNetwork()
	var obs []*pb.Observation

	if !net.GcpPrivateGoogleAccessV4 {
		ob := &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			ResourceRef:   utils.GetResourceRef(rsrc),
			Name:          r.Info().Name,
			ExpectedValue: structpb.NewStringValue("enabled"),
			ObservedValue: structpb.NewStringValue("disabled"),
			Remediation: &pb.Remediation{
				Description: fmt.Sprintf(
					"Network [%q](https://console.cloud.google.com/networking/networks/details/%s?project=%s) has [Private Google Access](https://cloud.google.com/vpc/docs/configure-private-google-access) disabled. Private Google Access allows the workloads to access Google APIs via a private network which is safer than going over the public Internet",
					getGcpReadableResourceName(rsrc.Name),
					getGcpReadableResourceName(rsrc.Name),
					constants.ResourceWithoutProjectsPrefix(rsrc.ResourceGroupName),
				),
				Recommendation: fmt.Sprintf(
					"Enable [Private Google Access](https://cloud.google.com/vpc/docs/configure-private-google-access) for Network [%q](https://console.cloud.google.com/networking/networks/details/%s?project=%s)",
					getGcpReadableResourceName(rsrc.Name),
					getGcpReadableResourceName(rsrc.Name),
					constants.ResourceWithoutProjectsPrefix(rsrc.ResourceGroupName),
				),
			},
			Severity: pb.Severity_SEVERITY_LOW,
		}
		obs = append(obs, ob)

	}
	return obs, nil
}

func (r *PrivateGoogleAccessDisabledRule) Info() *model.RuleInfo {
	return &r.info
}
