package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nianticlabs/modron/src/constants"
	modronmetric "github.com/nianticlabs/modron/src/metric"
	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/risk"
	"github.com/nianticlabs/modron/src/utils"

	"github.com/sirupsen/logrus"
)

var (
	log    = logrus.StandardLogger().WithField(constants.LogKeyPkg, "engine")
	meter  = otel.Meter("github.com/nianticlabs/modron/src/engine")
	tracer = otel.Tracer("github.com/nianticlabs/modron/src/engine")
)

const (
	checkRulesBufferSize = 100
)

type RuleEngine struct {
	excludedRules []string
	metrics       metrics
	rules         []model.Rule
	ruleConfigs   map[string]json.RawMessage
	storage       model.Storage

	tagConfig risk.TagConfig
	// memoizationMap is our caching layer that holds the resources that have been fetched from the storage.
	// In the current implementation, multiple engines will not share their cache.
	// Ideally, to avoid fetching multiple times the same resource, we would want to share the cache between engines.
	memoizationMap sync.Map
	// rgHierarchyMap is a map containing the resource group hierarchy for each collection ID.
	rgHierarchyMap sync.Map
}

var _ model.Engine = (*RuleEngine)(nil)

type metrics struct {
	RulesDuration      metric.Float64Histogram
	CheckRulesDuration metric.Float64Histogram
	CreateObservation  metric.Int64Counter
}

type CheckRuleResult struct {
	rule model.Rule
	obs  []*pb.Observation
	errs []error
}

func New(
	s model.Storage,
	rules []model.Rule,
	ruleConfigs map[string]json.RawMessage,
	excludedRules []string,
	tagConfig risk.TagConfig,
) (*RuleEngine, error) {
	log.Debugf("new rule engine with %d rules", len(rules))
	e := &RuleEngine{
		excludedRules: excludedRules,
		rules:         rules,
		storage:       s,
		ruleConfigs:   ruleConfigs,
		tagConfig:     tagConfig,
	}
	err := e.initMetrics()
	if err != nil {
		return nil, err
	}
	go e.startCacheCleanup()
	return e, nil
}

// Checks that the supplied rule applies to the provided resources.
func (e *RuleEngine) checkRule(ctx context.Context, r model.Rule, resources []*pb.Resource) (obs []*pb.Observation, errs []error) {
	ctx, span := tracer.Start(ctx, "checkRule")
	span.SetAttributes(
		attribute.String(constants.TraceKeyRule, r.Info().Name),
	)
	defer span.End()
	for _, rsrc := range resources {
		t, err := utils.TypeFromResource(rsrc)
		if err != nil {
			errs = append(errs, fmt.Errorf("could not retrieve type from resource %q: %w", rsrc, err))
			continue
		}
		acceptedResourceTypes := map[string]struct{}{}
		for _, at := range r.Info().AcceptedResourceTypes {
			acceptedResourceTypes[string(at.ProtoReflect().Type().Descriptor().FullName())] = struct{}{}
		}
		if _, ok := acceptedResourceTypes[t]; !ok {
			errs = append(errs, fmt.Errorf("resource type %q is not accepted by rule %s", t, r.Info().Name))
			continue
		}

		newObs, newErrs := r.Check(ctx, e, rsrc)
		if len(newErrs) > 0 {
			errs = append(errs, newErrs...)
		} else {
			obs = append(obs, newObs...)
		}
	}
	return
}

func (e *RuleEngine) checkRuleAsync(ctx context.Context, r model.Rule, resources []*pb.Resource, ch chan *CheckRuleResult) {
	ctx, span := tracer.Start(ctx, "checkRuleAsync",
		trace.WithAttributes(attribute.String(constants.TraceKeyRule, r.Info().Name)),
	)
	log := log.WithField("rule", r.Info().Name)
	defer span.End()
	start := time.Now()
	ret := &CheckRuleResult{
		rule: r,
		obs:  nil,
		errs: nil,
	}
	ret.obs, ret.errs = e.checkRule(ctx, r, resources)
	// Risk Score calculation
	for k, obs := range ret.obs {
		ret.obs[k].RiskScore, ret.obs[k].Impact, ret.obs[k].ImpactReason = e.computeRsImpact(ctx, obs)
	}

	status := modronmetric.StatusSuccess
	if len(ret.errs) > 0 {
		status = modronmetric.StatusError
		log.Errorf("rule execution failed: %v", ret.errs)
	}
	e.metrics.RulesDuration.
		Record(ctx, time.Since(start).Seconds(),
			metric.WithAttributes(
				attribute.String(modronmetric.KeyRule, r.Info().Name),
				attribute.String(modronmetric.KeyStatus, status),
			),
		)
	ch <- ret
}

