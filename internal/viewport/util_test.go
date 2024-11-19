package viewport

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-cmp/cmp"
	"testing"
)

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

func TestViewport_StringWidth(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expectedWidth int
	}{
		{
			"no ansi",
			"No ANSI here, just plain text",
			29,
		},
		{
			"ansi foo",
			"\x1b[38;2;214;125;17mfoo\x1b[0m",
			len("foo"),
		},
		{
			"hello world",
			"\x1b[31mHello, World!\x1b[0m",
			len("Hello, World!"),
		},
		{
			"bold text",
			"\x1b[1mBold Text\x1b[0m",
			9,
		},
		{
			"only bold and reset codes, no text",
			"\x1b[1m\x1b[0m",
			0,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result := lipgloss.Width(tt.input)
			if result != tt.expectedWidth {
				t.Errorf("For input '%s', expected width %d, but got %d", tt.input, tt.expectedWidth, result)
			}
		})
	}
}
