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

const ClusterNodesHavePublicIps = "CLUSTER_NODES_HAVE_PUBLIC_IPS"

type ClusterNodesHavePublicIpsRule struct {
	info model.RuleInfo
}

func init() {
	AddRule(NewClusterNodesHavePublicIpsRule())
}

func NewClusterNodesHavePublicIpsRule() model.Rule {
	return &ClusterNodesHavePublicIpsRule{
		info: model.RuleInfo{
			Name: ClusterNodesHavePublicIps,
			AcceptedResourceTypes: []proto.Message{
				&pb.KubernetesCluster{},
			},
		},
	}
}

func (r *ClusterNodesHavePublicIpsRule) Check(_ context.Context, _ model.Engine, rsrc *pb.Resource) ([]*pb.Observation, []error) {
	k8s := rsrc.GetKubernetesCluster()
	var obs []*pb.Observation
	if !k8s.PrivateCluster {
		ob := &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			ResourceRef:   utils.GetResourceRef(rsrc),
			Name:          r.Info().Name,
			ExpectedValue: structpb.NewStringValue("private"),
			ObservedValue: structpb.NewStringValue("public"),
			Remediation: &pb.Remediation{
				Description: fmt.Sprintf(
					"Cluster [%q](https://console.cloud.google.com/kubernetes/list/overview?project=%s) has a public IP, which could make it accessible by anyone on the internet",
					getGcpReadableResourceName(rsrc.Name),
					constants.ResourceWithoutProjectsPrefix(rsrc.ResourceGroupName),
				),
				Recommendation: fmt.Sprintf(
					"Unless strictly needed, redeploy cluster [%q](https://console.cloud.google.com/kubernetes/list/overview?project=%s) as a [private cluster](https://cloud.google.com/kubernetes-engine/docs/how-to/private-clusters)",
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

func (r *ClusterNodesHavePublicIpsRule) Info() *model.RuleInfo {
	return &r.info
}
