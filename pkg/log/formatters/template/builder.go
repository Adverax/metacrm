package templateFormatter

import (
	"errors"
	"strings"
	"text/template"

	"github.com/adverax/metacrm/pkg/log"
)

type Builder struct {
	formatter *Formatter
}

func NewBuilder() *Builder {
	return &Builder{
		formatter: &Formatter{
			timestampFormat:  log.DefaultTimestampFormat,
			disableTimestamp: false,
			template:         defaultTpl,
			systemFields:     defaultSystemFields,
			fieldMap:         log.FieldMap{},
		},
	}
}

func (that *Builder) WithTemplate(tpl *template.Template) *Builder {
	that.formatter.template = tpl
	return that
}

func (that *Builder) WithDisableTimestamp(disableTimestamp bool) *Builder {
	that.formatter.disableTimestamp = disableTimestamp
	return that
}

func (that *Builder) WithTimestampFormat(timestampFormat string) *Builder {
	that.formatter.timestampFormat = timestampFormat
	return that
}

func (that *Builder) WithFieldMap(fieldMap log.FieldMap) *Builder {
	that.formatter.fieldMap = fieldMap
	return that
}

func (that *Builder) WithDisableSorting(disableSorting bool) *Builder {
	that.formatter.disableSorting = disableSorting
	return that
}

func (that *Builder) WithSortingFunc(sortingFunc func([]string)) *Builder {
	that.formatter.sortingFunc = sortingFunc
	return that
}

func (that *Builder) WithDisableLevelTruncation(disableLevelTruncation bool) *Builder {
	that.formatter.disableLevelTruncation = disableLevelTruncation
	return that
}

func (that *Builder) WithPadLevelText(padLevelText bool) *Builder {
	that.formatter.padLevelText = padLevelText
	return that
}

func (that *Builder) WithPurifier(purifier Purifier) *Builder {
	that.formatter.purifier = purifier
	return that
}

func (that *Builder) Build() (*Formatter, error) {
	if err := that.checkRequiredFields(); err != nil {
		return nil, err
	}

	return that.formatter, nil
}

func (that *Builder) checkRequiredFields() error {
	if that.formatter.template == nil {
		return ErrTemplateRequired
	}

	return nil
}

var (
	ErrTemplateRequired = errors.New("template is required")
)

var funcMap = template.FuncMap{
	"ToUpper": strings.ToUpper,
}

var defaultTemplate = `{{.time}} {{.level | ToUpper}}{{if .trace_id}} #{{.trace_id}}{{end}}:{{.entity}} {{.msg}}{{.event}}{{if .details}} DETAILS {{.details}}{{end}}`

var defaultTpl = template.Must(template.New("log").Funcs(funcMap).Parse(defaultTemplate))

var defaultSystemFields = map[string]struct{}{
	log.FieldKeyTime:    {},
	log.FieldKeyLevel:   {},
	log.FieldKeyMsg:     {},
	log.FieldKeyTraceID: {},
	log.FieldKeyEntity:  {},
	log.FieldKeyAction:  {},
	log.FieldKeyMethod:  {},
	log.FieldKeySubject: {},
	log.FieldKeyData:    {},
}
