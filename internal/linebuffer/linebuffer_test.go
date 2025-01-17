package linebuffer

import (
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/kl/internal/constants"
	"github.com/robinovitch61/kl/internal/util"
	"strings"
	"testing"
)

func TestLineBuffer_TotalLines(t *testing.T) {
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
		{
			name:         "uneven number",
			s:            "1234567890",
			width:        3,
			continuation: "",
			expected:     4,
		},
		{
			name:         "unicode even",
			s:            "世界🌟世界",
			width:        2,
			continuation: "",
			expected:     5,
		},
		{
			name:         "unicode odd",
			s:            "世界🌟世界",
			width:        3,
			continuation: "",
			expected:     4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lb := New(tt.s)
			if lines := lb.TotalLines(tt.width); lines != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, lines)
			}
		})
	}
}

func TestLineBuffer_SeekToWidth(t *testing.T) {
	tests := []struct {
		name            string
		s               string
		width           int
		seekWidth       int
		continuation    string
		expectedPopLeft string
	}{
		{
			name:            "empty",
			s:               "",
			width:           10,
			seekWidth:       0,
			continuation:    "",
			expectedPopLeft: "",
		},
		{
			name:            "simple",
			s:               "1234567890",
			width:           10,
			seekWidth:       0,
			continuation:    "",
			expectedPopLeft: "1234567890",
		},
		{
			name:            "negative seekWidth",
			s:               "1234567890",
			width:           10,
			seekWidth:       -1,
			continuation:    "",
			expectedPopLeft: "1234567890",
		},
		{
			name:            "seek",
			s:               "1234567890",
			width:           10,
			seekWidth:       3,
			continuation:    "",
			expectedPopLeft: "4567890",
		},
		{
			name:            "seek to end",
			s:               "1234567890",
			width:           10,
			seekWidth:       10,
			continuation:    "",
			expectedPopLeft: "",
		},
		{
			name:            "seek past end",
			s:               "1234567890",
			width:           10,
			seekWidth:       11,
			continuation:    "",
			expectedPopLeft: "",
		},
		{
			name:            "continuation",
			s:               "1234567890",
			width:           7,
			seekWidth:       2,
			continuation:    "...",
			expectedPopLeft: "...6...",
		},
		{
			name:            "continuation past end",
			s:               "1234567890",
			width:           10,
			seekWidth:       11,
			continuation:    "...",
			expectedPopLeft: "",
		},
		{
			name:            "unicode",
			s:               "世界🌟世界🌟",
			width:           10,
			seekWidth:       0,
			continuation:    "",
			expectedPopLeft: "世界🌟世界",
		},
		{
			name:            "unicode seek past first rune",
			s:               "世界🌟世界🌟",
			width:           10,
			seekWidth:       2,
			continuation:    "",
			expectedPopLeft: "界🌟世界🌟",
		},
		{
			name:            "unicode seek past first 2 runes",
			s:               "世界🌟世界🌟",
			width:           10,
			seekWidth:       3,
			continuation:    "",
			expectedPopLeft: "🌟世界🌟",
		},
		{
			name:            "unicode seek past all but 1 rune",
			s:               "世界🌟世界🌟",
			width:           10,
			seekWidth:       10,
			continuation:    "",
			expectedPopLeft: "🌟",
		},
		{
			name:            "unicode seek almost to end",
			s:               "世界🌟世界🌟",
			width:           10,
			seekWidth:       11,
			continuation:    "",
			expectedPopLeft: "",
		},
		{
			name:            "unicode seek to end",
			s:               "世界🌟世界🌟",
			width:           10,
			seekWidth:       12,
			continuation:    "",
			expectedPopLeft: "",
		},
		{
			name:            "unicode insufficient width",
			s:               "世界🌟世界🌟",
			width:           1,
			seekWidth:       2,
			continuation:    "",
			expectedPopLeft: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lb := New(tt.s)
			lb.SeekToWidth(tt.seekWidth)
			// highlight tested in PopLeft tests
			if actual := lb.PopLeft(tt.width, tt.continuation, "", lipgloss.NewStyle()); actual != tt.expectedPopLeft {
				t.Errorf("expected %s, got %s", tt.expectedPopLeft, actual)
			}
		})
	}
}

