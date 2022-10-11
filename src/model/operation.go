package model

import "time"

type OperationStatus uint

const (
	OperationStarted OperationStatus = iota
	OperationCancelled
	OperationCompleted
	OperationFailed
)

func (s OperationStatus) String() string {
	statuses := []string{
		"STARTED",
		"CANCELLED",
		"COMPLETED",
		"FAILED",
	}
	if int(s) > len(statuses)-1 {
		return "UNKNOWN"
	}
	return statuses[s]
}

type Operation struct {
	ID            string
	ResourceGroup string
	OpsType       string
	StatusTime    time.Time
	Status        OperationStatus
}
