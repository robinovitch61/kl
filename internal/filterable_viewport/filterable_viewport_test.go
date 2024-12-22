package filterable_viewport

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/kl/internal/filter"
	"github.com/robinovitch61/kl/internal/keymap"
	"github.com/robinovitch61/kl/internal/style"
	"regexp"
	"strings"
	"testing"
)

var (
	filterKeyMsg = tea.KeyPressMsg{Code: '/', Text: "/"}
	enterKeyMsg  = tea.KeyPressMsg{Code: tea.KeyEnter, Text: "enter"}
	clearKeyMsg  = tea.KeyPressMsg{Code: tea.KeyEscape, Text: "esc"}
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
			TopHeader:             "Test Header",
			StartShowContext:      false,
			CanToggleShowContext:  true,
			StartSelectionEnabled: true,
			StartWrapOn:           false,
			KeyMap:                keymap.DefaultKeyMap(),
			Width:                 80,
			Height:                20,
			AllRows:               []TestItem{},
			MatchesFilter:         matchesFilter,
			ViewWhenEmpty:         "No items",
			Styles:                styles,
		},
	)
}

func getLines(fv FilterableViewport[TestItem]) []string {
	var lines []string
	for _, line := range strings.Split(fv.View(), "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, strings.TrimSpace(line))
		}
	}
	return lines
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
}