func TestLineBuffer_SeekToLine(t *testing.T) {
	tests := []struct {
		name            string
		s               string
		width           int
		continuation    string
		seekToLine      int
		expectedPopLeft string
	}{
		{
			name:            "empty",
			s:               "",
			width:           0,
			continuation:    "",
			seekToLine:      0,
			expectedPopLeft: "",
		},
		{
			name:            "seek to negative line",
			s:               "12345",
			width:           2,
			continuation:    "",
			seekToLine:      -1,
			expectedPopLeft: "12",
		},
		{
			name:            "seek to zero'th line",
			s:               "12345",
			width:           2,
			continuation:    "",
			seekToLine:      0,
			expectedPopLeft: "12",
		},
		{
			name:            "seek to first line",
			s:               "12345",
			width:           2,
			continuation:    "",
			seekToLine:      1,
			expectedPopLeft: "34",
		},
		{
			name:            "seek to second line",
			s:               "12345",
			width:           2,
			continuation:    "",
			seekToLine:      2,
			expectedPopLeft: "5",
		},
		{
			name:            "seek past end",
			s:               "12345",
			width:           2,
			continuation:    "",
			seekToLine:      3,
			expectedPopLeft: "",
		},
		{
			name:            "unicode zero'th line",
			s:               "世界🌟世界🌟",
			width:           2,
			continuation:    "",
			seekToLine:      0,
			expectedPopLeft: "世",
		},
		{
			name:            "unicode first line",
			s:               "世界🌟世界🌟",
			width:           2,
			continuation:    "",
			seekToLine:      1,
			expectedPopLeft: "界",
		},
		{
			name:            "unicode insufficient width",
			s:               "世界🌟世界🌟",
			width:           1,
			continuation:    "",
			seekToLine:      1,
			expectedPopLeft: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lb := New(tt.s)
			lb.SeekToLine(tt.seekToLine, tt.width)
			// highlight tested in PopLeft tests
			actual := lb.PopLeft(tt.width, tt.continuation, "", lipgloss.NewStyle())
			util.CmpStr(t, tt.expectedPopLeft, actual)
		})
	}
}

func TestLineBuffer_PopLeft(t *testing.T) {
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
			name:         "sufficient width, space at end preserved",
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
				"\x1b[48;2;255;0;0mhi the\x1b[m",
				"\x1b[48;2;255;0;0mre\x1b[m\x1b[38;2;0;0;255m re\x1b[m",
			},
		},
		{
			name:         "toHighlight, no continuation, overflows left and right one char, no ansi",
			s:            "hi there re",
			width:        7,
			continuation: "",
			toHighlight:  "hi there",
			numPopLefts:  2,
			expected: []string{
				highlightStyle.Render("hi ther"),
				highlightStyle.Render("e") + " re",
			},
		},
		{
			name:         "unicode toHighlight, no continuation, no overflow, no ansi",
			s:            "世界🌟世界🌟",
			width:        7,
			continuation: "",
			toHighlight:  "世界",
			numPopLefts:  2,
			expected: []string{
				highlightStyle.Render("世界") + "🌟",
				highlightStyle.Render("世界") + "🌟",
			},
		},
		{
			name:         "unicode toHighlight, no continuation, overflow, no ansi",
			s:            "世界🌟世界🌟",
			width:        7,
			continuation: "",
			toHighlight:  "世界🌟世",
			numPopLefts:  2,
			expected: []string{
				highlightStyle.Render("世界🌟"),
				highlightStyle.Render("世") + "界🌟",
			},
		},
		{
			name:         "unicode toHighlight, no continuation, overflow, ansi",
			s:            "\x1b[38;2;0;0;255m世界🌟世界🌟\x1b[m",
			width:        7,
			continuation: "",
			toHighlight:  "世界🌟世",
			numPopLefts:  2,
			expected: []string{
				highlightStyle.Render("世界🌟"),
				highlightStyle.Render("世") + "\x1b[38;2;0;0;255m界🌟\x1b[m",
			},
		},
		{
			name:         "unicode toHighlight, continuation, overflow, ansi",
			s:            "\x1b[38;2;0;0;255m世界🌟世界🌟\x1b[m",
			width:        7,
			continuation: "...",
			toHighlight:  "世界🌟世",
			numPopLefts:  2,
			expected: []string{
				"\x1b[38;2;0;0;255m世界..\x1b[m", // does not highlight continuation, could in future
				"\x1b[38;2;0;0;255m..界🌟\x1b[m", // does not highlight continuation, could in future
			},
		},
		{
			name: "unicode combining",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), e+ ́ (1w, 1b+2b) = 6w, 11b
			s:            "A💖中e\u0301A💖中e\u0301", // 12w total
			width:        10,
			continuation: "",
			numPopLefts:  2,
			expected: []string{
				"A💖中e\u0301A💖",
				"中e\u0301",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.expected) != tt.numPopLefts {
				t.Fatalf("num expected != num popLefts")
			}
			lb := New(tt.s)
			for i := 0; i < tt.numPopLefts; i++ {
				actual := lb.PopLeft(tt.width, tt.continuation, tt.toHighlight, highlightStyle)
				util.CmpStr(t, tt.expected[i], actual)
			}
		})
	}
}

