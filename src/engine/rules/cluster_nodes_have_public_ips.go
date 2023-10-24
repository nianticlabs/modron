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
			AcceptedResourceTypes: []string{
				common.ResourceKubernetesCluster,
			},
		},
	}
}

func (r *ClusterNodesHavePublicIpsRule) Check(ctx context.Context, rsrc *pb.Resource) ([]*pb.Observation, []error) {
	k8s := rsrc.GetKubernetesCluster()
	obs := []*pb.Observation{}
	if !k8s.PrivateCluster {
		ob := &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			Resource:      rsrc,
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
		}
		obs = append(obs, ob)
	}

	return obs, nil
}

func (r *ClusterNodesHavePublicIpsRule) Info() *model.RuleInfo {
	return &r.info
}
