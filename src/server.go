package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"golang.org/x/exp/maps"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	_ "github.com/lib/pq"

	"github.com/nianticlabs/modron/src/acl/fakeacl"
	"github.com/nianticlabs/modron/src/acl/gcpacl"
	"github.com/nianticlabs/modron/src/collector"
	"github.com/nianticlabs/modron/src/collector/gcpcollector"
	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/engine"
	"github.com/nianticlabs/modron/src/engine/rules"
	"github.com/nianticlabs/modron/src/lognotifier"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/nagatha"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/risk"
	"github.com/nianticlabs/modron/src/service"
	"github.com/nianticlabs/modron/src/statemanager/reqdepstatemanager"
	"github.com/nianticlabs/modron/src/storage"
	"github.com/nianticlabs/modron/src/storage/gormstorage"
	"github.com/nianticlabs/modron/src/storage/memstorage"
)

const (
	defaultCacheTimeout = 20 * time.Second
)

func newServer(ctx context.Context) (*service.Modron, error) {
	ctx, span := tracer.Start(ctx, "newServer")
	defer span.End()
	var st model.Storage
	var err error
	switch storage.Type(strings.ToLower(args.Storage)) {
	case storage.Memory:
		log.Warnf("using memory storage: this should never be used in production")
		st = memstorage.New()
	case storage.SQL:
		log.Tracef("setting up SQL")
		st, err = gormstorage.NewDB(
			args.SQLBackendDriver,
			args.SQLConnectionString,
			gormstorage.Config{
				BatchSize:     args.DbBatchSize,
				LogAllQueries: args.LogAllSQLQueries,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("sql storage creation: %w", err)
		}
	default:
		return nil, fmt.Errorf("invalid storage \"%s\" specified", args.Storage)
	}

	// Get Impact map
	impactMap, err := getImpactMap(args.ImpactMap)
	if err != nil {
		return nil, fmt.Errorf("getImpactMap: %w", err)
	}
	if len(maps.Keys(impactMap)) == 0 {
		log.Warn("Impact map is empty")
	}

	// Parse rule configs
	ruleConfigs, err := parseRuleConfigs(args.RuleConfigs)
	if err != nil {
		return nil, fmt.Errorf("parseRuleConfigs: %w", err)
	}

	additionalAdminRoles := map[constants.Role]struct{}{}
	for _, v := range args.AdditionalAdminRoles {
		theRole := constants.ToRole(v)
		additionalAdminRoles[theRole] = struct{}{}
	}

	log.Tracef("Purging incomplete operations")
	if err := st.PurgeIncompleteOperations(ctx); err != nil {
		log.Errorf("Purging incomplete operations: %v", err)
	}

	tagConfig := risk.TagConfig{
		ImpactMap:    impactMap,
		Environment:  args.TagEnvironment,
		EmployeeData: args.TagEmployeeData,
		CustomerData: args.TagCustomerData,
	}

	log.Tracef("Creating rule engine")
	ruleEngine, err := engine.New(st, rules.GetRules(), ruleConfigs, args.ExcludedRules, tagConfig)
	if err != nil {
		return nil, fmt.Errorf("creating rule engine: %w", err)
	}

	log.Tracef("Creating collector and ACL checker")
	coll, checker, err := getCollectorAndChecker(ctx, st, tagConfig)
	if err != nil {
		return nil, fmt.Errorf("getCollectorAndChecker: %w", err)
	}

	log.Tracef("Creating reqdepstate manager")
	stateManager, err := reqdepstatemanager.New()
	if err != nil {
		return nil, fmt.Errorf("creating reqdepstatemanager: %w", err)
	}

	log.Tracef("Creating Notification Service")
	var notificationSvc model.NotificationService
	notificationSvcAddr := args.NotificationService
	if notificationSvcAddr != "" {
		notificationSvc, err = setupNotificationService(ctx, notificationSvcAddr)
		if err != nil {
			return nil, fmt.Errorf("unable to setup notification service: %w", err)
		}
	} else {
		log.Tracef("Using lognotifier as notification service")
		log.Infof("NotificationService argument is empty, logging instead")
		notificationSvc = lognotifier.New()
	}

	labelToEmailRegexp, err := regexp.Compile(args.LabelToEmailRegexp)
	if err != nil {
		return nil, fmt.Errorf("regexp.Compile failed for contact label regex: %w", err)
	}

	return service.New(
		checker,
		args.CollectAndScanInterval,
		coll,
		args.NotificationInterval,
		notificationSvc,
		args.OrgSuffix,
		ruleEngine,
		args.SelfURL,
		stateManager,
		st,
		additionalAdminRoles,
		labelToEmailRegexp,
		args.LabelToEmailSubst,
	)
}

func parseRuleConfigs(configs string) (map[string]json.RawMessage, error) {
	var ruleConfigs map[string]json.RawMessage
	if err := json.Unmarshal([]byte(configs), &ruleConfigs); err != nil {
		return nil, fmt.Errorf("unable to decode rule configs: %w", err)
	}
	return ruleConfigs, nil
}

func getImpactMap(impactMapJSON string) (map[string]pb.Impact, error) {
	var impactMap map[string]string
	if err := json.Unmarshal([]byte(impactMapJSON), &impactMap); err != nil {
		return nil, fmt.Errorf("unable to decode impact map: %w", err)
	}
	finalImpactMap := map[string]pb.Impact{}
	for k, v := range impactMap {
		impactValue, ok := pb.Impact_value[v]
		if !ok {
			return nil, fmt.Errorf("invalid impact value: %s", v)
		}
		finalImpactMap[k] = pb.Impact(impactValue)
	}
	return finalImpactMap, nil
}

func setupNotificationService(ctx context.Context, notSvcAddr string) (notSvc model.NotificationService, err error) {
	log.Tracef("Using Nagatha as notification service")
	var tokenSource oauth2.TokenSource
	if args.IsE2EGrpcTest {
		tokenSource = oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: "e2e-test-token",
		})
	} else {
		tokenSource, err = idtoken.NewTokenSource(ctx, args.NotificationServiceClientID)
		if err != nil {
			return nil, fmt.Errorf("idtoken.NewTokenSource: %w", err)
		}
	}
	notSvc, err = nagatha.New(notSvcAddr, args.SelfURL, tokenSource)
	if err != nil {
		return nil, fmt.Errorf("nagatha service: %w", err)
	}
	return notSvc, nil
}