func TestLineBuffer_WrappedLines(t *testing.T) {
	tests := []struct {
		name            string
		s               string
		width           int
		maxLinesEachEnd int
		toHighlight     string
		highlightStyle  lipgloss.Style
		want            []string
	}{
		{
			name:            "empty string",
			s:               "",
			width:           10,
			maxLinesEachEnd: 2,
			want:            []string{""},
		},
		{
			name:            "single line within width",
			s:               "Hello",
			width:           10,
			maxLinesEachEnd: 2,
			want:            []string{"Hello"},
		},
		{
			name:            "zero width",
			s:               "Hello",
			width:           0,
			maxLinesEachEnd: 2,
			want:            []string{},
		},
		{
			name:            "zero maxLinesEachEnd",
			s:               "This is a very long line that needs wrapping",
			width:           10,
			maxLinesEachEnd: 0,
			want:            []string{"This is a ", "very long ", "line that ", "needs wrap", "ping"},
		},
		{
			name:            "negative maxLinesEachEnd",
			s:               "This is a very long line that needs wrapping",
			width:           10,
			maxLinesEachEnd: -1,
			want:            []string{"This is a ", "very long ", "line that ", "needs wrap", "ping"},
		},
		{
			name:            "limited by maxLinesEachEnd",
			s:               "This is a very long line that needs wrapping",
			width:           10,
			maxLinesEachEnd: 2,
			want: []string{
				"This is a ",
				"very long ",
				//"line that ",
				"needs wrap",
				"ping"},
		},
		{
			name:            "single chars",
			s:               strings.Repeat("Test \x1b[38;2;0;0;255mtest\x1b[m", 1),
			width:           1,
			maxLinesEachEnd: -1,
			want: []string{
				"T",
				"e",
				"s",
				"t",
				" ",
				"\x1b[38;2;0;0;255mt\x1b[m",
				"\x1b[38;2;0;0;255me\x1b[m",
				"\x1b[38;2;0;0;255ms\x1b[m",
				"\x1b[38;2;0;0;255mt\x1b[m",
			},
		},
		{
			name:            "long s with maxLinesEachEnd and space at end",
			s:               strings.Repeat("This \x1b[38;2;0;0;255mtest\x1b[m sentence. ", 200),
			width:           1,
			maxLinesEachEnd: 6,
			want: []string{
				"T",
				"h",
				"i",
				"s",
				" ",
				"\x1b[38;2;0;0;255mt\x1b[m",
				//"\x1b[38;2;0;0;255me\x1b[m",
				//"\x1b[38;2;0;0;255ms\x1b[m",
				//"\x1b[38;2;0;0;255mt\x1b[m",
				//" ",
				//"s",
				//"e",
				//"n",
				//"t",
				"e",
				"n",
				"c",
				"e",
				".",
				" ",
			},
		},
		{
			name:            "input with trailing spaces are not trimmed",
			s:               "Hello   ",
			width:           10,
			maxLinesEachEnd: 2,
			want:            []string{"Hello   "},
		},
		{
			name:            "input with only spaces is not trimmed",
			s:               "     ",
			width:           10,
			maxLinesEachEnd: 2,
			want:            []string{"     "},
		},
		{
			name:            "unicode characters",
			s:               "Hello 世界! This is a test with unicode characters 🌟",
			width:           10,
			maxLinesEachEnd: 2,
			want: []string{
				"Hello 世界",
				"! This is ",
				//"a test wit",
				//"h unicode ",
				"characters",
				" 🌟",
			},
		},
		{
			name:            "Width exactly matches s length",
			s:               "Hello World",
			width:           11,
			maxLinesEachEnd: 2,
			want:            []string{"Hello World"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lb := New(tt.s)
			got := lb.WrappedLines(tt.width, tt.maxLinesEachEnd, tt.toHighlight, tt.highlightStyle)
			if len(got) != len(tt.want) {
				t.Errorf("wrap() len = %d, want %d", len(got), len(tt.want))
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("wrap() line %d got %q, expected %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestLineBuffer_getLeftRuneIdx(t *testing.T) {
	tests := []struct {
		name     string
		w        int
		vals     []int
		expected int
	}{
		{
			name:     "empty",
			w:        0,
			vals:     []int{},
			expected: 0,
		},
		{
			name:     "step by 1",
			w:        2,
			vals:     []int{1, 2, 3},
			expected: 2,
		},
		{
			name:     "step by 2",
			w:        2,
			vals:     []int{1, 3, 5},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if actual := getLeftRuneIdx(tt.w, tt.vals); actual != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, actual)
			}
		})
	}
}

func TestLineBuffer_reapplyAnsi(t *testing.T) {
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
			actual := reapplyAnsi(tt.original, tt.truncated, tt.truncByteOffset, ansiCodeIndexes)
			util.CmpStr(t, tt.expected, actual)
		})
	}
}

