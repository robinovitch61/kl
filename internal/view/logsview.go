package view

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/bubbleo/filterableviewport"
	"github.com/robinovitch61/bubbleo/viewport"
	"github.com/robinovitch61/kl/internal/domain"
)

// LogsView displays the interleaved log stream
type LogsView struct {
	viewport        *filterableviewport.Model[LogRow]
	logs            domain.Logs
	ascending       bool
	timestampFormat TimestampFormat
	nameFormat      NameFormat
	containerColors map[domain.ContainerID]string
	width           int
	height          int
}

// NewLogsView creates a new logs view
func NewLogsView(width, height int, ascending bool) LogsView {
	vp := viewport.New[LogRow](width, height)
	fv := filterableviewport.New[LogRow](vp)
	return LogsView{
		viewport:        fv,
		ascending:       ascending,
		timestampFormat: TimestampShort,
		nameFormat:      NameShort,
		containerColors: make(map[domain.ContainerID]string),
		width:           width,
		height:          height,
	}
}

// Update handles messages
func (v LogsView) Update(msg tea.Msg) (LogsView, tea.Cmd) {
	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)
	return v, cmd
}

// View renders the logs view
func (v LogsView) View() string {
	return v.viewport.View()
}

// AppendLogs adds new logs and re-sorts
func (v LogsView) AppendLogs(logs []domain.Log) LogsView {
	v.logs = append(v.logs, logs...)
	v.refreshDisplay()
	return v
}

// ClearLogsForContainer removes logs for a specific container
func (v LogsView) ClearLogsForContainer(id domain.ContainerID) LogsView {
	var remaining domain.Logs
	for _, log := range v.logs {
		if log.ContainerID != id {
			remaining = append(remaining, log)
		}
	}
	v.logs = remaining
	v.refreshDisplay()
	return v
}

// SetAscending changes sort order
func (v LogsView) SetAscending(asc bool) LogsView {
	v.ascending = asc
	v.refreshDisplay()
	return v
}

// ToggleTimestampFormat cycles through formats
func (v LogsView) ToggleTimestampFormat() LogsView {
	v.timestampFormat = (v.timestampFormat + 1) % 3
	v.refreshDisplay()
	return v
}

// ToggleNameFormat cycles through formats
func (v LogsView) ToggleNameFormat() LogsView {
	v.nameFormat = (v.nameFormat + 1) % 3
	v.refreshDisplay()
	return v
}

// SetSize updates dimensions
func (v LogsView) SetSize(width, height int) LogsView {
	v.width = width
	v.height = height
	v.viewport.SetWidth(width)
	v.viewport.SetHeight(height)
	return v
}

// SelectedLog returns the currently selected log
func (v LogsView) SelectedLog() *domain.Log {
	row := v.viewport.GetSelectedItem()
	if row == nil {
		return nil
	}
	log := row.Log()
	return &log
}

func (v *LogsView) refreshDisplay() {
	var sorted domain.Logs
	if v.ascending {
		sorted = v.logs.SortAscending()
	} else {
		sorted = v.logs.SortDescending()
	}

	rows := make([]LogRow, len(sorted))
	for i, log := range sorted {
		color := v.containerColors[log.ContainerID]
		rows[i] = NewLogRow(log, v.timestampFormat, v.nameFormat, color)
	}
	v.viewport.SetObjects(rows)
}
