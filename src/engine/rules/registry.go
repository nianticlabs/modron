package rules

import (
	"fmt"
	"sync"

	"github.com/nianticlabs/modron/src/model"
)

var rules sync.Map

func AddRule(r model.Rule) {
	ruleName := r.Info().Name
	_, ok := rules.Load(ruleName)
	if ok {
		panic(fmt.Sprintf("rule %q already exists", ruleName))
	}
	rules.Store(r.Info().Name, r)
}

func GetRule(name string) (model.Rule, error) {
	if rule, ok := rules.Load(name); ok {
		return rule.(model.Rule), nil
	}
	return nil, fmt.Errorf("could not find rule %q", name)
}

func GetRules() []model.Rule {
	var rulesSnapshot []model.Rule
	rules.Range(func(_, rule any) bool {
		rulesSnapshot = append(rulesSnapshot, rule.(model.Rule))
		return true
	})

	return rulesSnapshot
}
