package style

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/kl/internal/dev"
	"image/color"
)

type TermStyleData struct {
	ForegroundDetected bool
	Foreground         color.Color
	ForegroundIsDark   bool
	BackgroundDetected bool
	Background         color.Color
	BackgroundIsDark   bool
}

func NewTermStyleData() TermStyleData {
	return TermStyleData{}
}

func (tsd *TermStyleData) SetBackground(msg tea.BackgroundColorMsg) {
	tsd.BackgroundDetected = true
	tsd.Background = msg.Color
	tsd.BackgroundIsDark = msg.IsDark()
}

func (tsd *TermStyleData) SetForeground(msg tea.ForegroundColorMsg) {
	tsd.ForegroundDetected = true
	tsd.Foreground = msg.Color
	tsd.ForegroundIsDark = msg.IsDark()
}

func (tsd TermStyleData) IsComplete() bool {
	return tsd.ForegroundDetected && tsd.BackgroundDetected
}

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

func NewStyles(data TermStyleData) Styles {
	if !data.IsComplete() {
		panic(fmt.Errorf("NewStyles called with incomplete TermStyleData"))
	}
	adjustment := uint32(30)
	lighterForeground := lipgloss.Color(lightenColor(data.Foreground, adjustment))
	darkerForeground := lipgloss.Color(lightenColor(data.Foreground, -adjustment))
	dev.Debug(fmt.Sprintf("foreground: %v, lighterForeground: %v, darkerForeground: %v", data.Foreground, lighterForeground, darkerForeground))

	fgLightDark := lipgloss.LightDark(data.ForegroundIsDark)
	bgLightDark := lipgloss.LightDark(data.BackgroundIsDark)

	lilac := lipgloss.Color("189")

	return Styles{
		Unset: lipgloss.NewStyle().Foreground(data.Foreground).Background(data.Background),

		Alt: lipgloss.NewStyle().Foreground(fgLightDark(lighterForeground, darkerForeground)),

		// need to specifically set the foreground color otherwise changes color on some terminals
		Bold: lipgloss.NewStyle().Bold(true).Foreground(data.Foreground),

		Inverse: lipgloss.NewStyle().Foreground(data.Background).Background(data.Foreground),

		// need to specifically set the foreground color otherwise changes color on some terminals
		BoldUnderline: lipgloss.NewStyle().Bold(true).Underline(true).Foreground(data.Foreground),

		InverseUnderline: lipgloss.NewStyle().Foreground(data.Background).Background(data.Foreground).Underline(true),

		AltInverse: lipgloss.NewStyle().Foreground(data.Background).Background(bgLightDark(lighterForeground, darkerForeground)),

		Underline: lipgloss.NewStyle().Underline(true),

		Blue: lipgloss.NewStyle().Background(lipgloss.Color("6")).Foreground(lipgloss.Color("#000000")),

		Lilac: lipgloss.NewStyle().Background(lilac).Foreground(lipgloss.Color("#000000")),

		Green: lipgloss.NewStyle().Background(lipgloss.Color("46")).Foreground(lipgloss.Color("#000000")),

		RightBorder: lipgloss.NewStyle().Border(lipgloss.ThickBorder(), false, true, false, false).BorderForeground(lilac),
	}
}

func lightenColor(c color.Color, offset uint32) string {
	r, g, b, _ := c.RGBA()
	// Convert from 0-65535 range to 0-255
	r = r >> 8
	g = g >> 8
	b = b >> 8

	r = clamp(r+offset, 0, 255)
	g = clamp(g+offset, 0, 255)
	b = clamp(b+offset, 0, 255)

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func clamp(value, min, max uint32) uint32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
