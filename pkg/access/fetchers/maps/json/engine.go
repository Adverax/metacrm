package jsonFetcher

import (
	"bytes"
	"encoding/json"
)

type Writer interface {
	Save(data []byte) error
}

type Fetcher interface {
	Fetch() ([]byte, error)
}

type Engine struct {
	fetcher Fetcher
}

func New(fetcher Fetcher) *Engine {
	return &Engine{
		fetcher: fetcher,
	}
}

func (that *Engine) Fetch() (map[string]interface{}, error) {
	data := make(map[string]interface{})

	source, err := that.fetcher.Fetch()
	if err != nil {
		return nil, err
	}

	if len(source) == 0 {
		return data, nil
	}

	decoder := json.NewDecoder(bytes.NewBuffer(source))
	err = decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (that *Engine) Save(data map[string]interface{}) error {
	if writer, ok := that.fetcher.(Writer); ok {
		bytes, err := json.Marshal(data)
		if err != nil {
			return err
		}
		return writer.Save(bytes)
	}

	return nil
}
