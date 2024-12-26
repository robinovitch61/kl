package viewport

import (
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/kl/internal/util"
	"strings"
	"testing"
)

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
		toHighlight     string
		highlightStyle  lipgloss.Style
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
		{
			name:            "Long input with truncation",
			input:           strings.Repeat("This is a \x1b[38;2;0;0;255mtest\x1b[0m sentence. ", 200),
			width:           1,
			maxLinesEachEnd: 10,
			want: []string{
				"T",
				"h",
				"i",
				"s",
				" ",
				"i",
				"s",
				" ",
				"a",
				" ",
				"s",
				"e",
				"n",
				"t",
				"e",
				"n",
				"c",
				"e",
				".",
			},
		},
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
		{
			name:            "Unicode characters",
			input:           "Hello 世界! This is a test with unicode characters 🌟",
			width:           10,
			maxLinesEachEnd: 1,
			want:            []string{"Hello 世界", "racters 🌟"},
		},
		{
			name:            "Width exactly matches input length",
			input:           "Hello World",
			width:           11,
			maxLinesEachEnd: 2,
			want:            []string{"Hello World"},
		},
		// TODO LEO: add tests for highlight
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrap(tt.input, tt.width, tt.maxLinesEachEnd, tt.toHighlight, tt.highlightStyle)
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
