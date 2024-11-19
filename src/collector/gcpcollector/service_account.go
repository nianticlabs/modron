package gcpcollector

import (
	"fmt"
	"time"

	"google.golang.org/api/iam/v1"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/constants"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/utils"

	"golang.org/x/net/context"
	"google.golang.org/api/monitoring/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	serviceAccountResourcePath      = "%s/serviceAccounts/%s"
	serviceAccountKeyUnusedMaxDelay = "100d"
	monitoringPageSize              = 1000
)

func toIamPolicy(policy *iam.Policy) *pb.IamPolicy {
	if policy == nil {
		return nil
	}
	iamPolicy := &pb.IamPolicy{}
	for _, binding := range policy.Bindings {
		role := constants.ToRole(binding.Role).String()
		iamPolicy.Permissions = append(iamPolicy.Permissions, &pb.Permission{
			Role:       role,
			Principals: binding.Members,
		})
	}
	return iamPolicy
}

func (collector *GCPCollector) ListServiceAccounts(ctx context.Context, rgName string) (serviceAccounts []*pb.Resource, err error) {
	name := constants.ResourceWithProjectsPrefix(rgName)
	serviceAccountsList, err := collector.api.ListServiceAccount(ctx, name)
	if err != nil {
		return nil, err
	}
	for _, account := range serviceAccountsList {
		iamPolicy, err := collector.api.GetServiceAccountIAMPolicy(ctx, account.Name)
		if err != nil {
			log.Warnf("cannot get IAM policies for service account %q: %v", account.Email, err)
		}
		serviceAccounts = append(serviceAccounts, &pb.Resource{
			Uid:               common.GetUUID(uuidGenRetries),
			ResourceGroupName: rgName,
			Name:              account.Email,
			Parent:            rgName,
			Type: &pb.Resource_ServiceAccount{
				ServiceAccount: &pb.ServiceAccount{
					ExportedCredentials: []*pb.ExportedCredentials{},
				},
			},
			IamPolicy: toIamPolicy(iamPolicy),
		})
	}
	resources := serviceAccounts
	for _, serviceAccount := range serviceAccounts {
		userKeys, err := collector.GetServiceAccountKeys(ctx, rgName, serviceAccount)
		if err != nil {
			return nil, err
		}
		resources = append(resources, userKeys...)
	}
	return resources, nil
}

// Note: also update the Service Account passed by reference by adding its service accountkeys
func (collector *GCPCollector) GetServiceAccountKeys(ctx context.Context, rgName string, serviceAccount *pb.Resource) (userKeys []*pb.Resource, err error) {
	name := fmt.Sprintf(serviceAccountResourcePath, constants.ResourceWithProjectsPrefix(rgName), serviceAccount.Name)

	keys, err := collector.api.ListServiceAccountKeys(ctx, name)
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		if key.KeyType != "USER_MANAGED" {
			continue
		}

		keyCreationDate, err := time.Parse(time.RFC3339, key.ValidAfterTime)
		if err != nil {
			return nil, fmt.Errorf("serviceAccountKey.CreationDate: %w", err)
		}
		keyExpirationDate, err := time.Parse(time.RFC3339, key.ValidBeforeTime)
		if err != nil {
			return nil, fmt.Errorf("serviceAccountKey.ExpirationDate: %w", err)
		}

		keyExported := &pb.ExportedCredentials{
			CreationDate:   timestamppb.New(keyCreationDate),
			ExpirationDate: timestamppb.New(keyExpirationDate),
		}

		keyID := utils.GetKeyID(key.Name)
		lastUsage, err := collector.GetServiceAccountKeyLastUsage(ctx, rgName, keyID)
		if err != nil {
			log.Warnf("cannot get key usage %q: %v", key.Name, err)
			// Need to return here as otherwise the object is created with the default time value EPOCH.
			return nil, fmt.Errorf("serviceAccountKey usage: %w", err)
		}
		keyExported.LastUsage = timestamppb.New(lastUsage)
		serviceAccount.GetServiceAccount().ExportedCredentials = append(serviceAccount.GetServiceAccount().ExportedCredentials, keyExported)
		userKeys = append(userKeys, &pb.Resource{
			Uid:               common.GetUUID(uuidGenRetries),
			ResourceGroupName: rgName,
			Name:              key.Name,
			Parent:            serviceAccount.Name,
			Type: &pb.Resource_ExportedCredentials{
				ExportedCredentials: keyExported,
			},
		})
	}
	return userKeys, nil
}

func (collector *GCPCollector) GetServiceAccountKeyLastUsage(ctx context.Context, rgName string, keyID string) (time.Time, error) {
	request := new(monitoring.QueryTimeSeriesRequest)
	request.Query = fmt.Sprintf(
		"fetch iam_service_account | metric 'iam.googleapis.com/service_account/key/authn_events_count' | filter (metric.key_id == '%s') | within %s",
		keyID,
		serviceAccountKeyUnusedMaxDelay,
	)
	request.PageSize = monitoringPageSize

	lastUsage := time.Time{}
	err := collector.api.ListServiceAccountKeyUsage(ctx, rgName, request).
		Pages(ctx, func(qtsr *monitoring.QueryTimeSeriesResponse) error {
			for _, timeData := range qtsr.TimeSeriesData {
				for _, pd := range timeData.PointData {
					timeF, err := time.Parse(time.RFC3339, pd.TimeInterval.EndTime)
					if err != nil {
						return fmt.Errorf("GetServiceAccountKeyUsage.convertingTime: %w", err)
					}
					if timeF.After(lastUsage) {
						lastUsage = timeF
					}
				}
			}
			return nil
		})
	if err != nil {
		return time.Time{}, fmt.Errorf("GetServiceAccountKeyUsage: query: %q, err: %w", request.Query, err)
	}
	return lastUsage, nil
}
