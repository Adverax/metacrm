package log

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type myExporter struct {
	entry *Entry
}

func (that *myExporter) Export(ctx context.Context, entry *Entry) {
	that.entry = entry.clone()
}

func TestLogger(t *testing.T) {
	exporter := &myExporter{}

	logger, err := NewBuilder().
		WithLevel(InfoLevel).
		WithExporter(exporter).
		WithHook(HookFunc(func(ctx context.Context, entry *Entry) error {
			entry.Time = time.Time{}
			return nil
		})).
		Build()
	require.NoError(t, err)

	ctx := context.Background()
	err = fmt.Errorf("invalid value")
	logger.
		WithFields(Fields{"key": "value"}).
		WithError(err).
		Error(ctx, "Hello, World2!")

	assert.Equal(t, 2, len(exporter.entry.Data))
	assert.Equal(t, "value", exporter.entry.Data["key"])
	assert.Equal(t, err, exporter.entry.Data["error"])
	assert.Equal(t, "Hello, World2!", exporter.entry.Message)
	assert.Equal(t, ErrorLevel, exporter.entry.Level)
}
