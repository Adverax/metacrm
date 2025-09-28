package log

import (
	"context"
	"time"

	"github.com/adverax/metacrm/pkg/enums"
)

type Level uint8

func (that Level) String() string {
	return Levels.DecodeOrDefault(that, "unknown")
}

const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel Level = iota
	// FatalLevel level. Logs and then calls `logger.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// NoticeLevel level. General operational entries about what's going on inside the
	// application.
	NoticeLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
	// TraceLevel level. Designates finer-grained informational events than the Debug.
	TraceLevel
)

var Levels = enums.New[Level](
	map[Level]string{
		PanicLevel:  "panic",
		FatalLevel:  "fatal",
		ErrorLevel:  "error",
		WarnLevel:   "warn",
		NoticeLevel: "notice",
		InfoLevel:   "info",
		DebugLevel:  "debug",
		TraceLevel:  "trace",
	},
)

type LogFunction func(ctx context.Context, logger Logger)

type Fields map[string]interface{}

func (that Fields) Fetch(key string) interface{} {
	if that == nil {
		return nil
	}
	if val, ok := that[key]; ok {
		return val
	}
	return nil
}

func (that Fields) Clone() Fields {
	clone := make(map[string]interface{}, len(that))
	for k, v := range that {
		clone[k] = v
	}

	return clone
}

func (that Fields) Expand() Fields {
	data := make(Fields, len(that)+4)
	for k, v := range that {
		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/sirupsen/logrus/issues/137
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}
	return data
}

type LoggerPiece interface {
	WithField(key string, value interface{}) LoggerPiece
	WithFields(fields Fields) LoggerPiece
	WithError(err error) LoggerPiece

	Message(ctx context.Context, msg string)
	Messagef(ctx context.Context, format string, args ...interface{})
}

type Logger interface {
	IsLevelEnabled(level Level) bool
	WithField(key string, value interface{}) Logger
	WithFields(fields Fields) Logger
	WithError(err error) Logger
	WithTime(t time.Time) Logger

	Panicf(ctx context.Context, format string, args ...interface{})
	Fatalf(ctx context.Context, format string, args ...interface{})
	Errorf(ctx context.Context, format string, args ...interface{})
	Warningf(ctx context.Context, format string, args ...interface{})
	Noticef(ctx context.Context, format string, args ...interface{})
	Infof(ctx context.Context, format string, args ...interface{})
	Debugf(ctx context.Context, format string, args ...interface{})
	Tracef(ctx context.Context, format string, args ...interface{})

	Panic(ctx context.Context, args ...interface{})
	Fatal(ctx context.Context, args ...interface{})
	Error(ctx context.Context, args ...interface{})
	Warning(ctx context.Context, args ...interface{})
	Notice(ctx context.Context, args ...interface{})
	Info(ctx context.Context, args ...interface{})
	Debug(ctx context.Context, args ...interface{})
	Trace(ctx context.Context, args ...interface{})

	PanicFn(ctx context.Context, fn LogFunction)
	FatalFn(ctx context.Context, fn LogFunction)
	ErrorFn(ctx context.Context, fn LogFunction)
	WarningFn(ctx context.Context, fn LogFunction)
	NoticeFn(ctx context.Context, fn LogFunction)
	InfoFn(ctx context.Context, fn LogFunction)
	DebugFn(ctx context.Context, fn LogFunction)
	TraceFn(ctx context.Context, fn LogFunction)
}
