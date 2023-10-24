package gcpcollector

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/nianticlabs/modron/src/pb"

	"golang.org/x/net/context"
	"google.golang.org/api/googleapi"
)

const (
	maxAttemptNumber     = 100
	maxSecBtwAttempts    = 30.
	maxSecAcrossAttempts = time.Duration(3600 * time.Second)
)

var (
	// Retry following errors:
	//   * 408: Request timeout
	//   * 429: Too many requests
	//   * 5XX: Server errors
	retryableErrorCode = []int{408, 429, 500, 502, 503, 504}
	// We are not interested in the following codes:
	//   * 403: Sometimes returned for non existing resources.
	//   * 404: A resource can be tracked by modron and then deleted.
	skipableErrorCode = []int{403, 404}
)

func getErrorCode(err error) int {
	if e, ok := err.(*googleapi.Error); ok {
		return e.Code
	}
	return 0
}

func isErrorCodeRetryable(errorCode int) bool {
	for _, code := range retryableErrorCode {
		if code == errorCode {
			return true
		}
	}
	return false
}

func isErrorCodeSkipable(errorCode int) bool {
	for _, code := range skipableErrorCode {
		if code == errorCode {
			return true
		}
	}
	return false
}

type CollectorCall func(context.Context, *pb.Resource) ([]*pb.Resource, error)

func (call CollectorCall) ExponentialBackoffRun(ctx context.Context, resource *pb.Resource) ([]*pb.Resource, error) {
	attemptNumber := 1
	waitSec := 0.
	start := time.Now()
	for {
		resources, err := call.Run(ctx, resource)
		if err == nil {
			return resources, nil
		}
		if !isErrorCodeRetryable(getErrorCode(err)) {
			return nil, fmt.Errorf("ExponentialBackoffRun.notRetryableCode: %w ", err)
		}
		if time.Since(start) > maxSecAcrossAttempts {
			return nil, fmt.Errorf("ExponentialBackoffRun.tooLong: after %v seconds %w ", time.Since(start), err)
		}
		if attemptNumber >= maxAttemptNumber {
			return nil, fmt.Errorf("ExponentialBackoffRun.maxAttempt: after %v attempts %w ", attemptNumber, err)
		}
		waitSec = math.Min((math.Pow(2, float64(attemptNumber)) + rand.Float64()), maxSecBtwAttempts)
		time.Sleep(time.Duration(waitSec) * time.Second)
		attemptNumber += 1
	}
}

func (call CollectorCall) Run(ctx context.Context, resource *pb.Resource) ([]*pb.Resource, error) {
	resources, err := call(ctx, resource)
	if isErrorCodeSkipable(getErrorCode(err)) {
		return []*pb.Resource{}, nil
	}
	return resources, err
}

// TODO: refactor with generics CollectorResourceGroupCall and CollectorCall are identical
type CollectorResourceGroupCall func(context.Context, string, string) (*pb.Resource, error)

func (call CollectorResourceGroupCall) ExponentialBackoffRun(ctx context.Context, collectId string, resource string) (*pb.Resource, error) {
	attemptNumber := 0
	waitSec := 0.
	start := time.Now()
	time.Sleep(time.Duration(rand.Float64()*2) * time.Second)
	for {
		resources, err := call.Run(ctx, collectId, resource)
		if err == nil {
			return resources, nil
		}
		if !isErrorCodeRetryable(getErrorCode(err)) {
			return nil, fmt.Errorf("ExponentialBackoffRun.notRetryableCode: %w ", err)
		}
		if time.Since(start) > maxSecAcrossAttempts {
			return nil, fmt.Errorf("ExponentialBackoffRun.tooLong: after %v seconds %w", time.Since(start), err)
		}
		if attemptNumber >= maxAttemptNumber {
			return nil, fmt.Errorf("ExponentialBackoffRun.maxAttempt: after %v attempts %w", attemptNumber, err)
		}
		waitSec = math.Min((math.Pow(2, float64(attemptNumber)) + rand.Float64()), maxSecBtwAttempts)
		time.Sleep(time.Duration(waitSec) * time.Second)
		attemptNumber += 1
	}
}

func (call CollectorResourceGroupCall) Run(ctx context.Context, collectId string, resource string) (*pb.Resource, error) {
	return call(ctx, collectId, resource)
}
