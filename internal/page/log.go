package page

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/help"
	"github.com/robinovitch61/kl/internal/keymap"
	"github.com/robinovitch61/kl/internal/model"
	"github.com/robinovitch61/kl/internal/style"
	"github.com/robinovitch61/viewport/filterableviewport"
	"github.com/robinovitch61/viewport/viewport"
	"github.com/robinovitch61/viewport/viewport/item"
)

// SingleLogLine wraps a string line for display in the viewport
type SingleLogLine struct {
	content string
}

func (s SingleLogLine) GetItem() item.Item {
	return item.NewItem(s.content)
}

type SingleLogPage struct {
	filterableViewport *filterableviewport.Model[SingleLogLine]
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
	addShift := func(keys []string) []string {
		shifted := make([]string, len(keys))
		for i, k := range keys {
			if !strings.Contains(k, "shift") {
				shifted[i] = "shift+" + k
			} else {
				shifted[i] = k
			}
		}
		return shifted
	}
	vp := viewport.New[SingleLogLine](width, height,
		viewport.WithKeyMap[SingleLogLine](viewport.KeyMap{
			PageDown:     key.NewBinding(key.WithKeys(addShift(keyMap.PageDown.Keys())...)),
			PageUp:       key.NewBinding(key.WithKeys(addShift(keyMap.PageUp.Keys())...)),
			HalfPageUp:   key.NewBinding(key.WithKeys(addShift(keyMap.HalfPageUp.Keys())...)),
			HalfPageDown: key.NewBinding(key.WithKeys(addShift(keyMap.HalfPageDown.Keys())...)),
			Up:           key.NewBinding(key.WithKeys(addShift(keyMap.Up.Keys())...)),
			Down:         key.NewBinding(key.WithKeys(addShift(keyMap.Down.Keys())...)),
			Left:         keyMap.Left,
			Right:        keyMap.Right,
			Top:          keyMap.Top,
			Bottom:       keyMap.Bottom,
		}),
		viewport.WithSelectionStyleOverridesItemStyle[SingleLogLine](false),
	)
	vp.SetSelectionEnabled(false)
	vp.SetWrapText(true)

	fvp := filterableviewport.New(vp,
		filterableviewport.WithKeyMap[SingleLogLine](filterableviewport.KeyMap{
			FilterKey:                  keyMap.Filter,
			RegexFilterKey:             keyMap.FilterRegex,
			CaseInsensitiveFilterKey:   keyMap.FilterCaseInsensitive,
			ApplyFilterKey:             keyMap.Enter,
			CancelFilterKey:            keyMap.Clear,
			ToggleMatchingItemsOnlyKey: keyMap.Context,
			NextMatchKey:               keyMap.FilterNextRow,
			PrevMatchKey:               keyMap.FilterPrevRow,
		}),
		filterableviewport.WithMatchingItemsOnly[SingleLogLine](false),
		filterableviewport.WithCanToggleMatchingItemsOnly[SingleLogLine](false),
		filterableviewport.WithEmptyText[SingleLogLine]("'/', 'r', or 'i' to filter"),
		filterableviewport.WithFilterLinePosition[SingleLogLine](filterableviewport.FilterLineTop),
		filterableviewport.WithFilterLinePrefix[SingleLogLine]("Single Log"),
		filterableviewport.WithStyles[SingleLogLine](filterableviewport.Styles{
			Match: filterableviewport.MatchStyles{
				Focused:   styles.Inverse,
				Unfocused: styles.AltInverse,
			},
		}),
	)

	p := SingleLogPage{
		filterableViewport: fvp,
		keyMap:             keyMap,
		styles:             styles,
	}
	p.updateStyles()

	return p
}

func (p SingleLogPage) Update(msg tea.Msg) (GenericPage, tea.Cmd) {
	dev.DebugUpdateMsg("SingleLogPage", msg)
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !p.HighjackingInput() {
			if key.Matches(msg, p.keyMap.Wrap) {
				p.filterableViewport.SetWrapText(!p.filterableViewport.GetWrapText())
				return p, nil
			}
		}
	}

	p.filterableViewport, cmd = p.filterableViewport.Update(msg)
	cmds = append(cmds, cmd)
	return p, tea.Batch(cmds...)
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
	// SingleLogPage doesn't support context toggling
	return p
}

func (p SingleLogPage) HasAppliedFilter() bool {
	return p.filterableViewport.GetFilterText() != ""
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
	p.updateStyles()
	return p
}

func (p SingleLogPage) WithBlur() GenericPage {
	p.focused = false
	p.updateStyles()
	return p
}

func (p SingleLogPage) WithStyles(styles style.Styles) GenericPage {
	p.styles = styles
	p.updateStyles()
	p.filterableViewport.SetFilterableViewportStyles(filterableviewport.Styles{
		Match: filterableviewport.MatchStyles{
			Focused:   styles.Inverse,
			Unfocused: styles.AltInverse,
		},
	})
	return p
}

func (p SingleLogPage) Help() string {
	return help.MakeHelp(p.keyMap, p.styles.InverseUnderline)
}

func (p *SingleLogPage) updateStyles() {
	p.filterableViewport.SetViewportStyles(viewportStylesForFocus(p.focused, p.styles))

	prefix := "Single Log"
	if p.focused {
		prefix = p.styles.Blue.Render(prefix)
	}
	p.filterableViewport.SetFilterLinePrefix(prefix)
}

func (p SingleLogPage) WithLog(log model.PageLog) SingleLogPage {
	needsUpdate := true

	if log.Log != nil && p.log.Log != nil {
		needsUpdate = log.Log.ContentItem.Content() != p.log.Log.ContentItem.Content()
	}

	if !needsUpdate {
		return p
	}

	p.log = log
	header, content := veryNicelyFormatThisLog(log, true)
	lines := []SingleLogLine{{content: header}}
	for _, c := range content {
		lines = append(lines, SingleLogLine{content: c})
	}
	p.filterableViewport.SetObjects(lines)
	return p
}

func veryNicelyFormatThisLog(log model.PageLog, styleHeader bool) (string, []string) {
	header := fmt.Sprintf("%s | %s", log.Log.Timestamps.Full, log.RenderName(log.ContainerNames.Full, styleHeader))
	return header, formatJSON(log.Log.ContentItem.Content())
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
