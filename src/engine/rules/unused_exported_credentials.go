package rules

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/engine"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

const (
	UnusedExportedCredentials = "UNUSED_EXPORTED_CREDENTIALS"
	// If you increase this value, also fetch the metric over a longer timeframe in the collector.
	oldestUsageVerificationMonths = 3
)

var (
	oldestUsage = time.Now().Add(time.Duration(-oldestUsageVerificationMonths) * time.Hour * 24 * 30)
)

type UnusedExportedCredentialsRule struct {
	info model.RuleInfo
}

func init() {
	AddRule(NewUnusedExportedCredentialsRule())
}

func NewUnusedExportedCredentialsRule() model.Rule {
	return &UnusedExportedCredentialsRule{
		info: model.RuleInfo{
			Name: UnusedExportedCredentials,
			AcceptedResourceTypes: []string{
				common.ResourceExportedCredentials,
			},
		},
	}
}

func (r *UnusedExportedCredentialsRule) Check(ctx context.Context, rsrc *pb.Resource) ([]*pb.Observation, []error) {
	ec := rsrc.GetExportedCredentials()
	obs := []*pb.Observation{}

	if ec.LastUsage.AsTime().Before(oldestUsage) {
		ob := &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			Resource:      rsrc,
			Name:          r.Info().Name,
			ExpectedValue: structpb.NewStringValue(fmt.Sprintf("%s <", oldestUsage.Format(time.RFC3339))),
			ObservedValue: structpb.NewStringValue(ec.LastUsage.AsTime().Format(time.RFC3339)),
			Remediation: &pb.Remediation{
				Description: fmt.Sprintf(
					"Exported key [%q](https://console.cloud.google.com/apis/credentials?project=%s) has not been used in the last %d months",
					engine.GetGcpReadableResourceName(rsrc.Name),
					rsrc.ResourceGroupName,
					oldestUsageVerificationMonths,
				),
				Recommendation: fmt.Sprintf(
					"Consider deleting the exported key [%q](https://console.cloud.google.com/apis/credentials?project=%s), which is no longer in use",
					engine.GetGcpReadableResourceName(rsrc.Name),
					rsrc.ResourceGroupName,
				),
			},
		}
		obs = append(obs, ob)
	}
	return obs, nil
}

func (r *UnusedExportedCredentialsRule) Info() *model.RuleInfo {
	return &r.info
}