func TestLineBuffer_highlightLine(t *testing.T) {
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

func TestLineBuffer_overflowsLeft(t *testing.T) {
	tests := []struct {
		name         string
		str          string
		startByteIdx int
		substr       string
		wantBool     bool
		wantInt      int
	}{
		{
			name:         "basic overflow case",
			str:          "my str here",
			startByteIdx: 3,
			substr:       "my str",
			wantBool:     true,
			wantInt:      6,
		},
		{
			name:         "no overflow case",
			str:          "my str here",
			startByteIdx: 6,
			substr:       "my str",
			wantBool:     false,
			wantInt:      0,
		},
		{
			name:         "empty string",
			str:          "",
			startByteIdx: 0,
			substr:       "test",
			wantBool:     false,
			wantInt:      0,
		},
		{
			name:         "empty substring",
			str:          "test string",
			startByteIdx: 0,
			substr:       "",
			wantBool:     false,
			wantInt:      0,
		},
		{
			name:         "startByteIdx out of bounds",
			str:          "test",
			startByteIdx: 10,
			substr:       "test",
			wantBool:     false,
			wantInt:      0,
		},
		{
			name:         "exact full match",
			str:          "hello world",
			startByteIdx: 0,
			substr:       "hello world",
			wantBool:     false,
			wantInt:      0,
		},
		{
			name:         "partial overflow at end",
			str:          "hello world",
			startByteIdx: 9,
			substr:       "dd",
			wantBool:     false,
			wantInt:      0,
		},
		{
			name:         "case sensitivity test - no match",
			str:          "Hello World",
			startByteIdx: 0,
			substr:       "hello",
			wantBool:     false,
			wantInt:      0,
		},
		{
			name:         "multiple character same overflow",
			str:          "aaaa",
			startByteIdx: 1,
			substr:       "aaa",
			wantBool:     true,
			wantInt:      3,
		},
		{
			name:         "multiple character same overflow but difference",
			str:          "aaaa",
			startByteIdx: 1,
			substr:       "baaa",
			wantBool:     false,
			wantInt:      0,
		},
		{
			name:         "special characters",
			str:          "test!@#$",
			startByteIdx: 4,
			substr:       "st!@#",
			wantBool:     true,
			wantInt:      7,
		},
		{
			name:         "false if does not overflow",
			str:          "some string",
			startByteIdx: 1,
			substr:       "ome",
			wantBool:     false,
			wantInt:      0,
		},
		{
			name:         "one char overflow",
			str:          "some string",
			startByteIdx: 1,
			substr:       "some",
			wantBool:     true,
			wantInt:      4,
		},
		// 世 is 3 bytes
		// 界 is 3 bytes
		// 🌟 is 4 bytes
		// "世界🌟世界🌟"[3:13] = "界🌟世"
		{
			name:         "unicode with ansi left not overflowing",
			str:          "世界🌟世界🌟",
			startByteIdx: 0,
			substr:       "世界🌟世",
			wantBool:     false,
			wantInt:      0,
		},
		{
			name:         "unicode with ansi left overflow 1 byte",
			str:          "世界🌟世界🌟",
			startByteIdx: 1,
			substr:       "世界🌟世",
			wantBool:     true,
			wantInt:      13,
		},
		{
			name:         "unicode with ansi left overflow 2 bytes",
			str:          "世界🌟世界🌟",
			startByteIdx: 2,
			substr:       "世界🌟世",
			wantBool:     true,
			wantInt:      13,
		},
		{
			name:         "unicode with ansi left overflow full rune",
			str:          "世界🌟世界🌟",
			startByteIdx: 3,
			substr:       "世界🌟世",
			wantBool:     true,
			wantInt:      13,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBool, gotInt := overflowsLeft(tt.str, tt.startByteIdx, tt.substr)
			if gotBool != tt.wantBool || gotInt != tt.wantInt {
				t.Errorf("overflowsLeft(%q, %d, %q) = (%v, %d), want (%v, %d)",
					tt.str, tt.startByteIdx, tt.substr, gotBool, gotInt, tt.wantBool, tt.wantInt)
			}
		})
	}
}

