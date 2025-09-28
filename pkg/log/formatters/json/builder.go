package jsonFormatter

import (
	"github.com/adverax/metacrm/pkg/log"
)

type Builder struct {
	formatter *Formatter
}

func NewBuilder() *Builder {
	return &Builder{
		formatter: &Formatter{
			timestampFormat:   log.DefaultTimestampFormat,
			disableTimestamp:  false,
			disableHTMLEscape: false,
			dataKey:           log.FieldKeyData,
			fieldMap:          nil,
			prettyPrint:       false,
		},
	}
}

func (that *Builder) WithDataKey(key string) *Builder {
	that.formatter.dataKey = key
	return that
}

func (that *Builder) WithFieldMap(fieldMap log.FieldMap) *Builder {
	that.formatter.fieldMap = fieldMap
	return that
}

func (that *Builder) WithPrettyPrint(prettyPrint bool) *Builder {
	that.formatter.prettyPrint = prettyPrint
	return that
}

func (that *Builder) WithTimestampFormat(timestampFormat string) *Builder {
	that.formatter.timestampFormat = timestampFormat
	return that
}

func (that *Builder) WithDisableTimestamp(disableTimestamp bool) *Builder {
	that.formatter.disableTimestamp = disableTimestamp
	return that
}

func (that *Builder) WithDisableHTMLEscape(disableHTMLEss bool) *Builder {
	that.formatter.disableHTMLEscape = disableHTMLEss
	return that
}

func (that *Builder) Build() (*Formatter, error) {
	return that.formatter, nil
}
