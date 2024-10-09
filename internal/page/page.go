package page

import (
	tea "github.com/charmbracelet/bubbletea"
)

type GenericPage interface {
	Update(msg tea.Msg) (GenericPage, tea.Cmd)
	View() string
	ContentToPersist() []string
	HighjackingInput() bool
	HasAppliedFilter() bool
	WithDimensions(width, height int) GenericPage
	Help() string
}

type Type int

const (
	EntitiesPageType Type = iota
	LogsPageType
	SingleLogPageType
)
