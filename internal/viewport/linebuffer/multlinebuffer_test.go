package linebuffer

import (
	"github.com/charmbracelet/lipgloss/v2"
	"testing"
)

func getEquivalentLineBuffers() map[string][]LineBufferer {
	return map[string][]LineBufferer{
		"hello world": {
			New("hello world"),
			NewMulti(New("hello world")),
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
		"ansi": {
			New(redBg.Render("hello") + " " + blueBg.Render("world")),
			NewMulti(New(redBg.Render("hello") + " " + blueBg.Render("world"))),
			NewMulti(
				New(redBg.Render("hello")+" "),
				New(blueBg.Render("world")),
			),
			NewMulti(
				New(redBg.Render("hello")),
				New(" "),
				New(blueBg.Render("world")),
			),
		},
		"unicode_ansi": {
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b) = 6w, 11b
			New(redBg.Render("A💖") + "中é"),
			NewMulti(New(redBg.Render("A💖") + "中é")),
			NewMulti(
				New(redBg.Render("A💖")),
				New("中"),
				New("é"),
			),
		}}
}

func TestMultiLineBuffer_Width(t *testing.T) {
	for _, eq := range getEquivalentLineBuffers() {
		for _, lb := range eq {
			if lb.Width() != eq[0].Width() {
				t.Errorf("expected %d, got %d for line buffer %s", eq[0].Width(), lb.Width(), lb.Repr())
			}
		}
	}
}

func TestMultiLineBuffer_Content(t *testing.T) {
	for _, eq := range getEquivalentLineBuffers() {
		for _, lb := range eq {
			if lb.Content() != eq[0].Content() {
				t.Errorf("expected %q, got %q for line buffer %s", eq[0].Content(), lb.Content(), lb.Repr())
			}
		}
	}
}

