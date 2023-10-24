package rules

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
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
			Resource: &pb.Resource{
				Name: "public-bucket-2-is-detected",
			},
			ObservedValue: structpb.NewStringValue("PUBLIC"),
			ExpectedValue: structpb.NewStringValue("PRIVATE"),
		},
		{
			Name: BucketIsPublicRuleName,
			Resource: &pb.Resource{
				Name: "public-bucket-1-is-detected",
			},
			ObservedValue: structpb.NewStringValue("PUBLIC"),
			ExpectedValue: structpb.NewStringValue("PRIVATE"),
		},
	}

	got := TestRuleRun(t, resources, []model.Rule{NewBucketIsPublicRule()})

	// Check that the observations are correct.
	if diff := cmp.Diff(want, got, cmp.Comparer(observationComparer), cmpopts.SortSlices(observationsSorter)); diff != "" {
		t.Errorf("CheckRules unexpected diff (-want, +got): %v", diff)
	}
}
