package filterable_viewport

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/kl/internal/filter"
	"github.com/robinovitch61/kl/internal/keymap"
	"github.com/robinovitch61/kl/internal/style"
	"github.com/robinovitch61/kl/internal/util"
	"regexp"
	"strings"
	"testing"
	"time"
)

var (
	focusFilterKeyMsg      = tea.KeyPressMsg{Code: '/', Text: "/"}
	focusRegexFilterKeyMsg = tea.KeyPressMsg{Code: 'r', Text: "r"}
	enterKeyMsg            = tea.KeyPressMsg{Code: tea.KeyEnter, Text: "enter"}
	clearKeyMsg            = tea.KeyPressMsg{Code: tea.KeyEscape, Text: "esc"}
	wrapKeyMsg             = tea.KeyPressMsg{Code: 'w', Text: "w"}
)

func makeKeyPressMsg(key rune) tea.Msg {
	return tea.KeyPressMsg{Code: key, Text: string(key)}
}

// TestItem implements RenderableComparable for testing
type TestItem struct {
	content string
}

func (t TestItem) Render() string {
	return t.content
}

func (t TestItem) Equals(other interface{}) bool {
	otherStr, ok := other.(TestItem)
	if !ok {
		return false
	}
	return t.content == otherStr.content
}

func newFilterableViewport() FilterableViewport[TestItem] {
	styles := style.NewStyles(style.TermStyleData{
		ForegroundDetected: true,
		Foreground:         lipgloss.Color("#ffffff"),
		ForegroundIsDark:   false,
		BackgroundDetected: true,
		Background:         lipgloss.Color("#000000"),
		BackgroundIsDark:   true,
	},
	)

	matchesFilter := func(item TestItem, f filter.Model) bool {
		if f.Value() == "" {
			return true
		}
		if f.IsRegex() {
			matched, err := regexp.MatchString(f.Value(), item.Render())
			if err != nil {
				return false
			}
			return matched
		}
		return strings.Contains(item.Render(), f.Value())
	}

	return NewFilterableViewport[TestItem](
		FilterableViewportConfig[TestItem]{
			TopHeader:            "Test Header",
			StartShowContext:     false,
			CanToggleShowContext: true,
			SelectionEnabled:     true,
			StartWrapOn:          false,
			KeyMap:               keymap.DefaultKeyMap(),
			Width:                80,
			Height:               20,
			AllRows:              []TestItem{},
			MatchesFilter:        matchesFilter,
			ViewWhenEmpty:        "No items",
			Styles:               styles,
		},
	)
}

func getTestLines(fv FilterableViewport[TestItem]) []string {
	var lines []string
	for _, line := range strings.Split(fv.View(), "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, strings.TrimSpace(line))
		}
	}
	return lines
}

func applyTestFilter(fv FilterableViewport[TestItem], km tea.KeyPressMsg, s string) FilterableViewport[TestItem] {
	fv, _ = fv.Update(km)
	if !fv.Filter.Focused() {
		panic("filter should be focused")
	}
	for _, r := range s {
		fv, _ = fv.Update(makeKeyPressMsg(r))
	}
	fv, _ = fv.Update(enterKeyMsg)
	return fv
}

func TestNewFilterableViewport(t *testing.T) {
	fv := newFilterableViewport()
	fv.SetAllRows([]TestItem{
		{content: "item one"},
		{content: "item two"},
		{content: "another item"},
	})

	if fv.viewport == nil {
		t.Error("viewport should not be nil")
	}
	if len(fv.allRows) != 3 {
		t.Errorf("expected 3 rows, got %d", len(fv.allRows))
	}
	if fv.topHeader != "Test Header" {
		t.Errorf("expected header 'Test Header', got '%s'", fv.topHeader)
	}
	if fv.Filter.ShowContext {
		t.Error("show context should be false by default")
	}
	if fv.whenEmpty != "No items" {
		t.Errorf("expected whenEmpty 'No items', got '%s'", fv.whenEmpty)
	}
	if fv.Filter.Focused() {
		t.Error("filter should not be focused")
	}
	if fv.focused {
		t.Error("viewport should start unfocused")
	}
}

