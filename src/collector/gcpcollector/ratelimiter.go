package gcpcollector

import (
	"context"
	"time"

	"golang.org/x/time/rate"
	"google.golang.org/api/cloudasset/v1"
)

const (
	// https://cloud.google.com/asset-inventory/docs/quota
	cloudAssetSearchResourcesQuotaPerMinute = 350
	cloudAssetSearchIamQuotaPerMinute       = 350
	cloudAssetSearchBurst                   = 1
	apiKeysPageSize                         = 300
	pageSize                                = 500  // Page size is capped at 500 even if a larger value is given.
	sccPageSize                             = 1000 // The maximum number of results to return in a single response. Default is 10, minimum is 1, maximum is 1000.
	sccReadBurst                            = 1
	sccReadRequestsPerMinute                = 1000
	listSubnetsRangeQuota                   = 10_000
	listSubnetsRangeBurst                   = 5_000
)

var (
	rlResources    = rate.NewLimiter(rate.Every(time.Minute/cloudAssetSearchResourcesQuotaPerMinute), cloudAssetSearchBurst)
	rlIam          = rate.NewLimiter(rate.Every(time.Minute/cloudAssetSearchIamQuotaPerMinute), cloudAssetSearchBurst)
	rlSCC          = rate.NewLimiter(rate.Every(time.Minute/sccReadRequestsPerMinute), sccReadBurst)
	rlSubnetRanges = rate.NewLimiter(rate.Every(time.Minute/listSubnetsRangeQuota), listSubnetsRangeBurst)
)

// newRateLimitedCloudAssetInventoryV1 returns a service similar to the CloudAssetV1Service that is rate limited according
// to the Cloud Asset Inventory API quotas.
func newRateLimitedCloudAssetInventoryV1(svc *cloudasset.V1Service) rateLimitedCloudAssetV1Service {
	return &rateLimitedCAI{
		svc: svc,
	}
}

type rateLimitedCAI struct {
	svc *cloudasset.V1Service
}

// rateLimitedCloudAssetV1Service is an interface that implements rate limiting on the original CloudAssetV1Service
type rateLimitedCloudAssetV1Service interface {
	SearchAllResources(ctx context.Context, scope string) (*cloudasset.V1SearchAllResourcesCall, error)
	SearchAllIamPolicies(ctx context.Context, scope string) (*cloudasset.V1SearchAllIamPoliciesCall, error)
}

func (r *rateLimitedCAI) SearchAllResources(ctx context.Context, scope string) (*cloudasset.V1SearchAllResourcesCall, error) {
	if err := rlResources.Wait(ctx); err != nil {
		return nil, err
	}
	return r.svc.SearchAllResources(scope).PageSize(pageSize), nil
}

func (r *rateLimitedCAI) SearchAllIamPolicies(ctx context.Context, scope string) (*cloudasset.V1SearchAllIamPoliciesCall, error) {
	if err := rlIam.Wait(ctx); err != nil {
		return nil, err
	}
	return r.svc.SearchAllIamPolicies(scope).PageSize(pageSize), nil
}
