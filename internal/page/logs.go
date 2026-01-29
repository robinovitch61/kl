package page

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/robinovitch61/bubbleo/filterableviewport"
	"github.com/robinovitch61/bubbleo/viewport"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/help"
	"github.com/robinovitch61/kl/internal/k8s/container"
	"github.com/robinovitch61/kl/internal/k8s/k8s_model"
	"github.com/robinovitch61/kl/internal/keymap"
	"github.com/robinovitch61/kl/internal/model"
	"github.com/robinovitch61/kl/internal/style"
)

var (
	timestampFormats = []string{"none", "short", "full"}
	nameFormats      = []string{"short", "none", "full"}
)

type LogsPage struct {
	filterableViewport *filterableviewport.Model[model.PageLog]
	viewport           *viewport.Model[model.PageLog]
	keyMap             keymap.KeyMap
	logContainer       *model.PageLogContainer
	timestampFormatIdx int
	nameFormatIdx      int
	styles             style.Styles
	focused            bool
	topHeader          string
}

// assert LogsPage implements GenericPage
var _ GenericPage = LogsPage{}

func NewLogsPage(
	keyMap keymap.KeyMap,
	width, height int,
	descending bool,
	styles style.Styles,
) LogsPage {
	lc := model.NewPageLogContainer(!descending)

	vp := viewport.New[model.PageLog](
		width,
		height-1, // -1 for filter line
		viewport.WithSelectionEnabled[model.PageLog](true),
		viewport.WithWrapText[model.PageLog](true),
		viewport.WithFooterEnabled[model.PageLog](true),
	)

	// Set up selection comparator to maintain selection when content changes
	vp.SetSelectionComparator(func(a, b model.PageLog) bool {
		return a.Equals(b)
	})

	fvp := filterableviewport.New[model.PageLog](
		vp,
		filterableviewport.WithPrefixText[model.PageLog](fmt.Sprintf("(L)ogs %s", getOrder(!descending))),
		filterableviewport.WithEmptyText[model.PageLog]("'/' or 'r' to filter"),
		filterableviewport.WithMatchingItemsOnly[model.PageLog](false),
		filterableviewport.WithCanToggleMatchingItemsOnly[model.PageLog](true),
	)
	// Set header to show when no logs
	vp.SetHeader([]string{"No logs yet"})
	fvp.SetObjects(lc.GetOrderedLogs())

	page := LogsPage{
		filterableViewport: fvp,
		viewport:           vp,
		keyMap:             keyMap,
		logContainer:       lc,
		timestampFormatIdx: 0,
		nameFormatIdx:      0,
		styles:             styles,
		topHeader:          fmt.Sprintf("(L)ogs %s", getOrder(!descending)),
	}
	page.setStickynessBasedOnOrder()
	return page
}

func (p LogsPage) Update(msg tea.Msg) (GenericPage, tea.Cmd) {
	dev.DebugUpdateMsg("LogsPage", msg)
	var cmd tea.Cmd

	p.filterableViewport, cmd = p.filterableViewport.Update(msg)
	return p, cmd
}

func (p LogsPage) View() string {
	return p.filterableViewport.View()
}

func (p LogsPage) HighjackingInput() bool {
	return p.filterableViewport.IsCapturingInput()
}

func (p LogsPage) ContentForFile() []string {
	var content []string
	for _, l := range p.logContainer.GetOrderedLogs() {
		content = append(content, l.GetItem().Content())
	}
	return content
}

func (p LogsPage) HasAppliedFilter() bool {
	return p.filterableViewport.FilterFocused()
}

func (p LogsPage) ToggleShowContext() GenericPage {
	// In bubbleo, this is handled by the 'o' key (ToggleMatchingItemsOnlyKey)
	// The filterableviewport handles this internally
	return p
}

func (p LogsPage) WithDimensions(width, height int) GenericPage {
	p.filterableViewport.SetWidth(width)
	p.filterableViewport.SetHeight(height)
	return p
}

func (p LogsPage) WithFocus() GenericPage {
	p.focused = true
	return p
}

func (p LogsPage) WithBlur() GenericPage {
	p.focused = false
	return p
}

func (p LogsPage) WithStyles(styles style.Styles) GenericPage {
	p.styles = styles
	// bubbleo's filterableviewport doesn't have the same style API
	// styles are set via options at construction time
	return p
}

func (p LogsPage) Help() string {
	return help.MakeHelp(p.keyMap, p.styles.InverseUnderline)
}

func (p LogsPage) WithLogFilter(lf model.LogFilter) LogsPage {
	// bubbleo's filterableviewport doesn't expose filter setting directly
	// The filter is managed internally via key bindings
	return p
}

func (p LogsPage) GetSelectedLog() *model.PageLog {
	return p.viewport.GetSelectedItem()
}

func (p LogsPage) ScrolledUpByOne() LogsPage {
	currentIdx := p.viewport.GetSelectedItemIdx()
	p.viewport.SetSelectedItemIdx(currentIdx - 1)
	return p
}

