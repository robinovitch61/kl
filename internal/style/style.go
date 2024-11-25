package style

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/robinovitch61/kl/internal/dev"
	"strconv"
	"strings"
)

var (
	// these are a bit of an approximation in some terminals
	backgroundHex = termenv.ConvertToRGB(termenv.BackgroundColor()).Hex()
	foregroundHex = termenv.ConvertToRGB(termenv.ForegroundColor()).Hex()

	adjustment           = 30
	lighterForegroundHex = lightenColor(foregroundHex, adjustment)
	darkerForegroundHex  = lightenColor(foregroundHex, -adjustment)
)

var (
	Unset = lipgloss.NewStyle().Foreground(lipgloss.Color(foregroundHex)).Background(lipgloss.Color(backgroundHex))

	Alt = lipgloss.NewStyle().Foreground(
		lipgloss.AdaptiveColor{
			Light: lighterForegroundHex,
			Dark:  darkerForegroundHex,
		},
	)

	// need to specifically set the foreground color otherwise changes color on some terminals
	Bold = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(foregroundHex))

	Inverse = lipgloss.NewStyle().Reverse(true)

	// need to specifically set the foreground color otherwise changes color on some terminals
	BoldUnderline = lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color(foregroundHex))

	// don't use .Reverse(true) as that messes with Underline
	InverseUnderline = lipgloss.NewStyle().Foreground(lipgloss.Color(backgroundHex)).Background(lipgloss.Color(foregroundHex)).Underline(true)

	// a slightly lighter or darker background depending on the user's background color
	AltInverse = lipgloss.NewStyle().Foreground(lipgloss.Color(backgroundHex)).Background(
		lipgloss.AdaptiveColor{
			Light: lighterForegroundHex,
			Dark:  darkerForegroundHex,
		})

	Underline = lipgloss.NewStyle().Underline(true)

	Blue = lipgloss.NewStyle().Background(lipgloss.Color("6")).Foreground(lipgloss.Color("#000000"))

	lilac = lipgloss.Color("189")

	Lilac = lipgloss.NewStyle().Background(lilac).Foreground(lipgloss.Color("#000000"))

	Green = lipgloss.NewStyle().Background(lipgloss.Color("46")).Foreground(lipgloss.Color("#000000"))

	RightBorder = lipgloss.NewStyle().Border(lipgloss.ThickBorder(), false, true, false, false).BorderForeground(lilac)
)

func lightenColor(hex string, offset int) string {
	hex = strings.TrimPrefix(hex, "#")

	r, _ := strconv.ParseInt(hex[0:2], 16, 0)
	g, _ := strconv.ParseInt(hex[2:4], 16, 0)
	b, _ := strconv.ParseInt(hex[4:6], 16, 0)

	r = clamp(r+int64(offset), 0, 255)
	g = clamp(g+int64(offset), 0, 255)
	b = clamp(b+int64(offset), 0, 255)

	hexRes := fmt.Sprintf("#%02x%02x%02x", r, g, b)
	dev.Debug(fmt.Sprintf("lightenColor(%s, %d) = %s", hex, offset, hexRes))
	return hexRes
}

func clamp(value, min, max int64) int64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
