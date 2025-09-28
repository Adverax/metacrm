package templateFormatter

import (
	"github.com/adverax/metacrm/pkg/log"
	"github.com/stretchr/testify/assert"
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
				Data: log.Fields{
					"key": "value",
				},
			},
			expected: "0001/01/01 00:00:00 INFO: Hello, World! DETAILS {\"key\":\"value\"}\n",
		},
	}

	formatter, err := NewBuilder().
		WithTimestampFormat("2006/01/02 15:04:05").
		Build()
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := formatter.Format(test.entry)
			require.NoError(t, err)
			assert.Equal(t, test.expected, string(data))
		})
	}
}
