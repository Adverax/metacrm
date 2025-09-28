package enums

import (
	"encoding/json"
	"fmt"
)

type Enum[T comparable] struct {
	encoders map[T]string
	decoders map[string]T
	keys     []T
	values   []string
}

func (that *Enum[T]) Encode(val string) (res T, err error) {
	if v, ok := that.decoders[val]; ok {
		return v, nil
	}
	return res, fmt.Errorf("unknown value %s of enum", val)
}

func (that *Enum[T]) Decode(val T) (string, error) {
	if v, ok := that.encoders[val]; ok {
		return v, nil
	}
	return "", fmt.Errorf("not a valid value of enum %v", val)
}

func (that *Enum[T]) EncodeOrDefault(val string, def T) T {
	if v, ok := that.decoders[val]; ok {
		return v
	}
	return def
}

func (that *Enum[T]) DecodeOrDefault(val T, def string) string {
	if v, ok := that.encoders[val]; ok {
		return v
	}
	return def
}

func (that *Enum[T]) Keys() []T {
	return that.keys
}

func (that *Enum[T]) Values() []string {
	return that.values
}

func (that *Enum[T]) UnmarshalText(text []byte, val *T) error {
	v, err := that.Encode(string(text))
	if err != nil {
		return err
	}

	*val = v
	return nil
}

func (that *Enum[T]) MarshalText(val T) ([]byte, error) {
	s, err := that.Decode(val)
	if err != nil {
		return nil, err
	}

	return []byte(s), nil
}

func (that *Enum[T]) MarshalJSON(val T) ([]byte, error) {
	s, err := that.Decode(val)
	if err != nil {
		return nil, err
	}

	return json.Marshal(s)
}

func (that *Enum[T]) UnmarshalJSON(data []byte, val *T) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	v, err := that.Encode(s)
	if err != nil {
		return err
	}

	*val = v
	return nil
}

func New[T comparable](encoders map[T]string) *Enum[T] {
	decoders := make(map[string]T, len(encoders))
	var keys []T
	var values []string
	for k, v := range encoders {
		decoders[v] = k
		keys = append(keys, k)
		values = append(values, v)
	}

	return &Enum[T]{
		encoders: encoders,
		decoders: decoders,
		keys:     keys,
		values:   values,
	}
}
