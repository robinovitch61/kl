package linebuffer

import (
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/kl/internal/constants"
	"github.com/robinovitch61/kl/internal/util"
	"strings"
	"testing"
)

func TestTotalLines(t *testing.T) {
	tests := []struct {
		name         string
		s            string
		width        int
		continuation string
		expected     int
	}{
		{
			name:         "simple",
			s:            "1234567890",
			width:        10,
			continuation: "",
			expected:     1,
		},
		{
			name:         "simple small width",
			s:            "1234567890",
			width:        1,
			continuation: "",
			expected:     10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lb := New(tt.s, tt.width, tt.continuation, "", lipgloss.NewStyle())
			if lb.TotalLines() != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, lb.TotalLines())
			}
		})
	}
}

func TestPopLeft(t *testing.T) {
	highlightStyle := lipgloss.NewStyle().Background(lipgloss.Color("#FF0000"))
	tests := []struct {
		name         string
		s            string
		width        int
		continuation string
		toHighlight  string
		numPopLefts  int
		expected     []string
	}{
		{
			name:         "no ansi, no continuation, no width",
			s:            "12345678901234",
			width:        0,
			continuation: "",
			numPopLefts:  3,
			expected: []string{
				"",
				"",
				"",
			},
		},
		{
			name:         "no ansi, continuation, no width",
			s:            "12345678901234",
			width:        0,
			continuation: "...",
			numPopLefts:  3,
			expected: []string{
				"",
				"",
				"",
			},
		},
		{
			name:         "no ansi, no continuation, width 1",
			s:            "12345678901234",
			width:        1,
			continuation: "",
			numPopLefts:  3,
			expected: []string{
				"1",
				"2",
				"3",
			},
		},
		{
			name:         "no ansi, continuation, width 1",
			s:            "12345678901234",
			width:        1,
			continuation: "...",
			numPopLefts:  3,
			expected: []string{
				".",
				".",
				".",
			},
		},
		{
			name:         "no ansi, no continuation",
			s:            "12345678901234",
			width:        5,
			continuation: "",
			numPopLefts:  4,
			expected: []string{
				"12345",
				"67890",
				"1234",
				"",
			},
		},
		{
			name:         "no ansi, continuation",
			s:            "12345678901234",
			width:        5,
			continuation: "...",
			numPopLefts:  4,
			expected: []string{
				"12...",
				".....",
				"...4",
				"",
			},
		},
		{
			name:         "no ansi, no continuation",
			s:            "12345678901234",
			width:        5,
			continuation: "",
			numPopLefts:  4,
			expected: []string{
				"12345",
				"67890",
				"1234",
				"",
			},
		},
		{
			name:         "no ansi, continuation",
			s:            "12345678901234",
			width:        5,
			continuation: "...",
			numPopLefts:  4,
			expected: []string{
				"12...",
				".....",
				"...4",
				"",
			},
		},
		{
			name:         "double width unicode, no continuation, no width",
			s:            "世界🌟", // each of these takes up 2 terminal cells
			width:        0,
			continuation: "",
			numPopLefts:  3,
			expected: []string{
				"",
				"",
				"",
			},
		},
		{
			name:         "double width unicode, continuation, no width",
			s:            "世界🌟", // each of these takes up 2 terminal cells
			width:        0,
			continuation: "...",
			numPopLefts:  3,
			expected: []string{
				"",
				"",
				"",
			},
		},
		{
			name:         "double width unicode, no continuation, width 1",
			s:            "世界🌟", // each of these takes up 2 terminal cells
			width:        1,
			continuation: "",
			numPopLefts:  3,
			expected: []string{
				"",
				"",
				"",
			},
		},
		{
			name:         "double width unicode, continuation, width 1",
			s:            "世界🌟", // each of these takes up 2 terminal cells
			width:        1,
			continuation: "...",
			numPopLefts:  3,
			expected: []string{
				"",
				"",
				"",
			},
		},
		{
			name:         "double width unicode, no continuation, width 2",
			s:            "世界🌟", // each of these takes up 2 terminal cells
			width:        2,
			continuation: "",
			numPopLefts:  4,
			expected: []string{
				"世",
				"界",
				"🌟",
				"",
			},
		},
		{
			name:         "double width unicode, continuation, width 2",
			s:            "世界🌟", // each of these takes up 2 terminal cells
			width:        2,
			continuation: "...",
			numPopLefts:  4,
			expected: []string{
				"..",
				"..",
				"..",
				"",
			},
		},
		{
			name:         "double width unicode, no continuation, width 3",
			s:            "世界🌟", // each of these takes up 2 terminal cells
			width:        3,
			continuation: "",
			numPopLefts:  4,
			expected: []string{
				"世",
				"界",
				"🌟",
				"",
			},
		},
		{
			name:         "double width unicode, continuation, width 3",
			s:            "世界🌟", // each of these takes up 2 terminal cells
			width:        3,
			continuation: "...",
			numPopLefts:  4,
			expected: []string{
				"..",
				"..",
				"..",
				"",
			},
		},
		{
			name:         "double width unicode, no continuation, width 4",
			s:            "世界🌟", // each of these takes up 2 terminal cells
			width:        4,
			continuation: "",
			numPopLefts:  3,
			expected: []string{
				"世界",
				"🌟",
				"",
			},
		},
		{
			name:         "double width unicode, continuation, width 3",
			s:            "世界🌟", // each of these takes up 2 terminal cells
			width:        4,
			continuation: "...",
			numPopLefts:  3,
			expected: []string{
				"世..",
				"..",
				"",
			},
		},
		{
			name:         "width equal to continuation",
			s:            "1234567890",
			width:        3,
			continuation: "...",
			numPopLefts:  4,
			expected: []string{
				"...",
				"...",
				"...",
				".",
			},
		},
		{
			name:         "width slightly bigger than continuation",
			s:            "1234567890",
			width:        4,
			continuation: "...",
			numPopLefts:  3,
			expected: []string{
				"1...",
				"....",
				"..",
			},
		},
		{
			name:         "width double continuation 1",
			s:            "123456789012345678",
			width:        6,
			continuation: "...",
			numPopLefts:  3,
			expected: []string{
				"123...",
				"......",
				"...678",
			},
		},
		{
			name:         "width double continuation 2",
			s:            "1234567890123456789",
			width:        6,
			continuation: "...",
			numPopLefts:  4,
			expected: []string{
				"123...",
				"......",
				"......",
				".",
			},
		},
		{
			name:         "small string",
			s:            "hi",
			width:        3,
			continuation: "...",
			numPopLefts:  1,
			expected:     []string{"hi"},
		},
		{
			name:         "lineContinuationIndicator longer than width",
			s:            "1234567890123456789012345",
			width:        1,
			continuation: "...",
			numPopLefts:  1,
			expected:     []string{"."},
		},
		{
			name:         "twice the lineContinuationIndicator longer than width",
			s:            "1234567",
			width:        5,
			continuation: "...",
			numPopLefts:  1,
			expected:     []string{"12..."},
		},
		{
			name:         "sufficient width",
			s:            "1234567890123456789012345",
			width:        30,
			continuation: "...",
			numPopLefts:  1,
			expected:     []string{"1234567890123456789012345"},
		},
		{
			name:         "sufficient width, space at end",
			s:            "1234567890123456789012345     ",
			width:        30,
			continuation: "...",
			numPopLefts:  1,
			expected:     []string{"1234567890123456789012345     "},
		},
		{
			name:         "insufficient width",
			s:            "1234567890123456789012345",
			width:        15,
			continuation: "...",
			numPopLefts:  1,
			expected:     []string{"123456789012..."},
		},
		{
			name:         "insufficient width",
			s:            "123456789012345678901234567890123456789012345",
			width:        15,
			continuation: "...",
			numPopLefts:  3,
			expected: []string{
				"123456789012...",
				"...901234567...",
				"...456789012345",
			},
		},
		{
			name:         "ansi simple, no continuation",
			s:            "\x1b[38;2;255;0;0ma really really long line\x1b[m",
			width:        15,
			continuation: "",
			numPopLefts:  2,
			expected: []string{
				"\x1b[38;2;255;0;0ma really really\x1b[m",
				"\x1b[38;2;255;0;0m long line\x1b[m",
			},
		},
		{
			name:         "ansi simple, continuation",
			s:            "\x1b[38;2;255;0;0m12345678901234567890123456789012345\x1b[m",
			width:        15,
			continuation: "...",
			numPopLefts:  3,
			expected: []string{
				"\x1b[38;2;255;0;0m123456789012...\x1b[m",
				"\x1b[38;2;255;0;0m...901234567...\x1b[m",
				"\x1b[38;2;255;0;0m...45\x1b[m",
			},
		},
		{
			name:         "inline ansi, no continuation",
			s:            "\x1b[38;2;255;0;0ma\x1b[m really really long line",
			width:        15,
			continuation: "",
			numPopLefts:  2,
			expected: []string{
				"\x1b[38;2;255;0;0ma\x1b[m really really",
				" long line",
			},
		},
		{
			name:         "inline ansi, continuation",
			s:            "|\x1b[38;2;169;15;15mfl..-1\x1b[m| {\"timestamp\": \"now\"}",
			width:        15,
			continuation: "...",
			numPopLefts:  3,
			expected: []string{
				"|\x1b[38;2;169;15;15mfl..-1\x1b[m| {\"t...",
				"...mp\": \"now\"}",
				"",
			},
		},
		{
			name:         "ansi short",
			s:            "\x1b[38;2;0;0;255mhi\x1b[m",
			width:        3,
			continuation: "...",
			numPopLefts:  1,
			expected: []string{
				"\x1b[38;2;0;0;255mhi\x1b[m",
			},
		},
		{
			name:         "multi-byte runes",
			s:            "├─flask",
			width:        6,
			continuation: "...",
			numPopLefts:  1,
			expected: []string{
				"├─f...",
			},
		},
		{
			name:         "multi-byte runes with ansi",
			s:            "\x1b[38;2;0;0;255m├─flask\x1b[m",
			width:        6,
			continuation: "...",
			numPopLefts:  1,
			expected: []string{
				"\x1b[38;2;0;0;255m├─f...\x1b[m",
			},
		},
		{
			name:         "width exceeds capacity",
			s:            "  │   └─[ ] local-path-provisioner (running for 11d)",
			width:        53,
			continuation: "",
			numPopLefts:  1,
			expected: []string{
				"  │   └─[ ] local-path-provisioner (running for 11d)",
			},
		},
		{
			name:         "toHighlight, no continuation, no overflow, no ansi",
			s:            "a very normal log",
			width:        15,
			continuation: "",
			toHighlight:  "very",
			numPopLefts:  1,
			expected: []string{
				"a " + highlightStyle.Render("very") + " normal l",
			},
		},
		{
			name:         "toHighlight, no continuation, no overflow, no ansi",
			s:            "a very normal log",
			width:        15,
			continuation: "",
			toHighlight:  "very",
			numPopLefts:  1,
			expected: []string{
				"a " + highlightStyle.Render("very") + " normal l",
			},
		},
		{
			name:         "toHighlight, continuation, no overflow, no ansi",
			s:            "a very normal log",
			width:        15,
			continuation: "...",
			toHighlight:  "l l",
			numPopLefts:  1,
			expected: []string{
				"a very norma...", // does not highlight continuation, could in future
			},
		},
		{
			name:         "toHighlight, no continuation, no overflow, no ansi, many matches",
			s:            strings.Repeat("r", 10),
			width:        6,
			continuation: "",
			toHighlight:  "r",
			numPopLefts:  2,
			expected: []string{
				strings.Repeat("\x1b[48;2;255;0;0mr\x1b[m", 6),
				strings.Repeat("\x1b[48;2;255;0;0mr\x1b[m", 4),
			},
		},
		{
			name:         "toHighlight, no continuation, no overflow, ansi",
			s:            "\x1b[38;2;0;0;255mhi \x1b[48;2;0;255;0mthere\x1b[m er",
			width:        15,
			continuation: "",
			toHighlight:  "er",
			numPopLefts:  1,
			expected: []string{
				"\x1b[38;2;0;0;255mhi \x1b[48;2;0;255;0mth\x1b[m\x1b[48;2;255;0;0mer\x1b[m\x1b[38;2;0;0;255m\x1b[48;2;0;255;0me\x1b[m \x1b[48;2;255;0;0mer\x1b[m",
			},
		},
		{
			name:         "toHighlight, no continuation, overflows left and right, no ansi",
			s:            "hi there re",
			width:        6,
			continuation: "",
			toHighlight:  "hi there",
			numPopLefts:  2,
			expected: []string{
				highlightStyle.Render("hi the"),
				highlightStyle.Render("re") + " re",
			},
		},
		{
			name:         "toHighlight, no continuation, overflows left and right, ansi",
			s:            "\x1b[38;2;0;0;255mhi there re\x1b[m",
			width:        6,
			continuation: "",
			toHighlight:  "hi there",
			numPopLefts:  2,
			expected: []string{
				"\x1b[38;2;0;0;255m\x1b[m\x1b[48;2;255;0;0mhi the\x1b[m\x1b[38;2;0;0;255m\x1b[m",
				"\x1b[38;2;0;0;255m\x1b[m\x1b[48;2;255;0;0mre\x1b[m\x1b[38;2;0;0;255m re\x1b[m",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.expected) != tt.numPopLefts {
				t.Fatalf("num expected != num popLefts")
			}
			lb := New(tt.s, tt.width, tt.continuation, tt.toHighlight, highlightStyle)
			for i := 0; i < tt.numPopLefts; i++ {
				actual := lb.PopLeft()
				util.CmpStr(t, tt.expected[i], actual)
			}
		})
	}
}

