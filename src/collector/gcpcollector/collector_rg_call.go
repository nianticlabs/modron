package gcpcollector

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	pb "github.com/nianticlabs/modron/src/proto/generated"
)

// TODO: find a way to avoid code duplication:
// (the implementation of CollectorResourceGroupCall.ExponentialBackoffRun is very similar to GenericCollector.ExponentialBackoffRun)
type CollectorResourceGroupCall func(ctx context.Context, collectID string, rgName string) (*pb.Resource, error)

func (call CollectorResourceGroupCall) ExponentialBackoffRun(ctx context.Context, collectID string, rgName string) (*pb.Resource, error) {
	attemptNumber := 0
	var waitSec float64
	start := time.Now()
	for {
		resources, err := call.Run(ctx, collectID, rgName)
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
		waitSec = math.Min(math.Pow(2, float64(attemptNumber))+rand.Float64(), maxSecBtwAttempts) //nolint:mnd,gosec
		time.Sleep(time.Duration(waitSec) * time.Second)
		attemptNumber++
	}
}

func (call CollectorResourceGroupCall) Run(ctx context.Context, collectID string, rgName string) (*pb.Resource, error) {
	return call(ctx, collectID, rgName)
}
