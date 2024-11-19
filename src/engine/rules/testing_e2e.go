package rules

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/nianticlabs/modron/src/collector/gcpcollector"
	"github.com/nianticlabs/modron/src/engine"
	"github.com/nianticlabs/modron/src/model"
	pb "github.com/nianticlabs/modron/src/proto/generated"
	"github.com/nianticlabs/modron/src/risk"
	"github.com/nianticlabs/modron/src/storage/memstorage"
)

func TestE2ERuleRun(t *testing.T, rules []model.Rule) ([]*pb.Observation, error) {
	t.Helper()
	logrus.StandardLogger().SetLevel(logrus.DebugLevel)
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		t.Skip("PROJECT_ID is not set")
	}
	orgID := os.Getenv("ORG_ID")
	if orgID == "" {
		t.Skip("ORG_ID is not set")
	}
	orgSuffix := os.Getenv("ORG_SUFFIX")
	if orgSuffix == "" {
		t.Skip("ORG_SUFFIX is not set")
	}
	tagConfig := risk.TagConfig{
		Environment:  "111111111111/environment",
		EmployeeData: "111111111111/employee_data",
		CustomerData: "111111111111/customer_data",
	}
	qualifiedProjectID := "projects/" + projectID
	ctx := context.Background()
	storage := memstorage.New()
	collector, err := gcpcollector.New(ctx, storage, orgID, orgSuffix, []string{}, tagConfig, []string{})
	if err != nil {
		t.Fatalf("NewCollector unexpected error: %v", err)
	}

	if err := collector.CollectAndStoreAll(ctx, "test-collect", []string{qualifiedProjectID}, nil); err != nil {
		t.Fatalf("collectAndStoreResources unexpected error: %v", err)
	}

	e, err := engine.New(storage, rules, map[string]json.RawMessage{}, []string{}, tagConfig)
	if err != nil {
		t.Fatalf("NewEngine unexpected error: %v", err)
	}
	obs, errArr := e.CheckRules(ctx, "unit-test-scan", "", []string{qualifiedProjectID}, nil)
	if errArr != nil {
		return nil, errors.Join(errArr...)
	}
	return obs, nil
}
