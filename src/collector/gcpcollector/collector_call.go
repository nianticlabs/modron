package gcpcollector

import (
	"errors"
	"time"

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
	//   * 403: Sometimes returned for non-existing resources.
	//   * 404: A resource can be tracked by modron and then deleted.
	skippableErrorCodes = []int{403, 404}
)

func getErrorCode(err error) int {
	var e *googleapi.Error
	if errors.As(err, &e) {
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

func isErrorCodeSkippable(errorCode int) bool {
	for _, code := range skippableErrorCodes {
		if code == errorCode {
			return true
		}
	}
	return false
}