func getCollectorAndChecker(ctx context.Context, st model.Storage, tagConfig risk.TagConfig) (c model.Collector, aclChecker model.Checker, err error) {
	switch collector.Type(strings.ToLower(string(args.Collector))) {
	case collector.Fake:
		log.Warnf("Using fake collector")
		c = gcpcollector.NewFake(ctx, st, tagConfig)
		log.Warnf("Using fake ACL")
		aclChecker = fakeacl.New()
	case collector.Gcp:
		log.Tracef("Creating GCP collector")
		c, err = gcpcollector.New(
			ctx,
			st,
			args.OrgID,
			args.OrgSuffix,
			args.AdditionalAdminRoles,
			tagConfig,
			args.AllowedSccCategories,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("NewGCPCollector: %w", err)
		}
		log.Tracef("Creating GCP ACL checker")
		if aclChecker, err = gcpacl.New(ctx, c, gcpacl.Config{
			AdminGroups:            args.AdminGroups,
			CacheTimeout:           defaultCacheTimeout,
			PersistentCache:        args.PersistentCache,
			PersistentCacheTimeout: args.PersistentCacheTimeout,
			SkipIap:                args.SkipIAP,
		}); err != nil {
			return nil, nil, fmt.Errorf("NewGcpChecker: %w", err)
		}
	default:
		return nil, nil, fmt.Errorf("invalid collector \"%s\" specified, use one of: %s",
			args.Collector, strings.Join(collector.ValidCollectors(), ", "))
	}
	return
}

func withCors() []grpcweb.Option {
	return []grpcweb.Option{
		grpcweb.WithOriginFunc(func(_ string) bool {
			return true
		}),
		grpcweb.WithAllowedRequestHeaders([]string{"*"}),
	}
}