func TestReapplyAnsi(t *testing.T) {
	tests := []struct {
		name            string
		original        string
		truncated       string
		truncByteOffset int
		expected        string
	}{
		{
			name:            "no ansi, no offset",
			original:        "1234567890123456789012345",
			truncated:       "12345",
			truncByteOffset: 0,
			expected:        "12345",
		},
		{
			name:            "no ansi, offset",
			original:        "1234567890123456789012345",
			truncated:       "2345",
			truncByteOffset: 1,
			expected:        "2345",
		},
		{
			name:            "multi ansi, no offset",
			original:        "\x1b[38;2;255;0;0m1\x1b[m\x1b[38;2;0;0;255m2\x1b[m\x1b[38;2;255;0;0m3\x1b[m45",
			truncated:       "123",
			truncByteOffset: 0,
			expected:        "\x1b[38;2;255;0;0m1\x1b[m\x1b[38;2;0;0;255m2\x1b[m\x1b[38;2;255;0;0m3\x1b[m",
		},
		{
			name:            "surrounding ansi, no offset",
			original:        "\x1b[38;2;255;0;0m12345\x1b[m",
			truncated:       "123",
			truncByteOffset: 0,
			expected:        "\x1b[38;2;255;0;0m123\x1b[m",
		},
		{
			name:            "surrounding ansi, offset",
			original:        "\x1b[38;2;255;0;0m12345\x1b[m",
			truncated:       "234",
			truncByteOffset: 1,
			expected:        "\x1b[38;2;255;0;0m234\x1b[m",
		},
		{
			name:            "left ansi, no offset",
			original:        "\x1b[38;2;255;0;0m1\x1b[m" + "2345",
			truncated:       "123",
			truncByteOffset: 0,
			expected:        "\x1b[38;2;255;0;0m1\x1b[m" + "23",
		},
		{
			name:            "left ansi, offset",
			original:        "\x1b[38;2;255;0;0m12\x1b[m" + "345",
			truncated:       "234",
			truncByteOffset: 1,
			expected:        "\x1b[38;2;255;0;0m2\x1b[m" + "34",
		},
		{
			name:            "right ansi, no offset",
			original:        "1" + "\x1b[38;2;255;0;0m2345\x1b[m",
			truncated:       "123",
			truncByteOffset: 0,
			expected:        "1" + "\x1b[38;2;255;0;0m23\x1b[m",
		},
		{
			name:            "right ansi, offset",
			original:        "12" + "\x1b[38;2;255;0;0m345\x1b[m",
			truncated:       "234",
			truncByteOffset: 1,
			expected:        "2" + "\x1b[38;2;255;0;0m34\x1b[m",
		},
		{
			name:            "left and right ansi, no offset",
			original:        "\x1b[38;2;255;0;0m1\x1b[m" + "2" + "\x1b[38;2;255;0;0m345\x1b[m",
			truncated:       "123",
			truncByteOffset: 0,
			expected:        "\x1b[38;2;255;0;0m1\x1b[m" + "2" + "\x1b[38;2;255;0;0m3\x1b[m",
		},
		{
			name:            "left and right ansi, offset",
			original:        "\x1b[38;2;255;0;0m12\x1b[m" + "3" + "\x1b[38;2;255;0;0m45\x1b[m",
			truncated:       "234",
			truncByteOffset: 1,
			expected:        "\x1b[38;2;255;0;0m2\x1b[m" + "3" + "\x1b[38;2;255;0;0m4\x1b[m",
		},
		{
			name:            "truncated right ansi, no offset",
			original:        "\x1b[38;2;255;0;0m1\x1b[m" + "234" + "\x1b[38;2;255;0;0m5\x1b[m",
			truncated:       "123",
			truncByteOffset: 0,
			expected:        "\x1b[38;2;255;0;0m1\x1b[m" + "23",
		},
		{
			name:            "truncated right ansi, offset",
			original:        "\x1b[38;2;255;0;0m12\x1b[m" + "34" + "\x1b[38;2;255;0;0m5\x1b[m",
			truncated:       "234",
			truncByteOffset: 1,
			expected:        "\x1b[38;2;255;0;0m2\x1b[m" + "34",
		},
		{
			name:            "truncated left ansi, offset",
			original:        "\x1b[38;2;255;0;0m1\x1b[m" + "23" + "\x1b[38;2;255;0;0m45\x1b[m",
			truncated:       "234",
			truncByteOffset: 1,
			expected:        "23" + "\x1b[38;2;255;0;0m4\x1b[m",
		},
		{
			name:            "nested color sequences",
			original:        "\x1b[31m1\x1b[32m2\x1b[33m3\x1b[m\x1b[m\x1b[m45",
			truncated:       "123",
			truncByteOffset: 0,
			expected:        "\x1b[31m1\x1b[32m2\x1b[33m3\x1b[m",
		},
		{
			name:            "nested color sequences with offset",
			original:        "\x1b[31m1\x1b[32m2\x1b[33m3\x1b[m\x1b[m\x1b[m45",
			truncated:       "234",
			truncByteOffset: 1,
			expected:        "\x1b[31m\x1b[32m2\x1b[33m3\x1b[m4",
		},
		{
			name:            "nested style sequences",
			original:        "\x1b[1m1\x1b[4m2\x1b[3m3\x1b[m\x1b[m\x1b[m45",
			truncated:       "123",
			truncByteOffset: 0,
			expected:        "\x1b[1m1\x1b[4m2\x1b[3m3\x1b[m",
		},
		{
			name:            "mixed nested sequences",
			original:        "\x1b[31m1\x1b[1m2\x1b[4;32m3\x1b[m\x1b[m\x1b[m45",
			truncated:       "234",
			truncByteOffset: 1,
			expected:        "\x1b[31m\x1b[1m2\x1b[4;32m3\x1b[m4",
		},
		{
			name:            "deeply nested sequences",
			original:        "\x1b[31m1\x1b[1m2\x1b[4m3\x1b[32m4\x1b[m\x1b[m\x1b[m\x1b[m5",
			truncated:       "123",
			truncByteOffset: 0,
			expected:        "\x1b[31m1\x1b[1m2\x1b[4m3\x1b[m",
		},
		{
			name:            "partial nested sequences",
			original:        "1\x1b[31m2\x1b[1m3\x1b[4m4\x1b[m\x1b[m\x1b[m5",
			truncated:       "234",
			truncByteOffset: 1,
			expected:        "\x1b[31m2\x1b[1m3\x1b[4m4\x1b[m",
		},
		{
			name:            "overlapping nested sequences",
			original:        "\x1b[31m1\x1b[1m2\x1b[m3\x1b[4m4\x1b[m5",
			truncated:       "234",
			truncByteOffset: 1,
			expected:        "\x1b[31m\x1b[1m2\x1b[m3\x1b[4m4\x1b[m",
		},
		{
			name:            "complex RGB nested sequences",
			original:        "\x1b[38;2;255;0;0m1\x1b[1m2\x1b[38;2;0;255;0m3\x1b[m\x1b[m45",
			truncated:       "123",
			truncByteOffset: 0,
			expected:        "\x1b[38;2;255;0;0m1\x1b[1m2\x1b[38;2;0;255;0m3\x1b[m",
		},
		{
			name:            "nested sequences with background colors",
			original:        "\x1b[31;44m1\x1b[1m2\x1b[32;45m3\x1b[m\x1b[m45",
			truncated:       "234",
			truncByteOffset: 1,
			expected:        "\x1b[31;44m\x1b[1m2\x1b[32;45m3\x1b[m4",
		},
		{
			name:            "emoji basic",
			original:        "1️⃣2️⃣3️⃣4️⃣5️⃣",
			truncated:       "1️⃣2️⃣3️⃣",
			truncByteOffset: 0,
			expected:        "1️⃣2️⃣3️⃣",
		},
		{
			name:            "emoji with ansi",
			original:        "\x1b[31m1️⃣\x1b[32m2️⃣\x1b[33m3️⃣\x1b[m",
			truncated:       "1️⃣2️⃣",
			truncByteOffset: 0,
			expected:        "\x1b[31m1️⃣\x1b[32m2️⃣\x1b[m",
		},
		{
			name:            "chinese characters",
			original:        "你好世界星星",
			truncated:       "你好世",
			truncByteOffset: 0,
			expected:        "你好世",
		},
		{
			name:            "simple with ansi and offset",
			original:        "\x1b[31ma\x1b[32mb\x1b[33mc\x1b[mde",
			truncated:       "bcd",
			truncByteOffset: 1,
			expected:        "\x1b[31m\x1b[32mb\x1b[33mc\x1b[md",
		},
		{
			name:            "chinese with ansi and offset",
			original:        "\x1b[31m你\x1b[32m好\x1b[33m世\x1b[m界星",
			truncated:       "好世界",
			truncByteOffset: 3, // 你 is 3 bytes
			expected:        "\x1b[31m\x1b[32m好\x1b[33m世\x1b[m界",
		},
		{
			name:            "lots of leading ansi",
			original:        "\x1b[38;2;255;0;0mr\x1b[m\x1b[38;2;255;0;0mr\x1b[m\x1b[38;2;255;0;0mr\x1b[m\x1b[38;2;255;0;0mr\x1b[m\x1b[38;2;255;0;0mr\x1b[m\x1b[38;2;255;0;0mr\x1b[m\x1b[38;2;255;0;0mr\x1b[m\x1b[38;2;255;0;0mr\x1b[m\x1b[38;2;255;0;0mr\x1b[m\x1b[38;2;255;0;0mr\x1b[m\x1b[38;2;255;0;0mr\x1b[m",
			truncated:       "r",
			truncByteOffset: 10,
			expected:        "\x1b[38;2;255;0;0mr\x1b[m",
		},
		{
			name:            "complex ansi, no offset",
			original:        "\x1b[38;2;0;0;255msome \x1b[m\x1b[38;2;255;0;0mred\x1b[m\x1b[38;2;0;0;255m t\x1b[m",
			truncated:       "some red t",
			truncByteOffset: 0,
			expected:        "\x1b[38;2;0;0;255msome \x1b[m\x1b[38;2;255;0;0mred\x1b[m\x1b[38;2;0;0;255m t\x1b[m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ansiCodeIndexes := constants.AnsiRegex.FindAllStringIndex(tt.original, -1)
			actual := reapplyANSI(tt.original, tt.truncated, tt.truncByteOffset, ansiCodeIndexes)
			util.CmpStr(t, tt.expected, actual)
		})
	}
}