// computeRsImpactSeverity calculates the risk score by taking into account some external factors such as
// the parent resource group labels (e.g: environment: prod).
// It then returns the determined impact and the original severity
func (e *RuleEngine) computeRsImpact(ctx context.Context, obs *pb.Observation) (riskScore pb.Severity, impact pb.Impact, reason string) {
	riskScore = pb.Severity_SEVERITY_UNKNOWN
	impact = pb.Impact_IMPACT_UNKNOWN

	collectID, ok := ctx.Value(constants.CollectIDKey).(string)
	if !ok {
		log.Errorf("no %s in context", constants.CollectIDKey)
		return
	}
	mapVal, ok := e.rgHierarchyMap.Load(collectID)
	if !ok {
		log.Errorf("no resource group hierarchy found for %q", collectID)
		return
	}
	rgHierarchy, ok := mapVal.(map[string]*pb.RecursiveResource)
	if !ok {
		log.Errorf("resource group hierarchy is not a map")
		return
	}
	if obs.ResourceRef == nil {
		log.Errorf("observation %q has no resource ref", obs.Uid)
		return
	}
	if obs.ResourceRef.GroupName == "" {
		log.Errorf("observation %q has no group name", obs.Uid)
		return
	}
	impact, reason = risk.GetImpact(e.tagConfig, rgHierarchy, obs.ResourceRef.GroupName)
	return risk.GetRiskScore(impact, obs.Severity), impact, reason
}

// Fetches accepted resources and runs each rule in the engine asynchronously.
func (e *RuleEngine) checkRulesAsync(ctx context.Context, resourceGroups []string, ch chan *CheckRuleResult) (errs []error) {
	ctx, span := tracer.Start(ctx, "checkRulesAsync")
	defer span.End()
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
			log.Infof("rule %s excluded", r.Info().Name)
			continue
		}
		wg.Add(1)
		go func(r model.Rule) {
			ctx, span := tracer.Start(ctx, "goCheckRuleAsync",
				trace.WithNewRoot(),
				trace.WithAttributes(
					attribute.String(constants.TraceKeyRule, r.Info().Name),
					attribute.StringSlice(constants.TraceKeyResourceGroupNames, resourceGroups),
				),
				trace.WithLinks(trace.LinkFromContext(ctx)),
			)
			defer span.End()
			types := r.Info().AcceptedResourceTypes
			acceptedTypes := utils.ProtoAcceptsTypes(types)
			filter := model.StorageFilter{
				ResourceTypes:      acceptedTypes,
				ResourceGroupNames: resourceGroups,
				OperationID:        ctx.Value(constants.CollectIDKey).(string),
			}
			if resources, err := e.storage.ListResources(ctx, filter); err != nil {
				log.Errorf("listing accepted resources: %v", err)
				span.RecordError(err)
				errs = append(errs, fmt.Errorf("listing accepted resources: %w", err))
			} else if len(resources) < 1 {
				log.Warnf("no resources found")
				span.RecordError(err)
				errs = append(errs, fmt.Errorf("no resources found"))
			} else {
				span.SetAttributes(attribute.Int(constants.TraceKeyNumResources, len(resources)))
				e.checkRuleAsync(ctx, r, resources, ch)
			}
			log.Infof("done with rule %s", r.Info().Name)
			wg.Done()
		}(r)
	}
	log.Infof("waiting for rules to finish")
	wg.Wait()
	return errs
}

