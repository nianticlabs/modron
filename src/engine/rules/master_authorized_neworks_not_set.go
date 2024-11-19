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
			AcceptedResourceTypes: []proto.Message{
				&pb.KubernetesCluster{},
			},
		},
	}
}

func (r *MasterAuthorizedNetworksNotSetRule) Check(_ context.Context, _ model.Engine, rsrc *pb.Resource) ([]*pb.Observation, []error) {
	k8s := rsrc.GetKubernetesCluster()
	var obs []*pb.Observation
	if len(k8s.MasterAuthorizedNetworks) < 1 {
		ob := &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			ResourceRef:   utils.GetResourceRef(rsrc),
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
			Severity: pb.Severity_SEVERITY_HIGH,
		}
		obs = append(obs, ob)
	}
	return obs, nil
}

func (r *MasterAuthorizedNetworksNotSetRule) Info() *model.RuleInfo {
	return &r.info
}
