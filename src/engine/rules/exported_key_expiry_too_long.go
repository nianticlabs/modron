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
	ExportedKeyIsTooOld = "EXPORTED_KEY_EXPIRY_TOO_LONG"
	timeFormat          = "2006-01-02 15:04:05 +0000 UTC"
)

type ExportedKeyIsTooOldRule struct {
	info model.RuleInfo
}

func init() {
	AddRule(NewExportedKeyIsTooOldRule())
}

func NewExportedKeyIsTooOldRule() model.Rule {
	return &ExportedKeyIsTooOldRule{
		info: model.RuleInfo{
			Name: ExportedKeyIsTooOld,
			AcceptedResourceTypes: []string{
				common.ResourceExportedCredentials,
			},
		},
	}
}

func (r *ExportedKeyIsTooOldRule) Check(ctx context.Context, rsrc *pb.Resource) ([]*pb.Observation, []error) {
	expiryMonths := 6
	ec := rsrc.GetExportedCredentials()
	obs := []*pb.Observation{}

	if ec.CreationDate.AsTime().Before(time.Now().AddDate(0, -expiryMonths, 1)) {
		ob := &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			Resource:      rsrc,
			Name:          r.Info().Name,
			ExpectedValue: structpb.NewStringValue("later creation date"),
			ObservedValue: structpb.NewStringValue(ec.CreationDate.AsTime().Format(timeFormat)),
			Remediation: &pb.Remediation{
				Description: fmt.Sprintf(
					"Exported key [%q](https://console.cloud.google.com/apis/credentials?project=%s) is too long lived",
					engine.GetGcpReadableResourceName(rsrc.Name),
					rsrc.ResourceGroupName,
				),
				Recommendation: fmt.Sprintf(
					"Rotate the exported key [%q](https://console.cloud.google.com/apis/credentials?project=%s) every %d months.",
					engine.GetGcpReadableResourceName(rsrc.Name),
					rsrc.ResourceGroupName,
					expiryMonths,
				),
			},
		}
		obs = append(obs, ob)

	}
	return obs, nil
}

func (r *ExportedKeyIsTooOldRule) Info() *model.RuleInfo {
	return &r.info
}
