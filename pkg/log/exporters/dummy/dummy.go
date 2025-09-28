package dummyExporter

import (
	"context"

	"github.com/adverax/metacrm/pkg/log"
)

type Exporter struct {
}

func New() *Exporter {
	return &Exporter{}
}

func (that *Exporter) Export(_ context.Context, _ *log.Entry) {
	// nothing
}
