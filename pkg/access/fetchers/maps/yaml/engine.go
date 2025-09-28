package yamlFetcher

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

type Writer interface {
	Save([]byte) error
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

	decoder := yaml.NewDecoder(bytes.NewBuffer(source))
	err = decoder.Decode(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (that *Engine) Save(data map[string]interface{}) error {
	if writer, ok := that.fetcher.(Writer); ok {
		buf := bytes.NewBuffer(nil)
		encoder := yaml.NewEncoder(buf)
		err := encoder.Encode(data)
		if err != nil {
			return err
		}
		return writer.Save(buf.Bytes())
	}

	return nil
}
