package style

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
)

// Theme contains every visual choice for the application.
// All styling must come from here — no ad-hoc styles anywhere else.
type Theme struct {
	// visual indicators
	SelectionPrefix string

	// viewport styles
	SelectedItem lipgloss.Style
	Footer       lipgloss.Style

	// filterable viewport match styles
	MatchFocused           lipgloss.Style
	MatchFocusedIfSelected lipgloss.Style
	MatchUnfocused         lipgloss.Style

	// container name coloring — deterministic foreground styles assigned via hash.
	// nil or empty disables container coloring.
	ContainerColors []lipgloss.Style

	// JSON syntax colorization
	JSONKey    lipgloss.Style
	JSONString lipgloss.Style
	JSONNumber lipgloss.Style
	JSONBool   lipgloss.Style
	JSONNull   lipgloss.Style

	// kl-specific styles
	TopBar              lipgloss.Style
	TopBarAccent        lipgloss.Style // e.g. [PAUSED]
	FilterPrefixFocused lipgloss.Style
	TimestampPrefix     lipgloss.Style
	HelpKeyColumn       lipgloss.Style
	EntityPaneBorder    lipgloss.Style
	PromptSelected      lipgloss.Style
	Error               lipgloss.Style
}

// ContainerColorStyle returns a foreground style for the given name,
// deterministically assigned via hash. Returns an empty style if the theme
// has no container colors.
func (t Theme) ContainerColorStyle(name string) lipgloss.Style {
	if len(t.ContainerColors) == 0 {
		return lipgloss.NewStyle()
	}
	hash := md5.Sum([]byte(name))
	hashStr := hex.EncodeToString(hash[:])
	var hashValue int64
	_, err := fmt.Sscanf(hashStr[:8], "%x", &hashValue)
	if err != nil {
		return t.ContainerColors[0]
	}
	return t.ContainerColors[hashValue%int64(len(t.ContainerColors))]
}

// DefaultTheme returns the default theme using only the most accessible ANSI colors and reverse video.
// Maximum compatibility across terminal themes (light, dark, Solarized, etc.).
// See https://blog.xoria.org/terminal-colors/ and https://jvns.ca/blog/2024/10/01/terminal-colours/
func DefaultTheme() Theme {
	return Theme{
		SelectionPrefix: "▍",

		ContainerColors: []lipgloss.Style{
			// these are the most accessible colors as per https://blog.xoria.org/terminal-colors/
			lipgloss.NewStyle().Foreground(lipgloss.Red),
			lipgloss.NewStyle().Foreground(lipgloss.Green),
			lipgloss.NewStyle().Foreground(lipgloss.Yellow),
			lipgloss.NewStyle().Foreground(lipgloss.Magenta),
			lipgloss.NewStyle().Foreground(lipgloss.Cyan),
			lipgloss.NewStyle().Foreground(lipgloss.BrightRed),
			lipgloss.NewStyle().Foreground(lipgloss.BrightGreen),
			lipgloss.NewStyle().Foreground(lipgloss.BrightMagenta),
			lipgloss.NewStyle().Foreground(lipgloss.BrightCyan),
		},
		SelectedItem: lipgloss.NewStyle(), // rely on SelectionPrefix
		Footer:       lipgloss.NewStyle(),

		MatchFocused:           lipgloss.NewStyle().Reverse(true).Foreground(lipgloss.Cyan),
		MatchFocusedIfSelected: lipgloss.NewStyle().Reverse(true).Foreground(lipgloss.Cyan),
		MatchUnfocused:         lipgloss.NewStyle().Reverse(true).Foreground(lipgloss.BrightRed),

		JSONKey:    lipgloss.NewStyle().Foreground(lipgloss.Cyan),
		JSONString: lipgloss.NewStyle(), //.Foreground(lipgloss.Green),
		JSONNumber: lipgloss.NewStyle().Foreground(lipgloss.Yellow),
		JSONBool:   lipgloss.NewStyle().Foreground(lipgloss.Magenta),
		JSONNull:   lipgloss.NewStyle().Foreground(lipgloss.Red),

		TopBar:              lipgloss.NewStyle().Foreground(lipgloss.Cyan).Reverse(true),
		TopBarAccent:        lipgloss.NewStyle().Foreground(lipgloss.Red),
		FilterPrefixFocused: lipgloss.NewStyle().Foreground(lipgloss.Yellow),
		TimestampPrefix:     lipgloss.NewStyle().Foreground(lipgloss.Green),
		HelpKeyColumn:       lipgloss.NewStyle().Reverse(true),
		EntityPaneBorder:    lipgloss.NewStyle().Border(lipgloss.ThickBorder(), false, true, false, false),
		PromptSelected:      lipgloss.NewStyle().Reverse(true),
		Error:               lipgloss.NewStyle().Foreground(lipgloss.Red),
	}
}

