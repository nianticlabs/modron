//go:build integration

package engine_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/nianticlabs/modron/src/collector/gcpcollector"
	"github.com/nianticlabs/modron/src/engine"
	"github.com/nianticlabs/modron/src/engine/rules"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/risk"
	"github.com/nianticlabs/modron/src/storage/memstorage"
)

func TestCrossEnvironmentRuleIntegration(t *testing.T) {
	rgNames := []string{
		"projects/modron-dev",
	}
	ctx := context.Background()
	st := memstorage.New()
	collectID := uuid.NewString()
	scanID := uuid.NewString()

	logrus.StandardLogger().SetLevel(logrus.DebugLevel)
	logrus.StandardLogger().SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})

	orgID := os.Getenv("ORG_ID")
	orgSuffix := os.Getenv("ORG_SUFFIX")
	if orgID == "" || orgSuffix == "" {
		t.Fatalf("ORG_ID and ORG_SUFFIX are required for this test")
	}
	tagConfig := risk.TagConfig{
		Environment:  "111111111111/environment",
		EmployeeData: "111111111111/employee_data",
		CustomerData: "111111111111/customer_data",
	}

	// Collect resources
	c, err := gcpcollector.New(ctx, st, orgID, orgSuffix, []string{}, tagConfig, []string{})
	if err != nil {
		t.Fatalf("failed to create collector: %v", err)
	}
	if err := c.CollectAndStoreAll(ctx, collectID, rgNames, nil); err != nil {
		t.Fatalf("failed to collect resources: %v", err)
	}

	e, _ := engine.New(st, []model.Rule{
		rules.NewCrossEnvironmentPermissionsRule(),
	}, map[string]json.RawMessage{}, []string{}, tagConfig)
	obs, errArr := e.CheckRules(ctx, scanID, collectID, rgNames, nil)
	err = errors.Join(errArr...)
	if err != nil {
		t.Fatalf("failed to check rules: %v", err)
	}
	t.Logf("Observations: %v", obs)
}
