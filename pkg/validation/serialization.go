package validation

import (
	"encoding/json"
	"fmt"
)

type RuleType string

type ValidatorDTO struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Unmarshaller func(data []byte) (RuleEx, error)

var unmarshalers = make(map[RuleType]Unmarshaller)

func RegisterUnmarshaller(t RuleType, unmarshaller Unmarshaller) {
	unmarshalers[t] = unmarshaller
}

func RegisterRule(rule RuleEx) {
	RegisterUnmarshaller(rule.RuleType(), func(data []byte) (RuleEx, error) {
		return rule, nil
	})
}

func MarshalRule(r RuleEx) ([]byte, error) {
	var dto ValidatorDTO
	dto.Type = string(r.RuleType())
	data, err := r.MarshalJSON()
	if err != nil {
		return nil, err
	}
	dto.Data = data
	return json.Marshal(dto)
}

func UnmarshalRule(data []byte) (RuleEx, error) {
	var dto ValidatorDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, err
	}
	if unm, ok := unmarshalers[RuleType(dto.Type)]; ok {
		return unm(dto.Data)
	}

	return nil, fmt.Errorf("unknown rule type: %s", dto.Type)
}

func UnmarshalTypedRule(data []byte, ruleType RuleType) (RuleEx, error) {
	if unm, ok := unmarshalers[ruleType]; ok {
		return unm(data)
	}
	return nil, fmt.Errorf("unknown rule type: %s", ruleType)
}

func MarshalRules(rules []RuleEx) ([]byte, error) {
	var rs []json.RawMessage
	for _, rule := range rules {
		data, err := MarshalRule(rule)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal rule: %w", err)
		}
		rs = append(rs, data)
	}
	return json.Marshal(rs)
}

func UnmarshalRules(data []byte) (rules []RuleEx, err error) {
	var rs []json.RawMessage
	if err := json.Unmarshal(data, &rs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
	}

	rules = make([]RuleEx, len(rs))
	for i, raw := range rs {
		rule, err := UnmarshalRule(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal rule: %w", err)
		}
		rules[i] = rule
	}
	return rules, nil
}

func rulesEx2Rules(rules []RuleEx) []Rule {
	var rs []Rule
	for _, rule := range rules {
		rs = append(rs, rule)
	}
	return rs
}

type RuleExs []RuleEx

func (r RuleExs) MarshalJSON() ([]byte, error) {
	var rs []json.RawMessage
	for _, rule := range r {
		data, err := MarshalRule(rule)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal rule: %w", err)
		}
		rs = append(rs, data)
	}
	return json.Marshal(rs)
}

func (r *RuleExs) UnmarshalJSON(data []byte) error {
	var rs []json.RawMessage
	if err := json.Unmarshal(data, &rs); err != nil {
		return fmt.Errorf("failed to unmarshal rules: %w", err)
	}

	*r = make(RuleExs, len(rs))
	for i, raw := range rs {
		rule, err := UnmarshalRule(raw)
		if err != nil {
			return fmt.Errorf("failed to unmarshal rule: %w", err)
		}
		(*r)[i] = rule
	}
	return nil
}
