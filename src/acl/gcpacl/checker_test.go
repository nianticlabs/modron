package gcpacl

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/nianticlabs/modron/src/collector/gcpcollector"
	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/storage/memstorage"

	"google.golang.org/grpc/metadata"
)

func TestMain(m *testing.M) {
	if orgIdEnv := os.Getenv(constants.OrgIdEnvVar); orgIdEnv == "" {
		os.Setenv(constants.OrgIdEnvVar, fmt.Sprintf("%s%s", constants.GCPOrgIdPrefix, "012345678912"))
		defer os.Unsetenv(constants.OrgIdEnvVar)
	}
	if orgSuffixEnv := os.Getenv(constants.OrgSuffixEnvVar); orgSuffixEnv == "" {
		os.Setenv(constants.OrgSuffixEnvVar, "example.com")
		defer os.Unsetenv(constants.OrgSuffixEnvVar)
	}
	os.Exit(m.Run())
}

func TestInvalidNoToken(t *testing.T) {
	ctx := context.Background()
	storage := memstorage.New()
	gcpCollector := gcpcollector.NewFake(ctx, storage)

	checker, err := New(ctx, gcpCollector, Config{})
	if err != nil {
		t.Error(err)
	}

	if _, err = checker.GetValidatedUser(ctx); err == nil {
		t.Error("expected error: the context does not have a tokenid but the checker authenticated a user")
	}
}

func TestInvalidParseToken(t *testing.T) {
	ctx := context.Background()
	ctx = metadata.NewIncomingContext(ctx, metadata.New(
		map[string]string{"Authorization": "QxMzZjMjAyYjhkMjkwNTgwYjE2NWMiLCJ0eXAiOiJKV1QifQ.eyJhdWQiOiIzMjU1NTk0MDU1OS5hcHBzLmdvb2dsZXVzZXJjb250ZW50LmNvbSIsImF6cCI6InNlcnZpY2VhY2NvdW50bmFtZUBzZWMtZXNhbGltYmVuaS1hcGkta2V5LXRlc3QuaWFtLmdzZXJ2aWNlYWNjb3VudC5jb20iLCJlbWFpbCI6InNlcnZpY2VhY2NvdW50bmFtZUBzZWMtZXNhbGltYmVuaS1hcGkta2V5LXRlc3QuaWFtLmdzZXJ2aWNlYWNjb3VudC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZXhwIjoxNjU5NTIxODcxLCJpYXQiOjE2NTk1MTgyNzEsImlzcyI6Imh0dHBzOi8vYWNjb3VudHMuZ29vZ2xlLmNvbSIsInN1YiI6IjEwNTM5OTA4NzA5OTk1NDg5MDA0MyJ9.LHwLeuAa6jZc7pFPhtFLvsKMg56vPHrm83dfVukLxycopz6CzkpFZoF_DXFRKQ-myQs4KMMd44loi29te5vfnAL9aMXwvySFEcjESOIE_SXPND3Q5FBlfRfWoSLFjGsGhqLhNwKQn-tvjynkdxmtopL4qVmhAFpgTqNA4u8b7l7cWsl3zoudPZMy8mi5pIUetWH5jpj7OaPyv9pVaQ-LaXaLUQkD8bx0bL3Tjhu9yu2IP2Z06jFR9mN-fJ60F05kMJ6Y4HquDXNjm8HCNrXfMGHBcKMUzE3wAaOIG4PoGI81t43dPWpUIUg07RS5tG5uxuWIrgJddxJYpYCWOhGjog"}))
	storage := memstorage.New()
	gcpCollector := gcpcollector.NewFake(ctx, storage)

	checker, err := New(ctx, gcpCollector, Config{})
	if err != nil {
		t.Error(err)
	}

	if _, err = checker.GetValidatedUser(ctx); err == nil {
		t.Error("expected error: checker parsed a jwt tokenId that is invalid")
	}
}

func TestCheckerReal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode: need GCP credentials")
	}

	ctx := context.Background()
	storage := memstorage.New()
	gcpCollector, err := gcpcollector.New(ctx, storage)
	if err != nil {
		t.Error(err)
	}

	checker, err := New(ctx, gcpCollector, Config{})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := checker.ListResourceGroupNamesOwned(ctx); err != nil {
		t.Error(err)
	}
}
