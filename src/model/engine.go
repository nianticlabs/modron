package model

import (
	"context"
	"encoding/json"

	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/risk"
)

type Engine interface {
	GetChildren(ctx context.Context, parent string) ([]*pb.Resource, error)
	GetResource(ctx context.Context, resourceName string) (*pb.Resource, error)
	GetHierarchy(ctx context.Context, collectionID string) (map[string]*pb.RecursiveResource, error)
	GetTagConfig() risk.TagConfig

	CheckRules(ctx context.Context, scanID string, collectID string, groups []string, preCollectedRGs []*pb.Resource) ([]*pb.Observation, []error)
	GetRuleConfig(ctx context.Context, ruleName string) (json.RawMessage, error)
	GetRules() []Rule
}
