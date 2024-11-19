package model

import "golang.org/x/net/context"

// ACLCache is a map of users to a map of resource names: {"user@example.com": {"projects/xyz": {}}}
// the reason why the last part is a struct{} is that we don't care about the value, we only care about the key
// and a struct{} is the smallest value we can use.
type ACLCache map[string]map[string]struct{}

type Checker interface {
	GetACL() ACLCache
	GetValidatedUser(ctx context.Context) (string, error)
	ListResourceGroupNamesOwned(ctx context.Context) (map[string]struct{}, error)
}
