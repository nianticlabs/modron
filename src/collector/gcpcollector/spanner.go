package gcpcollector

import (
	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/pb"

	"golang.org/x/net/context"
)

func (collector *GCPCollector) ListSpannerDatabases(ctx context.Context, resourceGroup *pb.Resource) ([]*pb.Resource, error) {
	name := constants.ResourceWithProjectsPrefix(resourceGroup.Name)
	dbs, err := collector.api.ListSpannerDatabases(ctx, name)
	if err != nil {
		return nil, err
	}
	resources := []*pb.Resource{}
	for _, database := range dbs {
		dbResource := &pb.Resource{
			// TODO: Collect IAM Policy
			Uid:               common.GetUUID(3),
			ResourceGroupName: resourceGroup.Name,
			Name:              database.Name,
			Parent:            resourceGroup.Name,
			Type: &pb.Resource_Database{
				Database: &pb.Database{
					Type:       "spanner",
					Address:    "spanner.googleapis.com",
					AutoResize: true,
					// Default on Spanner.
					TlsRequired: true,
				},
			},
		}
		if database.EncryptionConfig != nil {
			dbResource.GetDatabase().Encryption = pb.Database_ENCRYPTION_USER_MANAGED
		} else {
			dbResource.GetDatabase().Encryption = pb.Database_ENCRYPTION_MANAGED
		}

		resources = append(resources, dbResource)
	}

	return resources, nil
}
