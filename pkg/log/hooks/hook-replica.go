package hooks

import (
	"context"

	"github.com/adverax/metacrm/pkg/log"
)

type HookReplica struct {
	exporter log.Exporter
}

func NewHookReplica(
	exporter log.Exporter,
) *HookReplica {
	return &HookReplica{
		exporter: exporter,
	}
}

func (that *HookReplica) Fire(ctx context.Context, entry *log.Entry) error {
	that.exporter.Export(ctx, entry)
	return nil
}
