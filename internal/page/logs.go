package page

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/filter"
	"github.com/robinovitch61/kl/internal/filterable_viewport"
	"github.com/robinovitch61/kl/internal/help"
	"github.com/robinovitch61/kl/internal/keymap"
	"github.com/robinovitch61/kl/internal/model"
)

var (
	timestampFormats = []string{"none", "short", "full"}
	nameFormats      = []string{"short", "none", "full"}
)

type LogsPage struct {
	filterableViewport filterable_viewport.FilterableViewport[model.PageLog]
	keyMap             keymap.KeyMap
	logContainer       *model.PageLogContainer
	timestampFormatIdx int
	nameFormatIdx      int
}

// assert LogsPage implements GenericPage
var _ GenericPage = LogsPage{}

func NewLogsPage(keyMap keymap.KeyMap, width, height int, descending bool) LogsPage {
	lc := model.NewPageLogContainer(!descending)
	filterableViewport := filterable_viewport.NewFilterableViewport[model.PageLog](
		fmt.Sprintf("(L)ogs, %s", getOrder(!descending)),
		true,
		true,
		false,
		keyMap,
		width,
		height,
		lc.GetOrderedLogs(),
		func(log model.PageLog, filter filter.Model) bool {
			return filter.Matches(log)
		},
		"No logs yet",
	)
	filterableViewport.SetMaintainSelection(true)
	page := LogsPage{
		filterableViewport: filterableViewport,
		keyMap:             keyMap,
		logContainer:       lc,
		timestampFormatIdx: 0,
		nameFormatIdx:      0,
	}
	page.setStickynessBasedOnOrder()
	page.updateFilterLabel()
	return page
}

func (p LogsPage) Update(msg tea.Msg) (GenericPage, tea.Cmd) {
	dev.DebugUpdateMsg("LogsPage", msg)
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if p.HighjackingInput() {
			p.filterableViewport, cmd = p.filterableViewport.Update(msg)
			cmds = append(cmds, cmd)
			return p, tea.Batch(cmds...)
		}

		if key.Matches(msg, p.keyMap.Timestamps) {
			p.timestampFormatIdx = (p.timestampFormatIdx + 1) % len(timestampFormats)
			allLogs := p.logContainer.GetOrderedLogs()
			for i := range allLogs {
				allLogs[i].CurrentTimestamp = getLogTimestamp(allLogs[i], timestampFormats[p.timestampFormatIdx])
			}
			p.setLogs(allLogs)
			return p, nil
		}

		if key.Matches(msg, p.keyMap.Name) {
			p.nameFormatIdx = (p.nameFormatIdx + 1) % len(nameFormats)
			allLogs := p.logContainer.GetOrderedLogs()
			for i := range allLogs {
				allLogs[i].CurrentName = getContainerName(allLogs[i], nameFormats[p.nameFormatIdx])
			}
			p.setLogs(allLogs)
			return p, nil
		}

		if key.Matches(msg, p.keyMap.ReverseOrder) {
			// switch the log order
			p.logContainer.ToggleAscending()
			p.setStickynessBasedOnOrder()
			p.updateFilterLabel()

			// reinsert the logs with the updated comparator and update the viewport
			p.setLogs(p.logContainer.GetOrderedLogs())
			return p, nil
		}

		// toggle filtering with context
		if key.Matches(msg, p.keyMap.Context) {
			p.filterableViewport.ToggleFilteringWithContext()
			return p, nil
		}
	}

	p.filterableViewport, cmd = p.filterableViewport.Update(msg)
	cmds = append(cmds, cmd)
	return p, tea.Batch(cmds...)
}

func (p LogsPage) View() string {
	return p.filterableViewport.View()
}

func (p LogsPage) HighjackingInput() bool {
	return p.filterableViewport.HighjackingInput()
}

func (p LogsPage) ContentForFile() []string {
	var content []string
	for _, l := range p.logContainer.GetOrderedLogs() {
		if p.filterableViewport.Filter.FilteringWithContext {
			content = append(content, l.Render())
		} else if p.filterableViewport.Filter.Matches(l) {
			content = append(content, l.Render())
		}
	}
	return content
}

func (p LogsPage) WithDimensions(width, height int) GenericPage {
	p.filterableViewport = p.filterableViewport.WithDimensions(width, height)
	return p
}

func (p LogsPage) WithFocus() GenericPage {
	p.filterableViewport.SetFocus(true, true)
	return p
}

func (p LogsPage) WithBlur() GenericPage {
	p.filterableViewport.SetFocus(false, false)
	return p
}

func (p LogsPage) Help() string {
	return help.MakeHelp(p.keyMap)
}

func (p LogsPage) WithLogFilter(lf model.LogFilter) LogsPage {
	p.filterableViewport.Filter.SetValue(lf.Value)
	p.filterableViewport.Filter.SetIsRegex(lf.IsRegex)
	return p
}

