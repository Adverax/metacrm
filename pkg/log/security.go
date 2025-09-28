package log

import (
	"context"
	"encoding/json"
)

type Masker func(value any) string

type Security struct {
	secrets map[string]Masker // map of secret keys to their presence
	next    Exporter
}

func NewSecurityExporter(secrets map[string]Masker, next Exporter) Exporter {
	if next == nil {
		next = &dummyExporter{}
	}

	if secrets == nil {
		secrets = map[string]Masker{}
	}

	for k := range secrets {
		if secrets[k] == nil {
			secrets[k] = MaskAny
		}
	}

	return &Security{
		secrets: secrets,
		next:    next,
	}
}

func (that *Security) Export(ctx context.Context, entry *Entry) {
	MaskAll(that.secrets, entry.Data)
	that.next.Export(ctx, entry)
}

func MaskJson(secrets map[string]Masker, data []byte) []byte {
	var fields Fields
	err := json.Unmarshal(data, &fields)
	if err != nil {
		return data
	}

	MaskAll(secrets, fields)

	out, err := json.Marshal(fields)
	if err != nil {
		return data
	}

	return out
}

func MaskAll(secrets map[string]Masker, fields Fields) {
	for key, val := range fields {
		if v, ok := val.(map[string]any); ok {
			MaskAll(secrets, v)
			continue
		}
		if m, ok := secrets[key]; ok {
			fields[key] = m(val)
		}
	}
}

func MaskAny(value any) string {
	return "****"
}

func MaskPhone(value any) string {
	s, ok := value.(string)
	if !ok {
		return "****"
	}

	n := len(s)
	if n <= 4 {
		return "****"
	}

	if n <= 8 {
		return "****" + s[n-4:]
	}

	return s[:3] + "****" + s[n-4:]
}

func MaskEmail(value any) string {
	s, ok := value.(string)
	if !ok {
		return "****"
	}

	at := -1
	for i := 0; i < len(s); i++ {
		if s[i] == '@' {
			at = i
			break
		}
	}
	if at <= 1 {
		return "****"
	}

	if at <= 4 {
		return s[:1] + "****" + s[at:]
	}

	return s[:2] + "****" + s[at:]
}

func MaskIDCard(value any) string {
	s, ok := value.(string)
	if !ok {
		return "****"
	}

	n := len(s)
	if n <= 8 {
		return "****"
	}

	return s[:4] + "****" + s[n-4:]
}
