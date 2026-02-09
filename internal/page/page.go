package page

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/robinovitch61/kl/internal/style"
	"github.com/robinovitch61/viewport/viewport"
)

type GenericPage interface {
	Update(msg tea.Msg) (GenericPage, tea.Cmd)
	View() string
	ContentForFile() []string
	ToggleShowContext() GenericPage
	HasAppliedFilter() bool
	HighjackingInput() bool
	WithDimensions(width, height int) GenericPage
	WithFocus() GenericPage
	WithBlur() GenericPage
	WithStyles(style.Styles) GenericPage
	Help() string
}

type Type int

const (
	EntitiesPageType Type = iota
	LogsPageType
	SingleLogPageType
)

// viewportStylesForFocus returns the viewport styles for a page based on whether it is focused.
func viewportStylesForFocus(focused bool, styles style.Styles) viewport.Styles {
	if focused {
		return viewport.Styles{
			SelectedItemStyle:        styles.Inverse,
			HighlightStyle:           styles.Inverse,
			HighlightStyleIfSelected: styles.Unset,
		}
	}
	return viewport.Styles{
		FooterStyle:              styles.Alt,
		HighlightStyle:           styles.Inverse,
		HighlightStyleIfSelected: styles.Inverse,
	}
}
