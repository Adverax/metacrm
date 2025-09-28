package configs

import (
	"context"
	"encoding/json"
	"fmt"
)

type Loader struct {
	sources   []Source
	converter Converter
	distinct  bool
	hash      string
}

func (that *Loader) Load(config interface{}) error {
	ds, err := that.load(that.sources...)
	if err != nil {
		return err
	}

	if that.distinct {
		hash := that.hashOf(ds)
		if hash == that.hash {
			return ErrDistinct
		}
	}

	data := that.merge(ds)

	err = that.converter.Convert(config, data)
	if err != nil {
		return fmt.Errorf("error convert config: %w", err)
	}

	return nil
}

func (that *Loader) load(sources ...Source) ([]map[string]interface{}, error) {
	ds := make([]map[string]interface{}, 0, len(sources))
	for _, source := range sources {
		d, err := source.Fetch()
		if err != nil {
			return nil, fmt.Errorf("error in source: %w", err)
		}

		ds = append(ds, d)
	}
	return ds, nil
}

func (that *Loader) merge(ds []map[string]interface{}) map[string]interface{} {
	data := make(map[string]interface{})

	for _, d := range ds {
		override(data, d)
	}

	return data
}

func (that *Loader) hashOf(data []map[string]interface{}) string {
	hashs := make([]string, 0, len(data))
	for _, d := range data {
		hashs = append(hashs, hashOf(d))
	}

	bs, _ := json.Marshal(hashs)
	return digestOf(bs)
}

type defaultConverter struct{}

func (that *defaultConverter) Convert(dst interface{}, src map[string]interface{}) error {
	return Assign(context.Background(), dst, src)
}

var (
	DefaultConverter = &defaultConverter{}
)

var (
	ErrDistinct = fmt.Errorf("config without changes")
)
