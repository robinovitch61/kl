package page

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/filter"
	"github.com/robinovitch61/kl/internal/filterable_viewport"
	"github.com/robinovitch61/kl/internal/keymap"
	"github.com/robinovitch61/kl/internal/model"
	"github.com/robinovitch61/kl/internal/viewport"
	"strings"
)

type SingleLogPage struct {
	filterableViewport filterable_viewport.FilterableViewport[viewport.RenderableString]
	log                model.PageLog
	keyMap             keymap.KeyMap
}

// assert SingleLogPage implements GenericPage
var _ GenericPage = SingleLogPage{}

func NewSingleLogPage(keyMap keymap.KeyMap, width, height int) SingleLogPage {
	filterableViewport := filterable_viewport.NewFilterableViewport[viewport.RenderableString](
		fmt.Sprintf("Single Log - %s for Logs", strings.ToUpper(keyMap.Clear.Help().Key)),
		true,
		false,
		keyMap,
		width,
		height,
		[]viewport.RenderableString{},
		func(s viewport.RenderableString, filter filter.Model) bool {
			return filter.Matches(s)
		},
		"",
	)
	filterableViewport.SetUpDownMovementWithShift()
	return SingleLogPage{
		filterableViewport: filterableViewport,
		keyMap:             keyMap,
	}
}

func (p SingleLogPage) Update(msg tea.Msg) (GenericPage, tea.Cmd) {
	dev.DebugMsg("SingleLogPage", msg)
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	p.filterableViewport, cmd = p.filterableViewport.Update(msg)
	cmds = append(cmds, cmd)
	return p, tea.Batch(cmds...)
}

func (p SingleLogPage) View() string {
	return p.filterableViewport.View()
}

func (p SingleLogPage) HighjackingInput() bool {
	return p.filterableViewport.Filter.Focused()
}

func (p SingleLogPage) HasAppliedFilter() bool {
	return p.filterableViewport.Filter.Value() != ""
}

func (p SingleLogPage) ContentToPersist() []string {
	header, content := veryNicelyFormatThisLog(p.log)
	res := []string{header}
	return append(res, content...)
}

func (p SingleLogPage) WithDimensions(width, height int) GenericPage {
	p.filterableViewport = p.filterableViewport.WithDimensions(width, height)
	return p
}

func (p SingleLogPage) Help() string {
	local := []key.Binding{
		keymap.WithDesc(p.keyMap.Clear, "back to logs"),
		p.keyMap.Copy,
		p.keyMap.PrevLog,
		p.keyMap.NextLog,
		key.NewBinding(key.WithHelp("shift+↑/k", "scroll up within log")),
		key.NewBinding(key.WithHelp("shift+↓/j", "scroll down within log")),
	}
	return makePageHelp(
		"Single Log",
		keymap.GlobalKeyBindings(p.keyMap),
		local,
	)
}

func (p SingleLogPage) WithLog(log model.PageLog) SingleLogPage {
	if log == p.log {
		return p
	}
	p.log = log
	header, content := veryNicelyFormatThisLog(log)
	renderableStrings := []viewport.RenderableString{{Content: header}}
	for _, c := range content {
		renderableStrings = append(renderableStrings, viewport.RenderableString{Content: c})
	}
	p.filterableViewport.SetAllRows(renderableStrings)
	return p
}

func veryNicelyFormatThisLog(log model.PageLog) (string, []string) {
	header := fmt.Sprintf("%s | %s", log.Timestamps.Full, log.ContainerNames.Full)
	return header, formatJSON(log.Log.Content)
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
