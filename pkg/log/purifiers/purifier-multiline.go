package purifiers

import (
	"encoding/json"
	"strings"
)

type MultilinePurifier struct {
	next Purifier
}

func NewMultilinePurifier(next Purifier) *MultilinePurifier {
	return &MultilinePurifier{next: next}
}

func (that *MultilinePurifier) Purify(original, derivative string) string {
	derivative = that.purify(derivative)

	if that.next == nil {
		return derivative
	}

	return that.next.Purify(original, derivative)
}

func (that *MultilinePurifier) purify(s string) string {
	if s2, ok := that.purifyAsJson(s); ok {
		return s2
	}

	return that.purifyAsPlain(s)
}

func (that *MultilinePurifier) purifyAsJson(s string) (string, bool) {
	if !canBeJson.MatchString(s) {
		return s, false
	}

	var obj interface{}
	if err := json.Unmarshal([]byte(s), &obj); err == nil {
		b, _ := json.Marshal(obj)
		return string(b), true
	}

	return s, false
}

func (that *MultilinePurifier) purifyAsPlain(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.Trim(line, "\n\r\t ")
	}

	return strings.Join(lines, " ")
}
