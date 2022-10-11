package rules

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/google/uuid"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const BucketIsPublicRuleName = "BUCKET_IS_PUBLIC"

type BucketIsPublicRule struct {
	info model.RuleInfo
}

func init() {
	AddRule(NewBucketIsPublicRule())
}

func NewBucketIsPublicRule() model.Rule {
	return &BucketIsPublicRule{
		info: model.RuleInfo{
			Name: BucketIsPublicRuleName,
			AcceptedResourceTypes: []string{
				common.ResourceBucket,
			},
		},
	}
}

func (r *BucketIsPublicRule) Check(ctx context.Context, rsrc *pb.Resource) (obs []*pb.Observation, errs []error) {
	bk := rsrc.GetBucket()

	if bk.AccessType == pb.Bucket_PUBLIC {
		ob := &pb.Observation{
			Uid:           uuid.NewString(),
			Timestamp:     timestamppb.Now(),
			Resource:      rsrc,
			Name:          r.Info().Name,
			ExpectedValue: structpb.NewStringValue(pb.Bucket_PRIVATE.String()),
			ObservedValue: structpb.NewStringValue(bk.AccessType.String()),
			Remediation: &pb.Remediation{
				Description: fmt.Sprintf(
					"Bucket [%q](https://console.cloud.google.com/storage/browser/%s) is publicly accessible",
					rsrc.Name,
					rsrc.Name,
				),
				Recommendation: fmt.Sprintf(
					"Unless strictly needed, restrict the IAM policy of bucket [%q](https://console.cloud.google.com/storage/browser/%s) to prevent unconditional access by anyone. For more details, see [here](https://cloud.google.com/storage/docs/using-public-access-prevention)",
					rsrc.Name,
					rsrc.Name,
				),
			},
		}
		obs = append(obs, ob)
	} else if bk.AccessType == pb.Bucket_ACCESS_UNKNOWN {
		glog.Warningf("unknown access type for bucket %q", rsrc.Name)
	}

	return
}

func (r *BucketIsPublicRule) Info() *model.RuleInfo {
	return &r.info
}
