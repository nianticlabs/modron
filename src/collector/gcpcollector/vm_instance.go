package gcpcollector

import (
	"github.com/nianticlabs/modron/src/common"
	pb "github.com/nianticlabs/modron/src/proto/generated"

	"golang.org/x/net/context"
)

func (collector *GCPCollector) ListVMInstances(ctx context.Context, rgName string) (vmInstances []*pb.Resource, err error) {
	instances, err := collector.api.ListInstances(ctx, rgName)
	if err != nil {
		return nil, err
	}
	for _, instance := range instances {
		name := instance.Name
		privateIP, publicIP := "", ""
		for _, networkInterface := range instance.NetworkInterfaces {
			privateIP = networkInterface.NetworkIP
			for _, accessConfig := range networkInterface.AccessConfigs {
				publicIP = accessConfig.NatIP
			}
		}
		serviceAccountName := ""
		for _, sa := range instance.ServiceAccounts {
			serviceAccountName = sa.Email
		}
		vmInstances = append(vmInstances, &pb.Resource{
			Uid:               common.GetUUID(uuidGenRetries),
			ResourceGroupName: rgName,
			Name:              name,
			Parent:            rgName,
			Type: &pb.Resource_VmInstance{
				VmInstance: &pb.VmInstance{
					PublicIp:  publicIP,
					PrivateIp: privateIP,
					Identity:  serviceAccountName,
				},
			},
		})
	}
	return vmInstances, nil
}
