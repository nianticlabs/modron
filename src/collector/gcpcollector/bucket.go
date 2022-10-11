package gcpcollector

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/nianticlabs/modron/src/pb"
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
func (collector *GCPCollector) ListBuckets(ctx context.Context, resourceGroup *pb.Resource) ([]*pb.Resource, error) {
	res, err := collector.api.ListBuckets(resourceGroup.Name)
	if err != nil {
		return nil, err
	}

	buckets := []*pb.Resource{}
	for _, bucket := range res.Items {
		iamPolicy, err := collector.api.ListBucketsIamPolicy(bucket.Id)
		if err != nil {
			return nil, err
		}

		accessType := pb.Bucket_ACCESS_UNKNOWN
		permissions := []*pb.Permission{}
		for _, binding := range iamPolicy.Bindings {
			for i := range binding.Members {
				binding.Members[i] = strings.TrimPrefix(binding.Members[i], "projectViewer:")
				binding.Members[i] = strings.TrimPrefix(binding.Members[i], "projectOwner:")
				binding.Members[i] = strings.TrimPrefix(binding.Members[i], "projectEditor:")
			}
			permissions = append(permissions, &pb.Permission{
				Role:       strings.TrimPrefix(binding.Role, "roles/"),
				Principals: binding.Members,
			})
			if accessType != pb.Bucket_PUBLIC {
				accessType = getAccessType(binding.Members)
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
			Uid:               collector.getNewUid(),
			ResourceGroupName: resourceGroup.Name,
			Name:              formatResourceName(bucket.Name, bucket.Id),
			Parent:            resourceGroup.Name,
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