func TestMultiLineBuffer_Take(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		widthToLeft    int
		takeWidth      int
		continuation   string
		toHighlight    string
		highlightStyle lipgloss.Style
		expected       string
	}{
		{
			name:           "hello world start at 0",
			key:            "hello world",
			widthToLeft:    0,
			takeWidth:      7,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "hello w",
		},
		{
			name:           "hello world start at 1",
			key:            "hello world",
			widthToLeft:    1,
			takeWidth:      7,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "ello wo",
		},
		{
			name:           "hello world end",
			key:            "hello world",
			widthToLeft:    10,
			takeWidth:      3,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "d",
		},
		{
			name:           "hello world past end",
			key:            "hello world",
			widthToLeft:    11,
			takeWidth:      3,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "",
		},
		{
			name:           "hello world with continuation at end",
			key:            "hello world",
			widthToLeft:    0,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "hell...",
		},
		{
			name:           "hello world with continuation at start",
			key:            "hello world",
			widthToLeft:    4,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "...orld",
		},
		{
			name:           "hello world with continuation both ends",
			key:            "hello world",
			widthToLeft:    2,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "... ...",
		},
		{
			name:           "hello world with highlight whole word",
			key:            "hello world",
			widthToLeft:    0,
			takeWidth:      11,
			continuation:   "",
			toHighlight:    "hello",
			highlightStyle: redBg,
			expected:       redBg.Render("hello") + " world",
		},
		{
			name:           "hello world with highlight across buffer boundary",
			key:            "hello world",
			widthToLeft:    3,
			takeWidth:      6,
			continuation:   "",
			toHighlight:    "lo wo",
			highlightStyle: redBg,
			expected:       redBg.Render("lo wo") + "r",
		},
		{
			name:           "hello world with highlight and middle continuation",
			key:            "hello world",
			widthToLeft:    1,
			takeWidth:      7,
			continuation:   "..",
			toHighlight:    "lo ",
			highlightStyle: redBg,
			expected:       ".." + redBg.Render("lo ") + "..",
		},
		{
			name:           "hello world with highlight and overlapping continuation",
			key:            "hello world",
			widthToLeft:    1,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "lo ",
			highlightStyle: redBg,
			expected:       "...o...", // does not highlight continuation, could in future
		},
		{
			name:           "ansi start at 0",
			key:            "ansi",
			widthToLeft:    0,
			takeWidth:      7,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       redBg.Render("hello") + " " + blueBg.Render("w"),
		},
		{
			name:           "ansi start at 1",
			key:            "ansi",
			widthToLeft:    1,
			takeWidth:      7,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       redBg.Render("ello") + " " + blueBg.Render("wo"),
		},
		{
			name:           "ansi end",
			key:            "ansi",
			widthToLeft:    10,
			takeWidth:      3,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       blueBg.Render("d"),
		},
		{
			name:           "ansi past end",
			key:            "ansi",
			widthToLeft:    11,
			takeWidth:      3,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "",
		},
		{
			name:           "ansi with continuation at end",
			key:            "ansi",
			widthToLeft:    0,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       redBg.Render("hell.") + "." + blueBg.Render("."),
		},
		{
			name:           "ansi with continuation at start",
			key:            "ansi",
			widthToLeft:    4,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       redBg.Render(".") + "." + blueBg.Render(".orld"),
		},
		{
			name:           "ansi with continuation both ends",
			key:            "ansi",
			widthToLeft:    2,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       redBg.Render("...") + " " + blueBg.Render("..."),
		},
		{
			name:           "ansi with highlight whole word",
			key:            "ansi",
			widthToLeft:    0,
			takeWidth:      11,
			continuation:   "",
			toHighlight:    "hello",
			highlightStyle: greenBg,
			expected:       greenBg.Render("hello") + " " + blueBg.Render("world"),
		},
		{
			name:           "ansi with highlight partial word",
			key:            "ansi",
			widthToLeft:    0,
			takeWidth:      11,
			continuation:   "",
			toHighlight:    "ell",
			highlightStyle: greenBg,
			expected:       redBg.Render("h") + greenBg.Render("ell") + redBg.Render("o") + " " + blueBg.Render("world"),
		},
		{
			name:           "ansi with highlight across buffer boundary",
			key:            "ansi",
			widthToLeft:    0,
			takeWidth:      11,
			continuation:   "",
			toHighlight:    "lo wo",
			highlightStyle: greenBg,
			expected:       redBg.Render("hel") + greenBg.Render("lo wo") + blueBg.Render("rld"),
		},
		{
			name:           "ansi with highlight and middle continuation",
			key:            "ansi",
			widthToLeft:    1,
			takeWidth:      7,
			continuation:   "..",
			toHighlight:    "lo ",
			highlightStyle: greenBg,
			expected:       redBg.Render("..") + greenBg.Render("lo ") + blueBg.Render(".."),
		},
		{
			name:           "ansi with highlight and overlapping continuation",
			key:            "ansi",
			widthToLeft:    1,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "lo ",
			highlightStyle: redBg,
			expected:       redBg.Render("...o") + "." + blueBg.Render(".."), // does not highlight continuation, could in future
		},
		{
			name:           "unicode_ansi start at 0",
			key:            "unicode_ansi",
			widthToLeft:    0,
			takeWidth:      6,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       redBg.Render("A💖") + "中é",
		},
		{
			name:           "unicode_ansi start at 1",
			key:            "unicode_ansi",
			widthToLeft:    1,
			takeWidth:      5,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       redBg.Render("💖") + "中é",
		},
		{
			name:           "unicode_ansi end",
			key:            "unicode_ansi",
			widthToLeft:    5,
			takeWidth:      1,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "é",
		},
		{
			name:           "unicode_ansi past end",
			key:            "unicode_ansi",
			widthToLeft:    6,
			takeWidth:      3,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "",
		},
		{
			name:           "unicode_ansi with continuation at end",
			key:            "unicode_ansi",
			widthToLeft:    0,
			takeWidth:      5,
			continuation:   "...",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       redBg.Render("A💖") + "..", // bit of an edge cases, seems fine
		},
		{
			name:           "unicode_ansi with continuation at start",
			key:            "unicode_ansi",
			widthToLeft:    1,
			takeWidth:      5,
			continuation:   "...",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       redBg.Render("..") + "中é",
		},
		{
			name:           "unicode_ansi with highlight whole word",
			key:            "unicode_ansi",
			widthToLeft:    0,
			takeWidth:      6,
			continuation:   "",
			toHighlight:    "A💖",
			highlightStyle: greenBg,
			expected:       greenBg.Render("A💖") + "中é",
		},
		{
			name:           "unicode_ansi with highlight partial word",
			key:            "unicode_ansi",
			widthToLeft:    0,
			takeWidth:      6,
			continuation:   "",
			toHighlight:    "A",
			highlightStyle: greenBg,
			expected:       greenBg.Render("A") + redBg.Render("💖") + "中é",
		},
		{
			name:           "unicode_ansi with highlight across buffer boundary",
			key:            "unicode_ansi",
			widthToLeft:    0,
			takeWidth:      6,
			continuation:   "",
			toHighlight:    "💖中",
			highlightStyle: greenBg,
			expected:       redBg.Render("A") + greenBg.Render("💖中") + "é",
		},
		{
			name:           "unicode_ansi with highlight and overlapping continuation",
			key:            "unicode_ansi",
			widthToLeft:    1,
			takeWidth:      5,
			continuation:   "..",
			toHighlight:    "💖",
			highlightStyle: greenBg,
			expected:       redBg.Render("..") + "中é", // does not highlight continuation, could in future
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, eq := range getEquivalentLineBuffers()[tt.key] {
				actual, _ := eq.Take(tt.widthToLeft, tt.takeWidth, tt.continuation, tt.toHighlight, tt.highlightStyle)
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
		{
			name:            "hello world highlight",
			key:             "hello world",
			width:           5,
			maxLinesEachEnd: -1,
			toHighlight:     "lo",
			highlightStyle:  redBg,
			expected: []string{
				"hel" + redBg.Render("lo"),
				" worl",
				"d",
			},
		},
		{
			name:            "hello world highlight wrap",
			key:             "hello world",
			width:           4,
			maxLinesEachEnd: -1,
			toHighlight:     "lo",
			highlightStyle:  redBg,
			expected: []string{
				"hel" + redBg.Render("l"),
				redBg.Render("o") + " wo",
				"rld",
			},
		},
		{
			name:            "ansi full width",
			key:             "ansi",
			width:           11,
			maxLinesEachEnd: -1,
			toHighlight:     "",
			highlightStyle:  lipgloss.NewStyle(),
			expected:        []string{redBg.Render("hello") + " " + blueBg.Render("world")},
		},
		{
			name:            "ansi width 5",
			key:             "ansi",
			width:           5,
			maxLinesEachEnd: -1,
			toHighlight:     "",
			highlightStyle:  lipgloss.NewStyle(),
			expected: []string{
				redBg.Render("hello"),
				" " + blueBg.Render("worl"),
				blueBg.Render("d"),
			},
		},
		{
			name:            "ansi max 1 line each end",
			key:             "ansi",
			width:           5,
			maxLinesEachEnd: 1,
			toHighlight:     "",
			highlightStyle:  lipgloss.NewStyle(),
			expected: []string{
				redBg.Render("hello"),
				blueBg.Render("d"),
			},
		},
		{
			name:            "ansi width 0",
			key:             "ansi",
			width:           0,
			maxLinesEachEnd: -1,
			toHighlight:     "",
			highlightStyle:  lipgloss.NewStyle(),
			expected:        []string{},
		},
		{
			name:            "ansi highlight",
			key:             "ansi",
			width:           5,
			maxLinesEachEnd: -1,
			toHighlight:     "lo",
			highlightStyle:  greenBg,
			expected: []string{
				redBg.Render("hel") + greenBg.Render("lo"),
				" " + blueBg.Render("worl"),
				blueBg.Render("d"),
			},
		},
		{
			name:            "ansi highlight wrap",
			key:             "ansi",
			width:           4,
			maxLinesEachEnd: -1,
			toHighlight:     "lo",
			highlightStyle:  greenBg,
			expected: []string{
				redBg.Render("hel") + greenBg.Render("l"),
				greenBg.Render("o") + " " + blueBg.Render("wo"),
				blueBg.Render("rld"),
			},
		},
		{
			name:            "unicode_ansi full width",
			key:             "unicode_ansi",
			width:           6,
			maxLinesEachEnd: -1,
			toHighlight:     "",
			highlightStyle:  lipgloss.NewStyle(),
			expected:        []string{redBg.Render("A💖") + "中é"},
		},
		{
			name:            "unicode_ansi width 5",
			key:             "unicode_ansi",
			width:           5,
			maxLinesEachEnd: -1,
			toHighlight:     "",
			highlightStyle:  lipgloss.NewStyle(),
			expected: []string{
				redBg.Render("A💖") + "中",
				"é",
			},
		},
		{
			name:            "unicode_ansi max 1 line each end",
			key:             "unicode_ansi",
			width:           2,
			maxLinesEachEnd: 1,
			toHighlight:     "",
			highlightStyle:  lipgloss.NewStyle(),
			expected: []string{
				redBg.Render("A"),
				//redBg.Render("💖"),
				//"中",
				"é",
			},
		},
		{
			name:            "unicode_ansi width 0",
			key:             "unicode_ansi",
			width:           0,
			maxLinesEachEnd: -1,
			toHighlight:     "",
			highlightStyle:  lipgloss.NewStyle(),
			expected:        []string{},
		},
		{
			name:            "unicode_ansi highlight",
			key:             "unicode_ansi",
			width:           3,
			maxLinesEachEnd: -1,
			toHighlight:     "💖",
			highlightStyle:  greenBg,
			expected: []string{
				redBg.Render("A") + greenBg.Render("💖"),
				"中é",
			},
		},
		{
			name:            "unicode_ansi highlight wrap",
			key:             "unicode_ansi",
			width:           3,
			maxLinesEachEnd: -1,
			toHighlight:     "💖中",
			highlightStyle:  greenBg,
			expected: []string{
				redBg.Render("A") + greenBg.Render("💖"),
				greenBg.Render("中") + "é",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, eq := range getEquivalentLineBuffers()[tt.key] {
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
