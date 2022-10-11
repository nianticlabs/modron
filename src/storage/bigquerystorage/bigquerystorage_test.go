package bigquerystorage

import (
	"context"
	"os"
	"testing"

	"github.com/nianticlabs/modron/src/storage/test"
)

const (
	datasetIdEnvVar    = "DATASET_ID"
	gcpProjectIdEnvVar = "GCP_PROJECT_ID"

	observationTableIdEnvVar = "OBSERVATION_TABLE_ID"
	operationTableIdEnvVar   = "OPERATION_TABLE_ID"
	resourceTableIdEnvVar    = "RESOURCE_TABLE_ID"
)

var (
	requiredEnvVars = []string{datasetIdEnvVar, gcpProjectIdEnvVar, observationTableIdEnvVar, resourceTableIdEnvVar}
)

func TestBigQueryStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode: need GCP credentials")
	}
	for _, envVar := range requiredEnvVars {
		if env := os.Getenv(envVar); env == "" {
			t.Fatalf("environment variable %q is not set", envVar)
		}
	}
	ctx := context.Background()
	if s, err := New(
		ctx,
		os.Getenv(gcpProjectIdEnvVar),
		os.Getenv(datasetIdEnvVar),
		os.Getenv(resourceTableIdEnvVar),
		os.Getenv(observationTableIdEnvVar),
		os.Getenv(operationTableIdEnvVar),
	); err != nil {
		t.Errorf("BigQueryStorage.New unexpected error: %v", err)
	} else {
		test.TestBQStorageObservation(t, s)
		test.TestBQStorageResource(t, s)
	}
}
