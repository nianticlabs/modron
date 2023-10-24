package gcpcollector

import (
	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/pb"

	"golang.org/x/net/context"
)

func (collector *GCPCollector) ListVmInstances(ctx context.Context, resourceGroup *pb.Resource) ([]*pb.Resource, error) {
	instances, err := collector.api.ListInstances(ctx, resourceGroup.Name)
	if err != nil {
		return nil, err
	}
	vmInstances := []*pb.Resource{}
	for _, instance := range instances {
		name := instance.Name
		privateIp, publicIp := "", ""
		for _, networkInterface := range instance.NetworkInterfaces {
			privateIp = networkInterface.NetworkIP
			for _, accessConfig := range networkInterface.AccessConfigs {
				publicIp = accessConfig.NatIP
			}
		}
		serviceAccountName := ""
		for _, sa := range instance.ServiceAccounts {
			serviceAccountName = sa.Email
		}
		vmInstances = append(vmInstances, &pb.Resource{
			Uid:               common.GetUUID(3),
			ResourceGroupName: resourceGroup.Name,
			Name:              formatResourceName(name, instance.Id),
			Parent:            resourceGroup.Name,
			Type: &pb.Resource_VmInstance{
				VmInstance: &pb.VmInstance{
					PublicIp:  publicIp,
					PrivateIp: privateIp,
					Identity:  serviceAccountName,
				},
			},
		})
	}
	return vmInstances, nil
}
