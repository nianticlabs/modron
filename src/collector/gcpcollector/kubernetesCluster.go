package gcpcollector

import (
	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/pb"

	"golang.org/x/net/context"
)

func (collector *GCPCollector) ListKubernetesClusters(ctx context.Context, resourceGroup *pb.Resource) ([]*pb.Resource, error) {
	kubernetesClusters := []*pb.Resource{}
	clusters, err := collector.api.ListClustersByZone(resourceGroup.Name, "-")
	if err != nil {
		return nil, err
	}
	for _, cluster := range clusters {
		nodeVersion := ""
		for _, nodePool := range cluster.NodePools {
			nodeVersion = nodePool.Version
		}
		masterAuthorizedNetworks := []string{}
		if cluster.MasterAuthorizedNetworksConfig != nil {
			for _, cidrBlock := range cluster.MasterAuthorizedNetworksConfig.CidrBlocks {
				masterAuthorizedNetworks = append(masterAuthorizedNetworks, cidrBlock.CidrBlock)
			}
		}
		privateCluster := false
		if cluster.PrivateClusterConfig != nil {
			privateCluster = cluster.PrivateClusterConfig.EnablePrivateNodes
		}

		kubernetesClusters = append(kubernetesClusters, &pb.Resource{
			Uid:               common.GetUUID(3),
			ResourceGroupName: resourceGroup.Name,
			Name:              cluster.Name,
			Parent:            resourceGroup.Name,
			Type: &pb.Resource_KubernetesCluster{
				KubernetesCluster: &pb.KubernetesCluster{
					Location:                 cluster.Location,
					PrivateCluster:           privateCluster,
					MasterAuthorizedNetworks: masterAuthorizedNetworks,
					MasterVersion:            cluster.CurrentMasterVersion,
					NodesVersion:             nodeVersion,
				},
			},
		})
	}

	return kubernetesClusters, nil
}
