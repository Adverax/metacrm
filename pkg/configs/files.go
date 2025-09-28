package configs

import (
	"fmt"
	fileFetchers "github.com/adverax/metacrm/pkg/access/fetchers/bytes/files"
)

type SourceBuilder func(Fetcher) Source

type FileLoaderBuilder struct {
	sources   []Source
	builder   SourceBuilder
	converter Converter
	err       error
}

func NewFileLoaderBuilder() *FileLoaderBuilder {
	return &FileLoaderBuilder{
		converter: DefaultConverter,
	}
}

func (that *FileLoaderBuilder) WithSourceBuilder(builder SourceBuilder) *FileLoaderBuilder {
	that.builder = builder
	return that
}

func (that *FileLoaderBuilder) WithSource(sources ...Source) *FileLoaderBuilder {
	that.sources = append(that.sources, sources...)
	return that
}

func (that *FileLoaderBuilder) WithFile(file string, mustExists bool) *FileLoaderBuilder {
	fetcher, err := fileFetchers.NewBuilder().
		WithFilename(file).
		WithMustExists(mustExists).
		Build()
	if that.err != nil {
		that.err = err
		return that
	}

	that.sources = append(that.sources, that.builder(fetcher))
	return that
}

func (that *FileLoaderBuilder) WithConverter(converter Converter) *FileLoaderBuilder {
	that.converter = converter
	return that
}

func (that *FileLoaderBuilder) Build() (*Loader, error) {
	if err := that.checkRequiredFields(); err != nil {
		return nil, err
	}

	return NewBuilder().
		WithSource(that.sources...).
		WithConverter(that.converter).
		Build()
}

func (that *FileLoaderBuilder) checkRequiredFields() error {
	if that.err != nil {
		return that.err
	}

	if len(that.sources) == 0 {
		return ErrFieldFilesIsRequired
	}

	if that.builder == nil {
		return ErrFieldBuilderIsRequired
	}

	if that.converter == nil {
		return ErrFieldConverterIsRequired
	}

	return nil
}

var (
	ErrFieldFilesIsRequired     = fmt.Errorf("Files are required")
	ErrFieldBuilderIsRequired   = fmt.Errorf("Builder is required")
	ErrFieldConverterIsRequired = fmt.Errorf("Converter is required")
)