func TestFilterableViewport_FilterNoContext(t *testing.T) {
	fv := newFilterableViewport()
	fv.SetAllRows([]TestItem{
		{content: "item one"},
		{content: "item two"},
		{content: "another item"},
	})

	fv = applyTestFilter(fv, focusFilterKeyMsg, "one")

	// check filter correctly identifies lines in view
	lines := getTestLines(fv)
	util.CmpStr(
		t,
		"Test Header \x1b[48;2;225;225;225m \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m\x1b[38;2;0;0;0;48;2;225;225;225mfilter: \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225mone\x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m(matches only) \x1b[m",
		lines[0],
	)
	if len(lines) != 2 { // 1 for header
		t.Errorf("expected 1 visible item, got %d", len(lines)-1)
	}
	util.CmpStr(
		t,
		"item \x1b[38;2;0;0;0;48;2;255;255;255mone\x1b[m",
		lines[1],
	)

	// check adding matching lines
	newItems := []TestItem{
		{content: "item one"},
		{content: "another item one"},
		{content: "item two"},
		{content: "another item"},
	}
	fv.SetAllRows(newItems)
	lines = getTestLines(fv)
	if len(lines) != 3 { // 1 for header
		t.Errorf("expected 2 visible item, got %d", len(lines)-1)
	}
	if lines[2] != "another item \x1b[38;2;0;0;0;48;2;255;255;255mone\x1b[m" {
		t.Errorf("expected 'item one', got '%q'", lines[2])
	}

	// clear filter
	fv, _ = fv.Update(clearKeyMsg)
	lines = getTestLines(fv)
	if lines[0] != "Test Header  '/' or 'r' to filter" {
		t.Errorf("unexpected header with filter\n%q", fv.View())
	}
	if len(lines) != 5 { // 1 for header
		t.Errorf("expected 4 visible items, got %d", len(lines)-1)
	}
}

func TestFilterableViewport_FilterShowContext(t *testing.T) {
	fv := newFilterableViewport()
	fv.SetAllRows([]TestItem{
		{content: "item one"},
		{content: "item two"},
		{content: "another item"},
	})

	// apply filter
	fv = applyTestFilter(fv, focusFilterKeyMsg, "one")

	// check show context
	if fv.Filter.ShowContext {
		t.Error("contextual filtering should be disabled")
	}
	fv.ToggleShowContext()
	if !fv.Filter.ShowContext {
		t.Error("contextual filtering should be enabled")
	}

	lines := getTestLines(fv)
	if lines[0] != "Test Header \x1b[48;2;225;225;225m \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m\x1b[38;2;0;0;0;48;2;225;225;225mfilter: \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225mone\x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m(1/1, n/N to cycle) \x1b[m" {
		t.Errorf("unexpected header with show context filter\n%q", fv.View())
	}

	if len(lines) != 4 { // 1 for header
		t.Errorf("expected 3 visible item, got %d", len(lines)-1)
	}

	// check adding matching lines
	fv.SetAllRows([]TestItem{
		{content: "item one"},
		{content: "another item one"},
		{content: "item two"},
		{content: "another item"},
	})
	lines = getTestLines(fv)
	if lines[0] != "Test Header \x1b[48;2;225;225;225m \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m\x1b[38;2;0;0;0;48;2;225;225;225mfilter: \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225mone\x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m(1/2, n/N to cycle) \x1b[m" {
		t.Errorf("unexpected header with show context filter\n%q", fv.View())
	}
	if len(lines) != 5 { // 1 for header
		t.Errorf("expected 4 visible item, got %d", len(lines)-1)
	}

	// turn off context
	fv.ToggleShowContext()
	if fv.Filter.ShowContext {
		t.Error("contextual filtering should be disabled")
	}
	lines = getTestLines(fv)
	if lines[0] != "Test Header \x1b[48;2;225;225;225m \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m\x1b[38;2;0;0;0;48;2;225;225;225mfilter: \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225mone\x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m(matches only) \x1b[m" {
		t.Errorf("unexpected header with filter\n%q", fv.View())
	}

	// clear filter
	fv, _ = fv.Update(clearKeyMsg)
	lines = getTestLines(fv)
	if lines[0] != "Test Header  '/' or 'r' to filter" {
		t.Errorf("unexpected header with filter\n%q", fv.View())
	}
	if len(lines) != 5 { // 1 for header
		t.Errorf("expected 4 visible items, got %d", len(lines)-1)
	}
}

