package yamlConfig

import "gopkg.in/yaml.v3"

type Converter struct {
}

func NewConverter() *Converter {
	return &Converter{}
}

func (that *Converter) Convert(dst interface{}, src map[string]interface{}) error {
	raw, err := yaml.Marshal(src)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(raw, dst)
	if err != nil {
		return err
	}

	return nil
}
