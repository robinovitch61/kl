package internal

import (
	"context"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/keymap"
)

type CleanupCompleteMsg struct{}

type Model struct {
	config Config
	keyMap keymap.KeyMap
	cancel context.CancelFunc
}

func InitialModel(c Config) Model {
	return Model{
		config: c,
		keyMap: keymap.DefaultKeyMap(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	dev.DebugUpdateMsg("App", msg)
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keyMap.Quit) {
			return m, m.cleanupCmd()
		}

	case CleanupCompleteMsg:
		return m, tea.Quit
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return "TODO"
}

func (m Model) cleanupCmd() tea.Cmd {
	return func() tea.Msg {
		if m.cancel != nil {
			m.cancel()
		}
		return CleanupCompleteMsg{}
	}
}
