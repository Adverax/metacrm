package log

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"reflect"
	"time"
)

var ErrorKey = "error"

type Entry struct {
	Logger  *Log
	Data    Fields
	Time    time.Time
	Level   Level
	Message string
	Buffer  *bytes.Buffer
	LogErr  string
}

func NewEntry(logger *Log) *Entry {
	return &Entry{
		Logger: logger,
		Data:   make(Fields, 6),
	}
}

func (that *Entry) IsLevelEnabled(level Level) bool {
	return that.Logger.IsLevelEnabled(level)
}

func (that *Entry) clear() {
	that.Logger = nil
	that.Data = map[string]interface{}{}
	that.Time = time.Time{}
	that.Level = 0
	that.Message = ""
	that.Buffer = nil
	that.LogErr = ""
}

func (that *Entry) clone() *Entry {
	newEntry := that.Logger.newEntry()
	newEntry.Data = that.Data
	newEntry.Time = that.Time
	newEntry.Level = that.Level
	newEntry.Message = that.Message
	newEntry.Logger = that.Logger
	return newEntry
}

func (that *Entry) WithField(key string, value interface{}) Logger {
	return that.withFields(Fields{key: value})
}

func (that *Entry) withField(key string, value interface{}) *Entry {
	return that.withFields(Fields{key: value})
}

func (that *Entry) WithFields(fields Fields) Logger {
	return that.withFields(fields)
}

func (that *Entry) withFields(fields Fields) *Entry {
	data, err := that.expandData(fields)
	return &Entry{
		Logger: that.Logger,
		Data:   data,
		Time:   that.Time,
		LogErr: err,
	}
}

func (that *Entry) WithError(err error) Logger {
	return that.withField(ErrorKey, err)
}

func (that *Entry) withError(err error) *Entry {
	return that.withField(ErrorKey, err)
}

func (that *Entry) WithTime(t time.Time) Logger {
	return &Entry{
		Logger: that.Logger,
		Data:   that.Data.Clone(),
		Time:   t,
		LogErr: that.LogErr,
	}
}

func (that *Entry) Panicf(ctx context.Context, format string, args ...interface{}) {
	that.Logf(ctx, PanicLevel, format, args...)
}

func (that *Entry) Fatalf(ctx context.Context, format string, args ...interface{}) {
	that.Logf(ctx, FatalLevel, format, args...)
}

func (that *Entry) Errorf(ctx context.Context, format string, args ...interface{}) {
	that.Logf(ctx, ErrorLevel, format, args...)
}

func (that *Entry) Warningf(ctx context.Context, format string, args ...interface{}) {
	that.Logf(ctx, WarnLevel, format, args...)
}

func (that *Entry) Noticef(ctx context.Context, format string, args ...interface{}) {
	that.Logf(ctx, NoticeLevel, format, args...)
}

func (that *Entry) Infof(ctx context.Context, format string, args ...interface{}) {
	that.Logf(ctx, InfoLevel, format, args...)
}

func (that *Entry) Debugf(ctx context.Context, format string, args ...interface{}) {
	that.Logf(ctx, DebugLevel, format, args...)
}

func (that *Entry) Tracef(ctx context.Context, format string, args ...interface{}) {
	that.Logf(ctx, TraceLevel, format, args...)
}

func (that *Entry) Panic(ctx context.Context, args ...interface{}) {
	that.Log(ctx, PanicLevel, args...)
}

func (that *Entry) Fatal(ctx context.Context, args ...interface{}) {
	that.Log(ctx, FatalLevel, args...)
}

func (that *Entry) Error(ctx context.Context, args ...interface{}) {
	that.Log(ctx, ErrorLevel, args...)
}

func (that *Entry) Warning(ctx context.Context, args ...interface{}) {
	that.Log(ctx, WarnLevel, args...)
}

func (that *Entry) Notice(ctx context.Context, args ...interface{}) {
	that.Log(ctx, NoticeLevel, args...)
}

