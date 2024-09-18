package page

import (
	tea "github.com/charmbracelet/bubbletea"
)

type GenericPage interface {
	Update(msg tea.Msg) (GenericPage, tea.Cmd)
	View() string
	AllContent() []string
	HighjackingInput() bool
	WithDimensions(width, height int) GenericPage
	Help() string
}

type Type int

const (
	EntitiesPageType Type = iota
	LogsPageType
	SingleLogPageType
)
