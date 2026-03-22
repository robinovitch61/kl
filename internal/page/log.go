package page

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/help"
	"github.com/robinovitch61/kl/internal/keymap"
	"github.com/robinovitch61/kl/internal/model"
	"github.com/robinovitch61/kl/internal/style"
	"github.com/robinovitch61/kl/internal/util"
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
	theme              style.Theme
	focused            bool
}

// assert SingleLogPage implements GenericPage
var _ GenericPage = SingleLogPage{}

func NewSingleLogPage(
	keyMap keymap.KeyMap,
	width, height int,
	theme style.Theme,
) SingleLogPage {
	addShift := func(keys []string) []string {
		shifted := make([]string, len(keys))
		for i, k := range keys {
			if strings.Contains(k, "shift") {
				shifted[i] = k
			} else if len(k) == 1 {
				// single letter: use uppercase (v2 String() returns "F" not "shift+f")
				shifted[i] = strings.ToUpper(k)
			} else if strings.HasPrefix(k, "ctrl+") && len(k) == 6 {
				// ctrl+letter: use ctrl+shift+letter (v2 String() returns "ctrl+shift+f")
				shifted[i] = "ctrl+shift+" + k[5:]
			} else {
				// function keys: prefix with shift+ (v2 String() returns "shift+pgdown")
				shifted[i] = "shift+" + k
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
			SearchHistoryPrevKey:       keyMap.SearchHistoryPrev,
			SearchHistoryNextKey:       keyMap.SearchHistoryNext,
		}),
		filterableviewport.WithMatchingItemsOnly[SingleLogLine](false),
		filterableviewport.WithCanToggleMatchingItemsOnly[SingleLogLine](false),
		filterableviewport.WithEmptyText[SingleLogLine]("'/', 'r', or 'i' to filter"),
		filterableviewport.WithFilterLinePosition[SingleLogLine](filterableviewport.FilterLineTop),
		filterableviewport.WithFilterLinePrefix[SingleLogLine]("Single Log"),
		filterableviewport.WithHorizontalPad[SingleLogLine](100000), // effectively center matches when text is unwrapped
		filterableviewport.WithStyles[SingleLogLine](filterableviewport.Styles{
			Match: filterableviewport.MatchStyles{
				Focused:           theme.MatchFocused,
				FocusedIfSelected: theme.MatchFocusedIfSelected,
				Unfocused:         theme.MatchUnfocused,
			},
		}),
	)

	p := SingleLogPage{
		filterableViewport: fvp,
		keyMap:             keyMap,
		theme:              theme,
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
	header, content := veryNicelyFormatThisLog(p.log, false)
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

func (p SingleLogPage) WithTheme(theme style.Theme) GenericPage {
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

func (p SingleLogPage) Help() string {
	return help.MakeHelp(p.keyMap, p.theme.HelpKeyColumn)
}

func (p *SingleLogPage) updateStyles() {
	p.filterableViewport.SetViewportStyles(viewportStylesForFocus(p.focused, p.theme))

	prefix := "Single Log"
	if p.focused {
		prefix = p.theme.FilterPrefixFocused.Render(prefix)
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

func veryNicelyFormatThisLog(log model.PageLog, includeStyle bool) (string, []string) {
	header := fmt.Sprintf("%s | %s", log.Log.Timestamps.Full, log.RenderName(log.ContainerNames.Full, includeStyle))
	var colorize func(string) string
	if includeStyle {
		colorize = log.Log.Colorize()
	}
	return header, util.PrettyPrintJSON(log.Log.ContentItem.ContentNoAnsi(), colorize)
}
