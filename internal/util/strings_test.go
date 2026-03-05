package util_test

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/robinovitch61/kl/internal/util"
)

func TestGetUniqueShortNames(t *testing.T) {
	tests := []struct {
		nameSet   map[string]bool
		fromRight bool
		minChars  int
		expected  map[string]string
	}{
		{
			nameSet: map[string]bool{
				"apple":  true,
				"banana": true,
				"cherry": true,
			},
			fromRight: false,
			minChars:  2,
			expected: map[string]string{
				"apple":  "ap..",
				"banana": "ba..",
				"cherry": "ch..",
			},
		},
		{
			nameSet: map[string]bool{
				"apple":   true,
				"apricot": true,
				"banana":  true,
			},
			fromRight: false,
			minChars:  1,
			expected: map[string]string{
				"apple":   "app..",
				"apricot": "apr..",
				"banana":  "ban..",
			},
		},
		{
			nameSet: map[string]bool{
				"apple":  true,
				"papple": true,
				"grape":  true,
			},
			fromRight: true,
			minChars:  3,
			expected: map[string]string{
				"apple":  "apple",
				"papple": "papple",
				"grape":  "grape",
			},
		},
	}

	for _, test := range tests {
		result := util.GetUniqueShortNames(test.nameSet, test.fromRight, test.minChars)
		for k, v := range test.expected {
			if result[k] != v {
				t.Errorf("For name '%s', expected short name '%s' but got '%s'", k, v, result[k])
			}
		}
	}
}

// same test for GetUniqueShortNamesFromEdges
func TestGetUniqueShortNamesFromSides(t *testing.T) {
	tests := []struct {
		nameSet          map[string]bool
		numCharsEachSide int
		expected         map[string]string
	}{
		{
			nameSet: map[string]bool{
				"apple":  true,
				"banana": true,
				"cherry": true,
			},
			numCharsEachSide: 1,
			expected: map[string]string{
				"apple":  "a..e",
				"banana": "b..a",
				"cherry": "c..y",
			},
		},
		{
			nameSet: map[string]bool{
				"apple":  true,
				"banana": true,
				"cherry": true,
			},
			numCharsEachSide: 2,
			expected: map[string]string{
				"apple":  "ap..le",
				"banana": "ba..na",
				"cherry": "ch..ry",
			},
		},
		{
			nameSet: map[string]bool{
				"apple":  true,
				"banana": true,
				"cherry": true,
			},
			numCharsEachSide: 3,
			expected: map[string]string{
				"apple":  "apple",
				"banana": "banana",
				"cherry": "cherry",
			},
		},
		{
			nameSet: map[string]bool{
				"appsamele": true,
				"appdiffle": true,
			},
			numCharsEachSide: 1,
			expected: map[string]string{
				"appsamele": "app..ele",
				"appdiffle": "app..fle",
			},
		},
	}

	for _, test := range tests {
		result := util.GetUniqueShortNamesFromEdges(test.nameSet, test.numCharsEachSide)
		for k, v := range test.expected {
			if result[k] != v {
				t.Errorf("For name '%s', expected short name '%s' but got '%s'", k, v, result[k])
			}
		}
	}
}

func TestJoinWithEqualSpacing(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		items    []string
		expected string
	}{
		{
			name:     "two items",
			width:    20,
			items:    []string{"left", "right"},
			expected: "left           right",
		},
		{
			name:     "three items",
			width:    20,
			items:    []string{"a", "b", "c"},
			expected: "a         b        c",
		},
		{
			name:     "single item",
			width:    20,
			items:    []string{"hello"},
			expected: "hello",
		},
		{
			name:     "no items",
			width:    20,
			items:    []string{},
			expected: "",
		},
		{
			name:     "exact fit",
			width:    5,
			items:    []string{"ab", "cd"},
			expected: "ab cd",
		},
		{
			name:     "exceeds width",
			width:    5,
			items:    []string{"hello", "world"},
			expected: "hello",
		},
		{
			name:     "zero width",
			width:    0,
			items:    []string{"a", "b"},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.JoinWithEqualSpacing(tt.width, tt.items...)
			if result != tt.expected {
				t.Errorf("util.JoinWithEqualSpacing(%d, %v) = %q, want %q", tt.width, tt.items, result, tt.expected)
			}
		})
	}
}

func TestStyleStyledString(t *testing.T) {
	tests := []struct {
		name     string
		style    lipgloss.Style
		input    string
		expected string
	}{
		{
			name:     "no ansi",
			style:    lipgloss.NewStyle(),
			input:    "No ANSI here, just plain text",
			expected: "No ANSI here, just plain text",
		},
		{
			name:     "has ansi",
			style:    lipgloss.NewStyle().Foreground(lipgloss.Color("#0000ff")),
			input:    "some \x1b[31mred\x1b[m text",
			expected: "\x1b[38;2;0;0;255msome \x1b[m\x1b[31mred\x1b[m\x1b[38;2;0;0;255m text\x1b[m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.StyleStyledString(tt.input, tt.style)
			if result != tt.expected {
				t.Errorf("For input '%q', expected '%q', but got '%q'", tt.input, tt.expected, result)
			}
		})
	}
}

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
			got := util.SanitizeTerminalSequences(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeTerminalSequences(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
