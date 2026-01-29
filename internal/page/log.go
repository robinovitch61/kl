package page

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/robinovitch61/bubbleo/filterableviewport"
	"github.com/robinovitch61/bubbleo/viewport"
	"github.com/robinovitch61/bubbleo/viewport/item"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/help"
	"github.com/robinovitch61/kl/internal/keymap"
	"github.com/robinovitch61/kl/internal/model"
	"github.com/robinovitch61/kl/internal/style"
)

// LogLine is a simple wrapper around a string that implements viewport.Object
type LogLine struct {
	line item.SingleItem
}

// assert LogLine implements viewport.Object
var _ viewport.Object = LogLine{}

func NewLogLine(s string) LogLine {
	return LogLine{line: item.NewItem(s)}
}

func (l LogLine) GetItem() item.Item {
	return l.line
}

func (l LogLine) Equals(other LogLine) bool {
	return l.line.Content() == other.line.Content()
}

type SingleLogPage struct {
	filterableViewport *filterableviewport.Model[LogLine]
	viewport           *viewport.Model[LogLine]
	log                model.PageLog
	keyMap             keymap.KeyMap
	styles             style.Styles
	focused            bool
}

// assert SingleLogPage implements GenericPage
var _ GenericPage = SingleLogPage{}

func NewSingleLogPage(
	keyMap keymap.KeyMap,
	width, height int,
	styles style.Styles,
) SingleLogPage {
	vp := viewport.New[LogLine](
		width,
		height-1, // -1 for filter line
		viewport.WithSelectionEnabled[LogLine](false),
		viewport.WithWrapText[LogLine](true),
		viewport.WithFooterEnabled[LogLine](true),
	)

	fvp := filterableviewport.New[LogLine](
		vp,
		filterableviewport.WithPrefixText[LogLine]("Single Log"),
		filterableviewport.WithEmptyText[LogLine](""),
		filterableviewport.WithMatchingItemsOnly[LogLine](false),
		filterableviewport.WithCanToggleMatchingItemsOnly[LogLine](false),
	)

	return SingleLogPage{
		filterableViewport: fvp,
		viewport:           vp,
		keyMap:             keyMap,
		styles:             styles,
	}
}

func (p SingleLogPage) Update(msg tea.Msg) (GenericPage, tea.Cmd) {
	dev.DebugUpdateMsg("SingleLogPage", msg)
	var cmd tea.Cmd

	p.filterableViewport, cmd = p.filterableViewport.Update(msg)
	return p, cmd
}

func (p SingleLogPage) View() string {
	return p.filterableViewport.View()
}

func (p SingleLogPage) HighjackingInput() bool {
	return p.filterableViewport.IsCapturingInput()
}

func (p SingleLogPage) ContentForFile() []string {
	header, content := veryNicelyFormatThisLog(p.log, true)
	res := []string{header}
	return append(res, content...)
}

func (p SingleLogPage) ToggleShowContext() GenericPage {
	return p
}

func (p SingleLogPage) HasAppliedFilter() bool {
	return p.filterableViewport.FilterFocused()
}

func (p SingleLogPage) ContentForClipboard() []string {
	// don't include asci escape chars in header when copying single log to clipboard
	header, content := veryNicelyFormatThisLog(p.log, false)
	res := []string{header}
	return append(res, content...)
}

func (p SingleLogPage) WithDimensions(width, height int) GenericPage {
	p.filterableViewport.SetWidth(width)
	p.filterableViewport.SetHeight(height)
	return p
}

func (p SingleLogPage) WithFocus() GenericPage {
	p.focused = true
	return p
}

func (p SingleLogPage) WithBlur() GenericPage {
	p.focused = false
	return p
}

func (p SingleLogPage) WithStyles(styles style.Styles) GenericPage {
	p.styles = styles
	return p
}

func (p SingleLogPage) Help() string {
	return help.MakeHelp(p.keyMap, p.styles.InverseUnderline)
}

func (p SingleLogPage) WithLog(log model.PageLog) SingleLogPage {
	needsUpdate := true

	if log.Log != nil && p.log.Log != nil {
		needsUpdate = log.Log.Item.Content() != p.log.Log.Item.Content()
	}

	if !needsUpdate {
		return p
	}

	p.log = log
	header, content := veryNicelyFormatThisLog(log, true)
	logLines := []LogLine{NewLogLine(header)}
	for _, c := range content {
		logLines = append(logLines, NewLogLine(c))
	}
	p.filterableViewport.SetObjects(logLines)
	return p
}

func veryNicelyFormatThisLog(log model.PageLog, styleHeader bool) (string, []string) {
	header := fmt.Sprintf("%s | %s", log.Log.Timestamps.Full, log.RenderName(log.ContainerNames.Full, styleHeader))
	return header, formatJSON(log.Log.Item.Content())
}

func formatJSON(input string) []string {
	var raw map[string]interface{}

	err := json.Unmarshal([]byte(input), &raw)
	if err != nil {
		return []string{input}
	}

	var prettyJSON bytes.Buffer
	encoder := json.NewEncoder(&prettyJSON)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(raw)
	if err != nil {
		return []string{input}
	}

	lines := strings.Split(prettyJSON.String(), "\n")

	// remove trailing empty line if exists
	if len(lines) > 1 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	var result []string
	for i := range lines {
		if strings.Contains(lines[i], "\\n") || strings.Contains(lines[i], "\\t") {
			lines[i] = strings.ReplaceAll(lines[i], "\\t", "    ")
			parts := strings.Split(lines[i], "\\n")
			result = append(result, parts...)
		} else {
			result = append(result, lines[i])
		}
	}
	return result
}
