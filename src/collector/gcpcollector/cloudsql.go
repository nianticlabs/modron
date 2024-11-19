package gcpcollector

import (
	sqladmin "google.golang.org/api/sqladmin/v1beta4"

	"github.com/nianticlabs/modron/src/common"
	pb "github.com/nianticlabs/modron/src/proto/generated"

	"golang.org/x/net/context"
)

func (collector *GCPCollector) ListCloudSQLDatabases(ctx context.Context, rgName string) (resources []*pb.Resource, err error) {
	dbs, err := collector.api.ListCloudSQLDatabases(ctx, rgName)
	if err != nil {
		return nil, err
	}
	for _, instance := range dbs {
		dbResource := &pb.Resource{
			Uid:               common.GetUUID(uuidGenRetries),
			ResourceGroupName: rgName,
			Name:              instance.Name,
			Parent:            rgName,
			Type: &pb.Resource_Database{
				Database: &pb.Database{
					Type:    "cloudsql",
					Version: instance.DatabaseVersion,
					Address: instance.ConnectionName,
				},
			},
		}
		if instance.Settings != nil {
			setDbResourceSettings(instance, dbResource)
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

func setDbResourceSettings(instance *sqladmin.DatabaseInstance, dbResource *pb.Resource) {
	db := dbResource.GetDatabase()
	settings := instance.Settings
	ipConfig := settings.IpConfiguration
	if ipConfig != nil {
		db.TlsRequired = ipConfig.RequireSsl
		if ipConfig.AuthorizedNetworks == nil {
			db.AuthorizedNetworksSettingAvailable = pb.Database_AUTHORIZED_NETWORKS_NOT_SET
		} else {
			db.AuthorizedNetworksSettingAvailable = pb.Database_AUTHORIZED_NETWORKS_SET
			var authorizedNetworks []string
			for _, n := range ipConfig.AuthorizedNetworks {
				authorizedNetworks = append(authorizedNetworks, n.Value)
			}
			db.AuthorizedNetworks = authorizedNetworks
		}
		if ipConfig.Ipv4Enabled {
			db.IsPublic = true
		}
	}
	if settings.StorageAutoResize != nil {
		db.AutoResize = *settings.StorageAutoResize
	}
	if settings.BackupConfiguration != nil {
		db.BackupConfig = pb.Database_BACKUP_CONFIG_MANAGED
	} else {
		db.BackupConfig = pb.Database_BACKUP_CONFIG_DISABLED
	}
	switch settings.AvailabilityType {
	case "ZONAL":
		db.AvailabilityType = pb.Database_HA_ZONAL
	case "REGIONAL":
		db.AvailabilityType = pb.Database_HA_REGIONAL
	default:
		db.AvailabilityType = pb.Database_HA_UNKNOWN
	}
}
