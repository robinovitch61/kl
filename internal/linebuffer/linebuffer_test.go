package linebuffer

import (
	"github.com/robinovitch61/kl/internal/util"
	"testing"
)

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
			s:                         "\x1b[38;2;255;0;0ma really really long line\x1b[m",
			xOffset:                   15,
			width:                     15,
			lineContinuationIndicator: "",
			expected:                  "\x1b[38;2;255;0;0m long line\x1b[m",
		},
		{
			name:                      "zero offset, insufficient width, ansi",
			s:                         "\x1b[38;2;255;0;0m1234567890123456789012345\x1b[m",
			xOffset:                   0,
			width:                     15,
			lineContinuationIndicator: "...",
			expected:                  "\x1b[38;2;255;0;0m123456789012...\x1b[m",
		},
		{
			name:                      "positive offset, insufficient width, ansi",
			s:                         "\x1b[38;2;255;0;0m1234567890123456789012345\x1b[m",
			xOffset:                   5,
			width:                     15,
			lineContinuationIndicator: "...",
			expected:                  "\x1b[38;2;255;0;0m...901234567...\x1b[m",
		},
		{
			name:                      "no offset, insufficient width, inline ansi",
			s:                         "|\x1b[38;2;169;15;15mfl..-1\x1b[m| {\"timestamp\": \"2024-09-29T22:30:28.730520\"}",
			xOffset:                   0,
			width:                     15,
			lineContinuationIndicator: "...",
			expected:                  "|\x1b[38;2;169;15;15mfl..-1\x1b[m| {\"t...",
		},
		{
			name:                      "offset overflow, ansi",
			s:                         "\x1b[38;2;0;0;255mthird line that is fairly long\x1b[m",
			xOffset:                   41,
			width:                     10,
			lineContinuationIndicator: "...",
			expected:                  "",
		},
		{
			name:                      "offset overflow, ansi 2",
			s:                         "\x1b[38;2;0;0;255mfourth\x1b[m",
			xOffset:                   41,
			width:                     10,
			lineContinuationIndicator: "...",
			expected:                  "",
		},
		{
			name: "offset start space ansi",
			// 							   0123456789012345   67890
			//       									  0       123456789012345678901234
			s:                         "\x1b[38;2;255;0;0ma\x1b[m really really long line",
			xOffset:                   15,
			width:                     15,
			lineContinuationIndicator: "",
			expected:                  " long line",
		},
		{
			name:                      "ansi short",
			s:                         "\x1b[38;2;0;0;255mhi\x1b[m",
			xOffset:                   0,
			width:                     3,
			lineContinuationIndicator: "...",
			expected:                  "\x1b[38;2;0;0;255mhi\x1b[m",
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
			s:                         "\x1b[38;2;0;0;255m├─flask\x1b[m",
			xOffset:                   0,
			width:                     6,
			lineContinuationIndicator: "...",
			expected:                  "\x1b[38;2;0;0;255m├─f...\x1b[m",
		},
		{
			name:                      "width exceeds capacity",
			s:                         "  │   └─[ ] local-path-provisioner (running for 11d)",
			xOffset:                   0,
			width:                     53,
			lineContinuationIndicator: "",
			expected:                  "  │   └─[ ] local-path-provisioner (running for 11d)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lb := New(tt.s, tt.lineContinuationIndicator)
			actual := lb.Truncate(tt.xOffset, tt.width)
			util.CmpStr(t, tt.expected, actual)
		})
	}
}
