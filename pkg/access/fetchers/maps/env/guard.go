package envFetcher

import "strings"

type PrefixGuard struct {
	prefix string
}

func NewPrefixGuard(prefix string) *PrefixGuard {
	return &PrefixGuard{
		prefix: prefix,
	}
}

func (that *PrefixGuard) IsSatisfied(text string) (key string, matched bool) {
	if strings.HasPrefix(text, that.prefix) {
		return text[len(that.prefix):], true
	}

	return "", false
}
