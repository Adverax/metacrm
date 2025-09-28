package envFetcher

import (
	"os"
	"strings"
)

type Guard interface {
	IsSatisfied(text string) (key string, matched bool)
}

type Accumulator interface {
	Add(key, value string)
	Result() map[string]interface{}
}

type Engine struct {
	guard       Guard
	accumulator Accumulator
}

func New(guard Guard, accumulator Accumulator) *Engine {
	return &Engine{
		guard:       guard,
		accumulator: accumulator,
	}
}

func (that *Engine) Fetch() (map[string]interface{}, error) {
	return that.fetch(os.Environ())
}

func (that *Engine) fetch(es []string) (map[string]interface{}, error) {
	for _, e := range es {
		ss := strings.Split(e, "=")
		if len(ss) < 2 {
			continue
		}

		key, ok := that.guard.IsSatisfied(ss[0])
		if !ok {
			continue
		}

		that.accumulator.Add(key, ss[1])
	}

	return that.accumulator.Result(), nil
}
