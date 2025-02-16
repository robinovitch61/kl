package page

import (
	"bytes"
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/filter"
	"github.com/robinovitch61/kl/internal/filterable_viewport"
	"github.com/robinovitch61/kl/internal/help"
	"github.com/robinovitch61/kl/internal/keymap"
	"github.com/robinovitch61/kl/internal/model"
	"github.com/robinovitch61/kl/internal/style"
	"github.com/robinovitch61/kl/internal/viewport"
	"github.com/robinovitch61/kl/internal/viewport/linebuffer"
	"strings"
)

type SingleLogPage struct {
	filterableViewport filterable_viewport.FilterableViewport[viewport.RenderableString]
	log                model.PageLog
	keyMap             keymap.KeyMap
	styles             style.Styles
}

// assert SingleLogPage implements GenericPage
var _ GenericPage = SingleLogPage{}

func NewSingleLogPage(
	keyMap keymap.KeyMap,
	width, height int,
	styles style.Styles,
) SingleLogPage {
	filterableViewport := filterable_viewport.NewFilterableViewport[viewport.RenderableString](
		filterable_viewport.FilterableViewportConfig[viewport.RenderableString]{
			TopHeader:            "Single Log",
			StartShowContext:     true,
			CanToggleShowContext: false,
			SelectionEnabled:     false,
			StartWrapOn:          true,
			KeyMap:               keyMap,
			Width:                width,
			Height:               height,
			AllRows:              []viewport.RenderableString{},
			MatchesFilter: func(s viewport.RenderableString, filter filter.Model) bool {
				return s.Render().Matches(filter)
			},
			ViewWhenEmpty: "",
			Styles:        styles,
		},
	)
	filterableViewport.SetUpDownMovementWithShift()
	return SingleLogPage{
		filterableViewport: filterableViewport,
		keyMap:             keyMap,
	}
}

func (p SingleLogPage) Update(msg tea.Msg) (GenericPage, tea.Cmd) {
	dev.DebugUpdateMsg("SingleLogPage", msg)
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

func (p SingleLogPage) ContentForFile() []string {
	header, content := veryNicelyFormatThisLog(p.log, true)
	res := []string{header}
	return append(res, content...)
}

func (p SingleLogPage) ToggleShowContext() GenericPage {
	p.filterableViewport.ToggleShowContext()
	return p
}

func (p SingleLogPage) HasAppliedFilter() bool {
	return p.filterableViewport.Filter.Value() != ""
}

func (p SingleLogPage) ContentForClipboard() []string {
	// don't include asci escape chars in header when copying single log to clipboard
	header, content := veryNicelyFormatThisLog(p.log, false)
	res := []string{header}
	return append(res, content...)
}

func (p SingleLogPage) WithDimensions(width, height int) GenericPage {
	p.filterableViewport = p.filterableViewport.WithDimensions(width, height)
	return p
}

func (p SingleLogPage) WithFocus() GenericPage {
	p.filterableViewport.SetFocus(true)
	return p
}

func (p SingleLogPage) WithBlur() GenericPage {
	p.filterableViewport.SetFocus(false)
	return p
}

func (p SingleLogPage) WithStyles(styles style.Styles) GenericPage {
	p.styles = styles
	p.filterableViewport.SetStyles(styles)
	return p
}

func (p SingleLogPage) Help() string {
	return help.MakeHelp(p.keyMap, p.styles.InverseUnderline)
}

func (p SingleLogPage) WithLog(log model.PageLog) SingleLogPage {
	if log.Log.LineBuffer.Content() == p.log.Log.LineBuffer.Content() {
		return p
	}
	p.log = log
	header, content := veryNicelyFormatThisLog(log, true)
	renderableStrings := []viewport.RenderableString{{LineBuffer: linebuffer.New(header)}}
	for _, c := range content {
		renderableStrings = append(renderableStrings, viewport.RenderableString{LineBuffer: linebuffer.New(c)})
	}
	p.filterableViewport.SetAllRows(renderableStrings)
	return p
}

func veryNicelyFormatThisLog(log model.PageLog, styleHeader bool) (string, []string) {
	header := fmt.Sprintf("%s | %s", log.Timestamps.Full, log.RenderName(log.ContainerNames.Full, styleHeader))
	return header, formatJSON(log.Log.LineBuffer.Content())
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
