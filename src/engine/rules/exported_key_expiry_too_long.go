package rules

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"
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
			AcceptedResourceTypes: []proto.Message{
				&pb.ExportedCredentials{},
			},
		},
	}
}

func (r *ExportedKeyIsTooOldRule) Check(_ context.Context, _ model.Engine, rsrc *pb.Resource) ([]*pb.Observation, []error) {
	expiryMonths := 6
	ec := rsrc.GetExportedCredentials()
	obs := []*pb.Observation{}

	if ec.CreationDate.AsTime().Before(time.Now().AddDate(0, -expiryMonths, 1)) {
		ob := &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			ResourceRef:   utils.GetResourceRef(rsrc),
			Name:          r.Info().Name,
			ExpectedValue: structpb.NewStringValue("later creation date"),
			ObservedValue: structpb.NewStringValue(ec.CreationDate.AsTime().Format(timeFormat)),
			Remediation: &pb.Remediation{
				Description: fmt.Sprintf(
					"Exported key [%q](https://console.cloud.google.com/apis/credentials?project=%s) is too long lived",
					getGcpReadableResourceName(rsrc.Name),
					constants.ResourceWithoutProjectsPrefix(rsrc.ResourceGroupName),
				),
				Recommendation: fmt.Sprintf(
					"Rotate the exported key [%q](https://console.cloud.google.com/apis/credentials?project=%s) every %d months",
					getGcpReadableResourceName(rsrc.Name),
					constants.ResourceWithoutProjectsPrefix(rsrc.ResourceGroupName),
					expiryMonths,
				),
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		}
		obs = append(obs, ob)

	}
	return obs, nil
}

func (r *ExportedKeyIsTooOldRule) Info() *model.RuleInfo {
	return &r.info
}
