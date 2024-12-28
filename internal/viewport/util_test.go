package viewport

import (
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/kl/internal/util"
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
