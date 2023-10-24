// Package model is a shared set of models needed
package model

import (
	"context"
	"time"

	"github.com/nianticlabs/modron/src/pb"
)

type StorageFilter struct {
	Limit              int
	ResourceNames      []string
	ResourceTypes      []string
	ResourceGroupNames []string
	ResourceIDs        []string
	ParentNames        []string
	OperationID        string
	StartTime          time.Time
	TimeOffset         time.Duration
}

type Storage interface {
	BatchCreateResources(ctx context.Context, resource []*pb.Resource) ([]*pb.Resource, error)
	ListResources(ctx context.Context, filter StorageFilter) ([]*pb.Resource, error)
	BatchCreateObservations(ctx context.Context, observations []*pb.Observation) ([]*pb.Observation, error)
	ListObservations(ctx context.Context, filter StorageFilter) ([]*pb.Observation, error)

	AddOperationLog(ctx context.Context, ops []Operation) error
	FlushOpsLog(ctx context.Context) error
	PurgeIncompleteOperations(ctx context.Context) error
}
