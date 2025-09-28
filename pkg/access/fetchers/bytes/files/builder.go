package fileFetcher

import "errors"

type Builder struct {
	fetcher *Fetcher
}

func NewBuilder() *Builder {
	return &Builder{
		fetcher: &Fetcher{},
	}
}

func (that *Builder) WithFilename(filename string) *Builder {
	that.fetcher.filename = filename
	return that
}

func (that *Builder) WithMustExists(mustExists bool) *Builder {
	that.fetcher.mustExists = mustExists
	return that
}

func (that *Builder) Build() (*Fetcher, error) {
	if err := that.checkRequiredFields(); err != nil {
		return nil, err
	}

	return that.fetcher, nil
}

func (that *Builder) checkRequiredFields() error {
	if that.fetcher.filename == "" {
		return ErrFilenameRequired
	}
	return nil
}

var (
	ErrFilenameRequired = errors.New("filename is required")
)
