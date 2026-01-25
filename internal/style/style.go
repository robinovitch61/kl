package style

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	blue  = lipgloss.Color("6")
	lilac = lipgloss.Color("189")
	green = lipgloss.Color("46")
)

type Styles struct {
	Unset            lipgloss.Style
	Alt              lipgloss.Style
	Bold             lipgloss.Style
	Inverse          lipgloss.Style
	BoldUnderline    lipgloss.Style
	InverseUnderline lipgloss.Style
	AltInverse       lipgloss.Style
	Underline        lipgloss.Style
	Blue             lipgloss.Style
	Lilac            lipgloss.Style
	Green            lipgloss.Style
	RightBorder      lipgloss.Style
}

var DefaultStyles = Styles{
	Unset:            lipgloss.NewStyle(),
	Alt:              lipgloss.NewStyle().Foreground(lipgloss.Color("#cccccc")),
	Bold:             lipgloss.NewStyle().Bold(true),
	Inverse:          lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#ffffff")),
	BoldUnderline:    lipgloss.NewStyle().Bold(true).Underline(true),
	InverseUnderline: lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#ffffff")).Underline(true),
	AltInverse:       lipgloss.NewStyle().Foreground(lipgloss.Color("#141414")).Background(lipgloss.Color("#cbcbcb")),
	Underline:        lipgloss.NewStyle().Underline(true),
	Blue:             lipgloss.NewStyle().Background(blue).Foreground(lipgloss.Color("#000000")),
	Lilac:            lipgloss.NewStyle().Background(lilac).Foreground(lipgloss.Color("#000000")),
	Green:            lipgloss.NewStyle().Background(green).Foreground(lipgloss.Color("#000000")),
	RightBorder:      lipgloss.NewStyle().Border(lipgloss.ThickBorder(), false, true, false, false).BorderForeground(lilac),
}

func NewStyles() Styles {
	return DefaultStyles
}
