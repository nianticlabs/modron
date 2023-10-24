package gcpcollector

import (
	"fmt"
	"strings"
	"time"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/pb"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/api/monitoring/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	serviceAccountResourcePath      = "%s/serviceAccounts/%s"
	serviceAccountKeyUnusedMaxDelay = "100d"
)

func (collector *GCPCollector) ListServiceAccounts(ctx context.Context, resourceGroup *pb.Resource) ([]*pb.Resource, error) {
	name := constants.ResourceWithProjectsPrefix(resourceGroup.Name)
	serviceAccountsList, err := collector.api.ListServiceAccount(ctx, name)
	if err != nil {
		return nil, err
	}
	serviceAccounts := []*pb.Resource{}
	for _, account := range serviceAccountsList {
		serviceAccounts = append(serviceAccounts, &pb.Resource{
			// TODO: collect IAM policies
			Uid:               common.GetUUID(3),
			ResourceGroupName: resourceGroup.Name,
			Name:              account.Email,
			Parent:            resourceGroup.Name,
			Type: &pb.Resource_ServiceAccount{
				ServiceAccount: &pb.ServiceAccount{
					ExportedCredentials: []*pb.ExportedCredentials{},
				},
			},
		})
	}
	resources := serviceAccounts
	for _, serviceAccount := range serviceAccounts {
		userKeys, err := collector.GetServiceAccountKeys(ctx, resourceGroup, serviceAccount)
		if err != nil {
			return nil, err
		}
		resources = append(resources, userKeys...)
	}
	return resources, nil
}

// Note: also update the Service Account passed by reference by adding its service accountkeys
func (collector *GCPCollector) GetServiceAccountKeys(ctx context.Context, resourceGroup *pb.Resource, serviceAccount *pb.Resource) ([]*pb.Resource, error) {
	userKeys := []*pb.Resource{}

	name := fmt.Sprintf(serviceAccountResourcePath, constants.ResourceWithProjectsPrefix(resourceGroup.Name), serviceAccount.Name)

	keys, err := collector.api.ListServiceAccountKeys(name)
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

		splitName := strings.Split(key.Name, "/")
		keyId := splitName[len(splitName)-1]
		lastUsage, err := collector.GetServiceAccountKeyLastUsage(ctx, resourceGroup, keyId)
		if err != nil {
			glog.Warningf("cannot get key usage %q: %v", key.Name, err)
			// Need to return here as otherwise the object is created with the default time value EPOCH.
			return nil, fmt.Errorf("serviceAccountKey usage: %v", err)
		} else {
			keyExported.LastUsage = timestamppb.New(lastUsage)
		}

		serviceAccount.GetServiceAccount().ExportedCredentials = append(serviceAccount.GetServiceAccount().ExportedCredentials, keyExported)

		userKeys = append(userKeys, &pb.Resource{
			Uid:               common.GetUUID(3),
			ResourceGroupName: resourceGroup.Name,
			Name:              formatResourceName(key.Name, key.Name),
			Parent:            serviceAccount.Name,
			Type: &pb.Resource_ExportedCredentials{
				ExportedCredentials: keyExported,
			},
		})
	}
	return userKeys, nil
}

func (collector *GCPCollector) GetServiceAccountKeyLastUsage(ctx context.Context, resourceGroup *pb.Resource, keyID string) (time.Time, error) {
	request := new(monitoring.QueryTimeSeriesRequest)
	request.Query = fmt.Sprintf(
		"fetch iam_service_account | metric 'iam.googleapis.com/service_account/key/authn_events_count' | filter (metric.key_id == '%s') | within %s",
		keyID,
		serviceAccountKeyUnusedMaxDelay)

	lastUsage := time.Time{}
	err := collector.api.ListServiceAccountKeyUsage(ctx, resourceGroup.Name, request).Pages(ctx, func(qtsr *monitoring.QueryTimeSeriesResponse) error {
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
		return time.Time{}, fmt.Errorf("GetServiceAccountKeyUsage: query: %q, err: %v", request.Query, err)
	}
	return lastUsage, nil
}
