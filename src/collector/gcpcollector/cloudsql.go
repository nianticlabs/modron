package gcpcollector

import (
	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/pb"

	"golang.org/x/net/context"
)

func (collector *GCPCollector) ListCloudSqlDatabases(ctx context.Context, resourceGroup *pb.Resource) ([]*pb.Resource, error) {
	dbs, err := collector.api.ListCloudSqlDatabases(ctx, resourceGroup.Name)
	if err != nil {
		return nil, err
	}
	resources := []*pb.Resource{}
	for _, instance := range dbs {
		dbResource := &pb.Resource{
			Uid:               common.GetUUID(3),
			ResourceGroupName: resourceGroup.Name,
			Name:              instance.Name,
			Parent:            resourceGroup.Name,
			Type: &pb.Resource_Database{
				Database: &pb.Database{
					Type:    "cloudsql",
					Version: instance.DatabaseVersion,
					Address: instance.ConnectionName,
				},
			},
		}
		if instance.Settings != nil {
			if instance.Settings.IpConfiguration != nil {
				dbResource.GetDatabase().TlsRequired = instance.Settings.IpConfiguration.RequireSsl
				if instance.Settings.IpConfiguration.AuthorizedNetworks != nil {
					dbResource.GetDatabase().AuthorizedNetworksSettingAvailable = pb.Database_AUTHORIZED_NETWORKS_SET
					authorizedNetworks := []string{}
					for _, n := range instance.Settings.IpConfiguration.AuthorizedNetworks {
						authorizedNetworks = append(authorizedNetworks, n.Value)
					}
					dbResource.GetDatabase().AuthorizedNetworks = authorizedNetworks
				} else {
					dbResource.GetDatabase().AuthorizedNetworksSettingAvailable = pb.Database_AUTHORIZED_NETWORKS_NOT_SET
				}
				if instance.Settings.IpConfiguration.Ipv4Enabled {
					dbResource.GetDatabase().IsPublic = true
				}
			}
			if instance.Settings.StorageAutoResize != nil {
				dbResource.GetDatabase().AutoResize = *instance.Settings.StorageAutoResize
			}
			if instance.Settings.BackupConfiguration != nil {
				dbResource.GetDatabase().BackupConfig = pb.Database_BACKUP_CONFIG_MANAGED
			} else {
				dbResource.GetDatabase().BackupConfig = pb.Database_BACKUP_CONFIG_DISABLED
			}
			switch instance.Settings.AvailabilityType {
			case "ZONAL":
				dbResource.GetDatabase().AvailabilityType = pb.Database_HA_ZONAL
			case "REGIONAL":
				dbResource.GetDatabase().AvailabilityType = pb.Database_HA_REGIONAL
			default:
				dbResource.GetDatabase().AvailabilityType = pb.Database_HA_UNKNOWN
			}
		}
		if instance.DiskEncryptionStatus != nil {
			dbResource.GetDatabase().Encryption = pb.Database_ENCRYPTION_USER_MANAGED
		} else {
			dbResource.GetDatabase().Encryption = pb.Database_ENCRYPTION_MANAGED
		}

		resources = append(resources, dbResource)
	}

	return resources, nil
}
