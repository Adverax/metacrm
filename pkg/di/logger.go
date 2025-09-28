package di

import (
	"context"
	"fmt"
	"log"
	"time"
)

type Logger interface {
	WithError(err error) Logger
	Errorf(ctx context.Context, format string, args ...interface{})
	Debugf(ctx context.Context, format string, args ...interface{})
}

type loggerChunk struct {
	logger *defaultLogger
	err    error
}

func (that *loggerChunk) WithError(err error) Logger {
	return &loggerChunk{logger: that.logger, err: err}
}

func (that *loggerChunk) Errorf(ctx context.Context, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if that.err != nil {
		message = fmt.Sprintf("%s: %s", message, that.err.Error())
	}
	logMessage("ERROR", message)
}

func (that *loggerChunk) Debugf(ctx context.Context, format string, args ...interface{}) {
	logMessage("DEBUG", fmt.Sprintf(format, args...))
}

type defaultLogger struct{}

func (that *defaultLogger) WithError(err error) Logger {
	return &loggerChunk{logger: that, err: err}
}

func (that *defaultLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	logMessage("ERROR", fmt.Sprintf(format, args...))

}

func (that *defaultLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	logMessage("DEBUG", fmt.Sprintf(format, args...))
}

func logMessage(level, message string) {
	log.Printf("%s [%s] %s", time.Now().String(), level, message)
}

var DefaultLogger Logger = &dummyLogger{}

type dummyLogger struct {
}

func (that *dummyLogger) WithError(error) Logger {
	return that
}

func (that *dummyLogger) Errorf(context.Context, string, ...interface{}) {
	// empty
}

func (that *dummyLogger) Debugf(context.Context, string, ...interface{}) {
	// empty
}
