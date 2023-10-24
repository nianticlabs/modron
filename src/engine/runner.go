package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
	"golang.org/x/exp/slices"

	"github.com/nianticlabs/modron/src/common"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/pb"
)

type RuleEngine struct {
	excludedRules []string
	rules         []model.Rule
	storage       model.Storage
}

type CheckRuleResult struct {
	rule model.Rule
	obs  []*pb.Observation
	errs []error
}

func New(s model.Storage, rules []model.Rule, excludedRules []string) *RuleEngine {
	storage = &Storage{s}
	glog.Infof("new rule engine with %d rules", len(rules))
	return &RuleEngine{
		excludedRules: excludedRules,
		rules:         rules,
		storage:       s,
	}
}

// Checks that the supplied rule applies to the provided resources.
func (e *RuleEngine) checkRule(ctx context.Context, r model.Rule, resources []*pb.Resource) (obs []*pb.Observation, errs []error) {
	for _, rsrc := range resources {
		t, err := common.TypeFromResourceAsString(rsrc)
		if err != nil {
			errs = append(errs, fmt.Errorf("could not retrieve type from resource %q: %v", rsrc, err))
			continue
		}
		if !slices.Contains(r.Info().AcceptedResourceTypes, t) {
			errs = append(errs, fmt.Errorf("resource type %q is not accepted by rule %s", t, r.Info().Name))
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
	wg := sync.WaitGroup{}
	for _, r := range e.rules {
		isExcluded := false
		for _, er := range e.excludedRules {
			if er == r.Info().Name {
				isExcluded = true
				break
			}
		}
		if isExcluded {
			glog.V(5).Infof("rule %s excluded", r.Info().Name)
			continue
		}
		wg.Add(1)
		go func(r model.Rule) {
			types := r.Info().AcceptedResourceTypes
			filter := model.StorageFilter{
				ResourceTypes:      types,
				ResourceGroupNames: resourceGroups,
			}
			if resources, err := e.storage.ListResources(ctx, filter); err != nil {
				errs = append(errs, fmt.Errorf("listing accepted resources: %+v", err))
			} else if len(resources) < 1 {
				errs = append(errs, fmt.Errorf("no resources for %+v", filter))
			} else {
				e.checkRuleAsync(ctx, r, resources, ch)
			}
			glog.V(5).Infof("done with rule %s", r.Info().Name)
			wg.Done()
		}(r)
	}
	glog.V(5).Infof("waiting for rules to finish")
	wg.Wait()
	return errs
}

// Checks that all the supplied rules apply to resources belonging to `resourceGroups`.
func (e *RuleEngine) CheckRules(ctx context.Context, scanId string, resourceGroups []string) (obs []*pb.Observation, errs []error) {
	e.logScanStatus(ctx, scanId, resourceGroups, model.OperationStarted)
	checkCh := make(chan *CheckRuleResult, 100)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				errs = append(errs, fmt.Errorf("context cancelled: %w", ctx.Err()))
				e.logScanStatus(ctx, scanId, resourceGroups, model.OperationCancelled)
				if err := e.storage.FlushOpsLog(ctx); err != nil {
					glog.Errorf("flushing operation log: %v", err)
				}
				break
			case res, ok := <-checkCh:
				if !ok {
					checkCh = nil
					break
				}
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
			if checkCh == nil {
				break
			}
		}
		wg.Done()
	}()
	// For each rule, fetch the accepted resources and invoke the check.
	wg.Add(1)
	go func() {
		err := e.checkRulesAsync(ctx, resourceGroups, checkCh)
		if len(err) > 0 {
			errs = append(errs, err...)
			glog.Warningf("rules run for %v : %v", resourceGroups, err)
		}
		glog.V(5).Infof("closing channel")
		close(checkCh)
		wg.Done()
	}()
	glog.V(5).Infof("waiting for scan %s to finish", scanId)
	wg.Wait()
	e.logScanStatus(ctx, scanId, resourceGroups, model.OperationCompleted)
	return
}

func (e *RuleEngine) logScanStatus(ctx context.Context, scanId string, resourceGroups []string, status model.OperationStatus) {
	ops := []model.Operation{}
	glog.V(5).Infof("scan %s status %s for %+v", scanId, status, resourceGroups)
	for _, resourceGroup := range resourceGroups {
		ops = append(ops, model.Operation{ID: scanId, ResourceGroup: resourceGroup, OpsType: "scan", StatusTime: time.Now(), Status: status})
	}
	if err := e.storage.AddOperationLog(ctx, ops); err != nil {
		glog.Warningf("log operation: %v", err)
	}
}
