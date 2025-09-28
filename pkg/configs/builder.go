package configs

import "fmt"

type Builder struct {
	loader *Loader
}

func NewBuilder() *Builder {
	return &Builder{
		loader: &Loader{
			converter: DefaultConverter,
		},
	}
}

func (that *Builder) WithSource(sources ...Source) *Builder {
	that.loader.sources = append(that.loader.sources, sources...)
	return that
}

func (that *Builder) WithConverter(converter Converter) *Builder {
	that.loader.converter = converter
	return that
}

func (that *Builder) WithDistinct(distinct bool) *Builder {
	that.loader.distinct = distinct
	return that
}

func (that *Builder) Build() (*Loader, error) {
	if err := that.checkRequiredFields(); err != nil {
		return nil, err
	}

	return that.loader, nil
}

func (that *Builder) checkRequiredFields() error {
	if len(that.loader.sources) == 0 {
		return ErrFieldSourcesIsRequired
	}

	if that.loader.converter == nil {
		return ErrFieldConverterIsRequired
	}

	return nil
}

var (
	ErrFieldSourcesIsRequired = fmt.Errorf("Field sources is required")
)
