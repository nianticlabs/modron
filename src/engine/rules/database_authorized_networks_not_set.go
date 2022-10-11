package rules

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/engine"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

const DatabaseAuthorizedNetworksNotSet = "DATABASE_AUTHORIZED_NETWORKS_NOT_SET"

type DatabaseAuthorizedNetworksNotSetRule struct {
	info model.RuleInfo
}

func init() {
	AddRule(NewDatabaseAuthorizedNetworksNotSetRule())
}

func NewDatabaseAuthorizedNetworksNotSetRule() model.Rule {
	return &DatabaseAuthorizedNetworksNotSetRule{
		info: model.RuleInfo{
			Name: DatabaseAuthorizedNetworksNotSet,
			AcceptedResourceTypes: []string{
				common.ResourceDatabase,
			},
		},
	}
}

func (r *DatabaseAuthorizedNetworksNotSetRule) Check(ctx context.Context, rsrc *pb.Resource) ([]*pb.Observation, []error) {
	db := rsrc.GetDatabase()
	obs := []*pb.Observation{}

	if db.GetType() == "spanner" {
		return []*pb.Observation{}, nil
	}

	if db.AuthorizedNetworksSettingAvailable == pb.Database_AUTHORIZED_NETWORKS_NOT_SET {
		ob := &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			Resource:      rsrc,
			Name:          r.Info().Name,
			ExpectedValue: structpb.NewStringValue(pb.Database_AUTHORIZED_NETWORKS_SET.String()),
			ObservedValue: structpb.NewStringValue(pb.Database_AUTHORIZED_NETWORKS_NOT_SET.String()),
			Remediation: &pb.Remediation{
				Description: fmt.Sprintf(
					"Database %s is reachable from any IP on the Internet.",
					engine.GetGcpReadableResourceName(rsrc.Name),
				),
				Recommendation: fmt.Sprintf(
					"Enable the authorized network setting in the database settings to restrict what networks can access %s.",
					engine.GetGcpReadableResourceName(rsrc.Name),
				),
			},
		}
		obs = append(obs, ob)
	}
	return obs, nil
}

func (r *DatabaseAuthorizedNetworksNotSetRule) Info() *model.RuleInfo {
	return &r.info
}
