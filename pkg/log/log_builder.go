package log

import (
	"bytes"
	"errors"
)

type Builder struct {
	log *Log
}

func NewBuilder() *Builder {
	return &Builder{
		log: &Log{
			level:   InfoLevel,
			hooks:   NewHooks(),
			peaces:  newPool[Piece](),
			entries: newPool[Entry](),
			buffers: newPool[bytes.Buffer](),
		},
	}
}

func (that *Builder) WithLevel(level Level) *Builder {
	that.log.level = level
	return that
}

func (that *Builder) WithExporter(exporter Exporter) *Builder {
	that.log.exporter = exporter
	return that
}

func (that *Builder) WithHook(hook Hook) *Builder {
	that.log.AddHook(Levels.Keys(), hook)
	return that
}

func (that *Builder) WithHookForLevel(hook Hook, level Level) *Builder {
	that.log.AddHook([]Level{level}, hook)
	return that
}

func (that *Builder) WithHookForLevels(hook Hook, levels []Level) *Builder {
	that.log.AddHook(levels, hook)
	return that
}

func (that *Builder) Build() (*Log, error) {
	if err := that.checkRequiredFields(); err != nil {
		return nil, err
	}
	return that.log, nil
}

func (that *Builder) checkRequiredFields() error {
	if that.log.exporter == nil {
		return ErrRequiredFieldExporter
	}

	return nil
}

var (
	ErrRequiredFieldExporter = errors.New("exporter is required")
)

var DefaultLogger = newDummyLogger()

func newDummyLogger() *Log {
	l, err := NewBuilder().
		WithExporter(new(dummyExporter)).
		Build()
	if err != nil {
		panic(err)
	}
	return l
}