// CheckRules checks that all the supplied rules apply to resources belonging to `resourceGroups`.
func (e *RuleEngine) CheckRules(ctx context.Context, scanID string, collectID string, resourceGroups []string, preCollectedRgs []*pb.Resource) (obs []*pb.Observation, errs []error) {
	ctx, span := tracer.Start(ctx, "CheckRules")
	defer span.End()
	log := log.WithFields(logrus.Fields{
		constants.LogKeyScanID:             scanID,
		constants.LogKeyCollectID:          collectID,
		constants.LogKeyResourceGroupNames: resourceGroups,
	})
	log.Infof("Start CheckRules")
	defer log.Infof("End CheckRules")
	e.logScanStatus(ctx, scanID, resourceGroups, pb.Operation_STARTED)
	ctx = context.WithValue(ctx, constants.ScanIDKey, scanID)
	ctx = context.WithValue(ctx, constants.CollectIDKey, collectID)
	start := time.Now()
	failed := func() {
		e.logScanStatus(ctx, scanID, resourceGroups, pb.Operation_CANCELLED)
		e.metrics.CheckRulesDuration.Record(ctx,
			time.Since(start).Seconds(),
			metric.WithAttributes(
				attribute.String(modronmetric.KeyStatus, modronmetric.StatusCancelled),
			),
		)
	}

	// Get Resource Group hierarchy and store it in the scan-specific cache
	rgHierarchy, _ := utils.ComputeRgHierarchy(preCollectedRgs)
	e.rgHierarchyMap.Store(collectID, rgHierarchy)
	defer func() {
		e.rgHierarchyMap.Delete(collectID)
	}()

	checkCh := make(chan *CheckRuleResult, checkRulesBufferSize)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				errs = append(errs, fmt.Errorf("context cancelled: %w", ctx.Err()))
				failed()
				if err := e.storage.FlushOpsLog(ctx); err != nil {
					log.Errorf("flushing operation log: %v", err)
				}
				break
			case res, ok := <-checkCh:
				if !ok {
					checkCh = nil
					break
				}
				for _, err := range res.errs {
					errs = append(errs, fmt.Errorf("execution of rule %v failed: %w", res.rule.Info().Name, err))
				}
				for _, ob := range res.obs {
					ob.ScanUid = utils.RefOrNull(scanID)
					ob.Source = pb.Observation_SOURCE_MODRON

					if ob.Uid == "" {
						log.Errorf("observation from rule %s has no UUID", res.rule.Info().Name)
						ob.Uid = uuid.NewString()
					}
				}
				status := modronmetric.StatusSuccess
				if _, err := e.storage.BatchCreateObservations(ctx, res.obs); err != nil {
					status = modronmetric.StatusError
					errs = append(errs, fmt.Errorf("creation of observations for rule %v failed: %w", res.rule, err))
				}
				e.metrics.CreateObservation.Add(ctx, int64(len(res.obs)),
					metric.WithAttributes(
						attribute.String(modronmetric.KeyRule, res.rule.Info().Name),
						attribute.String(modronmetric.KeyStatus, status),
					),
				)
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
			log.WithError(errors.Join(errs...)).Warnf("rules run for with errors")
		}
		log.Tracef("closing channel")
		close(checkCh)
		wg.Done()
	}()
	log.Infof("waiting for scan %q to finish", scanID)
	wg.Wait()
	e.logScanStatus(ctx, scanID, resourceGroups, pb.Operation_COMPLETED)
	log.Infof("scan %q done", scanID)
	e.metrics.CheckRulesDuration.Record(ctx,
		time.Since(start).Seconds(),
		metric.WithAttributes(attribute.String(modronmetric.KeyStatus, modronmetric.StatusCompleted)),
	)
	return
}

func (e *RuleEngine) GetRules() []model.Rule {
	return e.rules
}

func (e *RuleEngine) GetRuleConfig(_ context.Context, ruleName string) (json.RawMessage, error) {
	v, ok := e.ruleConfigs[ruleName]
	if !ok {
		return nil, fmt.Errorf("no configuration found for rule %q", ruleName)
	}
	return v, nil
}

func (e *RuleEngine) GetHierarchy(_ context.Context, collID string) (map[string]*pb.RecursiveResource, error) {
	v, ok := e.rgHierarchyMap.Load(collID)
	if !ok {
		return nil, fmt.Errorf("no hierarchy found for %q", collID)
	}
	return v.(map[string]*pb.RecursiveResource), nil
}

func (e *RuleEngine) logScanStatus(ctx context.Context, scanID string, resourceGroups []string, status pb.Operation_Status) {
	var ops []*pb.Operation
	log.Infof("scan %q status %s for %+v", scanID, status, resourceGroups)
	for _, resourceGroup := range resourceGroups {
		ops = append(ops, &pb.Operation{
			Id:            scanID,
			ResourceGroup: resourceGroup,
			Type:          "scan",
			StatusTime:    timestamppb.New(time.Now()),
			Status:        status,
		})
	}
	if err := e.storage.AddOperationLog(ctx, ops); err != nil {
		log.Warnf("log operation: %v", err)
	}
}

func (e *RuleEngine) initMetrics() error {
	rulesDurationHist, err := meter.Float64Histogram(
		constants.MetricsPrefix+"rules_duration",
		metric.WithDescription("Duration of rules execution"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}
	checkRulesDurationHist, err := meter.Float64Histogram(
		constants.MetricsPrefix+"check_rules_duration",
		metric.WithDescription("Duration of check_rules operations"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}
	createObservationCounter, err := meter.Int64Counter(
		constants.MetricsPrefix+"create_observation",
		metric.WithDescription("Number of observations created"),
	)
	if err != nil {
		return err
	}
	e.metrics = metrics{
		RulesDuration:      rulesDurationHist,
		CheckRulesDuration: checkRulesDurationHist,
		CreateObservation:  createObservationCounter,
	}
	return nil
}

func (e *RuleEngine) GetTagConfig() risk.TagConfig {
	return e.tagConfig
}