func TestHighlightLine(t *testing.T) {
	red := lipgloss.Color("#ff0000")
	blue := lipgloss.Color("#0000ff")

	for _, tt := range []struct {
		name           string
		line           string
		highlight      string
		highlightStyle lipgloss.Style
		start          int
		end            int
		expected       string
	}{
		{
			name:           "empty",
			line:           "",
			highlight:      "",
			highlightStyle: lipgloss.NewStyle().Foreground(red),
			expected:       "",
		},
		{
			name:           "no highlight",
			line:           "hello",
			highlight:      "",
			highlightStyle: lipgloss.NewStyle().Foreground(red),
			expected:       "hello",
		},
		{
			name:           "highlight",
			line:           "hello",
			highlight:      "ell",
			highlightStyle: lipgloss.NewStyle().Foreground(red),
			expected:       "h\x1b[38;2;255;0;0mell\x1b[mo",
		},
		{
			name:           "highlight already styled line",
			line:           "\x1b[38;2;255;0;0mfirst line\x1b[m",
			highlight:      "first",
			highlightStyle: lipgloss.NewStyle().Foreground(blue),
			expected:       "\x1b[38;2;255;0;0m\x1b[m\x1b[38;2;0;0;255mfirst\x1b[m\x1b[38;2;255;0;0m line\x1b[m",
		},
		{
			name:           "highlight already partially styled line",
			line:           "hi a \x1b[38;2;255;0;0mstyled line\x1b[m cool \x1b[38;2;255;0;0mand styled\x1b[m more",
			highlight:      "style",
			highlightStyle: lipgloss.NewStyle().Foreground(blue),
			expected:       "hi a \x1b[38;2;255;0;0m\x1b[m\x1b[38;2;0;0;255mstyle\x1b[m\x1b[38;2;255;0;0md line\x1b[m cool \x1b[38;2;255;0;0mand \x1b[m\x1b[38;2;0;0;255mstyle\x1b[m\x1b[38;2;255;0;0md\x1b[m more",
		},
		{
			name:           "dont highlight ansi escape codes themselves",
			line:           "\x1b[38;2;255;0;0mhi\x1b[m",
			highlight:      "38",
			highlightStyle: lipgloss.NewStyle().Foreground(blue),
			expected:       "\x1b[38;2;255;0;0mhi\x1b[m",
		},
		{
			name:           "single letter in partially styled line",
			line:           "line \x1b[38;2;255;0;0mred\x1b[m e again",
			highlight:      "e",
			highlightStyle: lipgloss.NewStyle().Foreground(blue),
			expected:       "lin\x1b[38;2;0;0;255me\x1b[m \x1b[38;2;255;0;0mr\x1b[m\x1b[38;2;0;0;255me\x1b[m\x1b[38;2;255;0;0md\x1b[m \x1b[38;2;0;0;255me\x1b[m again",
		},
		{
			name:           "super long line",
			line:           strings.Repeat("python generator code world world world code text test code words random words generator hello python generator", 10000),
			highlight:      "e",
			highlightStyle: lipgloss.NewStyle().Foreground(red),
			expected:       strings.Repeat("python g\x1b[38;2;255;0;0me\x1b[mn\x1b[38;2;255;0;0me\x1b[mrator cod\x1b[38;2;255;0;0me\x1b[m world world world cod\x1b[38;2;255;0;0me\x1b[m t\x1b[38;2;255;0;0me\x1b[mxt t\x1b[38;2;255;0;0me\x1b[mst cod\x1b[38;2;255;0;0me\x1b[m words random words g\x1b[38;2;255;0;0me\x1b[mn\x1b[38;2;255;0;0me\x1b[mrator h\x1b[38;2;255;0;0me\x1b[mllo python g\x1b[38;2;255;0;0me\x1b[mn\x1b[38;2;255;0;0me\x1b[mrator", 10000),
		},
		{
			name:           "start and end",
			line:           "my line",
			highlight:      "line",
			highlightStyle: lipgloss.NewStyle().Foreground(red),
			start:          0,
			end:            2,
			expected:       "my line",
		},
		{
			name:           "start and end ansi, in range",
			line:           "\x1b[38;2;0;0;255mmy line\x1b[m",
			highlight:      "my",
			highlightStyle: lipgloss.NewStyle().Foreground(red),
			start:          0,
			end:            2,
			expected:       "\x1b[38;2;0;0;255m\x1b[m\x1b[38;2;255;0;0mmy\x1b[m\x1b[38;2;0;0;255m line\x1b[m",
		},
		{
			name:           "start and end ansi, out of range",
			line:           "\x1b[38;2;0;0;255mmy line\x1b[m",
			highlight:      "my",
			highlightStyle: lipgloss.NewStyle().Foreground(red),
			start:          2,
			end:            4,
			expected:       "\x1b[38;2;0;0;255mmy line\x1b[m",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.start == 0 && tt.end == 0 {
				tt.end = len(tt.line)
			}
			util.CmpStr(t, tt.expected, highlightLine(tt.line, tt.highlight, tt.highlightStyle, tt.start, tt.end))
		})
	}
}

