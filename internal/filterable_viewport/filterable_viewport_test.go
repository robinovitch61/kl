package filterable_viewport

import (
	"github.com/charmbracelet/bubbles/v2/key"
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
	red          = lipgloss.Color("#ff0000")
	blue         = lipgloss.Color("#0000ff")
	//green          = lipgloss.Color("#00ff00")
	//selectionStyle = lipgloss.NewStyle().Foreground(blue)
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
	testItems := []TestItem{
		{content: "item one"},
		{content: "item two"},
		{content: "another item"},
	}

	km := keymap.KeyMap{
		Filter:      key.NewBinding(key.WithKeys(filterKeyMsg.String())),
		FilterRegex: key.NewBinding(key.WithKeys("r")),
		Clear:       key.NewBinding(key.WithKeys("esc")),
		Enter:       key.NewBinding(key.WithKeys(enterKeyMsg.String())),
		Wrap:        key.NewBinding(key.WithKeys("w")),
	}

	styles := style.Styles{
		Blue:       lipgloss.NewStyle().Foreground(blue),
		Alt:        lipgloss.NewStyle().Background(red),
		Inverse:    lipgloss.NewStyle().Reverse(true),
		AltInverse: lipgloss.NewStyle().Background(red).Reverse(true),
		Unset:      lipgloss.NewStyle(),
	}

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
			TopHeader:                  "Test Header",
			StartFilterWithContext:     false,
			CanToggleFilterWithContext: true,
			StartSelectionEnabled:      true,
			StartWrapOn:                false,
			KeyMap:                     km,
			Width:                      80,
			Height:                     20,
			AllRows:                    testItems,
			MatchesFilter:              matchesFilter,
			ViewWhenEmpty:              "No items",
			Styles:                     styles,
		},
	)
}

func TestNewFilterableViewport(t *testing.T) {
	fv := newFilterableViewport()

	if fv.viewport == nil {
		t.Error("viewport should not be nil")
	}
	if len(fv.allRows) != 3 {
		t.Errorf("expected 3 rows, got %d", len(fv.allRows))
	}
	if fv.topHeader != "Test Header" {
		t.Errorf("expected header 'Test Header', got '%s'", fv.topHeader)
	}
	if fv.filterWithContext {
		t.Error("filterWithContext should be false by default")
	}
	if fv.whenEmpty != "No items" {
		t.Errorf("expected whenEmpty 'No items', got '%s'", fv.whenEmpty)
	}
}

func TestFilterableViewport_Filter(t *testing.T) {
	fv := newFilterableViewport()

	if fv.Filter.Focused() {
		t.Error("filter should not be focused")
	}

	// Test basic filtering
	fv, _ = fv.Update(filterKeyMsg)
	if !fv.Filter.Focused() {
		t.Errorf("filter should be focused after %s key", filterKeyMsg.String())
	}

	// Type "one" into filter
	for _, r := range "one" {
		fv, _ = fv.Update(makeKeyPressMsg(r))
	}

	// Press enter to apply filter
	fv, _ = fv.Update(enterKeyMsg)

	// Verify only items containing "one" are visible
	content := fv.viewport.View()
	println(content)
	//if len(content) != 1 {
	//	t.Errorf("expected 1 visible item, got %d", len(content))
	//}
	//if content[0].String() != "item one" {
	//	t.Errorf("expected 'item one', got '%s'", content[0].String())
	//}
}

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
//func TestFilterableViewport_ContextualFiltering(t *testing.T) {
//	fv := newFilterableViewport()
//
//	// Enable contextual filtering
//	fv.ToggleFilteringWithContext()
//	if !fv.Filter.FilteringWithContext {
//		t.Error("contextual filtering should be enabled")
//	}
//
//	// Set up a filter
//	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}
//	fv, _ = fv.Update(msg)
//	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("item")}
//	fv, _ = fv.Update(msg)
//	msg = tea.KeyMsg{Type: tea.KeyEnter}
//	fv, _ = fv.Update(msg)
//
//	// Verify all items are still visible but matches are highlighted
//	if len(fv.viewport.GetContent()) != 3 {
//		t.Errorf("expected all 3 items visible in contextual mode, got %d", len(fv.viewport.GetContent()))
//	}
//	if fv.viewport.GetStringToHighlight() != "item" {
//		t.Errorf("expected highlight string 'item', got '%s'", fv.viewport.GetStringToHighlight())
//	}
//}
//
//func TestFilterableViewport_UpdateContent(t *testing.T) {
//	fv := newFilterableViewport()
//
//	newItems := []TestItem{
//		{content: "new item one"},
//		{content: "new item two"},
//	}
//
//	fv.SetAllRows(newItems)
//
//	content := fv.viewport.GetContent()
//	if len(content) != 2 {
//		t.Errorf("expected 2 items after update, got %d", len(content))
//	}
//	if content[0].String() != "new item one" {
//		t.Errorf("expected 'new item one', got '%s'", content[0].String())
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
