package gcpcollector

import (
	"golang.org/x/net/context"
	"github.com/nianticlabs/modron/src/pb"
)

var (
	subnetworkPurposeList = map[string]struct{}{
		"PRIVATE": {},
	}
)

func (collector *GCPCollector) ListNetworks(ctx context.Context, resourceGroup *pb.Resource) ([]*pb.Resource, error) {
	resRegions, err := collector.api.ListRegions(resourceGroup.Name)
	if err != nil {
		return nil, err
	}

	networkIps := map[string][]string{}
	networkIds := map[string][]uint64{}
	networkGoogleAccessV4 := map[string]bool{}
	for _, region := range resRegions.Items {
		subNetworks, err := collector.api.ListSubNetworksByRegion(resourceGroup.Name, region.Name)
		if err != nil {
			return nil, err
		}
		for _, subNetwork := range subNetworks.Items {
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
			Uid:               collector.getNewUid(),
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
