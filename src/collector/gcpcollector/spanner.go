package gcpcollector

import (
	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/constants"
	pb "github.com/nianticlabs/modron/src/proto/generated"

	"golang.org/x/net/context"
)

func (collector *GCPCollector) ListSpannerDatabases(ctx context.Context, rgName string) (resources []*pb.Resource, err error) {
	name := constants.ResourceWithProjectsPrefix(rgName)
	dbs, err := collector.api.ListSpannerDatabases(ctx, name)
	if err != nil {
		return nil, err
	}
	for _, database := range dbs {
		dbResource := &pb.Resource{
			// TODO: Collect IAM Policy
			Uid:               common.GetUUID(uuidGenRetries),
			ResourceGroupName: rgName,
			Name:              database.Name,
			Parent:            rgName,
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
