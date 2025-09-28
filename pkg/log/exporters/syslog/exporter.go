package syslogExporter

import (
	"context"
	"fmt"
	"log/syslog"
	"os"

	"github.com/adverax/metacrm/pkg/log"
)

type Exporter struct {
	formatter log.Formatter
	out       *syslog.Writer
}

func New(formatter log.Formatter, out *syslog.Writer) *Exporter {
	return &Exporter{
		formatter: formatter,
		out:       out,
	}
}

func (that *Exporter) Export(ctx context.Context, entry *log.Entry) {
	buffer := entry.Logger.GetBuffer()
	defer func() {
		entry.Buffer = nil
		buffer.Reset()
		entry.Logger.FreeBuffer(buffer)
	}()
	buffer.Reset()
	entry.Buffer = buffer

	that.export(entry)

	entry.Buffer = nil
}

func (that *Exporter) export(entry *log.Entry) {
	serialized, err := that.formatter.Format(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to obtain reader, %v\n", err)
		return
	}

	err = that.put(entry.Level, string(serialized))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write to log, %v\n", err)
	}
}

func (that *Exporter) put(level log.Level, msg string) error {
	switch level {
	case log.TraceLevel:
		return that.out.Debug(msg)
	case log.DebugLevel:
		return that.out.Debug(msg)
	case log.InfoLevel:
		return that.out.Info(msg)
	case log.WarnLevel:
		return that.out.Warning(msg)
	case log.ErrorLevel:
		return that.out.Err(msg)
	case log.FatalLevel:
		return that.out.Crit(msg)
	case log.PanicLevel:
		return that.out.Emerg(msg)
	default:
		return nil
	}
}
