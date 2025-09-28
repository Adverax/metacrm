package envFetcher

import "strings"

type KeyPathAccumulator struct {
	data  map[string]interface{}
	delim string
}

func NewKeyPathAccumulator(delim string) *KeyPathAccumulator {
	return &KeyPathAccumulator{
		data:  make(map[string]interface{}),
		delim: delim,
	}
}

func (that *KeyPathAccumulator) Add(key, value string) {
	keys := strings.Split(key, that.delim)
	that.add(that.data, keys, value)
}

func (that *KeyPathAccumulator) add(data map[string]interface{}, keys []string, val string) {
	if len(keys) == 0 {
		return
	}

	key := keys[0]
	if len(keys) == 1 {
		data[key] = val
		return
	}

	if _, ok := data[key]; !ok {
		data[key] = make(map[string]interface{})
	}

	if space, ok := data[key].(map[string]interface{}); ok {
		that.add(space, keys[1:], val)
		return
	}

	space := make(map[string]interface{})
	data[key] = space
	that.add(space, keys[1:], val)
}

func (that *KeyPathAccumulator) Result() map[string]interface{} {
	return that.data
}