func (that *Entry) Info(ctx context.Context, args ...interface{}) {
	that.Log(ctx, InfoLevel, args...)
}

func (that *Entry) Debug(ctx context.Context, args ...interface{}) {
	that.Log(ctx, DebugLevel, args...)
}

func (that *Entry) Trace(ctx context.Context, args ...interface{}) {
	that.Log(ctx, TraceLevel, args...)
}

func (that *Entry) Logf(ctx context.Context, level Level, format string, args ...interface{}) {
	if that.Logger.IsLevelEnabled(level) {
		that.log(ctx, level, fmt.Sprintf(format, args...))
	}
}

func (that *Entry) PanicFn(ctx context.Context, fn LogFunction) {
	that.LogFn(ctx, PanicLevel, fn)
}

func (that *Entry) FatalFn(ctx context.Context, fn LogFunction) {
	that.LogFn(ctx, FatalLevel, fn)
}

func (that *Entry) ErrorFn(ctx context.Context, fn LogFunction) {
	that.LogFn(ctx, ErrorLevel, fn)
}

func (that *Entry) WarningFn(ctx context.Context, fn LogFunction) {
	that.LogFn(ctx, WarnLevel, fn)
}

func (that *Entry) NoticeFn(ctx context.Context, fn LogFunction) {
	that.LogFn(ctx, NoticeLevel, fn)
}

func (that *Entry) InfoFn(ctx context.Context, fn LogFunction) {
	that.LogFn(ctx, InfoLevel, fn)
}

func (that *Entry) DebugFn(ctx context.Context, fn LogFunction) {
	that.LogFn(ctx, DebugLevel, fn)
}

func (that *Entry) TraceFn(ctx context.Context, fn LogFunction) {
	that.LogFn(ctx, TraceLevel, fn)
}

func (that *Entry) Log(ctx context.Context, level Level, args ...interface{}) {
	if that.Logger.IsLevelEnabled(level) {
		that.log(ctx, level, fmt.Sprint(args...))
	}
}

func (that *Entry) log(ctx context.Context, level Level, msg string) {
	entry := that.clone()
	defer that.Logger.freeEntry(entry)

	entry.prepare(level, msg)
	entry.fire(ctx)
	entry.Logger.exporter.Export(ctx, entry)

	if entry.Level <= PanicLevel {
		panic(entry)
	}
}

func (that *Entry) LogFn(ctx context.Context, level Level, fn LogFunction) {
	if that.Logger.IsLevelEnabled(level) {
		entry := that.Logger.newEntry()
		defer that.Logger.freeEntry(entry)

		entry.Level = level
		entry.Time = that.Time
		entry.LogErr = that.LogErr
		entry.Data = make(Fields, len(that.Data))
		for k, v := range that.Data {
			entry.Data[k] = v
		}

		fn(ctx, entry)
	}
}

func (that *Entry) prepare(level Level, msg string) {
	if that.Time.IsZero() {
		that.Time = time.Now()
	}

	that.Level = level
	that.Message = msg
}

func (that *Entry) fire(ctx context.Context) {
	err := that.Logger.hooks.Fire(ctx, that.Level, that)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to fire hook: %v\n", err)
	}
}

func (that *Entry) expandData(fields Fields) (Fields, string) {
	data := make(Fields, len(that.Data)+len(fields))
	for k, v := range that.Data {
		data[k] = v
	}

	fieldErr := that.LogErr
	for k, v := range fields {
		isErrField := false
		if t := reflect.TypeOf(v); t != nil {
			switch {
			case t.Kind() == reflect.Func, t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Func:
				isErrField = true
			}
		}
		if isErrField {
			tmp := fmt.Sprintf("can not add field %q", k)
			if fieldErr != "" {
				fieldErr = that.LogErr + ", " + tmp
			} else {
				fieldErr = tmp
			}
		} else {
			data[k] = v
		}
	}

	return data, fieldErr
}
