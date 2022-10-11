package model

import "golang.org/x/net/context"

type Checker interface {
	GetAcl() map[string]map[string]struct{}
	GetValidatedUser(ctx context.Context) (string, error)
	ListResourceGroupNamesOwned(ctx context.Context) (map[string]struct{}, error)
}