func TestOverflowsLeft(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		index    int
		substr   string
		wantBool bool
		wantInt  int
	}{
		{
			name:     "basic overflow case",
			str:      "my str here",
			index:    3,
			substr:   "my str",
			wantBool: true,
			wantInt:  6,
		},
		{
			name:     "no overflow case",
			str:      "my str here",
			index:    6,
			substr:   "my str",
			wantBool: false,
			wantInt:  0,
		},
		{
			name:     "empty string",
			str:      "",
			index:    0,
			substr:   "test",
			wantBool: false,
			wantInt:  0,
		},
		{
			name:     "empty substring",
			str:      "test string",
			index:    0,
			substr:   "",
			wantBool: false,
			wantInt:  0,
		},
		{
			name:     "index out of bounds",
			str:      "test",
			index:    10,
			substr:   "test",
			wantBool: false,
			wantInt:  0,
		},
		{
			name:     "exact full match",
			str:      "hello world",
			index:    0,
			substr:   "hello world",
			wantBool: false,
			wantInt:  0,
		},
		{
			name:     "partial overflow at end",
			str:      "hello world",
			index:    9,
			substr:   "dd",
			wantBool: false,
			wantInt:  0,
		},
		{
			name:     "case sensitivity test - no match",
			str:      "Hello World",
			index:    0,
			substr:   "hello",
			wantBool: false,
			wantInt:  0,
		},
		{
			name:     "multiple character same overflow",
			str:      "aaaa",
			index:    1,
			substr:   "aaa",
			wantBool: true,
			wantInt:  3,
		},
		{
			name:     "multiple character same overflow but difference",
			str:      "aaaa",
			index:    1,
			substr:   "baaa",
			wantBool: false,
			wantInt:  0,
		},
		{
			name:     "special characters",
			str:      "test!@#$",
			index:    4,
			substr:   "st!@#",
			wantBool: true,
			wantInt:  7,
		},
		{
			name:     "false if does not overflow",
			str:      "some string",
			index:    1,
			substr:   "ome",
			wantBool: false,
			wantInt:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBool, gotInt := overflowsLeft(tt.str, tt.index, tt.substr)
			if gotBool != tt.wantBool || gotInt != tt.wantInt {
				t.Errorf("overflowsLeft(%q, %d, %q) = (%v, %d), want (%v, %d)",
					tt.str, tt.index, tt.substr, gotBool, gotInt, tt.wantBool, tt.wantInt)
			}
		})
	}
}