func TestLineBuffer_overflowsRight(t *testing.T) {
	tests := []struct {
		name       string
		s          string
		endByteIdx int
		substr     string
		wantBool   bool
		wantInt    int
	}{
		{
			name:       "example 1",
			s:          "my str here",
			endByteIdx: 3,
			substr:     "y str",
			wantBool:   true,
			wantInt:    1,
		},
		{
			name:       "example 2",
			s:          "my str here",
			endByteIdx: 3,
			substr:     "y strong",
			wantBool:   false,
			wantInt:    0,
		},
		{
			name:       "example 3",
			s:          "my str here",
			endByteIdx: 6,
			substr:     "tr here",
			wantBool:   true,
			wantInt:    4,
		},
		{
			name:       "empty string",
			s:          "",
			endByteIdx: 0,
			substr:     "test",
			wantBool:   false,
			wantInt:    0,
		},
		{
			name:       "empty substring",
			s:          "test string",
			endByteIdx: 0,
			substr:     "",
			wantBool:   false,
			wantInt:    0,
		},
		{
			name:       "end index out of bounds",
			s:          "test",
			endByteIdx: 10,
			substr:     "test",
			wantBool:   false,
			wantInt:    0,
		},
		{
			name:       "exact full match",
			s:          "hello world",
			endByteIdx: 11,
			substr:     "hello world",
			wantBool:   false,
			wantInt:    0,
		},
		{
			name:       "case sensitivity test - no match",
			s:          "Hello World",
			endByteIdx: 4,
			substr:     "hello",
			wantBool:   false,
			wantInt:    0,
		},
		{
			name:       "multiple character same overflow",
			s:          "aaaa",
			endByteIdx: 2,
			substr:     "aaa",
			wantBool:   true,
			wantInt:    0,
		},
		{
			name:       "multiple character same overflow but difference",
			s:          "aaaa",
			endByteIdx: 2,
			substr:     "aaab",
			wantBool:   false,
			wantInt:    0,
		},
		{
			name:       "false if does not overflow",
			s:          "some string",
			endByteIdx: 5,
			substr:     "ome ",
			wantBool:   false,
			wantInt:    0,
		},
		{
			name:       "one char overflow",
			s:          "some string",
			endByteIdx: 5,
			substr:     "ome s",
			wantBool:   true,
			wantInt:    1,
		},
		// 世 is 3 bytes
		// 界 is 3 bytes
		// 🌟 is 4 bytes
		// "世界🌟世界🌟"[3:10] = "界🌟"
		{
			name:       "unicode with ansi no overflow",
			s:          "世界🌟世界🌟",
			endByteIdx: 13,
			substr:     "界🌟世",
			wantBool:   false,
			wantInt:    0,
		},
		{
			name:       "unicode with ansi overflow right one byte",
			s:          "世界🌟世界🌟",
			endByteIdx: 12,
			substr:     "界🌟世",
			wantBool:   true,
			wantInt:    3,
		},
		{
			name:       "unicode with ansi overflow right two bytes",
			s:          "世界🌟世界🌟",
			endByteIdx: 11,
			substr:     "界🌟世",
			wantBool:   true,
			wantInt:    3,
		},
		{
			name:       "unicode with ansi overflow right full rune",
			s:          "世界🌟世界🌟",
			endByteIdx: 10,
			substr:     "界🌟世",
			wantBool:   true,
			wantInt:    3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBool, gotInt := overflowsRight(tt.s, tt.endByteIdx, tt.substr)
			if gotBool != tt.wantBool || gotInt != tt.wantInt {
				t.Errorf("overflowsRight(%q, %d, %q) = (%v, %d), want (%v, %d)",
					tt.s, tt.endByteIdx, tt.substr, gotBool, gotInt, tt.wantBool, tt.wantInt)
			}
		})
	}
}