func TestFilterableViewport_Focus(t *testing.T) {
	fv := newFilterableViewport()
	fv.SetAllRows([]TestItem{
		{content: "item one"},
	})

	fv.SetFocus(true)
	if !fv.focused {
		t.Error("viewport should be focused after SetFocus(true)")
	}

	// header should be styled when focused
	lines := getTestLines(fv)
	if lines[0] != "\x1b[38;2;0;0;0;46mTest Header\x1b[m  '/' or 'r' to filter" {
		t.Errorf("unexpected header\n%q", fv.View())
	}

	// selection should be styled when focused
	if lines[1] != "\x1b[38;2;0;0;0;48;2;255;255;255mitem one\x1b[m" {
		t.Errorf("unexpected selection\n%q", fv.View())
	}

	// apply filter
	fv = applyTestFilter(fv, focusFilterKeyMsg, "one")

	// selection should have styled filtered line
	lines = getTestLines(fv)
	if lines[1] != "\x1b[38;2;0;0;0;48;2;255;255;255mitem \x1b[m\x1b[38;2;255;255;255;48;2;0;0;0mone\x1b[m" {
		t.Errorf("unexpected selection\n%q", fv.View())
	}
}

func TestFilterableViewport_Clear(t *testing.T) {
	fv := newFilterableViewport()
	fv.SetAllRows([]TestItem{
		{content: "item one"},
		{content: "item two"},
	})

	// apply filter, focus filter, and clear
	fv = applyTestFilter(fv, focusFilterKeyMsg, "one")
	fv, _ = fv.Update(focusFilterKeyMsg)
	fv, _ = fv.Update(clearKeyMsg)
	lines := getTestLines(fv)
	if lines[0] != "Test Header  '/' or 'r' to filter" {
		t.Errorf("unexpected header\n%q", fv.View())
	}

	// apply filter and clear
	fv = applyTestFilter(fv, focusFilterKeyMsg, "one")
	fv, _ = fv.Update(clearKeyMsg)
	lines = getTestLines(fv)
	if lines[0] != "Test Header  '/' or 'r' to filter" {
		t.Errorf("unexpected header\n%q", fv.View())
	}
}

func TestFilterableViewport_ToggleWrap(t *testing.T) {
	fv := newFilterableViewport()

	initialWrap := fv.viewport.GetWrapText()

	fv, _ = fv.Update(wrapKeyMsg)

	if fv.viewport.GetWrapText() == initialWrap {
		t.Error("wrap text should have toggled")
	}
}

func TestFilterableViewport_FilterRegex(t *testing.T) {
	fv := newFilterableViewport()
	fv.SetAllRows([]TestItem{
		{content: "item one"},
		{content: "item two"},
	})

	fv = applyTestFilter(fv, focusRegexFilterKeyMsg, "i.*m")
	lines := getTestLines(fv)
	if lines[0] != "Test Header \x1b[48;2;225;225;225m \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m\x1b[38;2;0;0;0;48;2;225;225;225mregex filter: \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225mi.*m\x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m(matches only) \x1b[m" {
		t.Errorf("unexpected header with regex filter\n%q", fv.View())
	}
	if len(lines) != 3 { // 1 for header
		t.Errorf("expected 2 visible item, got %d", len(lines)-1)
	}
}

func TestFilterableViewport_LongLineManyMatches(t *testing.T) {
	runTest := func(t *testing.T) {
		fv := newFilterableViewport()
		fv, _ = fv.Update(wrapKeyMsg)
		if !fv.viewport.GetWrapText() {
			t.Error("wrap text should be enabled")
		}
		fv.SetAllRows([]TestItem{
			{content: strings.Repeat("rick ross really rad rebel arrr", 10000)},
		})
		fv = applyTestFilter(fv, focusRegexFilterKeyMsg, "r")
		lines := getTestLines(fv)
		if len(lines) != 20 {
			t.Errorf("expected 20 lines, got %d", len(lines))
		}
	}

	util.RunWithTimeout(t, runTest, 3500*time.Millisecond)
}
