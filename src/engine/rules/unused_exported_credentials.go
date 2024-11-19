package rules

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"
)

const (
	unusedExportedCredentials = "UNUSED_EXPORTED_CREDENTIALS" //nolint:gosec
	// If you increase this value, also fetch the metric over a longer timeframe in the collector.
	oldestUsageVerificationMonths = 3
	sevenDays                     = 7 * time.Hour * 24
)

const (
	oneMonth = time.Hour * 24 * 30
)

var oldestUsage = time.Now().Add(time.Duration(-oldestUsageVerificationMonths) * oneMonth)

type UnusedExportedCredentialsRule struct {
	info model.RuleInfo
}

func init() {
	AddRule(NewUnusedExportedCredentialsRule())
}

func NewUnusedExportedCredentialsRule() model.Rule {
	return &UnusedExportedCredentialsRule{
		info: model.RuleInfo{
			Name: unusedExportedCredentials,
			AcceptedResourceTypes: []proto.Message{
				&pb.ExportedCredentials{},
			},
		},
	}
}

func (r *UnusedExportedCredentialsRule) Check(_ context.Context, _ model.Engine, rsrc *pb.Resource) ([]*pb.Observation, []error) {
	ec := rsrc.GetExportedCredentials()
	var obs []*pb.Observation
	if ec.LastUsage == nil {
		// If there is no last usage value, we don't report anything.
		return []*pb.Observation{}, []error{}
	}
	if time.Since(ec.CreationDate.AsTime()) > sevenDays && ec.LastUsage.AsTime().Before(oldestUsage) {
		ob := &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			ResourceRef:   utils.GetResourceRef(rsrc),
			Name:          r.Info().Name,
			ExpectedValue: structpb.NewStringValue(fmt.Sprintf("%s <", oldestUsage.Format(time.RFC3339))),
			ObservedValue: structpb.NewStringValue(ec.LastUsage.AsTime().Format(time.RFC3339)),
			Remediation: &pb.Remediation{
				Description: fmt.Sprintf(
					"Exported key `%s` of [%q](https://console.cloud.google.com/iam-admin/serviceaccounts/details/%s/keys?project=%s) has not been used in the last %d months",
					utils.GetKeyID(rsrc.Name),
					utils.GetServiceAccountNameFromKeyRef(rsrc.Name),
					utils.GetServiceAccountNameFromKeyRef(rsrc.Name),
					utils.StripProjectsPrefix(rsrc.ResourceGroupName),
					oldestUsageVerificationMonths,
				),
				Recommendation: fmt.Sprintf(
					"Consider deleting the exported key `%s` of [%q](https://console.cloud.google.com/iam-admin/serviceaccounts/details/%s/keys?project=%s) which is no longer in use",
					utils.GetKeyID(rsrc.Name),
					utils.GetServiceAccountNameFromKeyRef(rsrc.Name),
					utils.GetServiceAccountNameFromKeyRef(rsrc.Name),
					utils.StripProjectsPrefix(rsrc.ResourceGroupName),
				),
			},
			Severity: pb.Severity_SEVERITY_HIGH,
		}
		obs = append(obs, ob)
	}
	return obs, nil
}

func (r *UnusedExportedCredentialsRule) Info() *model.RuleInfo {
	return &r.info
}
