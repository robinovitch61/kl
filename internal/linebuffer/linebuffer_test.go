package linebuffer

import (
	"github.com/robinovitch61/kl/internal/constants"
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
			name:                      "zero width zero truncOffset",
			s:                         "1234567890123456789012345",
			xOffset:                   0,
			width:                     0,
			lineContinuationIndicator: "...",
			expected:                  "",
		},
		{
			name:                      "zero width positive truncOffset",
			s:                         "1234567890123456789012345",
			xOffset:                   5,
			width:                     0,
			lineContinuationIndicator: "...",
			expected:                  "",
		},
		{
			name:                      "zero width negative truncOffset",
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
			name:                      "zero truncOffset, sufficient width",
			s:                         "1234567890123456789012345",
			xOffset:                   0,
			width:                     30,
			lineContinuationIndicator: "...",
			expected:                  "1234567890123456789012345",
		},
		{
			name:                      "zero truncOffset, sufficient width, space at end",
			s:                         "1234567890123456789012345     ",
			xOffset:                   0,
			width:                     30,
			lineContinuationIndicator: "...",
			expected:                  "1234567890123456789012345     ",
		},
		{
			name:                      "zero truncOffset, insufficient width",
			s:                         "1234567890123456789012345",
			xOffset:                   0,
			width:                     15,
			lineContinuationIndicator: "...",
			expected:                  "123456789012...",
		},
		{
			name:                      "positive truncOffset, insufficient width",
			s:                         "1234567890123456789012345",
			xOffset:                   5,
			width:                     15,
			lineContinuationIndicator: "...",
			expected:                  "...901234567...",
		},
		{
			name:                      "positive truncOffset, exactly at end",
			s:                         "1234567890123456789012345",
			xOffset:                   15,
			width:                     10,
			lineContinuationIndicator: "...",
			expected:                  "...9012345",
		},
		{
			name:                      "positive truncOffset, over the end",
			s:                         "1234567890123456789012345",
			xOffset:                   20,
			width:                     10,
			lineContinuationIndicator: "...",
			expected:                  "...45",
		},
		{
			name:                      "positive truncOffset, ansi",
			s:                         "\x1b[38;2;255;0;0ma really really long line\x1b[m",
			xOffset:                   15,
			width:                     15,
			lineContinuationIndicator: "",
			expected:                  "\x1b[38;2;255;0;0m long line\x1b[m",
		},
		{
			name:                      "zero truncOffset, insufficient width, ansi",
			s:                         "\x1b[38;2;255;0;0m1234567890123456789012345\x1b[m",
			xOffset:                   0,
			width:                     15,
			lineContinuationIndicator: "...",
			expected:                  "\x1b[38;2;255;0;0m123456789012...\x1b[m",
		},
		{
			name:                      "positive truncOffset, insufficient width, ansi",
			s:                         "\x1b[38;2;255;0;0m1234567890123456789012345\x1b[m",
			xOffset:                   5,
			width:                     15,
			lineContinuationIndicator: "...",
			expected:                  "\x1b[38;2;255;0;0m...901234567...\x1b[m",
		},
		{
			name:                      "no truncOffset, insufficient width, inline ansi",
			s:                         "|\x1b[38;2;169;15;15mfl..-1\x1b[m| {\"timestamp\": \"2024-09-29T22:30:28.730520\"}",
			xOffset:                   0,
			width:                     15,
			lineContinuationIndicator: "...",
			expected:                  "|\x1b[38;2;169;15;15mfl..-1\x1b[m| {\"t...",
		},
		{
			name:                      "truncOffset overflow, ansi",
			s:                         "\x1b[38;2;0;0;255mthird line that is fairly long\x1b[m",
			xOffset:                   41,
			width:                     10,
			lineContinuationIndicator: "...",
			expected:                  "",
		},
		{
			name:                      "truncOffset overflow, ansi 2",
			s:                         "\x1b[38;2;0;0;255mfourth\x1b[m",
			xOffset:                   41,
			width:                     10,
			lineContinuationIndicator: "...",
			expected:                  "",
		},
		{
			name: "truncOffset start space ansi",
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

func TestReapplyAnsi(t *testing.T) {
	tests := []struct {
		name        string
		original    string
		truncated   string
		truncOffset int
		expected    string
	}{
		{
			name:        "no ansi, no truncOffset",
			original:    "1234567890123456789012345",
			truncated:   "12345",
			truncOffset: 0,
			expected:    "12345",
		},
		{
			name:        "no ansi, truncOffset",
			original:    "1234567890123456789012345",
			truncated:   "2345",
			truncOffset: 1,
			expected:    "2345",
		},
		{
			name:        "surrounding ansi, no truncOffset",
			original:    "\x1b[38;2;255;0;0m12345\x1b[m",
			truncated:   "123",
			truncOffset: 0,
			expected:    "\x1b[38;2;255;0;0m123\x1b[m",
		},
		{
			name:        "surrounding ansi, truncOffset",
			original:    "\x1b[38;2;255;0;0m12345\x1b[m",
			truncated:   "234",
			truncOffset: 1,
			expected:    "\x1b[38;2;255;0;0m234\x1b[m",
		},
		{
			name:        "left ansi, no truncOffset",
			original:    "\x1b[38;2;255;0;0m1\x1b[m" + "2345",
			truncated:   "123",
			truncOffset: 0,
			expected:    "\x1b[38;2;255;0;0m1\x1b[m" + "23",
		},
		{
			name:        "left ansi, truncOffset",
			original:    "\x1b[38;2;255;0;0m12\x1b[m" + "345",
			truncated:   "234",
			truncOffset: 1,
			expected:    "\x1b[38;2;255;0;0m2\x1b[m" + "34",
		},
		{
			name:        "right ansi, no truncOffset",
			original:    "1" + "\x1b[38;2;255;0;0m2345\x1b[m",
			truncated:   "123",
			truncOffset: 0,
			expected:    "1" + "\x1b[38;2;255;0;0m23\x1b[m",
		},
		{
			name:        "right ansi, truncOffset",
			original:    "12" + "\x1b[38;2;255;0;0m345\x1b[m",
			truncated:   "234",
			truncOffset: 1,
			expected:    "2" + "\x1b[38;2;255;0;0m34\x1b[m",
		},
		{
			name:        "left and right ansi, no truncOffset",
			original:    "\x1b[38;2;255;0;0m1\x1b[m" + "2" + "\x1b[38;2;255;0;0m345\x1b[m",
			truncated:   "123",
			truncOffset: 0,
			expected:    "\x1b[38;2;255;0;0m1\x1b[m" + "2" + "\x1b[38;2;255;0;0m3\x1b[m",
		},
		{
			name:        "left and right ansi, truncOffset",
			original:    "\x1b[38;2;255;0;0m12\x1b[m" + "3" + "\x1b[38;2;255;0;0m45\x1b[m",
			truncated:   "234",
			truncOffset: 1,
			expected:    "\x1b[38;2;255;0;0m2\x1b[m" + "3" + "\x1b[38;2;255;0;0m4\x1b[m",
		},
		{
			name:        "truncated right ansi, no truncOffset",
			original:    "\x1b[38;2;255;0;0m1\x1b[m" + "234" + "\x1b[38;2;255;0;0m5\x1b[m",
			truncated:   "123",
			truncOffset: 0,
			expected:    "\x1b[38;2;255;0;0m1\x1b[m" + "23",
		},
		{
			name:        "truncated right ansi, truncOffset",
			original:    "\x1b[38;2;255;0;0m12\x1b[m" + "34" + "\x1b[38;2;255;0;0m5\x1b[m",
			truncated:   "234",
			truncOffset: 1,
			expected:    "\x1b[38;2;255;0;0m2\x1b[m" + "34",
		},
		{
			name:        "truncated left ansi, truncOffset",
			original:    "\x1b[38;2;255;0;0m1\x1b[m" + "23" + "\x1b[38;2;255;0;0m45\x1b[m",
			truncated:   "234",
			truncOffset: 1,
			expected:    "23" + "\x1b[38;2;255;0;0m4\x1b[m",
		},
		{
			name:        "nested color sequences",
			original:    "\x1b[31m1\x1b[32m2\x1b[33m3\x1b[m\x1b[m\x1b[m45",
			truncated:   "123",
			truncOffset: 0,
			expected:    "\x1b[31m1\x1b[32m2\x1b[33m3\x1b[m",
		},
		{
			name:        "nested color sequences with truncOffset",
			original:    "\x1b[31m1\x1b[32m2\x1b[33m3\x1b[m\x1b[m\x1b[m45",
			truncated:   "234",
			truncOffset: 1,
			expected:    "\x1b[31m\x1b[32m2\x1b[33m3\x1b[m4",
		},
		{
			name:        "nested style sequences",
			original:    "\x1b[1m1\x1b[4m2\x1b[3m3\x1b[m\x1b[m\x1b[m45",
			truncated:   "123",
			truncOffset: 0,
			expected:    "\x1b[1m1\x1b[4m2\x1b[3m3\x1b[m",
		},
		{
			name:        "mixed nested sequences",
			original:    "\x1b[31m1\x1b[1m2\x1b[4;32m3\x1b[m\x1b[m\x1b[m45",
			truncated:   "234",
			truncOffset: 1,
			expected:    "\x1b[31m\x1b[1m2\x1b[4;32m3\x1b[m4",
		},
		{
			name:        "deeply nested sequences",
			original:    "\x1b[31m1\x1b[1m2\x1b[4m3\x1b[32m4\x1b[m\x1b[m\x1b[m\x1b[m5",
			truncated:   "123",
			truncOffset: 0,
			expected:    "\x1b[31m1\x1b[1m2\x1b[4m3\x1b[m",
		},
		{
			name:        "partial nested sequences",
			original:    "1\x1b[31m2\x1b[1m3\x1b[4m4\x1b[m\x1b[m\x1b[m5",
			truncated:   "234",
			truncOffset: 1,
			expected:    "\x1b[31m2\x1b[1m3\x1b[4m4\x1b[m",
		},
		{
			name:        "overlapping nested sequences",
			original:    "\x1b[31m1\x1b[1m2\x1b[m3\x1b[4m4\x1b[m5",
			truncated:   "234",
			truncOffset: 1,
			expected:    "\x1b[31m\x1b[1m2\x1b[m3\x1b[4m4\x1b[m",
		},
		{
			name:        "complex RGB nested sequences",
			original:    "\x1b[38;2;255;0;0m1\x1b[1m2\x1b[38;2;0;255;0m3\x1b[m\x1b[m45",
			truncated:   "123",
			truncOffset: 0,
			expected:    "\x1b[38;2;255;0;0m1\x1b[1m2\x1b[38;2;0;255;0m3\x1b[m",
		},
		{
			name:        "nested sequences with background colors",
			original:    "\x1b[31;44m1\x1b[1m2\x1b[32;45m3\x1b[m\x1b[m45",
			truncated:   "234",
			truncOffset: 1,
			expected:    "\x1b[31;44m\x1b[1m2\x1b[32;45m3\x1b[m4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ansiCodeIndexes := constants.AnsiRegex.FindAllStringIndex(tt.original, -1)
			actual := reapplyANSI(tt.original, tt.truncated, tt.truncOffset, ansiCodeIndexes)
			util.CmpStr(t, tt.expected, actual)
		})
	}
}
