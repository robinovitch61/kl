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
	SelectedItem        lipgloss.Style
	Highlight           lipgloss.Style
	HighlightIfSelected lipgloss.Style
	Footer              lipgloss.Style

	// filterable viewport match styles
	MatchFocused           lipgloss.Style
	MatchFocusedIfSelected lipgloss.Style
	MatchUnfocused         lipgloss.Style

	// container name coloring — deterministic foreground styles assigned via hash.
	// nil or empty disables container coloring.
	ContainerColors []lipgloss.Style

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

// DefaultTheme returns a theme using only most visible ANSI colors and reverse video.
// See https://blog.xoria.org/terminal-colors/ and https://jvns.ca/blog/2024/10/01/terminal-colours/
func DefaultTheme() Theme {
	return Theme{
		SelectionPrefix: "",

		ContainerColors: []lipgloss.Style{
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
		SelectedItem:        lipgloss.NewStyle().Reverse(true),
		Highlight:           lipgloss.NewStyle().Reverse(true),
		HighlightIfSelected: lipgloss.NewStyle(),
		Footer:              lipgloss.NewStyle(),

		MatchFocused:           lipgloss.NewStyle().Reverse(true).Foreground(lipgloss.Cyan),
		MatchFocusedIfSelected: lipgloss.NewStyle(),
		MatchUnfocused:         lipgloss.NewStyle().Foreground(lipgloss.BrightRed),

		TopBar:              lipgloss.NewStyle().Reverse(true),
		TopBarAccent:        lipgloss.NewStyle().Foreground(lipgloss.Red),
		FilterPrefixFocused: lipgloss.NewStyle().Reverse(true),
		TimestampPrefix:     lipgloss.NewStyle().Foreground(lipgloss.Green),
		HelpKeyColumn:       lipgloss.NewStyle().Reverse(true),
		EntityPaneBorder:    lipgloss.NewStyle().Border(lipgloss.ThickBorder(), false, true, false, false),
		PromptSelected:      lipgloss.NewStyle().Reverse(true),
		Error:               lipgloss.NewStyle().Foreground(lipgloss.Red),
	}
}

// NoColorTheme returns a theme with all styles empty.
func NoColorTheme() Theme {
	s := lipgloss.NewStyle()
	return Theme{
		// provide visual differentiation without color
		SelectionPrefix: "> ",

		ContainerColors:     nil,
		SelectedItem:        s,
		Highlight:           s,
		HighlightIfSelected: s,
		Footer:              s,

		MatchFocused:           s,
		MatchFocusedIfSelected: s,
		MatchUnfocused:         s,

		TopBar:              s,
		TopBarAccent:        s,
		FilterPrefixFocused: s,
		TimestampPrefix:     s,
		HelpKeyColumn:       s,
		EntityPaneBorder:    lipgloss.NewStyle().Border(lipgloss.ThickBorder(), false, true, false, false),
		PromptSelected:      s,
		Error:               s,
	}
}

// PickTheme returns the appropriate theme based on NO_COLOR environment variable.
func PickTheme() Theme {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return NoColorTheme()
	}
	return DefaultTheme()
}
