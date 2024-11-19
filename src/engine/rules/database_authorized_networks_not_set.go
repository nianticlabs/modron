package rules

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"
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
			AcceptedResourceTypes: []proto.Message{
				&pb.Database{},
			},
		},
	}
}

func (r *DatabaseAuthorizedNetworksNotSetRule) Check(_ context.Context, _ model.Engine, rsrc *pb.Resource) ([]*pb.Observation, []error) {
	db := rsrc.GetDatabase()
	obs := []*pb.Observation{}

	if db.GetType() == "spanner" {
		return []*pb.Observation{}, nil
	}

	if db.IsPublic && db.AuthorizedNetworksSettingAvailable == pb.Database_AUTHORIZED_NETWORKS_NOT_SET {
		ob := &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			ResourceRef:   utils.GetResourceRef(rsrc),
			Name:          r.Info().Name,
			ExpectedValue: structpb.NewStringValue(pb.Database_AUTHORIZED_NETWORKS_SET.String()),
			ObservedValue: structpb.NewStringValue(pb.Database_AUTHORIZED_NETWORKS_NOT_SET.String()),
			Remediation: &pb.Remediation{
				Description: fmt.Sprintf(
					"Database %s is reachable from any IP on the Internet.",
					getGcpReadableResourceName(rsrc.Name),
				),
				Recommendation: fmt.Sprintf(
					"Enable the authorized network setting in the database settings to restrict what networks can access %s.",
					getGcpReadableResourceName(rsrc.Name),
				),
			},
			Severity: pb.Severity_SEVERITY_HIGH,
		}
		obs = append(obs, ob)
	}
	return obs, nil
}

func (r *DatabaseAuthorizedNetworksNotSetRule) Info() *model.RuleInfo {
	return &r.info
}
