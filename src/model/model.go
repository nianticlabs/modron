package model

import (
	"context"

	"github.com/nianticlabs/modron/src/pb"
)

// Base interface to be implemented by rules. A `Rule` takes a resource, checks
// its observed values against an expected reference value, and creates an
// observation if it identifies a discrepancy, which may include a remediation for
// resolving it.
type Rule interface {
	// Performs a rule-dependent check on a resource and, in case it detects an anomaly,
	// returns a list of observations. The method MUST return nil in case either it did
	// not create any observations or detect any errors.
	Check(ctx context.Context, rsrc *pb.Resource) (obs []*pb.Observation, errs []error)
	// Returns the associated `RuleInfo` data.
	Info() *RuleInfo
}

type RuleInfo struct {
	// Human readable name of the rule, e.g., "EXPOSED_INFRASTRUCTURE_WITH_ADMIN_PRIVILEGES".
	Name string
	// Types of resource this rule accepts as an input to `Check`. This helps the rule engine
	// fetch in advance all the resources the rule needs to perform check(s) against.
	AcceptedResourceTypes []string
}

type RuleEngine interface {
	CheckRules(ctx context.Context, scanId string, resourceGroups []string) (obs []*pb.Observation, errs []error)
}
