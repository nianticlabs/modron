//go:build integration

package gcpcollector

import (
	"context"
	"os"
	"testing"

	"github.com/nianticlabs/modron/src/constants"
)

func TestListServiceAccounts(t *testing.T) {
	ctx := context.Background()
	coll, _ := getCollector(ctx, t)
	project := constants.GCPProjectsNamePrefix + os.Getenv("PROJECT_ID")
	accounts, err := coll.(*GCPCollector).ListServiceAccounts(ctx, project)
	if err != nil {
		t.Fatalf("ListServiceAccounts failed: %v", err)
	}
	if len(accounts) == 0 {
		t.Fatalf("ListServiceAccounts returned 0 accounts")
	}
	for _, account := range accounts {
		t.Logf("ServiceAccount: %s", account.Name)
		if account.IamPolicy != nil && len(account.IamPolicy.Permissions) > 0 {
			t.Logf("\tIAM Policy: %v", account.IamPolicy)
		}
	}
}
