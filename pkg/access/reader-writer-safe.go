package access

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

type SafeReaderWriter struct {
	sync.RWMutex
	ReaderWriter
}

func (that *SafeReaderWriter) GetProperty(
	ctx context.Context,
	name string,
) (interface{}, error) {
	that.RLock()
	defer that.RUnlock()

	return that.ReaderWriter.GetProperty(ctx, name)
}

func (that *SafeReaderWriter) SetProperty(
	ctx context.Context,
	name string,
	value interface{},
) error {
	that.Lock()
	defer that.Unlock()

	return that.ReaderWriter.SetProperty(ctx, name, value)
}

func (that *SafeReaderWriter) GetBoolean(
	ctx context.Context,
	name string,
	defVal bool,
) (bool, error) {
	that.RLock()
	defer that.RUnlock()

	return that.ReaderWriter.GetBoolean(ctx, name, defVal)
}

func (that *SafeReaderWriter) GetString(
	ctx context.Context,
	name string,
	defVal string,
) (string, error) {
	that.RLock()
	defer that.RUnlock()

	return that.ReaderWriter.GetString(ctx, name, defVal)
}

func (that *SafeReaderWriter) GetInteger(
	ctx context.Context,
	name string,
	defVal int64,
) (int64, error) {
	that.RLock()
	defer that.RUnlock()

	return that.ReaderWriter.GetInteger(ctx, name, defVal)
}

func (that *SafeReaderWriter) GetFloat(
	ctx context.Context,
	name string,
	defVal float64,
) (float64, error) {
	that.RLock()
	defer that.RUnlock()

	return that.ReaderWriter.GetFloat(ctx, name, defVal)
}

func (that *SafeReaderWriter) GetDuration(
	ctx context.Context,
	name string,
	defVal time.Duration,
) (time.Duration, error) {
	that.RLock()
	defer that.RUnlock()

	return that.ReaderWriter.GetDuration(ctx, name, defVal)
}

func (that *SafeReaderWriter) GetJson(
	ctx context.Context,
	name string,
	defVal json.RawMessage,
) (json.RawMessage, error) {
	that.RLock()
	defer that.RUnlock()

	return that.ReaderWriter.GetJson(ctx, name, defVal)
}

func (that *SafeReaderWriter) SetBoolean(
	ctx context.Context,
	name string,
	value bool,
) error {
	that.Lock()
	defer that.Unlock()

	return that.ReaderWriter.SetBoolean(ctx, name, value)
}

func (that *SafeReaderWriter) SetString(
	ctx context.Context,
	name string,
	value string,
) error {
	that.Lock()
	defer that.Unlock()

	return that.ReaderWriter.SetString(ctx, name, value)
}

func (that *SafeReaderWriter) SetInteger(
	ctx context.Context,
	name string,
	value int64,
) error {
	that.Lock()
	defer that.Unlock()

	return that.ReaderWriter.SetInteger(ctx, name, value)
}

func (that *SafeReaderWriter) SetFloat(
	ctx context.Context,
	name string,
	value float64,
) error {
	that.Lock()
	defer that.Unlock()

	return that.ReaderWriter.SetFloat(ctx, name, value)
}

func (that *SafeReaderWriter) SetDuration(
	ctx context.Context,
	name string,
	value time.Duration,
) error {
	that.Lock()
	defer that.Unlock()

	return that.ReaderWriter.SetDuration(ctx, name, value)
}

func (that *SafeReaderWriter) Transaction(
	ctx context.Context,
	action func(ctx context.Context, rw ReaderWriter) error,
) error {
	that.Lock()
	defer that.Unlock()

	return action(ctx, that.ReaderWriter)
}

func (that *SafeReaderWriter) SetJson(
	ctx context.Context,
	name string,
	value json.RawMessage,
) error {
	that.Lock()
	defer that.Unlock()

	return that.ReaderWriter.SetJson(ctx, name, value)
}

func NewSafeReaderWriter(rw ReaderWriter) *SafeReaderWriter {
	return &SafeReaderWriter{
		ReaderWriter: rw,
	}
}
