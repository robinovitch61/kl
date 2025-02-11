package linebuffer

import (
	"github.com/charmbracelet/lipgloss/v2"
	"testing"
)

var (
	redBg = lipgloss.NewStyle().Background(lipgloss.Color("#FF0000"))
)

var equivalentLineBuffers = map[string][]LineBufferer{
	// TODO LEO: add ansi, unicode
	"hello world": {
		New("hello world"),
		NewMulti(
			New("hello"),
			New(" world"),
		),
		NewMulti(
			New("hel"),
			New("lo "),
			New("wo"),
			New("rld"),
		),
		NewMulti(
			New("h"),
			New("e"),
			New("l"),
			New("l"),
			New("o"),
			New(" "),
			New("w"),
			New("o"),
			New("r"),
			New("l"),
			New("d"),
		),
	},
}

func TestMultiLineBuffer_Width(t *testing.T) {
	for _, eq := range equivalentLineBuffers {
		for i, lb := range eq {
			if lb.Width() != eq[0].Width() {
				t.Errorf("expected %d, got %d for line buffer %d", eq[0].Width(), lb.Width(), i)
			}
		}
	}
}

func TestMultiLineBuffer_Content(t *testing.T) {
	for _, eq := range equivalentLineBuffers {
		for i, lb := range eq {
			if lb.Content() != eq[0].Content() {
				t.Errorf("expected %q, got %q for line buffer %d", eq[0].Content(), lb.Content(), i)
			}
		}
	}
}

func TestMultiLineBuffer_Take(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		startWidth     int
		takeWidth      int
		continuation   string
		toHighlight    string
		highlightStyle lipgloss.Style
		expected       string
	}{
		{
			name:           "hello world 0",
			key:            "hello world",
			startWidth:     0,
			takeWidth:      7,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "hello w",
		},
		{
			name:           "hello world 1",
			key:            "hello world",
			startWidth:     1,
			takeWidth:      7,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "ello wo",
		},
		{
			name:           "hello world end",
			key:            "hello world",
			startWidth:     10,
			takeWidth:      3,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "d",
		},
		{
			name:           "hello world past end",
			key:            "hello world",
			startWidth:     11,
			takeWidth:      3,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "",
		},
		{
			name:           "hello world with continuation at end",
			key:            "hello world",
			startWidth:     0,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "hell...",
		},
		{
			name:           "hello world with continuation at start",
			key:            "hello world",
			startWidth:     4,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "...orld",
		},
		{
			name:           "hello world with continuation both ends",
			key:            "hello world",
			startWidth:     2,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "... ...",
		},
		{
			name:           "hello world with highlight whole word",
			key:            "hello world",
			startWidth:     0,
			takeWidth:      11,
			continuation:   "",
			toHighlight:    "hello",
			highlightStyle: redBg,
			expected:       redBg.Render("hello") + " world",
		},
		{
			name:           "hello world with highlight across buffer boundary",
			key:            "hello world",
			startWidth:     3,
			takeWidth:      6,
			continuation:   "",
			toHighlight:    "lo wo",
			highlightStyle: redBg,
			expected:       "lo" + redBg.Render("lo wo") + "rl",
		},
		{
			name:           "hello world with highlight and continuation",
			key:            "hello world",
			startWidth:     2,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "lo",
			highlightStyle: redBg,
			expected:       "..." + redBg.Render("lo") + "...",
		},
		// TODO LEO: highlight style, other keys
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, eq := range equivalentLineBuffers[tt.key] {
				actual, _ := eq.Take(tt.startWidth, tt.takeWidth, tt.continuation, "", lipgloss.NewStyle())
				if actual != tt.expected {
					t.Errorf("for %s, expected %q, got %q", eq.Repr(), tt.expected, actual)
				}
			}
		})
	}
}

func TestMultiLineBuffer_WrappedLines(t *testing.T) {
	tests := []struct {
		name            string
		key             string
		width           int
		maxLinesEachEnd int
		toHighlight     string
		highlightStyle  lipgloss.Style
		expected        []string
	}{
		{
			name:            "hello world full width",
			key:             "hello world",
			width:           11,
			maxLinesEachEnd: -1,
			toHighlight:     "",
			highlightStyle:  lipgloss.NewStyle(),
			expected:        []string{"hello world"},
		},
		{
			name:            "hello world width 5",
			key:             "hello world",
			width:           5,
			maxLinesEachEnd: -1,
			toHighlight:     "",
			highlightStyle:  lipgloss.NewStyle(),
			expected:        []string{"hello", " worl", "d"},
		},
		{
			name:            "hello world max 1 line each end",
			key:             "hello world",
			width:           5,
			maxLinesEachEnd: 1,
			toHighlight:     "",
			highlightStyle:  lipgloss.NewStyle(),
			expected:        []string{"hello", "d"},
		},
		{
			name:            "hello world width 0",
			key:             "hello world",
			width:           0,
			maxLinesEachEnd: -1,
			toHighlight:     "",
			highlightStyle:  lipgloss.NewStyle(),
			expected:        []string{},
		},
		// TODO LEO: highlight style, other keys
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, eq := range equivalentLineBuffers[tt.key] {
				actual := eq.WrappedLines(tt.width, tt.maxLinesEachEnd, tt.toHighlight, tt.highlightStyle)

				if len(actual) != len(tt.expected) {
					t.Errorf("for %s, expected %d lines, got %d lines", eq.Repr(), len(tt.expected), len(actual))
					continue
				}

				for i := range actual {
					if actual[i] != tt.expected[i] {
						t.Errorf("for %s, line %d: expected %q, got %q", eq.Repr(), i, tt.expected[i], actual[i])
					}
				}
			}
		})
	}
}
