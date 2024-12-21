package page

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/robinovitch61/kl/internal/style"
)

type GenericPage interface {
	Update(msg tea.Msg) (GenericPage, tea.Cmd)
	View() string
	ContentForFile() []string
	ToggleFilteringWithContext() GenericPage
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
