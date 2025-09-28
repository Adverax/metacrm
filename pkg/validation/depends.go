package validation

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	RuleTypeDependsOn = "depends_on"
)

// DependsOn returns a validation rule that executes the given list of rules dependsOn the condition is true.
func DependsOn(condition string, rules ...RuleEx) DependsOnRule {
	return DependsOnRule{
		dependsOnRuleOptions: dependsOnRuleOptions{
			Condition: condition,
			Rules:     rules,
		},
		rules:     rulesEx2Rules(rules),
		elseRules: []Rule{},
		errs:      defaultCELErrors,
	}
}

type dependsOnRuleOptions struct {
	Condition string  `json:"condition"` // Condition to check
	Rules     RuleExs `json:"rules"`
	ElseRules RuleExs `json:"else_rules,omitempty"` // Rules to execute if the condition is false
}

// DependsOnRule is a validation rule that executes the given list of rules dependsOn the condition is true.
type DependsOnRule struct {
	dependsOnRuleOptions
	rules     []Rule
	elseRules []Rule
	errs      celErrors
}

func (r DependsOnRule) RuleType() RuleType {
	return RuleTypeDependsOn
}

func (r *DependsOnRule) MarshalJSON() ([]byte, error) {
	var rule struct {
		Condition string          `json:"condition"`
		Rules     json.RawMessage `json:"rules"`
		ElseRules json.RawMessage `json:"else_rules,omitempty"`
	}

	var err error
	rule.Condition = r.Condition
	rule.Rules, err = MarshalRules(r.dependsOnRuleOptions.Rules)
	if err != nil {
		return nil, err
	}
	rule.ElseRules, err = MarshalRules(r.dependsOnRuleOptions.ElseRules)
	if err != nil {
		return nil, err
	}

	return json.Marshal(rule)
}

func (r *DependsOnRule) UnmarshalJSON(data []byte) error {
	var rule struct {
		Condition string          `json:"condition"`
		Rules     json.RawMessage `json:"rules"`
		ElseRules json.RawMessage `json:"else_rules,omitempty"`
	}
	if err := json.Unmarshal(data, &rule); err != nil {
		return err
	}

	var err error
	r.Condition = rule.Condition
	r.dependsOnRuleOptions.Rules, err = UnmarshalRules(rule.Rules)
	if err != nil {
		return err
	}
	r.dependsOnRuleOptions.ElseRules, err = UnmarshalRules(rule.ElseRules)
	if err != nil {
		return err
	}

	r.rules = rulesEx2Rules(r.Rules)
	r.elseRules = rulesEx2Rules(r.ElseRules)
	r.errs = defaultCELErrors

	return nil
}

// Validate checks if the condition is true and if so, it validates the value using the specified rules.
func (r DependsOnRule) Validate(ctx context.Context, value interface{}) error {
	condition, err := validateCriteria(ctx, r.Condition, value, &r.errs)
	if err != nil {
		return err
	}

	if condition {
		return Validate(ctx, value, r.rules...)
	}

	return Validate(ctx, value, r.elseRules...)
}

// Else returns a validation rule that executes the given list of rules dependsOn the condition is false.
func (r DependsOnRule) Else(rules ...RuleEx) DependsOnRule {
	r.elseRules = rulesEx2Rules(rules)
	r.ElseRules = rules
	return r
}

func init() {
	RegisterUnmarshaller(RuleTypeDependsOn, func(data []byte) (RuleEx, error) {
		rule := DependsOn("")
		err := json.Unmarshal(data, &rule)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal rule: %w", err)
		}
		return &rule, nil
	})
}
