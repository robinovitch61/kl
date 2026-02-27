package page

import (
	tea "charm.land/bubbletea/v2"
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
	WithTheme(style.Theme) GenericPage
	Help() string
}

type Type int

const (
	EntitiesPageType Type = iota
	LogsPageType
	SingleLogPageType
)

// viewportStylesForFocus returns the viewport styles for a page based on whether it is focused.
func viewportStylesForFocus(focused bool, theme style.Theme) viewport.Styles {
	if focused {
		return viewport.Styles{
			SelectionPrefix:          theme.SelectionPrefix,
			SelectedItemStyle:        theme.SelectedItem,
			HighlightStyle:           theme.Highlight,
			HighlightStyleIfSelected: theme.HighlightIfSelected,
		}
	}
	return viewport.Styles{
		FooterStyle:              theme.Footer,
		HighlightStyle:           theme.Highlight,
		HighlightStyleIfSelected: theme.Highlight,
	}
}
