package gcpcollector

import (
	"golang.org/x/net/context"
	"github.com/nianticlabs/modron/src/pb"
)

func (collector *GCPCollector) ListKubernetesClusters(ctx context.Context, resourceGroup *pb.Resource) ([]*pb.Resource, error) {
	kubernetesClusters := []*pb.Resource{}
	clusters, err := collector.api.ListClustersByZone(resourceGroup.Name, "-")
	if err != nil {
		return nil, err
	}
	for _, cluster := range clusters.Clusters {
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
			Uid:               collector.getNewUid(),
			ResourceGroupName: resourceGroup.Name,
			Name:              formatResourceName(cluster.Name, cluster.Id),
			Parent:            resourceGroup.Name,
			Type: &pb.Resource_KubernetesCluster{
				KubernetesCluster: &pb.KubernetesCluster{
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
