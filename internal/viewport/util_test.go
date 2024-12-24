package viewport

import (
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/kl/internal/util"
	"strings"
	"testing"
)

func TestHighlightLine(t *testing.T) {
	for _, tc := range []struct {
		name           string
		line           string
		highlight      string
		highlightStyle lipgloss.Style
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
			expected:       "\x1b[38;2;0;0;255mfirst\x1b[m\x1b[38;2;255;0;0m line\x1b[m",
		},
		{
			name:           "highlight already partially styled line",
			line:           "hi a \x1b[38;2;255;0;0mstyled line\x1b[m cool \x1b[38;2;255;0;0mand styled\x1b[m more",
			highlight:      "style",
			highlightStyle: lipgloss.NewStyle().Foreground(blue),
			expected:       "hi a \x1b[38;2;0;0;255mstyle\x1b[m\x1b[38;2;255;0;0md line\x1b[m cool \x1b[38;2;255;0;0mand \x1b[m\x1b[38;2;0;0;255mstyle\x1b[m\x1b[38;2;255;0;0md\x1b[m more",
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
	} {
		t.Run(tc.name, func(t *testing.T) {
			util.CmpStr(t, tc.expected, highlightLine(tc.line, tc.highlight, tc.highlightStyle))
		})
	}
}

func TestPad(t *testing.T) {
	width, height := 5, 4
	lines := []string{"a", "b", "c"}
	expected := `a    
b    
c    
     `
	util.CmpStr(t, expected, pad(width, height, lines))
}

func TestPad_OverflowWidth(t *testing.T) {
	width, height := 5, 4
	lines := []string{"123456", "b", "c"}
	expected := `123456
b    
c    
     `
	util.CmpStr(t, expected, pad(width, height, lines))
}

func TestPad_Ansi(t *testing.T) {
	width, height := 5, 4
	lines := []string{lipgloss.NewStyle().Foreground(red).Render("a"), "b", "c"}
	expected := "\x1b[38;2;255;0;0ma\x1b[m    \nb    \nc    \n     "
	util.CmpStr(t, expected, pad(width, height, lines))
}

func TestWrap(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		width           int
		maxLinesEachEnd int
		want            []string
	}{
		{
			name:            "Empty string",
			input:           "",
			width:           10,
			maxLinesEachEnd: 2,
			want:            []string{""},
		},
		{
			name:            "Single line within width",
			input:           "Hello",
			width:           10,
			maxLinesEachEnd: 2,
			want:            []string{"Hello"},
		},
		{
			name:            "Zero width",
			input:           "Hello",
			width:           0,
			maxLinesEachEnd: 2,
			want:            []string{},
		},
		{
			name:            "Zero maxLinesEachEnd",
			input:           "This is a very long line that needs wrapping",
			width:           10,
			maxLinesEachEnd: 0,
			want:            []string{"This is a ", "very long ", "line that ", "needs wrap", "ping"},
		},
		{
			name:            "Negative maxLinesEachEnd",
			input:           "This is a very long line that needs wrapping",
			width:           10,
			maxLinesEachEnd: -1,
			want:            []string{"This is a ", "very long ", "line that ", "needs wrap", "ping"},
		},
		{
			name:            "Limited by maxLinesEachEnd",
			input:           "This is a very long line that needs wrapping",
			width:           10,
			maxLinesEachEnd: 2,
			want:            []string{"This is a ", "very long ", " that need", "s wrapping"},
		},
		{
			name:            "Single chars",
			input:           strings.Repeat("Test \x1b[38;2;0;0;255mtest\x1b[m", 1),
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
		// TODO LEO: complete
		//{
		//	name:            "Long input with truncation",
		//	input:           strings.Repeat("This is a \x1b[38;2;0;0;255mtest\x1b[0m sentence. ", 200),
		//	width:           1,
		//	maxLinesEachEnd: -1,
		//	want:            []string{},
		//},
		{
			name:            "Input with trailing spaces are trimmed",
			input:           "Hello   ",
			width:           10,
			maxLinesEachEnd: 2,
			want:            []string{"Hello"},
		},
		{
			name:            "Input with only spaces is not trimmed",
			input:           "     ",
			width:           10,
			maxLinesEachEnd: 2,
			want:            []string{"     "},
		},
		// TODO LEO: fix this
		//{
		//	name:            "Unicode characters",
		//	input:           "Hello 世界! This is a test with unicode characters 🌟",
		//	width:           10,
		//	maxLinesEachEnd: 3,
		//	want:            []string{"Hello 世界! ", "This is a ", "test with ", "unicode ch", "aracters 🌟"},
		//},
		{
			name:            "Width exactly matches input length",
			input:           "Hello World",
			width:           11,
			maxLinesEachEnd: 2,
			want:            []string{"Hello World"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrap(tt.input, tt.width, tt.maxLinesEachEnd)
			if len(got) != len(tt.want) {
				t.Errorf("wrap() len = %d, want %d", len(got), len(tt.want))
			}

			for i := range got {
				if i < len(tt.want) {
					if got[i] != tt.want[i] {
						t.Errorf("wrap() line %d got %q, expected %q", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}
