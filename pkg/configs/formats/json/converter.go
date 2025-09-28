package jsonConfig

import "encoding/json"

type Converter struct {
}

func NewConverter() *Converter {
	return &Converter{}
}

func (that *Converter) Convert(dst interface{}, src map[string]interface{}) error {
	raw, err := json.Marshal(src)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw, dst)
	if err != nil {
		return err
	}

	return nil
}