func TestOverflowsRight(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		index    int
		substr   string
		wantBool bool
		wantInt  int
	}{
		{
			name:     "example 1",
			str:      "my str here",
			index:    3,
			substr:   "y str",
			wantBool: true,
			wantInt:  1,
		},
		{
			name:     "example 2",
			str:      "my str here",
			index:    3,
			substr:   "y strong",
			wantBool: false,
			wantInt:  0,
		},
		{
			name:     "example 3",
			str:      "my str here",
			index:    6,
			substr:   "tr here",
			wantBool: true,
			wantInt:  4,
		},
		{
			name:     "empty string",
			str:      "",
			index:    0,
			substr:   "test",
			wantBool: false,
			wantInt:  0,
		},
		{
			name:     "empty substring",
			str:      "test string",
			index:    0,
			substr:   "",
			wantBool: false,
			wantInt:  0,
		},
		{
			name:     "index out of bounds",
			str:      "test",
			index:    10,
			substr:   "test",
			wantBool: false,
			wantInt:  0,
		},
		{
			name:     "exact full match",
			str:      "hello world",
			index:    10,
			substr:   "hello world",
			wantBool: false,
			wantInt:  0,
		},
		{
			name:     "case sensitivity test - no match",
			str:      "Hello World",
			index:    4,
			substr:   "hello",
			wantBool: false,
			wantInt:  0,
		},
		{
			name:     "multiple character same overflow",
			str:      "aaaa",
			index:    2,
			substr:   "aaa",
			wantBool: true,
			wantInt:  1,
		},
		{
			name:     "multiple character same overflow but difference",
			str:      "aaaa",
			index:    2,
			substr:   "aaab",
			wantBool: false,
			wantInt:  0,
		},
		{
			name:     "special characters",
			str:      "test!@#$",
			index:    6,
			substr:   "@#$",
			wantBool: true,
			wantInt:  5,
		},
		{
			name:     "false if does not overflow",
			str:      "some string",
			index:    4,
			substr:   "ome ",
			wantBool: false,
			wantInt:  0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBool, gotInt := overflowsRight(tt.str, tt.index, tt.substr)
			if gotBool != tt.wantBool || gotInt != tt.wantInt {
				t.Errorf("overflowsRight(%q, %d, %q) = (%v, %d), want (%v, %d)",
					tt.str, tt.index, tt.substr, gotBool, gotInt, tt.wantBool, tt.wantInt)
			}
		})
	}
}
