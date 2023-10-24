package model

import (
	"fmt"
	"time"
)

type OperationStatus uint

const (
	OperationStarted OperationStatus = iota
	OperationCancelled
	OperationCompleted
	OperationFailed
	OperationUnknown
)

var statuses = []string{
	"STARTED",
	"CANCELLED",
	"COMPLETED",
	"FAILED",
	"UNKNOWN",
}

func (s OperationStatus) String() string {
	if int(s) > len(statuses)-1 {
		return "UNKNOWN"
	}
	return statuses[s]
}

func StatusFromString(s string) (OperationStatus, error) {
	switch s {
	case OperationCancelled.String():
		return OperationCancelled, nil
	case OperationFailed.String():
		return OperationFailed, nil
	case OperationCompleted.String():
		return OperationCompleted, nil
	case OperationStarted.String():
		return OperationStarted, nil
	default:
		return OperationUnknown, fmt.Errorf("unknown status %q", s)
	}
}

type Operation struct {
	ID            string
	ResourceGroup string
	OpsType       string
	StatusTime    time.Time
	Status        OperationStatus
	Reason        string
}
