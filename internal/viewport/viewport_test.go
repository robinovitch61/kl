package viewport

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-cmp/cmp"
	"github.com/muesli/termenv"
	"io"
	"strings"
	"testing"
)

var (
	keyMap     = DefaultKeyMap()
	downKeyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(keyMap.Down.Keys()[0])}
	upKeyMsg   = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(keyMap.Up.Keys()[0])}
	red        = lipgloss.Color("#ff0000")
)

func renderer() *lipgloss.Renderer {
	r := lipgloss.NewRenderer(io.Discard)
	r.SetColorProfile(termenv.TrueColor)
	r.SetHasDarkBackground(true)
	return r
}

// pad pads the given lines to the given width and height.
// for example, pad(5, 4, []string{"a", "b", "c"}) will be padded to:
// "a    "
// "b    "
// "c    "
// "     "
// as a single string
func pad(width, height int, lines []string) string {
	var res []string
	for _, line := range lines {
		resLine := line
		numSpaces := width - lipgloss.Width(line)
		if numSpaces > 0 {
			resLine += strings.Repeat(" ", numSpaces)
		}
		res = append(res, resLine)
	}
	numEmptyLines := height - len(lines)
	for i := 0; i < numEmptyLines; i++ {
		res = append(res, strings.Repeat(" ", width))
	}
	return strings.Join(res, "\n")
}

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

func newViewport(width, height int) Model[RenderableString] {
	return New[RenderableString](width, height)
}

func compare(t *testing.T, expected, actual string) {
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}
}

// # HELPER FUNCTIONS THAT ARE TRICKY

func TestGetVisiblePartOfLine(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := getVisiblePartOfLine(tt.s, tt.xOffset, tt.width, tt.lineContinuationIndicator)
			if diff := cmp.Diff(tt.expected, actual); diff != "" {
				t.Errorf("Mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// # SELECTION DISABLED, WRAP OFF

func TestViewport_SelectionDisabled_WrapOff_Basic(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetContent([]RenderableString{
		{Content: "first line"},
		//{Content: renderer().NewStyle().Foreground(red).Render("second") + " line"},
		{Content: renderer().NewStyle().Foreground(red).Render("a really really long line")},
		//{Content: renderer().NewStyle().Foreground(red).Render("a") + " really really long line"},
	})
	expectedView := pad(w, h, []string{
		"first line",
		//"\x1b[38;2;255;0;0msecond\x1b[0m line",
		"\x1b[38;2;255;0;0ma really rea...\x1b[0m",
	})
	if diff := cmp.Diff(expectedView, vp.View()); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestViewport_SelectionDisabled_WrapOff_OverflowLine(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetContent([]RenderableString{
		{Content: "123456789012345"},
		{Content: "1234567890123456"},
	})
	expectedView := pad(w, h, []string{
		"123456789012345",
		"123456789012...",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionDisabled_WrapOff_OverflowHeight(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetContent([]RenderableString{
		{Content: "123456789012345"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
	})
	expectedView := pad(w, h, []string{
		"123456789012345",
		"123456789012...",
		"123456789012...",
		"123456789012...",
		"66% (4/6)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionDisabled_WrapOff_Scrolling(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
		{Content: "third"},
		{Content: "fourth"},
		{Content: "fifth"},
		{Content: "sixth"},
	})
	expectedView := pad(w, h, []string{
		"first",
		"second",
		"third",
		"fourth",
		"66% (4/6)",
	})
	compare(t, expectedView, vp.View())

	// scrolling up past top is no-op
	vp, _ = vp.Update(upKeyMsg)
	compare(t, expectedView, vp.View())

	// scrolling down by one
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(w, h, []string{
		"second",
		"third",
		"fourth",
		"fifth",
		"83% (5/6)",
	})
	compare(t, expectedView, vp.View())

	// scrolling down by one again
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(w, h, []string{
		"third",
		"fourth",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())

	// scrolling down past bottom is no-op
	vp, _ = vp.Update(downKeyMsg)
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionDisabled_WrapOff_Panning(t *testing.T) {
	t.Errorf("TODO")
}

// # SELECTION DISABLED, WRAP ON

func TestViewport_OverflowLineWrap(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "123456789012345"},
		{Content: "1234567890123456"},
	})
	expectedView := pad(w, h, []string{
		"123456789012345",
		"123456789012345",
		"6",
	})
	compare(t, expectedView, vp.View())
}

// TODO:
// adding lipgloss style to a word at the start of a line should not shorten the line's view
// transitioning between wrap/no wrap
// adding new content should preserve selected line
// test with a bunch of spaces at end of line(s)
