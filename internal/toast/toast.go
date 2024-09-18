package toast

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/kl/internal/dev"
	"sync"
)

var (
	lastID int
	idMtx  sync.Mutex
)

type Model struct {
	ID           int
	message      string
	Visible      bool
	messageStyle lipgloss.Style
}

func New(message string) Model {
	return Model{
		ID:           nextID(),
		message:      message,
		Visible:      true,
		messageStyle: lipgloss.NewStyle(),
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	dev.DebugMsg("Toast", msg)
	switch msg := msg.(type) {
	case TimeoutMsg:
		if msg.ID > 0 && msg.ID != m.ID {
			return m, nil
		}
		m.Visible = false
	}
	return m, nil
}

func (m Model) View() string {
	if m.Visible {
		return m.messageStyle.Render(m.message)
	}
	return ""
}

func (m Model) ViewHeight() int {
	return lipgloss.Height(m.View())
}

type TimeoutMsg struct {
	ID int
}

func nextID() int {
	idMtx.Lock()
	defer idMtx.Unlock()
	lastID++
	return lastID
}
