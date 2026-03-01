package page

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/filter"
	"github.com/robinovitch61/kl/internal/help"
	"github.com/robinovitch61/kl/internal/k8s/container"
	"github.com/robinovitch61/kl/internal/k8s/k8s_model"
	"github.com/robinovitch61/kl/internal/keymap"
	"github.com/robinovitch61/kl/internal/model"
	"github.com/robinovitch61/kl/internal/style"
	"github.com/robinovitch61/viewport/filterableviewport"
	"github.com/robinovitch61/viewport/viewport"
)

var (
	timestampFormats = []string{"none", "short", "full"}
	nameFormats      = []string{"short", "none", "full"}
)

type LogsPage struct {
	filterableViewport *filterableviewport.Model[model.PageLog]
	keyMap             keymap.KeyMap
	logContainer       *model.PageLogContainer
	timestampFormatIdx int
	nameFormatIdx      int
	prettyPrint        bool
	theme              style.Theme
	focused            bool
	viewWhenEmpty      string
}

// assert LogsPage implements GenericPage
var _ GenericPage = LogsPage{}

func NewLogsPage(
	keyMap keymap.KeyMap,
	width, height int,
	descending bool,
	theme style.Theme,
) LogsPage {
	lc := model.NewPageLogContainer(!descending)

	vp := viewport.New[model.PageLog](width, height,
		viewport.WithKeyMap[model.PageLog](viewport.KeyMap{
			PageDown:     keyMap.PageDown,
			PageUp:       keyMap.PageUp,
			HalfPageUp:   keyMap.HalfPageUp,
			HalfPageDown: keyMap.HalfPageDown,
			Up:           keyMap.Up,
			Down:         keyMap.Down,
			Left:         keyMap.Left,
			Right:        keyMap.Right,
			Top:          keyMap.Top,
			Bottom:       keyMap.Bottom,
		}),
		viewport.WithSelectionStyleOverridesItemStyle[model.PageLog](false),
	)
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)

	fvp := filterableviewport.New(vp,
		filterableviewport.WithKeyMap[model.PageLog](filterableviewport.KeyMap{
			FilterKey:                  keyMap.Filter,
			RegexFilterKey:             keyMap.FilterRegex,
			CaseInsensitiveFilterKey:   keyMap.FilterCaseInsensitive,
			ApplyFilterKey:             keyMap.Enter,
			CancelFilterKey:            keyMap.Clear,
			ToggleMatchingItemsOnlyKey: keyMap.Context,
			NextMatchKey:               keyMap.FilterNextRow,
			PrevMatchKey:               keyMap.FilterPrevRow,
		}),
		filterableviewport.WithMatchingItemsOnly[model.PageLog](false), // ShowContext=true equivalent
		filterableviewport.WithCanToggleMatchingItemsOnly[model.PageLog](true),
		filterableviewport.WithEmptyText[model.PageLog]("'/', 'r', or 'i' to filter"),
		filterableviewport.WithFilterLinePosition[model.PageLog](filterableviewport.FilterLineTop),
		filterableviewport.WithFilterLinePrefix[model.PageLog](fmt.Sprintf("(L)ogs, %s", getOrder(!descending))),
		filterableviewport.WithStyles[model.PageLog](filterableviewport.Styles{
			Match: filterableviewport.MatchStyles{
				Focused:           theme.MatchFocused,
				FocusedIfSelected: theme.MatchFocusedIfSelected,
				Unfocused:         theme.MatchUnfocused,
			},
		}),
	)

	// Set initial logs
	fvp.SetObjects(lc.GetOrderedLogs())

	// Set selection comparator for maintaining selection
	fvp.SetSelectionComparator(func(a, b model.PageLog) bool {
		return a.Equals(b)
	})

	page := LogsPage{
		filterableViewport: fvp,
		keyMap:             keyMap,
		logContainer:       lc,
		timestampFormatIdx: 0,
		nameFormatIdx:      0,
		theme:              theme,
		viewWhenEmpty:      "No logs yet",
	}
	page.setStickynessBasedOnOrder()
	page.updateStyles()

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
		if key.Matches(msg, p.keyMap.Wrap) {
			p.filterableViewport.SetWrapText(!p.filterableViewport.GetWrapText())
			if !p.filterableViewport.GetWrapText() && p.prettyPrint {
				p.prettyPrint = false
				p.updatePrettyPrintOnLogs()
			}
			return p, nil
		}
		if key.Matches(msg, p.keyMap.PrettyPrint) {
			p.prettyPrint = !p.prettyPrint
			p.updatePrettyPrintOnLogs()
			if p.prettyPrint && !p.filterableViewport.GetWrapText() {
				p.filterableViewport.SetWrapText(true)
			}
			return p, nil
		}
	}

	p.filterableViewport, cmd = p.filterableViewport.Update(msg)
	cmds = append(cmds, cmd)
	return p, tea.Batch(cmds...)
}

func (p LogsPage) View() string {
	if len(p.logContainer.GetOrderedLogs()) == 0 {
		if p.focused {
			return p.theme.FilterPrefixFocused.Render(p.viewWhenEmpty)
		}
		return p.viewWhenEmpty
	}
	return p.filterableViewport.View()
}

func (p LogsPage) HighjackingInput() bool {
	return p.filterableViewport.IsCapturingInput()
}

