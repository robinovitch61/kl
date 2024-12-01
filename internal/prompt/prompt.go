package prompt

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/kl/internal/dev"
)

type Model struct {
	Visible           bool
	proceedIsSelected bool
	width, height     int
	text              []string
	selectedStyle     lipgloss.Style
}

func New(visible bool, width, height int, text []string, selectedStyle lipgloss.Style) Model {
	return Model{
		Visible:       visible,
		width:         width,
		height:        height,
		text:          text,
		selectedStyle: selectedStyle,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	dev.DebugUpdateMsg("Prompt", msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "down", "left", "right", "h", "j", "k", "l", "tab":
			m.proceedIsSelected = !m.proceedIsSelected
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.Visible {
		// intentionally padded
		cancel := " NO, CANCEL "
		proceed := " YES, PROCEED "
		if m.proceedIsSelected {
			proceed = m.selectedStyle.Render(proceed)
		} else {
			cancel = m.selectedStyle.Render(cancel)
		}
		view := lipgloss.JoinVertical(
			lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center, m.text...),
			"\n",
			lipgloss.JoinHorizontal(lipgloss.Center, cancel, "   ", proceed),
		)
		view = lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).Padding(1, 1).Render(view)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, view)
	}
	return ""
}

func (m Model) ProceedIsSelected() bool {
	return m.proceedIsSelected
}

func (m *Model) SetWidthAndHeight(width, height int) {
	m.width = width
	m.height = height
}
