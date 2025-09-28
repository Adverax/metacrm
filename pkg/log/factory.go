package log

import "sync"

type Constructor func(exporter Exporter, level Level) (Logger, error)

type Factory struct {
	sync.Mutex
	logs        map[Level]Logger
	exporter    Exporter
	constructor Constructor
}

func (that *Factory) NewLogger(level Level) (Logger, error) {
	that.Lock()
	defer that.Unlock()

	if l, ok := that.logs[level]; ok {
		return l, nil
	}

	l, err := that.constructor(that.exporter, level)
	if err != nil {
		return nil, err
	}

	that.logs[level] = l
	return l, nil
}

func NewFactory(exporter Exporter, constructor Constructor) *Factory {
	if constructor == nil {
		constructor = DefaultConstructor
	}

	return &Factory{
		constructor: constructor,
		logs:        make(map[Level]Logger),
		exporter:    exporter,
	}
}

var DefaultConstructor Constructor = func(exporter Exporter, level Level) (Logger, error) {
	return NewBuilder().
		WithExporter(exporter).
		WithLevel(level).
		Build()
}
