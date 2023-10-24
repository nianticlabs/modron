package fakeacl

import (
	"os"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"github.com/nianticlabs/modron/src/constants"
	"github.com/nianticlabs/modron/src/model"
)

type GcpCheckerFake struct{}

func New() model.Checker {
	glog.Warningf("If you see this on production, contact security%s", os.Getenv(constants.OrgSuffixEnvVar))
	return &GcpCheckerFake{}
}

func (checker *GcpCheckerFake) GetAcl() map[string]map[string]struct{} {
	return nil
}

func (checker *GcpCheckerFake) GetValidatedUser(ctx context.Context) (string, error) {
	return "", nil
}

func (checker *GcpCheckerFake) ListResourceGroupNamesOwned(ctx context.Context) (map[string]struct{}, error) {
	return map[string]struct{}{"projects/modron-test": {}}, nil
}
