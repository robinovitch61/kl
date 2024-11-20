package page

import (
	tea "github.com/charmbracelet/bubbletea"
)

type GenericPage interface {
	Update(msg tea.Msg) (GenericPage, tea.Cmd)
	View() string
	ContentForFile() []string
	HighjackingInput() bool
	WithDimensions(width, height int) GenericPage
	WithFocus() GenericPage
	WithBlur() GenericPage
	Help() string
}

type Type int

const (
	EntitiesPageType Type = iota
	LogsPageType
	SingleLogPageType
)
