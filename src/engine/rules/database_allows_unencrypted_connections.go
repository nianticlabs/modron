package rules

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

const DatabaseAllowsUnencryptedConnections = "DATABASE_ALLOWS_UNENCRYPTED_CONNECTIONS"

type DatabaseAllowsUnencryptedConnectionsRule struct {
	info model.RuleInfo
}

func init() {
	AddRule(NewDatabaseAllowsUnencryptedConnectionsRule())
}

func NewDatabaseAllowsUnencryptedConnectionsRule() model.Rule {
	return &DatabaseAllowsUnencryptedConnectionsRule{
		info: model.RuleInfo{
			Name: DatabaseAllowsUnencryptedConnections,
			AcceptedResourceTypes: []string{
				common.ResourceDatabase,
			},
		},
	}
}

func (r *DatabaseAllowsUnencryptedConnectionsRule) Check(ctx context.Context, rsrc *pb.Resource) ([]*pb.Observation, []error) {
	db := rsrc.GetDatabase()
	obs := []*pb.Observation{}

	if db.GetType() == "spanner" {
		return []*pb.Observation{}, nil
	}

	if !db.GetTlsRequired() {
		ob := &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			Resource:      rsrc,
			Name:          r.Info().Name,
			ExpectedValue: structpb.NewBoolValue(true),
			ObservedValue: structpb.NewBoolValue(false),
			Remediation: &pb.Remediation{
				Description: fmt.Sprintf(
					"Database %s allows for unencrypted connections.",
					getGcpReadableResourceName(rsrc.Name),
				),
				Recommendation: fmt.Sprintf(
					"Enable the require SSL setting in the database settings to allow only encrypted connections to %s.",
					getGcpReadableResourceName(rsrc.Name),
				),
			},
		}
		obs = append(obs, ob)
	}
	return obs, nil
}

func (r *DatabaseAllowsUnencryptedConnectionsRule) Info() *model.RuleInfo {
	return &r.info
}