func (p LogsPage) ContentForFile() []string {
	var content []string
	matchingOnly := p.filterableViewport.GetMatchingItemsOnly()
	filterText := p.filterableViewport.GetFilterText()

	var f filter.Model
	if matchingOnly && filterText != "" {
		f = filter.NewFromText(filterText, p.filterableViewport.IsRegexMode(), p.keyMap)
	}

	for _, l := range p.logContainer.GetOrderedLogs() {
		if !matchingOnly || filterText == "" {
			content = append(content, l.GetItem().Content())
		} else if f.Matches(l.GetItem().ContentNoAnsi()) {
			content = append(content, l.GetItem().Content())
		}
	}
	return content
}

func (p LogsPage) HasAppliedFilter() bool {
	return p.filterableViewport.GetFilterText() != ""
}

func (p LogsPage) ToggleShowContext() GenericPage {
	currentValue := p.filterableViewport.GetMatchingItemsOnly()
	p.filterableViewport.SetMatchingItemsOnly(!currentValue)
	return p
}

func (p LogsPage) WithDimensions(width, height int) GenericPage {
	p.filterableViewport.SetWidth(width)
	p.filterableViewport.SetHeight(height)
	return p
}

func (p LogsPage) WithFocus() GenericPage {
	p.focused = true
	p.updateStyles()
	return p
}

func (p LogsPage) WithBlur() GenericPage {
	p.focused = false
	p.updateStyles()
	return p
}

func (p LogsPage) WithTheme(theme style.Theme) GenericPage {
	p.theme = theme
	p.updateStyles()
	p.filterableViewport.SetFilterableViewportStyles(filterableviewport.Styles{
		Match: filterableviewport.MatchStyles{
			Focused:           theme.MatchFocused,
			FocusedIfSelected: theme.MatchFocusedIfSelected,
			Unfocused:         theme.MatchUnfocused,
		},
	})
	return p
}

func (p LogsPage) Help() string {
	return help.MakeHelp(p.keyMap, p.theme.HelpKeyColumn)
}

func (p LogsPage) WithLogFilter(lf model.LogFilter) LogsPage {
	p.filterableViewport.SetFilter(lf.Value, lf.IsRegex)
	return p
}

func (p LogsPage) GetSelectedLog() *model.PageLog {
	return p.filterableViewport.GetSelectedItem()
}

func (p LogsPage) ScrolledUpByOne() LogsPage {
	currentIdx := p.filterableViewport.GetSelectedItemIdx()
	p.filterableViewport.SetSelectedItemIdx(currentIdx - 1)
	return p
}

func (p LogsPage) ScrolledDownByOne() LogsPage {
	currentIdx := p.filterableViewport.GetSelectedItemIdx()
	p.filterableViewport.SetSelectedItemIdx(currentIdx + 1)
	return p
}

func (p LogsPage) WithAppendedLogs(logs []model.PageLog) LogsPage {
	dev.Debug(fmt.Sprintf("Appending %d logs", len(logs)))
	defer dev.Debug("Done appending logs")

	prevLen := p.logContainer.Len()

	// Check if all new logs are >= the current last timestamp.
	// If ascending and this holds, new logs land at the end of the ordered
	// list and we can use the more efficient AppendObjects path.
	canAppend := p.logContainer.Ascending() && prevLen > 0 && len(logs) > 0
	if canAppend {
		if lastTS, ok := p.logContainer.LastTimestamp(); ok {
			for i := range logs {
				if logs[i].Log.Timestamp.Before(lastTS) {
					canAppend = false
					break
				}
			}
		}
	}

	for i := range logs {
		logs[i].CurrentTimestamp = getLogTimestamp(logs[i], timestampFormats[p.timestampFormatIdx])
		logs[i].CurrentName = getContainerName(logs[i], nameFormats[p.nameFormatIdx])
		logs[i].PrettyPrinted = p.prettyPrint
		logs[i].BuildPrettyItemWithPrefix()
		p.logContainer.AppendLog(logs[i], nil)
	}

	orderedLogs := p.logContainer.GetOrderedLogs()

	if canAppend && len(orderedLogs) == prevLen+len(logs) {
		p.filterableViewport.AppendObjects(orderedLogs[prevLen:])
	} else {
		p.filterableViewport.SetObjects(orderedLogs)
	}

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
	p.updateFilterLabel()

	// reinsert the logs with the updated comparator and update the viewport
	p.setLogs(p.logContainer.GetOrderedLogs())
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
		newLogs[i].BuildPrettyItemWithPrefix()
		p.logContainer.AppendLog(newLogs[i], nil)
	}
	p.filterableViewport.SetObjects(p.logContainer.GetOrderedLogs())
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
	prefix := fmt.Sprintf("(L)ogs, %s", getOrder(p.logContainer.Ascending()))
	if p.focused {
		prefix = p.theme.FilterPrefixFocused.Render(prefix)
	}
	p.filterableViewport.SetFilterLinePrefix(prefix)
}

func (p *LogsPage) updatePrettyPrintOnLogs() {
	allLogs := p.logContainer.GetOrderedLogs()
	for i := range allLogs {
		allLogs[i].PrettyPrinted = p.prettyPrint
	}
	p.setLogs(allLogs)
}

func (p *LogsPage) updateStyles() {
	p.filterableViewport.SetViewportStyles(viewportStylesForFocus(p.focused, p.theme))
	p.updateFilterLabel()
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
