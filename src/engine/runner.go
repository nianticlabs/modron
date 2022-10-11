package engine

import (
	"context"
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

type RuleEngine struct {
	storage model.Storage
	rules   []model.Rule
}

type CheckRuleResult struct {
	rule model.Rule
	obs  []*pb.Observation
	errs []error
}

func New(s model.Storage, rules []model.Rule) *RuleEngine {
	storage = &Storage{s}
	return &RuleEngine{
		storage: s,
		rules:   rules,
	}
}

// Checks that the supplied rule applies to the provided resources.
func (e *RuleEngine) checkRule(ctx context.Context, r model.Rule, resources []*pb.Resource) (obs []*pb.Observation, errs []error) {
	for _, rsrc := range resources {
		t, err := common.TypeFromResourceAsString(rsrc)
		if err != nil {
			errs = append(errs, fmt.Errorf(`could not retrieve type from resource %q: %v`, rsrc, err))
			return
		}
		if !slices.Contains(r.Info().AcceptedResourceTypes, t) {
			errs = append(errs, fmt.Errorf(`resource type "%s" is not accepted by rule "%v"`, t, r.Info().Name))
			continue
		}

		newObs, newErrs := r.Check(ctx, rsrc)
		if len(newErrs) > 0 {
			errs = append(errs, newErrs...)
		} else {
			obs = append(obs, newObs...)
		}
	}

	return
}

func (e *RuleEngine) checkRuleAsync(ctx context.Context, r model.Rule, resources []*pb.Resource, ch chan *CheckRuleResult) {
	ret := &CheckRuleResult{
		rule: r,
		obs:  nil,
		errs: nil,
	}

	ret.obs, ret.errs = e.checkRule(ctx, r, resources)

	ch <- ret
}

// Fetches accepted resources and runs each rule in the engine asynchronously.
func (e *RuleEngine) checkRulesAsync(ctx context.Context, resourceGroups []string, ch chan *CheckRuleResult) (errs []error) {
fetchResourcesAndCheck:
	for _, r := range e.rules {
		typesInt := []int{}
		for _, t := range r.Info().AcceptedResourceTypes {
			tInt, err := common.TypeFromString(t)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid resource type: %v", err))
				continue fetchResourcesAndCheck
			}
			typesInt = append(typesInt, tInt)
		}
		filter := model.StorageFilter{
			ResourceTypes:      &typesInt,
			ResourceGroupNames: &resourceGroups,
		}
		if resources, err := e.storage.ListResources(ctx, filter); err != nil {
			errs = append(errs, fmt.Errorf("listing accepted resources: %v", err))
		} else {
			go e.checkRuleAsync(ctx, r, resources, ch)
		}
	}
	return
}

// Checks that all the supplied rules apply to resources belonging to `resourceGroups`.
func (e *RuleEngine) CheckRules(ctx context.Context, scanId string, resourceGroups []string) (obs []*pb.Observation, errs []error) {
	checkCh := make(chan *CheckRuleResult)

	// For each rule, fetch the accepted resources and invoke the check.
	e.checkRulesAsync(ctx, resourceGroups, checkCh)

	// Wait on all the checks to either terminate gracefully or be
	// canceled.
	for range e.rules {
		select {
		case <-ctx.Done():
			errs = append(errs, fmt.Errorf("execution of rule was cancelled: %w", ctx.Err()))
		case res := <-checkCh:
			for _, err := range res.errs {
				errs = append(errs, fmt.Errorf("execution of rule %v failed: %w", res.rule, err))
			}
			for _, ob := range res.obs {
				ob.ScanUid = scanId
			}
			if _, err := e.storage.BatchCreateObservations(ctx, res.obs); err != nil {
				errs = append(errs, fmt.Errorf("creation of observations for rule %v failed: %w", res.rule, err))
			}
			obs = append(obs, res.obs...)
		}
	}

	return
}
