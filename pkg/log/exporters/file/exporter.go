package fileExporter

import (
	"context"
	"fmt"
	"os"

	"github.com/adverax/metacrm/pkg/log"
)

type Exporter struct {
	formatter log.Formatter
	file      *os.File
}

func New(file *os.File, formatter log.Formatter) *Exporter {
	return &Exporter{
		file:      file,
		formatter: formatter,
	}
}

func (that *Exporter) Export(_ context.Context, entry *log.Entry) {
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
		serialized = []byte(fmt.Sprintf("Failed to format log entry, %s\n", err.Error()))
	}

	_, _ = that.file.Write(serialized)
}
