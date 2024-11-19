package gcpcollector

import (
	"fmt"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/sync/errgroup"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/constants"
	pb "github.com/nianticlabs/modron/src/proto/generated"

	"golang.org/x/net/context"
)

var subnetworkPurposeList = map[string]struct{}{
	"PRIVATE": {},
}

func (collector *GCPCollector) ListNetworks(ctx context.Context, rgName string) (networks []*pb.Resource, err error) {
	ctx, span := tracer.Start(ctx, "ListNetworks")
	span.SetAttributes(
		attribute.String(constants.TraceKeyResourceGroup, rgName),
	)
	defer span.End()
	regions, err := collector.api.ListRegions(ctx, rgName)
	if err != nil {
		return nil, err
	}
	errGroup := new(errgroup.Group)
	networkIPs := sync.Map{}
	networkGoogleAccessV4 := sync.Map{}
	for _, region := range regions {
		errGroup.Go(func() error {
			return collector.fetchRegion(ctx, rgName, region.Name, &networkIPs, &networkGoogleAccessV4)
		})
	}
	if err := errGroup.Wait(); err != nil {
		return nil, fmt.Errorf("failed to fetch regions: %w", err)
	}
	networkIPs.Range(func(netName, value interface{}) bool {
		hasGoogleAccess, ok := networkGoogleAccessV4.Load(netName)
		if !ok {
			hasGoogleAccess = false
		}
		networks = append(networks, &pb.Resource{
			Uid:               common.GetUUID(uuidGenRetries),
			ResourceGroupName: rgName,
			Name:              netName.(string),
			Parent:            rgName,
			Type: &pb.Resource_Network{
				Network: &pb.Network{
					GcpPrivateGoogleAccessV4: hasGoogleAccess.(bool),
					Ips:                      value.([]string),
				},
			},
		})
		return true
	})
	return networks, nil
}

func (collector *GCPCollector) fetchRegion(
	ctx context.Context,
	rgName, regionName string,
	networkIPs, networkGoogleAccessV4 *sync.Map,
) error {
	subNetworks, err := collector.api.ListSubNetworksByRegion(ctx, rgName, regionName)
	if err != nil {
		return fmt.Errorf("failed to list subnetworks in region %s: %w", regionName, err)
	}
	for _, subNetwork := range subNetworks {
		netIPs, _ := networkIPs.LoadOrStore(subNetwork.Name, []string{})
		netIPs = append(netIPs.([]string), subNetwork.IpCidrRange)
		networkIPs.Store(subNetwork.Name, netIPs)

		netGoogleAccessV4, ok := networkGoogleAccessV4.Load(subNetwork.Name)
		if !ok {
			netGoogleAccessV4 = false
		}
		if _, ok := subnetworkPurposeList[subNetwork.Purpose]; ok {
			networkGoogleAccessV4.Store(subNetwork.Name, netGoogleAccessV4.(bool) || subNetwork.PrivateIpGoogleAccess)
		} else {
			networkGoogleAccessV4.Store(subNetwork.Name, true)
		}
	}
	return nil
}
