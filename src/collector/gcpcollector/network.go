package gcpcollector

import (
	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/pb"

	"golang.org/x/net/context"
)

var subnetworkPurposeList = map[string]struct{}{
	"PRIVATE": {},
}

func (collector *GCPCollector) ListNetworks(ctx context.Context, resourceGroup *pb.Resource) ([]*pb.Resource, error) {
	regions, err := collector.api.ListRegions(ctx, resourceGroup.Name)
	if err != nil {
		return nil, err
	}

	networkIps := map[string][]string{}
	networkIds := map[string][]uint64{}
	networkGoogleAccessV4 := map[string]bool{}
	for _, region := range regions {
		subNetworks, err := collector.api.ListSubNetworksByRegion(ctx, resourceGroup.Name, region.Name)
		if err != nil {
			return nil, err
		}
		for _, subNetwork := range subNetworks {
			networkIds[subNetwork.Name] = append(networkIds[subNetwork.Name], subNetwork.Id)
			networkIps[subNetwork.Name] = append(networkIps[subNetwork.Name], subNetwork.IpCidrRange)
			if _, ok := subnetworkPurposeList[subNetwork.Purpose]; ok {
				networkGoogleAccessV4[subNetwork.Name] = networkGoogleAccessV4[subNetwork.Name] || subNetwork.PrivateIpGoogleAccess
			} else {
				networkGoogleAccessV4[subNetwork.Name] = true
			}
		}
	}
	networks := []*pb.Resource{}
	for netName, Ips := range networkIps {
		networks = append(networks, &pb.Resource{
			Uid:               common.GetUUID(3),
			ResourceGroupName: resourceGroup.Name,
			Name:              formatResourceName(netName, networkIds[netName]),
			Parent:            resourceGroup.Name,
			Type: &pb.Resource_Network{
				Network: &pb.Network{
					GcpPrivateGoogleAccessV4: networkGoogleAccessV4[netName],
					Ips:                      Ips,
				},
			},
		})
	}
	return networks, nil
}