func TestLineBuffer_replaceStartWithContinuation(t *testing.T) {
	tests := []struct {
		name         string
		s            string
		continuation string
		expected     string
	}{
		{
			name:         "empty",
			s:            "",
			continuation: "",
			expected:     "",
		},
		{
			name:         "empty continuation",
			s:            "my string",
			continuation: "",
			expected:     "my string",
		},
		{
			name:         "simple",
			s:            "my string",
			continuation: "...",
			expected:     "...string",
		},
		{
			name: "unicode",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), e+ ́ (1w, 1b+2b)
			s:            "A💖中e\u0301",
			continuation: "...",
			expected:     "...中e\u0301",
		},
		{
			name: "unicode leading combined",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), e+ ́ (1w, 1b+2b)
			s:            "e\u0301💖中",
			continuation: "...",
			expected:     "...中",
		},
		{
			name: "unicode combined",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), e+ ́ (1w, 1b+2b)
			s:            "💖e\u0301💖中",
			continuation: "...",
			expected:     "...💖中",
		},
		{
			name: "unicode width overlap",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), e+ ́ (1w, 1b+2b)
			s:            "中💖中e\u0301",
			continuation: "...",
			expected:     "..💖中e\u0301", // continuation shrinks by 1
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if r := replaceStartWithContinuation(tt.s, []rune(tt.continuation)); r != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, r)
			}
		})
	}
}

func TestLineBuffer_replaceEndWithContinuation(t *testing.T) {
	tests := []struct {
		name         string
		s            string
		continuation string
		expected     string
	}{
		{
			name:         "empty",
			s:            "",
			continuation: "",
			expected:     "",
		},
		{
			name:         "empty continuation",
			s:            "my string",
			continuation: "",
			expected:     "my string",
		},
		{
			name:         "simple",
			s:            "my string",
			continuation: "...",
			expected:     "my str...",
		},
		{
			name: "unicode",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), e+ ́ (1w, 1b+2b)
			s:            "A💖中e",
			continuation: "...",
			expected:     "A💖...",
		},
		{
			name: "unicode trailing combined",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), e+ ́ (1w, 1b+2b)
			s:            "A💖中e\u0301",
			continuation: "...",
			expected:     "A💖...",
		},
		{
			name: "unicode combined",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), e+ ́ (1w, 1b+2b)
			s:            "A💖e\u0301中",
			continuation: "...",
			expected:     "A💖...",
		},
		{
			name: "unicode width overlap",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), e+ ́ (1w, 1b+2b)
			s:            "💖中",
			continuation: "...",
			expected:     "💖..", // continuation shrinks by 1
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if r := replaceEndWithContinuation(tt.s, []rune(tt.continuation)); r != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, r)
			}
		})
	}
}
