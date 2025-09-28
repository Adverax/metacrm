package log

import (
	"bytes"
	"context"
	"sync"
	"time"
)

type Log struct {
	exporter Exporter
	level    Level
	mu       sync.Mutex
	hooks    *Hooks
	peaces   *pool[Piece]
	entries  *pool[Entry]
	buffers  *pool[bytes.Buffer]
}

func (that *Log) WithField(key string, value interface{}) Logger {
	entry := that.newEntry()
	defer that.freeEntry(entry)
	return entry.WithField(key, value)
}

func (that *Log) WithFields(fields Fields) Logger {
	entry := that.newEntry()
	defer that.freeEntry(entry)
	return entry.WithFields(fields)
}

func (that *Log) WithError(err error) Logger {
	entry := that.newEntry()
	defer that.freeEntry(entry)
	return entry.WithError(err)
}

func (that *Log) WithTime(t time.Time) Logger {
	entry := that.newEntry()
	defer that.freeEntry(entry)
	return entry.WithTime(t)
}

func (that *Log) PanicFn(ctx context.Context, fn LogFunction) {
	that.LogFn(ctx, PanicLevel, fn)
}

func (that *Log) FatalFn(ctx context.Context, fn LogFunction) {
	that.LogFn(ctx, FatalLevel, fn)
}

func (that *Log) ErrorFn(ctx context.Context, fn LogFunction) {
	that.LogFn(ctx, ErrorLevel, fn)
}

func (that *Log) WarningFn(ctx context.Context, fn LogFunction) {
	that.LogFn(ctx, WarnLevel, fn)
}

func (that *Log) NoticeFn(ctx context.Context, fn LogFunction) {
	that.LogFn(ctx, NoticeLevel, fn)
}

func (that *Log) InfoFn(ctx context.Context, fn LogFunction) {
	that.LogFn(ctx, InfoLevel, fn)
}

func (that *Log) DebugFn(ctx context.Context, fn LogFunction) {
	that.LogFn(ctx, DebugLevel, fn)
}

func (that *Log) TraceFn(ctx context.Context, fn LogFunction) {
	that.LogFn(ctx, TraceLevel, fn)
}

func (that *Log) Panicf(ctx context.Context, format string, args ...interface{}) {
	that.Logf(ctx, PanicLevel, format, args...)
}

func (that *Log) Fatalf(ctx context.Context, format string, args ...interface{}) {
	that.Logf(ctx, FatalLevel, format, args...)
}

func (that *Log) Errorf(ctx context.Context, format string, args ...interface{}) {
	that.Logf(ctx, ErrorLevel, format, args...)
}

func (that *Log) Warningf(ctx context.Context, format string, args ...interface{}) {
	that.Logf(ctx, WarnLevel, format, args...)
}

func (that *Log) Noticef(ctx context.Context, format string, args ...interface{}) {
	that.Logf(ctx, NoticeLevel, format, args...)
}

func (that *Log) Infof(ctx context.Context, format string, args ...interface{}) {
	that.Logf(ctx, InfoLevel, format, args...)
}

func (that *Log) Debugf(ctx context.Context, format string, args ...interface{}) {
	that.Logf(ctx, DebugLevel, format, args...)
}

func (that *Log) Tracef(ctx context.Context, format string, args ...interface{}) {
	that.Logf(ctx, TraceLevel, format, args...)
}

func (that *Log) Panic(ctx context.Context, args ...interface{}) {
	that.Log(ctx, PanicLevel, args...)
}

func (that *Log) Fatal(ctx context.Context, args ...interface{}) {
	that.Log(ctx, FatalLevel, args...)
}

func (that *Log) Error(ctx context.Context, args ...interface{}) {
	that.Log(ctx, ErrorLevel, args...)
}

func (that *Log) Warning(ctx context.Context, args ...interface{}) {
	that.Log(ctx, WarnLevel, args...)
}

func (that *Log) Notice(ctx context.Context, args ...interface{}) {
	that.Log(ctx, NoticeLevel, args...)
}

func (that *Log) Info(ctx context.Context, args ...interface{}) {
	that.Log(ctx, InfoLevel, args...)
}

func (that *Log) Debug(ctx context.Context, args ...interface{}) {
	that.Log(ctx, DebugLevel, args...)
}

func (that *Log) Trace(ctx context.Context, args ...interface{}) {
	that.Log(ctx, TraceLevel, args...)
}

func (that *Log) LogFn(ctx context.Context, level Level, fn LogFunction) {
	if that.IsLevelEnabled(level) {
		entry := that.newEntry()
		defer that.freeEntry(entry)

		entry.Level = level
		fn(ctx, entry)
	}
}

func (that *Log) Logf(ctx context.Context, level Level, format string, args ...interface{}) {
	if that.IsLevelEnabled(level) {
		entry := that.newEntry()
		defer that.freeEntry(entry)

		entry.Logf(ctx, level, format, args...)
	}
}

func (that *Log) Log(ctx context.Context, level Level, args ...interface{}) {
	if that.IsLevelEnabled(level) {
		entry := that.newEntry()
		defer that.freeEntry(entry)

		entry.Log(ctx, level, args...)
	}
}

func (that *Log) IsLevelEnabled(level Level) bool {
	return that.level >= level
}

func (that *Log) AddHook(levels []Level, hook Hook) {
	that.hooks.Add(levels, hook)
}

func (that *Log) newEntry() *Entry {
	entry := that.entries.Get()
	entry.clear()
	entry.Logger = that
	return entry
}

func (that *Log) freeEntry(entry *Entry) {
	entry.Data = map[string]interface{}{}
	that.entries.Put(entry)
}

func (that *Log) newPeace() *Piece {
	peace := that.peaces.Get()
	peace.entry = that.newEntry()
	return peace
}

func (that *Log) freePeace(peace *Piece) {
	entry := peace.entry
	peace.entry = nil
	that.freeEntry(entry)
	that.peaces.Put(peace)
}

func (that *Log) GetBuffer() *bytes.Buffer {
	return that.buffers.Get()
}

func (that *Log) FreeBuffer(buffer *bytes.Buffer) {
	that.buffers.Put(buffer)
}
