package viewport

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestPad(t *testing.T) {
	width, height := 5, 4
	lines := []string{"a", "b", "c"}
	expected := `a    
b    
c    
     `
	if diff := cmp.Diff(expected, pad(width, height, lines)); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestPad_OverflowWidth(t *testing.T) {
	width, height := 5, 4
	lines := []string{"123456", "b", "c"}
	expected := `123456
b    
c    
     `
	if diff := cmp.Diff(expected, pad(width, height, lines)); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestPad_Ansi(t *testing.T) {
	width, height := 5, 4
	lines := []string{renderer().NewStyle().Foreground(red).Render("a"), "b", "c"}
	expected := "\x1b[38;2;255;0;0ma\x1b[0m    \nb    \nc    \n     "
	if diff := cmp.Diff(expected, pad(width, height, lines)); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestTruncateLine(t *testing.T) {
	tests := []struct {
		name                      string
		s                         string
		xOffset                   int
		width                     int
		lineContinuationIndicator string
		expected                  string
	}{
		{
			name:                      "zero width zero offset",
			s:                         "1234567890123456789012345",
			xOffset:                   0,
			width:                     0,
			lineContinuationIndicator: "...",
			expected:                  "",
		},
		{
			name:                      "zero width positive offset",
			s:                         "1234567890123456789012345",
			xOffset:                   5,
			width:                     0,
			lineContinuationIndicator: "...",
			expected:                  "",
		},
		{
			name:                      "zero width negative offset",
			s:                         "1234567890123456789012345",
			xOffset:                   -5,
			width:                     0,
			lineContinuationIndicator: "...",
			expected:                  "",
		},
		{
			name:                      "start near end of string",
			s:                         "1234567890",
			xOffset:                   9,
			width:                     5,
			lineContinuationIndicator: "...",
			expected:                  ".",
		},
		{
			name:                      "small string",
			s:                         "hi",
			xOffset:                   0,
			width:                     3,
			lineContinuationIndicator: "...",
			expected:                  "hi",
		},
		{
			name:                      "lineContinuationIndicator longer than width",
			s:                         "1234567890123456789012345",
			xOffset:                   0,
			width:                     1,
			lineContinuationIndicator: "...",
			expected:                  ".",
		},
		{
			name:                      "twice the lineContinuationIndicator longer than width",
			s:                         "1234567",
			xOffset:                   1,
			width:                     5,
			lineContinuationIndicator: "...",
			expected:                  ".....",
		},
		{
			name:                      "zero offset, sufficient width",
			s:                         "1234567890123456789012345",
			xOffset:                   0,
			width:                     30,
			lineContinuationIndicator: "...",
			expected:                  "1234567890123456789012345",
		},
		{
			name:                      "zero offset, sufficient width, space at end",
			s:                         "1234567890123456789012345     ",
			xOffset:                   0,
			width:                     30,
			lineContinuationIndicator: "...",
			expected:                  "1234567890123456789012345     ",
		},
		{
			name:                      "zero offset, insufficient width",
			s:                         "1234567890123456789012345",
			xOffset:                   0,
			width:                     15,
			lineContinuationIndicator: "...",
			expected:                  "123456789012...",
		},
		{
			name:                      "positive offset, insufficient width",
			s:                         "1234567890123456789012345",
			xOffset:                   5,
			width:                     15,
			lineContinuationIndicator: "...",
			expected:                  "...901234567...",
		},
		{
			name:                      "positive offset, exactly at end",
			s:                         "1234567890123456789012345",
			xOffset:                   15,
			width:                     10,
			lineContinuationIndicator: "...",
			expected:                  "...9012345",
		},
		{
			name:                      "positive offset, over the end",
			s:                         "1234567890123456789012345",
			xOffset:                   20,
			width:                     10,
			lineContinuationIndicator: "...",
			expected:                  "...45",
		},
		{
			name:                      "positive offset, ansi",
			s:                         "\x1b[38;2;255;0;0ma really really long line\x1b[0m",
			xOffset:                   15,
			width:                     15,
			lineContinuationIndicator: "",
			expected:                  "\x1b[38;2;255;0;0m long line\x1b[0m",
		},
		{
			name:                      "zero offset, insufficient width, ansi",
			s:                         "\x1b[38;2;255;0;0m1234567890123456789012345\x1b[0m",
			xOffset:                   0,
			width:                     15,
			lineContinuationIndicator: "...",
			expected:                  "\x1b[38;2;255;0;0m123456789012...\x1b[0m",
		},
		{
			name:                      "positive offset, insufficient width, ansi",
			s:                         "\x1b[38;2;255;0;0m1234567890123456789012345\x1b[0m",
			xOffset:                   5,
			width:                     15,
			lineContinuationIndicator: "...",
			expected:                  "\x1b[38;2;255;0;0m...901234567...\x1b[0m",
		},
		{
			name:                      "no offset, insufficient width, inline ansi",
			s:                         "|\x1b[38;2;169;15;15mfl..-1\x1b[0m| {\"timestamp\": \"2024-09-29T22:30:28.730520\"}",
			xOffset:                   0,
			width:                     15,
			lineContinuationIndicator: "...",
			expected:                  "|\x1b[38;2;169;15;15mfl..-1\x1b[0m| {\"t...",
		},
		{
			name:                      "offset overflow, ansi",
			s:                         "\x1b[38;2;0;0;255mthird line that is fairly long\x1b[0m",
			xOffset:                   41,
			width:                     10,
			lineContinuationIndicator: "...",
			expected:                  "\x1b[38;2;0;0;255m\x1b[0m",
		},
		{
			name:                      "offset overflow, ansi 2",
			s:                         "\x1b[38;2;0;0;255mfourth\x1b[0m",
			xOffset:                   41,
			width:                     10,
			lineContinuationIndicator: "...",
			expected:                  "\x1b[38;2;0;0;255m\x1b[0m",
		},
		{
			name: "offset start space ansi",
			// 							   0123456789012345   67890
			//       									  0       123456789012345678901234
			s:                         "\x1b[38;2;255;0;0ma\x1b[0m really really long line",
			xOffset:                   15,
			width:                     15,
			lineContinuationIndicator: "",
			expected:                  "\x1b[38;2;255;0;0m\x1b[0m long line",
		},
		{
			name:                      "ansi short",
			s:                         "\x1b[38;2;0;0;255mhi\x1b[0m",
			xOffset:                   0,
			width:                     3,
			lineContinuationIndicator: "...",
			expected:                  "\x1b[38;2;0;0;255mhi\x1b[0m",
		},
		{
			name:                      "multi-byte chars",
			s:                         "├─flask",
			xOffset:                   0,
			width:                     6,
			lineContinuationIndicator: "...",
			expected:                  "├─f...",
		},
		{
			name:                      "multi-byte chars with ansi",
			s:                         "\x1b[38;2;0;0;255m├─flask\x1b[0m",
			xOffset:                   0,
			width:                     6,
			lineContinuationIndicator: "...",
			expected:                  "\x1b[38;2;0;0;255m├─f...\x1b[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := truncateLine(tt.s, tt.xOffset, tt.width, tt.lineContinuationIndicator)
			if diff := cmp.Diff(tt.expected, actual); diff != "" {
				t.Errorf("Mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestViewport_StringWidth(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expectedWidth int
	}{
		{
			"no ansi",
			"No ANSI here, just plain text",
			29,
		},
		{
			"ansi foo",
			"\x1b[38;2;214;125;17mfoo\x1b[0m",
			len("foo"),
		},
		{
			"hello world",
			"\x1b[31mHello, World!\x1b[0m",
			len("Hello, World!"),
		},
		{
			"bold text",
			"\x1b[1mBold Text\x1b[0m",
			9,
		},
		{
			"only bold and reset codes, no text",
			"\x1b[1m\x1b[0m",
			0,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result := lipgloss.Width(tt.input)
			if result != tt.expectedWidth {
				t.Errorf("For input '%s', expected width %d, but got %d", tt.input, tt.expectedWidth, result)
			}
		})
	}
}
