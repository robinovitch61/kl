package style

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/robinovitch61/kl/internal/dev"
	"image/color"
	"math"
	"strconv"
	"strings"
)

var (
	output               = termenv.DefaultOutput()
	foregroundHex        = termenv.ConvertToRGB(output.ForegroundColor()).Hex()
	lighterForegroundHex = adjustColor(foregroundHex, 1.7)
	darkerForegroundHex  = adjustColor(foregroundHex, 0.1)
	backgroundHex        = termenv.ConvertToRGB(output.BackgroundColor()).Hex()
	lighterBackgroundHex = adjustColor(backgroundHex, 1.7)
	darkerBackgroundHex  = adjustColor(backgroundHex, 0.1)
	foreground           = lipgloss.Color(foregroundHex)
	altForeground        = lipgloss.AdaptiveColor{
		Light: lighterForegroundHex,
		Dark:  darkerForegroundHex,
	}
	background    = lipgloss.Color(backgroundHex)
	altBackground = lipgloss.AdaptiveColor{
		Light: lighterBackgroundHex,
		Dark:  darkerBackgroundHex,
	}
)

func DebugColors() {
	darkBg := termenv.HasDarkBackground()
	dev.Debug(fmt.Sprintf("has dark background: %t", darkBg))
	dev.Debug(fmt.Sprintf("foreground: %s", foregroundHex))
	dev.Debug(fmt.Sprintf("lighterForegroundHex: %s", lighterForegroundHex))
	dev.Debug(fmt.Sprintf("darkerForegroundHex: %s", darkerForegroundHex))
	dev.Debug(fmt.Sprintf("background: %s", backgroundHex))
	dev.Debug(fmt.Sprintf("lighterBackgroundHex: %s", lighterBackgroundHex))
	dev.Debug(fmt.Sprintf("darkerBackgroundHex: %s", darkerBackgroundHex))
	if darkBg {
		dev.Debug(fmt.Sprintf("altForeground: %s", altForeground.Dark))
	} else {
		dev.Debug(fmt.Sprintf("altForeground: %s", altForeground.Light))
	}
	if darkBg {
		dev.Debug(fmt.Sprintf("altBackground: %s", altBackground.Dark))
	} else {
		dev.Debug(fmt.Sprintf("altBackground: %s", altBackground.Light))
	}
}

var (
	Regular                          = lipgloss.NewStyle().Foreground(foreground).Background(background).BorderForeground(foreground).BorderBackground(background).ColorWhitespace(true)
	Bold                             = Regular.Copy().Bold(true)
	Inverse                          = Regular.Copy().Foreground(background).Background(foreground)
	ViewportBackgroundStyle          = Regular.Copy()
	ViewportHeaderStyle              = Bold.Copy()
	ViewportSelectedRowStyle         = Inverse.Copy()
	ViewportHighlightStyle           = Inverse.Copy().Background(altForeground)
	ViewportHighlightIfSelectedStyle = Regular.Copy().Background(altBackground)
	ContentStyle                     = Regular.Copy()
	ViewportFooterStyle              = Bold.Copy()
	ModalOptionStyle                 = Regular.Copy().Margin(0, 1).Padding(0, 1)
	ModalSelectedStyle               = Inverse.Copy().Margin(0, 1).Padding(0, 1)
	UnderlineStyle                   = Regular.Copy().Underline(true)
	FilterPrefix                     = Bold.Copy()
	FilterEditing                    = Inverse.Copy()
	FilterApplied                    = Inverse.Copy().Background(altForeground)
	FilterCursor                     = Inverse.Copy()
	KeyHelpStyle                     = Bold.Copy().Foreground(background).Background(foreground).Underline(true)
)

type hsl struct {
	H, S, L float64
}

func hexToRGBA(hex string) color.RGBA {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		panic(fmt.Sprintf("invalid hex color: %s", hex))
	}

	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)

	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
}

func rgbaToHex(c color.RGBA) string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

func rgbaToHSL(rgba color.RGBA) hsl {
	r := float64(rgba.R) / 255
	g := float64(rgba.G) / 255
	b := float64(rgba.B) / 255

	mx := math.Max(math.Max(r, g), b)
	mn := math.Min(math.Min(r, g), b)
	l := (mx + mn) / 2

	var h, s float64
	if mx == mn {
		h, s = 0, 0 // achromatic
	} else {
		d := mx - mn
		if l > 0.5 {
			s = d / (2 - mx - mn)
		} else {
			s = d / (mx + mn)
		}
		switch mx {
		case r:
			h = (g - b) / d
			if g < b {
				h += 6
			}
		case g:
			h = (b-r)/d + 2
		case b:
			h = (r-g)/d + 4
		}
		h /= 6
	}

	return hsl{h, s, l}
}

func hslToRGBA(hsl hsl) color.RGBA {
	var r, g, b float64

	if hsl.S == 0 {
		r, g, b = hsl.L, hsl.L, hsl.L
	} else {
		var q float64
		if hsl.L < 0.5 {
			q = hsl.L * (1 + hsl.S)
		} else {
			q = hsl.L + hsl.S - hsl.L*hsl.S
		}
		p := 2*hsl.L - q

		r = hueToRGB(p, q, hsl.H+1.0/3.0)
		g = hueToRGB(p, q, hsl.H)
		b = hueToRGB(p, q, hsl.H-1.0/3.0)
	}

	return color.RGBA{
		R: uint8(math.Round(r * 255)),
		G: uint8(math.Round(g * 255)),
		B: uint8(math.Round(b * 255)),
		A: 255,
	}
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

func adjustColor(hexColor string, factor float64) string {
	rgba := hexToRGBA(hexColor)
	hsl := rgbaToHSL(rgba)

	// lightness
	adjustmentFactor := (factor - 1) * 0.1
	hsl.L = math.Max(0, math.Min(1, hsl.L+adjustmentFactor))

	// saturation for non-grayscale colors
	if hsl.S > 0 {
		hsl.S = math.Max(0, math.Min(1, hsl.S*factor))
	}

	adjustedRGBA := hslToRGBA(hsl)
	res := rgbaToHex(adjustedRGBA)
	dev.Debug(fmt.Sprintf("adjustColor(%s, %f) = %s", hexColor, factor, res))
	return res
}
