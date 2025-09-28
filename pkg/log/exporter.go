package log

import (
	"context"
)

type Exporter interface {
	Export(ctx context.Context, entry *Entry)
}

type dummyExporter struct{}

func (that *dummyExporter) Export(ctx context.Context, entry *Entry) {
	// nothing
}
