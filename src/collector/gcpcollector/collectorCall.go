package gcpcollector

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/api/googleapi"
	"github.com/nianticlabs/modron/src/pb"
)

const (
	maxAttemptNumber    = 100
	maxSecBtwAttempts   = 30.
	maxSecAcrossAttemps = 3600.
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
	attemptNumber := 0
	waitSec := 0.
	waitSecTot := 0.
	var errTemp error
	for {
		if waitSecTot >= maxSecAcrossAttemps {
			return nil, fmt.Errorf("ExponentialBackoffRun.tooLong error: after %v seconds %v ", waitSecTot, errTemp)
		}
		if attemptNumber >= maxAttemptNumber {
			return nil, fmt.Errorf("ExponentialBackoffRun.maxAttempt error: after %v  attempts %v ", attemptNumber, errTemp)
		}
		resources, err := call.Run(ctx, resource)
		if errTemp = err; err == nil {
			return resources, nil
		}
		if !isErrorCodeRetryable(getErrorCode(errTemp)) {
			return nil, fmt.Errorf("ExponentialBackoffRun.notRetryableCode error:  %v ", errTemp)
		}
		waitSec = math.Min((math.Pow(2, float64(attemptNumber)) + rand.Float64()), maxSecBtwAttempts)
		time.Sleep(time.Duration(waitSec) * time.Second)
		attemptNumber += 1
		waitSecTot += waitSec
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
	waitSecTot := 0.
	var errTemp error
	time.Sleep(time.Duration(rand.Float64()*2) * time.Second)
	for {
		if waitSecTot >= maxSecAcrossAttemps {
			return nil, fmt.Errorf("ExponentialBackoffRun.tooLong error: after %v seconds %v ", waitSecTot, errTemp)
		}
		if attemptNumber >= maxAttemptNumber {
			return nil, fmt.Errorf("ExponentialBackoffRun.maxAttempt error: after %v  attempts %v ", attemptNumber, errTemp)
		}
		resources, err := call.Run(ctx, collectId, resource)
		if errTemp = err; err == nil {
			return resources, nil
		}
		if !isErrorCodeRetryable(getErrorCode(errTemp)) {
			return nil, fmt.Errorf("ExponentialBackoffRun.notRetryableCode error:  %v ", errTemp)
		}
		waitSec = math.Min((math.Pow(2, float64(attemptNumber)) + rand.Float64()), maxSecBtwAttempts)
		time.Sleep(time.Duration(waitSec) * time.Second)
		attemptNumber += 1
		waitSecTot += waitSec
	}
}

func (call CollectorResourceGroupCall) Run(ctx context.Context, collectId string, resource string) (*pb.Resource, error) {
	return call(ctx, collectId, resource)
}
