package k8s_log

import "testing"

func TestSanitizeTerminalSequences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain text unchanged",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "SGR color preserved",
			input:    "\x1b[31mred text\x1b[0m",
			expected: "\x1b[31mred text\x1b[0m",
		},
		{
			name:     "SGR bold preserved",
			input:    "\x1b[1mbold\x1b[0m",
			expected: "\x1b[1mbold\x1b[0m",
		},
		{
			name:     "SGR 256-color preserved",
			input:    "\x1b[38;5;196mcolor\x1b[0m",
			expected: "\x1b[38;5;196mcolor\x1b[0m",
		},
		{
			name:     "SGR RGB color preserved",
			input:    "\x1b[38;2;255;0;0mrgb\x1b[0m",
			expected: "\x1b[38;2;255;0;0mrgb\x1b[0m",
		},
		{
			name:     "cursor up removed",
			input:    "before\x1b[Aafter",
			expected: "beforeafter",
		},
		{
			name:     "cursor movement removed",
			input:    "a\x1b[2Bb\x1b[3Cc\x1b[4D",
			expected: "abc",
		},
		{
			name:     "cursor position removed",
			input:    "\x1b[10;20Htext",
			expected: "text",
		},
		{
			name:     "clear screen removed",
			input:    "\x1b[2Jtext",
			expected: "text",
		},
		{
			name:     "clear line removed",
			input:    "text\x1b[K",
			expected: "text",
		},
		{
			name:     "OSC title sequence removed",
			input:    "\x1b]0;window title\x07text",
			expected: "text",
		},
		{
			name:     "OSC with ST removed",
			input:    "\x1b]0;title\x1b\\text",
			expected: "text",
		},
		{
			name:     "mixed styling kept and control removed",
			input:    "\x1b[31mred\x1b[0m\x1b[2Jcleared\x1b[1mbold\x1b[0m",
			expected: "\x1b[31mred\x1b[0mcleared\x1b[1mbold\x1b[0m",
		},
		{
			name:     "BEL removed",
			input:    "alert\x07text",
			expected: "alerttext",
		},
		{
			name:     "backspace removed",
			input:    "ab\x08c",
			expected: "abc",
		},
		{
			name:     "DEL removed",
			input:    "ab\x7fc",
			expected: "abc",
		},
		{
			name:     "ESC c reset removed",
			input:    "\x1bctext",
			expected: "text",
		},
		{
			name:     "DCS sequence removed",
			input:    "\x1bPsome data\x1b\\text",
			expected: "text",
		},
		{
			name:     "private mode sequence removed",
			input:    "\x1b[?25htext\x1b[?25l",
			expected: "text",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "lone ESC at end",
			input:    "text\x1b",
			expected: "text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeTerminalSequences(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizeTerminalSequences(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