func (p LogsPage) ScrolledDownByOne() LogsPage {
	currentIdx := p.viewport.GetSelectedItemIdx()
	p.viewport.SetSelectedItemIdx(currentIdx + 1)
	return p
}

func (p LogsPage) WithAppendedLogs(logs []model.PageLog) LogsPage {
	dev.Debug(fmt.Sprintf("Appending %d logs", len(logs)))
	defer dev.Debug("Done appending logs")
	for i := range logs {
		logs[i].CurrentTimestamp = getLogTimestamp(logs[i], timestampFormats[p.timestampFormatIdx])
		logs[i].CurrentName = getContainerName(logs[i], nameFormats[p.nameFormatIdx])
		p.logContainer.AppendLog(logs[i], nil)
	}
	orderedLogs := p.logContainer.GetOrderedLogs()
	// Clear header once we have logs
	if len(orderedLogs) > 0 {
		p.viewport.SetHeader(nil)
	}
	p.filterableViewport.SetObjects(orderedLogs)
	return p
}

func (p LogsPage) WithContainerColors(containerIdToColor map[string]container.ContainerColors) LogsPage {
	allLogs := p.logContainer.GetOrderedLogs()
	for i := range allLogs {
		color, ok := containerIdToColor[allLogs[i].Log.Container.ID()]
		if ok {
			allLogs[i].ContainerColors = &color
		}
	}
	p.setLogs(allLogs)
	return p
}

func (p LogsPage) WithUpdatedShortNames(f func(container.Container) (k8s_model.ContainerNameAndPrefix, error)) (LogsPage, error) {
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

func (p LogsPage) WithLogsRemovedForContainer(containerSpec container.Container) LogsPage {
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

func (p LogsPage) WithLogsTerminatedForContainer(containerSpec container.Container) LogsPage {
	allLogs := p.logContainer.GetOrderedLogs()
	for i := range allLogs {
		if allLogs[i].Log.Container.Equals(containerSpec) {
			allLogs[i].Terminated = true
		}
	}
	p.setLogs(allLogs)
	return p
}

func (p LogsPage) WithNewTimestampFormat() LogsPage {
	p.timestampFormatIdx = (p.timestampFormatIdx + 1) % len(timestampFormats)
	allLogs := p.logContainer.GetOrderedLogs()
	for i := range allLogs {
		allLogs[i].CurrentTimestamp = getLogTimestamp(allLogs[i], timestampFormats[p.timestampFormatIdx])
	}
	p.setLogs(allLogs)
	return p
}

func (p LogsPage) WithNewNameFormat() LogsPage {
	p.nameFormatIdx = (p.nameFormatIdx + 1) % len(nameFormats)
	allLogs := p.logContainer.GetOrderedLogs()
	for i := range allLogs {
		allLogs[i].CurrentName = getContainerName(allLogs[i], nameFormats[p.nameFormatIdx])
	}
	p.setLogs(allLogs)
	return p
}

func (p LogsPage) WithReversedLogOrder() LogsPage {
	// switch the log order
	p.logContainer.ToggleAscending()
	p.setStickynessBasedOnOrder()

	// reinsert the logs with the updated comparator and update the viewport
	p.setLogs(p.logContainer.GetOrderedLogs())
	return p
}

func (p LogsPage) WithStickyness() LogsPage {
	p.setStickynessBasedOnOrder()
	return p
}

func (p LogsPage) WithNoStickyness() LogsPage {
	p.viewport.SetTopSticky(false)
	p.viewport.SetBottomSticky(false)
	return p
}

func (p *LogsPage) setLogs(newLogs []model.PageLog) {
	p.logContainer.RemoveAllLogs()
	for i := range newLogs {
		p.logContainer.AppendLog(newLogs[i], nil)
	}
	p.filterableViewport.SetObjects(p.logContainer.GetOrderedLogs())
}

// setStickynessBasedOnOrder sets viewport stickyness so selection stays at most recent log
func (p *LogsPage) setStickynessBasedOnOrder() {
	if p.logContainer.Ascending() {
		p.viewport.SetTopSticky(false)
		p.viewport.SetBottomSticky(true)
	} else {
		p.viewport.SetTopSticky(true)
		p.viewport.SetBottomSticky(false)
	}
}

func getOrder(ascending bool) string {
	if ascending {
		return "Ascending"
	}
	return "Descending"
}

func getLogTimestamp(log model.PageLog, format string) string {
	if format == "short" {
		return log.Log.Timestamps.Short
	}
	if format == "full" {
		return log.Log.Timestamps.Full
	}
	return ""
}

func getContainerName(log model.PageLog, format string) *k8s_model.ContainerNameAndPrefix {
	var name k8s_model.ContainerNameAndPrefix
	if format == "short" {
		name = log.ContainerNames.Short
	}
	if format == "full" {
		name = log.ContainerNames.Full
	}
	if log.Terminated && len(name.ContainerName) > 0 {
		name.ContainerName += " [TERMINATED]"
	}
	return &name
}
