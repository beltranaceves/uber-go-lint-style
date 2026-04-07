package rules

import "github.com/mgechev/revive/lint"

// NewRule creates a new rule instance by name
func NewRule(name string) (lint.Rule, bool) {
	factories := map[string]func() lint.Rule{
		"atomic":       func() lint.Rule { return &AtomicRule{} },
		"error-wrap":   func() lint.Rule { return &ErrorWrapRule{} },
		"error-name":   func() lint.Rule { return &ErrorNameRule{} },
		"struct-embed": func() lint.Rule { return &StructEmbedRule{} },
		"global-mut":   func() lint.Rule { return &GlobalMutRule{} },
	}

	factory, ok := factories[name]
	if !ok {
		return nil, false
	}
	return factory(), true
}

// GetAllRuleNames returns all available rule names
func GetAllRuleNames() []string {
	return []string{
		"atomic",
		"error-wrap",
		"error-name",
		"struct-embed",
		"global-mut",
	}
}

// GetAllRules returns all rule instances
func GetAllRules() []lint.Rule {
	rules := make([]lint.Rule, 0, 5)
	for _, name := range GetAllRuleNames() {
		if rule, ok := NewRule(name); ok {
			rules = append(rules, rule)
		}
	}
	return rules
}
