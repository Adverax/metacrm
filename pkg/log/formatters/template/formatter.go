package templateFormatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"text/template"

	"github.com/adverax/metacrm/pkg/log"
)

type Purifier interface {
	Purify(original, derivative string) string
}

// Formatter formats logs into text
type Formatter struct {
	purifier               Purifier
	disableTimestamp       bool
	timestampFormat        string
	disableSorting         bool
	sortingFunc            func([]string)
	disableLevelTruncation bool
	padLevelText           bool
	fieldMap               log.FieldMap
	template               *template.Template
	systemFields           map[string]struct{}
}

// Format renders a single log entry
func (that *Formatter) Format(entry *log.Entry) ([]byte, error) {
	data := make(log.Fields)
	for k, v := range entry.Data {
		data[k] = v
	}
	that.fieldMap.EncodePrefixFieldClashes(data)
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	fixedKeys := make([]string, 0, 4+len(data))
	if !that.disableTimestamp {
		fixedKeys = append(fixedKeys, that.fieldMap.Resolve(log.FieldKeyTime))
	}
	fixedKeys = append(fixedKeys, that.fieldMap.Resolve(log.FieldKeyLevel))
	if entry.Message != "" {
		fixedKeys = append(fixedKeys, that.fieldMap.Resolve(log.FieldKeyMsg))
	}
	if entry.LogErr != "" {
		fixedKeys = append(fixedKeys, that.fieldMap.Resolve(log.FieldKeyLoggerError))
	}

	if !that.disableSorting {
		if that.sortingFunc == nil {
			sort.Strings(keys)
			fixedKeys = append(fixedKeys, keys...)
		} else {
			fixedKeys = append(fixedKeys, keys...)
			that.sortingFunc(fixedKeys)

		}
	} else {
		fixedKeys = append(fixedKeys, keys...)
	}

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	timestampFormat := that.timestampFormat
	if timestampFormat == "" {
		timestampFormat = log.DefaultTimestampFormat
	}

	systemFields := that.systemFields
	if systemFields == nil {
		systemFields = defaultSystemFields
	}

	params := make(map[string]interface{})
	rest := make(map[string]interface{})
	var entity, action, method string
	var subject, body string

	for _, key := range fixedKeys {
		var value interface{}
		switch {
		case key == that.fieldMap.Resolve(log.FieldKeyTime):
			value = entry.Time.Format(timestampFormat)
		case key == that.fieldMap.Resolve(log.FieldKeyLevel):
			value = entry.Level.String()
		case key == that.fieldMap.Resolve(log.FieldKeyMsg):
			value = that.purify(entry.Message)
		case key == that.fieldMap.Resolve(log.FieldKeyLoggerError):
			value = entry.LogErr
		case key == that.fieldMap.Resolve(log.FieldKeyTraceID):
			value, _ = data[key]
			continue
		case key == that.fieldMap.Resolve(log.FieldKeyEntity):
			value, _ = data[key]
			entity = that.value2string(value)
			continue
		case key == that.fieldMap.Resolve(log.FieldKeyAction):
			value, _ = data[key]
			action = that.value2string(value)
			continue
		case key == that.fieldMap.Resolve(log.FieldKeyMethod):
			value, _ = data[key]
			method = that.value2string(value)
			continue
		case key == that.fieldMap.Resolve(log.FieldKeySubject):
			value, _ = data[key]
			subject = that.value2string(value)
			continue
		case key == that.fieldMap.Resolve(log.FieldKeyData):
			value, _ = data[key]
			body = that.value2string(value)
			continue
		default:
			value = data[key]
		}

		val := that.value2string(value)
		if _, ok := systemFields[key]; ok {
			params[key] = val
		} else {
			rest[key] = val
		}
	}

	params["entity"] = that.formatEntity(entity, action)
	params["event"] = that.formatEvent(method, subject, body)

	if _, ok := params[log.FieldKeyTraceID]; !ok {
		params[log.FieldKeyTraceID] = ""
	}

	if len(rest) > 0 {
		var details []byte
		details, _ = json.Marshal(rest)
		params["details"] = string(details)
	}

	tpl := that.template
	if tpl == nil {
		tpl = defaultTpl
	}
	_ = tpl.Execute(b, params)

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (that *Formatter) value2string(value interface{}) string {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}

	return stringVal
}

func (that *Formatter) formatEntity(entity, action string) string {
	if entity == "" {
		return ""
	}

	var result bytes.Buffer
	result.WriteByte(' ')
	result.WriteString(entity)
	if action == "" {
		result.WriteByte(':')
	} else {
		result.WriteByte(' ')
		result.WriteString(action)
	}

	return result.String()
}

func (that *Formatter) formatEvent(method, subject, body string) string {
	if body == "" && subject == "" {
		return ""
	}

	var wantSpace bool
	var result bytes.Buffer
	result.WriteByte(' ')

	if method != "" {
		result.WriteString(method)
		wantSpace = true
	}

	if subject != "" {
		if wantSpace {
			result.WriteByte(' ')
		}
		result.WriteString(subject)
		wantSpace = true
	}

	if body != "" {
		if wantSpace {
			result.WriteByte(' ')
		}
		result.WriteString(that.purify(body))
	}

	return result.String()
}

func (that *Formatter) purify(s string) string {
	if that.purifier == nil {
		return s
	}

	return that.purifier.Purify(s, s)
}
