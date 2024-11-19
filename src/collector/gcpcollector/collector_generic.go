package gcpcollector

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

type GenericCollector[T any] func(ctx context.Context, rgName string) ([]T, error)

func (call GenericCollector[T]) ExponentialBackoffRun(ctx context.Context, rgName string) ([]T, error) {
	attemptNumber := 1
	var waitSec float64
	start := time.Now()
	for {
		collected, err := call.Run(ctx, rgName)
		if err == nil {
			return collected, nil
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
		waitSec = math.Min(math.Pow(2, float64(attemptNumber))+rand.Float64(), maxSecBtwAttempts) //nolint:mnd,gosec
		time.Sleep(time.Duration(waitSec) * time.Second)
		attemptNumber++
	}
}

func (call GenericCollector[T]) Run(ctx context.Context, rgName string) ([]T, error) {
	resources, err := call(ctx, rgName)
	if isErrorCodeSkippable(getErrorCode(err)) {
		return []T{}, nil
	}
	return resources, err
}
