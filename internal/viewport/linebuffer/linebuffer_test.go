package linebuffer

import (
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/kl/internal/util"
	"strings"
	"testing"
)

func TestLineBuffer_Width(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		expected int
	}{
		{
			name:     "empty",
			s:        "",
			expected: 0,
		},
		{
			name:     "simple",
			s:        "1234567890",
			expected: 10,
		},
		{
			name:     "unicode",
			s:        "世界🌟世界a",
			expected: 11,
		},
		{
			name:     "ansi",
			s:        "\x1b[38;2;255;0;0mhi\x1b[m",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lb := New(tt.s)
			if actual := lb.Width(); actual != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, actual)
			}
		})
	}
}

func TestLineBuffer_Content(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		expected string
	}{
		{
			name:     "empty",
			s:        "",
			expected: "",
		},
		{
			name:     "simple",
			s:        "1234567890",
			expected: "1234567890",
		},
		{
			name:     "unicode",
			s:        "世界🌟世界",
			expected: "世界🌟世界",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lb := New(tt.s)
			if actual := lb.Content(); actual != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, actual)
			}
		})
	}
}

func TestLineBuffer_Take(t *testing.T) {
	highlightStyle := lipgloss.NewStyle().Background(lipgloss.Color("#FF0000"))
	tests := []struct {
		name         string
		s            string
		width        int
		continuation string
		toHighlight  string
		startWidth   int
		numTakes     int
		expected     []string
	}{
		{
			name:         "empty",
			s:            "",
			width:        10,
			continuation: "",
			startWidth:   0,
			numTakes:     1,
			expected:     []string{""},
		},
		{
			name:         "simple",
			s:            "1234567890",
			width:        10,
			continuation: "",
			startWidth:   0,
			numTakes:     1,
			expected:     []string{"1234567890"},
		},
		{
			name:         "negative startWidth",
			s:            "1234567890",
			width:        10,
			continuation: "",
			startWidth:   -1,
			numTakes:     1,
			expected:     []string{"1234567890"},
		},
		{
			name:         "seek",
			s:            "1234567890",
			width:        10,
			continuation: "",
			startWidth:   3,
			numTakes:     1,
			expected:     []string{"4567890"},
		},
		{
			name:         "seek to end",
			s:            "1234567890",
			width:        10,
			continuation: "",
			startWidth:   10,
			numTakes:     1,
			expected:     []string{""},
		},
		{
			name:         "seek past end",
			s:            "1234567890",
			width:        10,
			continuation: "",
			startWidth:   11,
			numTakes:     1,
			expected:     []string{""},
		},
		{
			name:         "continuation",
			s:            "1234567890",
			width:        7,
			continuation: "...",
			startWidth:   2,
			numTakes:     1,
			expected:     []string{"...6..."},
		},
		{
			name:         "continuation past end",
			s:            "1234567890",
			width:        10,
			continuation: "...",
			startWidth:   11,
			numTakes:     1,
			expected:     []string{""},
		},
		{
			name:         "unicode",
			s:            "世界🌟世界🌟",
			width:        10,
			continuation: "",
			startWidth:   0,
			numTakes:     1,
			expected:     []string{"世界🌟世界"},
		},
		{
			name:         "unicode seek past first rune",
			s:            "世界🌟世界🌟",
			width:        10,
			continuation: "",
			startWidth:   2,
			numTakes:     1,
			expected:     []string{"界🌟世界🌟"},
		},
		{
			name:         "unicode seek past first 2 runes",
			s:            "世界🌟世界🌟",
			width:        10,
			continuation: "",
			startWidth:   3,
			numTakes:     1,
			expected:     []string{"🌟世界🌟"},
		},
		{
			name:         "unicode seek past all but 1 rune",
			s:            "世界🌟世界🌟",
			width:        10,
			continuation: "",
			startWidth:   10,
			numTakes:     1,
			expected:     []string{"🌟"},
		},
		{
			name:         "unicode seek almost to end",
			s:            "世界🌟世界🌟",
			width:        10,
			continuation: "",
			startWidth:   11,
			numTakes:     1,
			expected:     []string{""},
		},
		{
			name:         "unicode seek to end",
			s:            "世界🌟世界🌟",
			width:        10,
			continuation: "",
			startWidth:   12,
			numTakes:     1,
			expected:     []string{""},
		},
		{
			name:         "unicode insufficient width",
			s:            "世界🌟世界🌟",
			width:        1,
			continuation: "",
			startWidth:   2,
			numTakes:     1,
			expected:     []string{""},
		},
		{
			name:         "no ansi, no continuation, no width",
			s:            "12345678901234",
			width:        0,
			continuation: "",
			numTakes:     3,
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
			numTakes:     3,
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
			numTakes:     3,
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
			numTakes:     3,
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
			numTakes:     4,
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
			numTakes:     4,
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
			numTakes:     4,
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
			numTakes:     4,
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
			numTakes:     3,
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
			numTakes:     3,
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
			numTakes:     3,
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
			numTakes:     3,
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
			numTakes:     4,
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
			numTakes:     4,
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
			numTakes:     4,
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
			numTakes:     4,
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
			numTakes:     3,
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
			numTakes:     3,
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
			numTakes:     4,
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
			numTakes:     3,
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
			numTakes:     3,
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
			numTakes:     4,
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
			numTakes:     1,
			expected:     []string{"hi"},
		},
		{
			name:         "lineContinuationIndicator longer than width",
			s:            "1234567890123456789012345",
			width:        1,
			continuation: "...",
			numTakes:     1,
			expected:     []string{"."},
		},
		{
			name:         "twice the lineContinuationIndicator longer than width",
			s:            "1234567",
			width:        5,
			continuation: "...",
			numTakes:     1,
			expected:     []string{"12..."},
		},
		{
			name:         "sufficient width",
			s:            "1234567890123456789012345",
			width:        30,
			continuation: "...",
			numTakes:     1,
			expected:     []string{"1234567890123456789012345"},
		},
		{
			name:         "sufficient width, space at end preserved",
			s:            "1234567890123456789012345     ",
			width:        30,
			continuation: "...",
			numTakes:     1,
			expected:     []string{"1234567890123456789012345     "},
		},
		{
			name:         "insufficient width",
			s:            "1234567890123456789012345",
			width:        15,
			continuation: "...",
			numTakes:     1,
			expected:     []string{"123456789012..."},
		},
		{
			name:         "insufficient width",
			s:            "123456789012345678901234567890123456789012345",
			width:        15,
			continuation: "...",
			numTakes:     3,
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
			numTakes:     2,
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
			numTakes:     3,
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
			numTakes:     2,
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
			numTakes:     3,
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
			numTakes:     1,
			expected: []string{
				"\x1b[38;2;0;0;255mhi\x1b[m",
			},
		},
		{
			name:         "multi-byte runes",
			s:            "├─flask",
			width:        6,
			continuation: "...",
			numTakes:     1,
			expected: []string{
				"├─f...",
			},
		},
		{
			name:         "multi-byte runes with ansi",
			s:            "\x1b[38;2;0;0;255m├─flask\x1b[m",
			width:        6,
			continuation: "...",
			numTakes:     1,
			expected: []string{
				"\x1b[38;2;0;0;255m├─f...\x1b[m",
			},
		},
		{
			name:         "width exceeds capacity",
			s:            "  │   └─[ ] local-path-provisioner (running for 11d)",
			width:        53,
			continuation: "",
			numTakes:     1,
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
			numTakes:     1,
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
			numTakes:     1,
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
			numTakes:     1,
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
			numTakes:     2,
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
			numTakes:     1,
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
			numTakes:     2,
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
			numTakes:     2,
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
			numTakes:     2,
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
			numTakes:     2,
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
			numTakes:     2,
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
			numTakes:     2,
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
			numTakes:     2,
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
			numTakes:     2,
			expected: []string{
				"A💖中e\u0301A💖",
				"中e\u0301",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.expected) != tt.numTakes {
				t.Fatalf("num expected != num popLefts")
			}
			lb := New(tt.s)
			startWidth := tt.startWidth
			for i := 0; i < tt.numTakes; i++ {
				actual, actualWidth := lb.Take(startWidth, tt.width, tt.continuation, tt.toHighlight, highlightStyle)
				util.CmpStr(t, tt.expected[i], actual)
				startWidth += actualWidth
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
