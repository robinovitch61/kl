package view

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/bubbleo/viewport"
	"github.com/robinovitch61/bubbleo/viewport/item"
	"github.com/robinovitch61/kl/internal/domain"
)

// SingleLogLine wraps a line for single log display
type SingleLogLine struct {
	content string
}

// NewSingleLogLine creates a new single log line
func NewSingleLogLine(content string) SingleLogLine {
	return SingleLogLine{content: content}
}

// GetItem implements viewport.Object
func (l SingleLogLine) GetItem() item.Item {
	return item.NewItem(l.content)
}

// SingleLogView displays a single expanded log entry
type SingleLogView struct {
	viewport *viewport.Model[SingleLogLine]
	log      *domain.Log
	width    int
	height   int
}

// NewSingleLogView creates a new single log view
func NewSingleLogView(width, height int) SingleLogView {
	vp := viewport.New[SingleLogLine](width, height)
	return SingleLogView{
		viewport: vp,
		width:    width,
		height:   height,
	}
}

// Update handles messages
func (v SingleLogView) Update(msg tea.Msg) (SingleLogView, tea.Cmd) {
	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)
	return v, cmd
}

// View renders the single log view
func (v SingleLogView) View() string {
	return v.viewport.View()
}

// SetLog sets the log to display
func (v SingleLogView) SetLog(log domain.Log) SingleLogView {
	v.log = &log
	v.refreshDisplay()
	return v
}

// SetSize updates dimensions
func (v SingleLogView) SetSize(width, height int) SingleLogView {
	v.width = width
	v.height = height
	v.viewport.SetWidth(width)
	v.viewport.SetHeight(height)
	return v
}

// PlainText returns the log content without ANSI codes for clipboard
func (v SingleLogView) PlainText() string {
	if v.log == nil {
		return ""
	}
	// TODO: implement proper formatting (JSON expansion, escape sequences)
	return v.log.Content
}

// Log returns the currently displayed log
func (v SingleLogView) Log() *domain.Log {
	return v.log
}

func (v *SingleLogView) refreshDisplay() {
	if v.log == nil {
		v.viewport.SetObjects(nil)
		return
	}

	// TODO: implement JSON formatting and escape sequence expansion
	lines := []SingleLogLine{
		NewSingleLogLine(v.log.Content),
	}
	v.viewport.SetObjects(lines)
}
