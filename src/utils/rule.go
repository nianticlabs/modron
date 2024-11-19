package utils

import (
	"context"
	"encoding/json"

	"github.com/nianticlabs/modron/src/model"
)

func GetRuleConfig[T any](ctx context.Context, e model.Engine, name string, c *T) error {
	v, err := e.GetRuleConfig(ctx, name)
	if err != nil {
		log.Errorf("no config found for rule %q: %v", name, err)
		return err
	}
	return json.Unmarshal(v, c)
}
