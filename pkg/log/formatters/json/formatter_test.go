package jsonFormatter

import (
	"github.com/adverax/metacrm/pkg/log"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestFormatter(t *testing.T) {
	type Test struct {
		name     string
		entry    *log.Entry
		expected string
	}

	tests := []Test{
		{
			name: "info",
			entry: &log.Entry{
				Time:    time.Time{},
				Level:   log.InfoLevel,
				Message: "Hello, World!",
				Data:    log.Fields{},
			},
			expected: "{\"level\":\"info\",\"msg\":\"Hello, World!\",\"time\":\"0001-01-01 00:00:00\"}\n",
		},
		{
			name: "error",
			entry: &log.Entry{
				Time:    time.Time{},
				Level:   log.ErrorLevel,
				Message: "Hello, World2!",
				Data:    log.Fields{"key": "value"},
			},
			expected: "{\"data\":{\"key\":\"value\"},\"level\":\"error\",\"msg\":\"Hello, World2!\",\"time\":\"0001-01-01 00:00:00\"}\n",
		},
	}

	formatter, err := NewBuilder().Build()
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := formatter.Format(test.entry)
			require.NoError(t, err)
			require.Equal(t, test.expected, string(data))
		})
	}
}