func TestFilterableViewport_FilterNoContext(t *testing.T) {
	fv := newFilterableViewport()
	fv.SetAllRows([]TestItem{
		{content: "item one"},
		{content: "item two"},
		{content: "another item"},
	})

	// apply filter
	fv, _ = fv.Update(filterKeyMsg)
	if !fv.Filter.Focused() {
		t.Errorf("filter should be focused after %s key", filterKeyMsg.String())
	}
	for _, r := range "one" {
		fv, _ = fv.Update(makeKeyPressMsg(r))
	}
	fv, _ = fv.Update(enterKeyMsg)

	// check filter correctly identifies lines in view
	lines := getLines(fv)
	if lines[0] != "Test Header \x1b[48;2;225;225;225m \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m\x1b[38;2;0;0;0;48;2;225;225;225mfilter: \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225mone\x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m(matches only) \x1b[m\x1b[m" {
		t.Errorf("unexpected header with filter\n%q", fv.View())
	}
	if len(lines) != 2 { // 1 for header
		t.Errorf("expected 1 visible item, got %d", len(lines)-1)
	}
	if lines[1] != "item \x1b[38;2;0;0;0;48;2;255;255;255mone\x1b[m" {
		t.Errorf("expected 'item one', got '%q'", lines[1])
	}

	// check adding matching lines
	newItems := []TestItem{
		{content: "item one"},
		{content: "another item one"},
		{content: "item two"},
		{content: "another item"},
	}
	fv.SetAllRows(newItems)
	lines = getLines(fv)
	if len(lines) != 3 { // 1 for header
		t.Errorf("expected 2 visible item, got %d", len(lines)-1)
	}
	if lines[2] != "another item \x1b[38;2;0;0;0;48;2;255;255;255mone\x1b[m" {
		t.Errorf("expected 'item one', got '%q'", lines[2])
	}

	// clear filter
	fv, _ = fv.Update(clearKeyMsg)
	lines = getLines(fv)
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
	fv, _ = fv.Update(filterKeyMsg)
	if !fv.Filter.Focused() {
		t.Errorf("filter should be focused after %s key", filterKeyMsg.String())
	}
	for _, r := range "one" {
		fv, _ = fv.Update(makeKeyPressMsg(r))
	}
	fv, _ = fv.Update(enterKeyMsg)

	// check show context
	if fv.Filter.ShowContext {
		t.Error("contextual filtering should be disabled")
	}
	fv.ToggleShowContext()
	if !fv.Filter.ShowContext {
		t.Error("contextual filtering should be enabled")
	}

	lines := getLines(fv)
	if lines[0] != "Test Header \x1b[48;2;225;225;225m \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m\x1b[38;2;0;0;0;48;2;225;225;225mfilter: \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225mone\x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m(1/1, n/N to cycle) \x1b[m\x1b[m" {
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
	lines = getLines(fv)
	if lines[0] != "Test Header \x1b[48;2;225;225;225m \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m\x1b[38;2;0;0;0;48;2;225;225;225mfilter: \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225mone\x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m(1/2, n/N to cycle) \x1b[m\x1b[m" {
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
	lines = getLines(fv)
	if lines[0] != "Test Header \x1b[48;2;225;225;225m \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m\x1b[38;2;0;0;0;48;2;225;225;225mfilter: \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225mone\x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m \x1b[m\x1b[38;2;0;0;0;48;2;225;225;225m(matches only) \x1b[m\x1b[m" {
		t.Errorf("unexpected header with filter\n%q", fv.View())
	}

	// clear filter
	fv, _ = fv.Update(clearKeyMsg)
	lines = getLines(fv)
	if lines[0] != "Test Header  '/' or 'r' to filter" {
		t.Errorf("unexpected header with filter\n%q", fv.View())
	}
	if len(lines) != 5 { // 1 for header
		t.Errorf("expected 4 visible items, got %d", len(lines)-1)
	}
}

// TODO focus filterable viewport

// TODO regex filter

// TODO clear filter

// TODO filter filterable viewport

//func TestFilterableViewport_RegexFilter(t *testing.T) {
//	fv := newFilterableViewport()
//
//	// Activate regex filter
//	msg := tea.KeyMsg{Type: tea.KeyCtrl, Runes: []rune("/")}
//	fv, _ = fv.Update(msg)
//	if !fv.Filter.IsRegex() {
//		t.Error("filter should be in regex mode")
//	}
//
//	// Type regex pattern
//	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("^item")}
//	fv, _ = fv.Update(msg)
//
//	// Apply filter
//	msg = tea.KeyMsg{Type: tea.KeyEnter}
//	fv, _ = fv.Update(msg)
//
//	// Verify only items starting with "item" are visible
//	content := fv.viewport.GetContent()
//	if len(content) != 2 {
//		t.Errorf("expected 2 visible items, got %d", len(content))
//	}
//	if content[0].String() != "item one" {
//		t.Errorf("expected 'item one', got '%s'", content[0].String())
//	}
//	if content[1].String() != "item two" {
//		t.Errorf("expected 'item two', got '%s'", content[1].String())
//	}
//}
//
//func TestFilterableViewport_ClearFilter(t *testing.T) {
//	fv := newFilterableViewport()
//
//	// Set up a filter first
//	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}
//	fv, _ = fv.Update(msg)
//	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("one")}
//	fv, _ = fv.Update(msg)
//	msg = tea.KeyMsg{Type: tea.KeyEnter}
//	fv, _ = fv.Update(msg)
//
//	// Clear the filter
//	msg = tea.KeyMsg{Type: tea.KeyEsc}
//	fv, _ = fv.Update(msg)
//
//	// Verify all items are visible again
//	if len(fv.viewport.GetContent()) != 3 {
//		t.Errorf("expected 3 visible items after clear, got %d", len(fv.viewport.GetContent()))
//	}
//	if fv.Filter.Value() != "" {
//		t.Errorf("expected empty filter value, got '%s'", fv.Filter.Value())
//	}
//}
//
//func TestFilterableViewport_ToggleWrap(t *testing.T) {
//	fv := newFilterableViewport()
//
//	initialWrap := fv.viewport.GetWrapText()
//
//	// Toggle wrap
//	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("w")}
//	fv, _ = fv.Update(msg)
//
//	if fv.viewport.GetWrapText() == initialWrap {
//		t.Error("wrap text should have toggled")
//	}
//}
//
//func TestFilterableViewport_Focus(t *testing.T) {
//	fv := newFilterableViewport()
//
//	// Test unfocused state
//	if fv.focused {
//		t.Error("viewport should start unfocused")
//	}
//
//	// Focus the viewport
//	fv.SetFocus(true)
//	if !fv.focused {
//		t.Error("viewport should be focused after SetFocus(true)")
//	}
//
//	// Verify styles are updated
//	if fv.viewport.SelectedItemStyle != fv.styles.Inverse {
//		t.Error("focused viewport should have inverse selection style")
//	}
//}
//
//func TestFilterableViewport_Selection(t *testing.T) {
//	fv := newFilterableViewport()
//
//	// Test initial selection
//	selection := fv.GetSelection()
//	if selection == nil {
//		t.Fatal("initial selection should not be nil")
//	}
//	if selection.Render() != "item one" {
//		t.Errorf("expected initial selection 'item one', got '%s'", selection.Render())
//	}
//
//	// Change selection
//	fv.SetSelectedContentIdx(1)
//	selection = fv.GetSelection()
//	if selection.Render() != "item two" {
//		t.Errorf("expected selection 'item two', got '%s'", selection.Render())
//	}
//}