func (p LogsPage) GetSelectedLog() *model.PageLog {
	return p.filterableViewport.GetSelection()
}

func (p LogsPage) ScrolledUpByOne() LogsPage {
	currentIdx := p.filterableViewport.GetSelectionIdx()
	p.filterableViewport.SetSelectedContentIdx(currentIdx - 1)
	return p
}

func (p LogsPage) ScrolledDownByOne() LogsPage {
	currentIdx := p.filterableViewport.GetSelectionIdx()
	p.filterableViewport.SetSelectedContentIdx(currentIdx + 1)
	return p
}

func (p LogsPage) WithAppendedLogs(logs []model.PageLog) LogsPage {
	dev.Debug(fmt.Sprintf("Appending %d logs", len(logs)))
	defer dev.Debug("Done appending logs")
	for _, log := range logs {
		log.CurrentTimestamp = getLogTimestamp(log, timestampFormats[p.timestampFormatIdx])
		log.CurrentName = getContainerName(log, nameFormats[p.nameFormatIdx])
		p.logContainer.AppendLog(log, nil)
	}
	p.filterableViewport.SetAllRows(p.logContainer.GetOrderedLogs())
	return p
}

func (p LogsPage) WithContainerColors(containerIdToColor map[string]model.ContainerColors) LogsPage {
	allLogs := p.logContainer.GetOrderedLogs()
	for i := range allLogs {
		color, ok := containerIdToColor[allLogs[i].Log.Container.ID()]
		if ok {
			allLogs[i].ContainerColors = color
		}
	}
	p.setLogs(allLogs)
	return p
}

func (p LogsPage) WithUpdatedShortNames(f func(model.Container) (model.PageLogContainerName, error)) (LogsPage, error) {
	allLogs := p.logContainer.GetOrderedLogs()
	for i := range allLogs {
		short, err := f(allLogs[i].Log.Container)
		if err != nil {
			return p, err
		}
		allLogs[i].ContainerNames.Short = short
		allLogs[i].CurrentName = getContainerName(allLogs[i], nameFormats[p.nameFormatIdx])
	}
	p.setLogs(allLogs)
	return p, nil
}

func (p LogsPage) WithLogsRemovedForContainer(containerSpec model.Container) LogsPage {
	allLogs := p.logContainer.GetOrderedLogs()
	var newLogs []model.PageLog
	for _, log := range allLogs {
		if !log.Log.Container.Equals(containerSpec) {
			newLogs = append(newLogs, log)
		}
	}
	p.setLogs(newLogs)
	return p
}

func (p LogsPage) WithLogsTerminatedForContainer(containerSpec model.Container) LogsPage {
	allLogs := p.logContainer.GetOrderedLogs()
	for i := range allLogs {
		if allLogs[i].Log.Container.Equals(containerSpec) {
			allLogs[i].Terminated = true
		}
	}
	p.setLogs(allLogs)
	return p
}

func (p LogsPage) WithStickyness() LogsPage {
	p.setStickynessBasedOnOrder()
	return p
}

func (p LogsPage) WithNoStickyness() LogsPage {
	p.filterableViewport.SetTopSticky(false)
	p.filterableViewport.SetBottomSticky(false)
	return p
}

func (p *LogsPage) setLogs(newLogs []model.PageLog) {
	p.logContainer.RemoveAllLogs()
	for i := range newLogs {
		p.logContainer.AppendLog(newLogs[i], nil)
	}
	p.filterableViewport.SetAllRows(p.logContainer.GetOrderedLogs())
}

// setStickynessBasedOnOrder sets viewport stickyness so selection stays at most recent log
func (p *LogsPage) setStickynessBasedOnOrder() {
	if p.logContainer.Ascending() {
		p.filterableViewport.SetTopSticky(false)
		p.filterableViewport.SetBottomSticky(true)
	} else {
		p.filterableViewport.SetTopSticky(true)
		p.filterableViewport.SetBottomSticky(false)
	}
}

func (p *LogsPage) updateFilterLabel() {
	label := fmt.Sprintf("(L)ogs %s", getOrder(p.logContainer.Ascending()))
	p.filterableViewport.SetTopHeader(label)
}

func getOrder(ascending bool) string {
	if ascending {
		return "Ascending"
	}
	return "Descending"
}

func getLogTimestamp(log model.PageLog, format string) string {
	if format == "short" {
		return log.Timestamps.Short
	}
	if format == "full" {
		return log.Timestamps.Full
	}
	return ""
}

func getContainerName(log model.PageLog, format string) model.PageLogContainerName {
	var name model.PageLogContainerName
	if format == "short" {
		name = log.ContainerNames.Short
	}
	if format == "full" {
		name = log.ContainerNames.Full
	}
	if log.Terminated && len(name.ContainerName) > 0 {
		name.ContainerName += " [TERMINATED]"
	}
	return name
}
