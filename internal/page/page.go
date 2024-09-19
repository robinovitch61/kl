package page

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/robinovitch61/kl/internal/style"
)

type GenericPage interface {
	Update(msg tea.Msg) (GenericPage, tea.Cmd)
	View() string
	AllContent() []string
	HighjackingInput() bool
	WithDimensions(width, height int) GenericPage
	WithStyles(styles style.Styles) GenericPage
	Help(styles style.Styles) string
}

type Type int

const (
	EntitiesPageType Type = iota
	LogsPageType
	SingleLogPageType
)
