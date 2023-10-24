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

const MasterAuthorizedNetworksNotSet = "MASTER_AUTHORIZED_NETWORKS_NOT_SET"

type MasterAuthorizedNetworksNotSetRule struct {
	info model.RuleInfo
}

func init() {
	AddRule(NewMasterAuthorizedNetworksNotSetRule())
}

func NewMasterAuthorizedNetworksNotSetRule() model.Rule {
	return &MasterAuthorizedNetworksNotSetRule{
		info: model.RuleInfo{
			Name: MasterAuthorizedNetworksNotSet,
			AcceptedResourceTypes: []string{
				common.ResourceKubernetesCluster,
			},
		},
	}
}

func (r *MasterAuthorizedNetworksNotSetRule) Check(ctx context.Context, rsrc *pb.Resource) ([]*pb.Observation, []error) {
	k8s := rsrc.GetKubernetesCluster()
	obs := []*pb.Observation{}
	if len(k8s.MasterAuthorizedNetworks) < 1 {
		ob := &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			Resource:      rsrc,
			Name:          r.Info().Name,
			ExpectedValue: structpb.NewStringValue("not empty"),
			ObservedValue: structpb.NewStringValue("empty"),
			Remediation: &pb.Remediation{
				Description: fmt.Sprintf(
					"Cluster [%q](https://console.cloud.google.com/kubernetes/list/overview?project=%s) does not have a [Master Authorized Network](https://cloud.google.com/kubernetes-engine/docs/how-to/authorized-networks#create_cluster) set. Without this setting, the cluster control plane is accessible to anyone",
					getGcpReadableResourceName(rsrc.Name),
					constants.ResourceWithoutProjectsPrefix(rsrc.ResourceGroupName),
				),
				Recommendation: fmt.Sprintf(
					"Set a [Master Authorized Network](https://cloud.google.com/kubernetes-engine/docs/how-to/authorized-networks#create_cluster) network range for cluster [%q](https://console.cloud.google.com/kubernetes/list/overview?project=%s)",
					getGcpReadableResourceName(rsrc.Name),
					constants.ResourceWithoutProjectsPrefix(rsrc.ResourceGroupName),
				),
			},
		}
		obs = append(obs, ob)
	}
	return obs, nil
}

func (r *MasterAuthorizedNetworksNotSetRule) Info() *model.RuleInfo {
	return &r.info
}
