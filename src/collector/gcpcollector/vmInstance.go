package gcpcollector

import (
	"golang.org/x/net/context"
	"github.com/nianticlabs/modron/src/pb"
)

func (collector *GCPCollector) ListVmInstances(ctx context.Context, resourceGroup *pb.Resource) ([]*pb.Resource, error) {
	resInstances, err := collector.api.ListInstances(resourceGroup.Name)
	if err != nil {
		return nil, err
	}
	vmInstances := []*pb.Resource{}
	for _, instanceList := range resInstances.Items {
		for _, instance := range instanceList.Instances {
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
				Uid:               collector.getNewUid(),
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
	}
	return vmInstances, nil
}
