package gcpcollector

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/nianticlabs/modron/src/common"
	pb "github.com/nianticlabs/modron/src/proto/generated"

	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	customerManagedKeyRegex = regexp.MustCompile("^projects/.+/locations/.+/keyRings/.+/cryptoKeys/.+$")
	isPublicRegex           = regexp.MustCompile("^(allUsers|allAuthenticatedUsers)$")
)

func getAccessType(members []string) pb.Bucket_AccessType {
	for _, member := range members {
		if isPublicRegex.Match([]byte(member)) {
			return pb.Bucket_PUBLIC
		}
	}
	return pb.Bucket_PRIVATE
}

// TODO: Check the ACL to detect if the bucket is public if uniform bucket-level access is disabled.
func (collector *GCPCollector) ListBuckets(ctx context.Context, rgName string) (buckets []*pb.Resource, err error) {
	res, err := collector.api.ListBuckets(ctx, rgName)
	if err != nil {
		return nil, err
	}

	removeDefaultBindings := func(members []string) (filteredList []string) {
		for _, member := range members {
			if strings.HasPrefix(member, "projectViewer:") || strings.HasPrefix(member, "projectOwner:") || strings.HasPrefix(member, "projectEditor:") {
				continue
			}
			filteredList = append(filteredList, member)
		}
		return
	}
	for _, bucket := range res {
		iamPolicy, err := collector.api.ListBucketsIamPolicy(bucket.Id)
		if err != nil {
			return nil, err
		}

		accessType := pb.Bucket_ACCESS_UNKNOWN
		var permissions []*pb.Permission
		for _, binding := range iamPolicy.Bindings {
			bindingMembers := removeDefaultBindings(binding.Members)
			permissions = append(permissions, &pb.Permission{
				Role:       strings.TrimPrefix(binding.Role, "roles/"),
				Principals: bindingMembers,
			})
			if accessType != pb.Bucket_PUBLIC {
				accessType = getAccessType(bindingMembers)
			}
		}
		creationDate, err := time.Parse(time.RFC3339, bucket.TimeCreated)
		if err != nil {
			return nil, fmt.Errorf("unable to parse creation timestamp of bucket %q", bucket.Name)
		}
		isKeyCustomerManaged := false
		if bucket.Encryption != nil {
			isKeyCustomerManaged = customerManagedKeyRegex.Match([]byte(bucket.Encryption.DefaultKmsKeyName))
		}
		var retentionPolicy *pb.Bucket_RetentionPolicy
		if bucket.RetentionPolicy != nil {
			retentionPeriod, err := time.ParseDuration(fmt.Sprintf("%ds", bucket.RetentionPolicy.RetentionPeriod))
			if err != nil {
				return nil, fmt.Errorf("unable to parse retention period of bucket %q", bucket.Name)
			}
			retentionPolicy = &pb.Bucket_RetentionPolicy{
				Period:   durationpb.New(retentionPeriod),
				IsLocked: bucket.RetentionPolicy.IsLocked,
			}
		}
		accessControlType := pb.Bucket_ACCESS_CONTROL_UNKNOWN
		if iamConfig := bucket.IamConfiguration; iamConfig != nil && iamConfig.UniformBucketLevelAccess != nil {
			if iamConfig.UniformBucketLevelAccess.Enabled {
				accessControlType = pb.Bucket_UNIFORM
			} else {
				accessControlType = pb.Bucket_NON_UNIFORM
			}
		}
		buckets = append(buckets, &pb.Resource{
			Uid:               common.GetUUID(uuidGenRetries),
			ResourceGroupName: rgName,
			Name:              bucket.Name,
			Parent:            rgName,
			IamPolicy: &pb.IamPolicy{
				Permissions: permissions,
			},
			Type: &pb.Resource_Bucket{
				Bucket: &pb.Bucket{
					CreationDate:    timestamppb.New(creationDate),
					RetentionPolicy: retentionPolicy,
					EncryptionPolicy: &pb.Bucket_EncryptionPolicy{
						// SSE is always on in GCP.
						IsEnabled:            true,
						IsKeyCustomerManaged: isKeyCustomerManaged,
					},
					AccessType:        accessType,
					AccessControlType: accessControlType,
				},
			},
		})
	}
	return buckets, nil
}
