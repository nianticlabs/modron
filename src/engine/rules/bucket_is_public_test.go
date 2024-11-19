package rules

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
)

func TestCheckDetectsPublicBucket(t *testing.T) {
	resources := []*pb.Resource{
		{
			Name:              testProjectName,
			Parent:            "",
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_ResourceGroup{
				ResourceGroup: &pb.ResourceGroup{},
			},
		},
		{
			Name:              "public-bucket-1-is-detected",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_Bucket{
				Bucket: &pb.Bucket{
					CreationDate:      &timestamppb.Timestamp{},
					RetentionPolicy:   &pb.Bucket_RetentionPolicy{},
					EncryptionPolicy:  &pb.Bucket_EncryptionPolicy{},
					AccessType:        pb.Bucket_PUBLIC,
					AccessControlType: pb.Bucket_NON_UNIFORM,
				},
			},
		},
		{
			Name:              "private-bucket-is-not-detected",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_Bucket{
				Bucket: &pb.Bucket{
					CreationDate:      &timestamppb.Timestamp{},
					RetentionPolicy:   &pb.Bucket_RetentionPolicy{},
					EncryptionPolicy:  &pb.Bucket_EncryptionPolicy{},
					AccessType:        pb.Bucket_PRIVATE,
					AccessControlType: pb.Bucket_UNIFORM,
				},
			},
		},
		{
			Name:              "public-bucket-2-is-detected",
			Parent:            testProjectName,
			ResourceGroupName: testProjectName,
			IamPolicy:         &pb.IamPolicy{},
			Type: &pb.Resource_Bucket{
				Bucket: &pb.Bucket{
					CreationDate:      &timestamppb.Timestamp{},
					RetentionPolicy:   &pb.Bucket_RetentionPolicy{},
					EncryptionPolicy:  &pb.Bucket_EncryptionPolicy{},
					AccessType:        pb.Bucket_PUBLIC,
					AccessControlType: pb.Bucket_UNIFORM,
				},
			},
		},
	}

	want := []*pb.Observation{
		{
			Name: BucketIsPublicRuleName,
			ResourceRef: &pb.ResourceRef{
				Uid:           proto.String("uuid-1"),
				GroupName:     "projects/project-0",
				ExternalId:    proto.String("public-bucket-2-is-detected"),
				CloudPlatform: pb.CloudPlatform_GCP,
			},
			ObservedValue: structpb.NewStringValue("PUBLIC"),
			ExpectedValue: structpb.NewStringValue("PRIVATE"),
			Remediation: &pb.Remediation{
				Description:    "Bucket [\"public-bucket-2-is-detected\"](https://console.cloud.google.com/storage/browser/public-bucket-2-is-detected) is publicly accessible",
				Recommendation: "Unless strictly needed, restrict the IAM policy of bucket [\"public-bucket-2-is-detected\"](https://console.cloud.google.com/storage/browser/public-bucket-2-is-detected) to prevent unconditional access by anyone. For more details, see [here](https://cloud.google.com/storage/docs/using-public-access-prevention)",
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		},
		{
			Name: BucketIsPublicRuleName,
			ResourceRef: &pb.ResourceRef{
				Uid:           proto.String("uuid-2"),
				GroupName:     "projects/project-0",
				ExternalId:    proto.String("public-bucket-1-is-detected"),
				CloudPlatform: pb.CloudPlatform_GCP,
			},
			ObservedValue: structpb.NewStringValue("PUBLIC"),
			ExpectedValue: structpb.NewStringValue("PRIVATE"),
			Remediation: &pb.Remediation{
				Description:    "Bucket [\"public-bucket-1-is-detected\"](https://console.cloud.google.com/storage/browser/public-bucket-1-is-detected) is publicly accessible",
				Recommendation: "Unless strictly needed, restrict the IAM policy of bucket [\"public-bucket-1-is-detected\"](https://console.cloud.google.com/storage/browser/public-bucket-1-is-detected) to prevent unconditional access by anyone. For more details, see [here](https://cloud.google.com/storage/docs/using-public-access-prevention)",
			},
			Severity: pb.Severity_SEVERITY_MEDIUM,
		},
	}

	TestRuleRun(t, resources, []model.Rule{NewBucketIsPublicRule()}, want)
}
