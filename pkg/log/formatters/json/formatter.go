package jsonFormatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/adverax/metacrm/pkg/log"
)

type Formatter struct {
	timestampFormat   string
	disableTimestamp  bool
	disableHTMLEscape bool
	dataKey           string
	fieldMap          log.FieldMap
	prettyPrint       bool
}

// Format renders a single log entry
func (that *Formatter) Format(entry *log.Entry) ([]byte, error) {
	data := entry.Data.Expand()

	newData := make(log.Fields, 4)
	if len(data) > 0 {
		newData[log.FieldKeyData] = data
	}
	data = newData

	that.fieldMap.EncodePrefixFieldClashes(data)

	timestampFormat := that.timestampFormat

	if entry.LogErr != "" {
		data[that.fieldMap.Resolve(log.FieldKeyLoggerError)] = entry.LogErr
	}
	if !that.disableTimestamp {
		data[that.fieldMap.Resolve(log.FieldKeyTime)] = entry.Time.Format(timestampFormat)
	}
	data[that.fieldMap.Resolve(log.FieldKeyMsg)] = entry.Message
	data[that.fieldMap.Resolve(log.FieldKeyLevel)] = entry.Level.String()

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	encoder := json.NewEncoder(b)
	encoder.SetEscapeHTML(!that.disableHTMLEscape)
	if that.prettyPrint {
		encoder.SetIndent("", "  ")
	}
	if err := encoder.Encode(data); err != nil {
		return nil, fmt.Errorf("failed to marshal fields to JSON: %w", err)
	}

	return b.Bytes(), nil
}
