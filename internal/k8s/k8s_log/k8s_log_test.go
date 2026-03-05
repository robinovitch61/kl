package k8s_log_test

import (
	"testing"

	"github.com/robinovitch61/kl/internal/k8s/k8s_log"
)

func TestPrettyPrintJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "not JSON",
			input:    "just plain text",
			expected: []string{"just plain text"},
		},
		{
			name:     "invalid JSON",
			input:    `{"key": broken}`,
			expected: []string{`{"key": broken}`},
		},
		{
			name:     "JSON array not pretty-printed",
			input:    `[1, 2, 3]`,
			expected: []string{`[1, 2, 3]`},
		},
		{
			name:  "simple object",
			input: `{"key":"value"}`,
			expected: []string{
				`{`,
				`    "key": "value"`,
				`}`,
			},
		},
		{
			name:  "nested object",
			input: `{"a":{"b":"c"}}`,
			expected: []string{
				`{`,
				`    "a": {`,
				`        "b": "c"`,
				`    }`,
				`}`,
			},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{""},
		},
		{
			name:     "empty object",
			input:    `{}`,
			expected: []string{`{}`},
		},
		{
			name:  "escaped newlines split into lines",
			input: `{"msg":"line1\nline2\nline3"}`,
			expected: []string{
				`{`,
				`    "msg": "line1`,
				`line2`,
				`line3"`,
				`}`,
			},
		},
		{
			name:  "escaped tabs replaced with spaces",
			input: `{"msg":"before\tafter"}`,
			expected: []string{
				`{`,
				`    "msg": "before    after"`,
				`}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := k8s_log.PrettyPrintJSON(tt.input)
			if len(got) != len(tt.expected) {
				t.Fatalf("expected %d lines, got %d\nexpected: %q\ngot:      %q", len(tt.expected), len(got), tt.expected, got)
			}
			for i := range tt.expected {
				if got[i] != tt.expected[i] {
					t.Errorf("line %d: expected %q, got %q", i, tt.expected[i], got[i])
				}
			}
		})
	}
}