// VividTheme uses 256-color and true-color values for maximum visual richness.
func VividTheme() Theme {
	lilac := lipgloss.Color("189")
	return Theme{
		SelectionPrefix: "▍",

		ContainerColors: []lipgloss.Style{
			lipgloss.NewStyle().Background(lipgloss.Color("#58A2EE")).Foreground(lipgloss.Color("#000000")),
			lipgloss.NewStyle().Background(lipgloss.Color("#3FE34B")).Foreground(lipgloss.Color("#000000")),
			lipgloss.NewStyle().Background(lipgloss.Color("#7c60d7")).Foreground(lipgloss.Color("#000000")),
			lipgloss.NewStyle().Background(lipgloss.Color("#FD2C4C")).Foreground(lipgloss.Color("#000000")),
			lipgloss.NewStyle().Background(lipgloss.Color("#FE7A00")).Foreground(lipgloss.Color("#000000")),
			lipgloss.NewStyle().Background(lipgloss.Color("#FAF81C")).Foreground(lipgloss.Color("#000000")),
			lipgloss.NewStyle().Background(lipgloss.Color("#56EBD3")).Foreground(lipgloss.Color("#000000")),
			lipgloss.NewStyle().Background(lipgloss.Color("#42952E")).Foreground(lipgloss.Color("#000000")),
			lipgloss.NewStyle().Background(lipgloss.Color("#FFACE6")).Foreground(lipgloss.Color("#000000")),
			lipgloss.NewStyle().Background(lipgloss.Color("#FE16F4")).Foreground(lipgloss.Color("#000000")),
			lipgloss.NewStyle().Background(lipgloss.Color("#D6A112")).Foreground(lipgloss.Color("#000000")),
			lipgloss.NewStyle().Background(lipgloss.Color("#FFDAB9")).Foreground(lipgloss.Color("#000000")),
			lipgloss.NewStyle().Background(lipgloss.Color("#FF7E6A")).Foreground(lipgloss.Color("#000000")),
		},
		SelectedItem: lipgloss.NewStyle(), // rely on SelectionPrefix
		Footer:       lipgloss.NewStyle(),

		MatchFocused:           lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#ffffff")),
		MatchFocusedIfSelected: lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#ffffff")),
		MatchUnfocused:         lipgloss.NewStyle().Foreground(lipgloss.Color("#141414")).Background(lipgloss.Color("#cbcbcb")),

		JSONKey:    lipgloss.NewStyle().Foreground(lipgloss.Color("#88C0D0")),
		JSONString: lipgloss.NewStyle().Foreground(lipgloss.Color("#A3BE8C")),
		JSONNumber: lipgloss.NewStyle().Foreground(lipgloss.Color("#B48EAD")),
		JSONBool:   lipgloss.NewStyle().Foreground(lipgloss.Color("#D08770")),
		JSONNull:   lipgloss.NewStyle().Foreground(lipgloss.Color("#BF616A")),

		TopBar:              lipgloss.NewStyle().Background(lilac).Foreground(lipgloss.Color("#000000")),
		TopBarAccent:        lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#ffffff")),
		FilterPrefixFocused: lipgloss.NewStyle().Background(lipgloss.Color("6")).Foreground(lipgloss.Color("#000000")),
		TimestampPrefix:     lipgloss.NewStyle().Background(lipgloss.Color("46")).Foreground(lipgloss.Color("#000000")),
		HelpKeyColumn:       lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#ffffff")),
		EntityPaneBorder:    lipgloss.NewStyle().Border(lipgloss.ThickBorder(), false, true, false, false).BorderForeground(lilac),
		PromptSelected:      lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#ffffff")),
		Error:               lipgloss.NewStyle().Foreground(lipgloss.Red),
	}
}

// NoColorTheme returns a theme with all styles empty.
func NoColorTheme() Theme {
	noStyle := lipgloss.NewStyle()
	return Theme{
		// provide visual differentiation without color
		SelectionPrefix: "▍",

		ContainerColors: nil,
		SelectedItem:    noStyle,
		Footer:          noStyle,

		MatchFocused:           noStyle,
		MatchFocusedIfSelected: noStyle,
		MatchUnfocused:         noStyle,

		JSONKey:    noStyle,
		JSONString: noStyle,
		JSONNumber: noStyle,
		JSONBool:   noStyle,
		JSONNull:   noStyle,

		TopBar:              noStyle,
		TopBarAccent:        noStyle,
		FilterPrefixFocused: noStyle,
		TimestampPrefix:     noStyle,
		HelpKeyColumn:       noStyle,
		EntityPaneBorder:    lipgloss.NewStyle().Border(lipgloss.ThickBorder(), false, true, false, false),
		PromptSelected:      noStyle,
		Error:               noStyle,
	}
}

// PickTheme returns the appropriate theme based on the theme name and NO_COLOR env var.
// Valid theme names: "" (default/ansi), "vivid", "none".
func PickTheme(themeName string) Theme {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return NoColorTheme()
	}
	switch themeName {
	case "vivid":
		return VividTheme()
	case "none":
		return NoColorTheme()
	default:
		return DefaultTheme()
	}
}
