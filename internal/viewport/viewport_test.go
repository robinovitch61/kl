package viewport

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"io"
	"strings"
	"testing"
)

var (
	testKeyMap       = DefaultKeyMap()
	downKeyMsg       = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(testKeyMap.Down.Keys()[0])}
	halfPgDownKeyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(testKeyMap.HalfPageDown.Keys()[0])}
	fullPgDownKeyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(testKeyMap.PageDown.Keys()[0])}
	upKeyMsg         = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(testKeyMap.Up.Keys()[0])}
	halfPgUpKeyMsg   = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(testKeyMap.HalfPageUp.Keys()[0])}
	fullPgUpKeyMsg   = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(testKeyMap.PageUp.Keys()[0])}
	goToTopKeyMsg    = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(testKeyMap.Top.Keys()[0])}
	goToBottomKeyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(testKeyMap.Bottom.Keys()[0])}
	red              = lipgloss.Color("#ff0000")
	blue             = lipgloss.Color("#0000ff")
	green            = lipgloss.Color("#00ff00")
	selectionStyle   = renderer().NewStyle().Foreground(blue)
)

func renderer() *lipgloss.Renderer {
	r := lipgloss.NewRenderer(io.Discard)
	r.SetColorProfile(termenv.TrueColor)
	r.SetHasDarkBackground(true)
	return r
}

func newViewport(width, height int) Model[RenderableString] {
	vp := New[RenderableString](width, height)
	vp.SelectedItemStyle = selectionStyle
	return vp
}

// # SELECTION DISABLED, WRAP OFF

func TestViewport_SelectionOff_WrapOff_Empty(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	expectedView := pad(vp.width, vp.height, []string{})
	compare(t, expectedView, vp.View())
	vp.SetHeader([]string{"header"})
	expectedView = pad(vp.width, vp.height, []string{"header"})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_SmolDimensions(t *testing.T) {
	w, h := 0, 0
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetContent([]RenderableString{{Content: "hi"}})
	expectedView := pad(vp.width, vp.height, []string{""})
	compare(t, expectedView, vp.View())

	vp.SetWidth(1)
	vp.SetHeight(1)
	expectedView = pad(vp.width, vp.height, []string{"."})
	compare(t, expectedView, vp.View())

	vp.SetWidth(2)
	vp.SetHeight(2)
	expectedView = pad(vp.width, vp.height, []string{"..", ""})
	compare(t, expectedView, vp.View())

	vp.SetWidth(3)
	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{"...", "hi", "..."})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_Basic(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetContent([]RenderableString{
		{Content: "first line"},
		{Content: renderer().NewStyle().Foreground(red).Render("second") + " line"},
		{Content: renderer().NewStyle().Foreground(red).Render("a really really long line")},
		{Content: renderer().NewStyle().Foreground(red).Render("a") + " really really long line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond\x1b[0m line",
		"\x1b[38;2;255;0;0ma really rea...\x1b[0m",
		"\x1b[38;2;255;0;0ma\x1b[0m really rea...",
		"100% (4/4)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_GetConfigs(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
	})
	if selectionEnabled := vp.GetSelectionEnabled(); selectionEnabled {
		t.Errorf("expected selection to be disabled, got %v", selectionEnabled)
	}
	if wrapText := vp.GetWrapText(); wrapText {
		t.Errorf("expected text wrapping to be disabled, got %v", wrapText)
	}
	if selectedItemIdx := vp.GetSelectedItemIdx(); selectedItemIdx != 0 {
		t.Errorf("expected selected item index to be 0, got %v", selectedItemIdx)
	}
	if selectedItem := vp.GetSelectedItem(); selectedItem != nil {
		t.Errorf("expected selected item to be nil, got %v", selectedItem)
	}
}

func TestViewport_SelectionOff_WrapOff_ShowFooter(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetContent([]RenderableString{
		{Content: "first line"},
		{Content: renderer().NewStyle().Foreground(red).Render("second") + " line"},
		{Content: renderer().NewStyle().Foreground(red).Render("a really really long line")},
		{Content: renderer().NewStyle().Foreground(red).Render("a") + " really really long line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond\x1b[0m line",
		"\x1b[38;2;255;0;0ma really rea...\x1b[0m",
		"75% (3/4)",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond\x1b[0m line",
		"\x1b[38;2;255;0;0ma really rea...\x1b[0m",
		"\x1b[38;2;255;0;0ma\x1b[0m really rea...",
		"100% (4/4)",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(7)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond\x1b[0m line",
		"\x1b[38;2;255;0;0ma really rea...\x1b[0m",
		"\x1b[38;2;255;0;0ma\x1b[0m really rea...",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_FooterStyle(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.FooterStyle = renderer().NewStyle().Foreground(red)
	vp.SetContent([]RenderableString{
		{Content: "1"},
		{Content: "2"},
		{Content: "3"},
		{Content: "4"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"1",
		"2",
		"3",
		"\x1b[38;2;255;0;0m75% (3/4)\x1b[0m",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_FooterDisabled(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetContent([]RenderableString{
		{Content: "first line"},
		{Content: "second line"},
		{Content: "third line"},
		{Content: "fourth line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"second line",
		"third line",
		"75% (3/4)",
	})
	compare(t, expectedView, vp.View())

	vp.SetFooterVisible(false)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"second line",
		"third line",
		"fourth line",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_SpaceAround(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetContent([]RenderableString{
		{Content: "    first line     "},
		{Content: "          first line          "},
		{Content: "               first line               "},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"    first li...",
		"          fi...",
		"            ...",
		"100% (3/3)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_MultiHeader(t *testing.T) {
	w, h := 15, 2
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header1", "header2"})
	vp.SetContent([]RenderableString{
		{Content: "line1"},
		{Content: "line2"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header1",
		"header2",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(4)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line1",
		"50% (1/2)",
	})
	compare(t, expectedView, vp.View())

	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line2",
		"100% (2/2)",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(5)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line1",
		"line2",
		"100% (2/2)",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line1",
		"line2",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_OverflowLine(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"long header overflows"})
	vp.SetContent([]RenderableString{
		{Content: "123456789012345"},
		{Content: "1234567890123456"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"long header ...",
		"123456789012345",
		"123456789012...",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_OverflowHeight(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetContent([]RenderableString{
		{Content: "123456789012345"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"123456789012345",
		"123456789012...",
		"123456789012...",
		"123456789012...",
		"66% (4/6)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_Scrolling(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent := func() {
		vp.SetContent([]RenderableString{
			{Content: "first"},
			{Content: "second"},
			{Content: "third"},
			{Content: "fourth"},
			{Content: "fifth"},
			{Content: "sixth"},
		})
	}
	validate := func(expectedView string) {
		// set content multiple times to confirm no side effects of doing it
		compare(t, expectedView, vp.View())
		setContent()
		compare(t, expectedView, vp.View())
	}
	setContent()
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
		"third",
		"fourth",
		"66% (4/6)",
	})
	validate(expectedView)

	// scrolling up past top is no-op
	vp, _ = vp.Update(upKeyMsg)
	validate(expectedView)

	// scrolling down by one
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"second",
		"third",
		"fourth",
		"fifth",
		"83% (5/6)",
	})
	validate(expectedView)

	// scrolling down by one again
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"third",
		"fourth",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	validate(expectedView)

	// scrolling down past bottom when at bottom is no-op
	vp, _ = vp.Update(downKeyMsg)
	validate(expectedView)
}

func TestViewport_SelectionOff_WrapOff_ScrollToItem(t *testing.T) {
	w, h := 15, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
		{Content: "third"},
		{Content: "fourth"},
		{Content: "fifth"},
		{Content: "sixth"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
		"33% (2/6)",
	})
	compare(t, expectedView, vp.View())

	// scroll so last item in view
	vp.ScrollSoItemIdxInView(5)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())

	// scroll so second item in view
	vp.ScrollSoItemIdxInView(1)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"second",
		"third",
		"50% (3/6)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_BulkScrolling(t *testing.T) {
	w, h := 15, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
		{Content: "third"},
		{Content: "fourth"},
		{Content: "fifth"},
		{Content: "sixth"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
		"33% (2/6)",
	})
	compare(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"third",
		"fourth",
		"66% (4/6)",
	})
	compare(t, expectedView, vp.View())

	// half page down
	vp, _ = vp.Update(halfPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"fourth",
		"fifth",
		"83% (5/6)",
	})
	compare(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"third",
		"fourth",
		"66% (4/6)",
	})
	compare(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"second",
		"third",
		"50% (3/6)",
	})
	compare(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
		"33% (2/6)",
	})
	compare(t, expectedView, vp.View())

	// go to bottom
	vp, _ = vp.Update(goToBottomKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())

	// go to top
	vp, _ = vp.Update(goToTopKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
		"33% (2/6)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_Panning(t *testing.T) {
	w, h := 10, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header long"})
	setContent := func() {
		vp.SetContent([]RenderableString{
			{Content: "first line that is fairly long"},
			{Content: "second line that is even much longer than the first"},
			{Content: "third line that is fairly long"},
			{Content: "fourth"},
			{Content: "fifth line that is fairly long"},
			{Content: "sixth"},
		})
	}
	validate := func(expectedView string) {
		// set content multiple times to confirm no side effects of doing it
		compare(t, expectedView, vp.View())
		setContent()
		compare(t, expectedView, vp.View())
	}
	setContent()
	expectedView := pad(vp.width, vp.height, []string{
		"header ...",
		"first l...",
		"second ...",
		"third l...",
		"fourth",
		"66% (4/6)",
	})
	validate(expectedView)

	// pan right
	vp.safelySetXOffset(5)
	expectedView = pad(vp.width, vp.height, []string{
		"header ...",
		"...ne t...",
		"...ine ...",
		"...ne t...",
		".",
		"66% (4/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header ...",
		"...ine ...",
		"...ne t...",
		".",
		"...ne t...",
		"83% (5/6)",
	})
	validate(expectedView)

	// pan all the way right
	vp.safelySetXOffset(41)
	expectedView = pad(vp.width, vp.height, []string{
		"header ...",
		"...e first",
		"...",
		"...",
		"...",
		"83% (5/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header ...",
		"...ly long",
		"...",
		"...ly long",
		"...",
		"100% (6/6)",
	})
	validate(expectedView)

	// set shorter content
	vp.SetContent([]RenderableString{
		{Content: "the first one"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header ...",
		"...rst one",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_ChangeHeight(t *testing.T) {
	w, h := 10, 3
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
		{Content: "third"},
		{Content: "fourth"},
		{Content: "fifth"},
		{Content: "sixth"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(6)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
		"third",
		"fourth",
		"66% (4/6)",
	})
	compare(t, expectedView, vp.View())

	// scroll to bottom
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"third",
		"fourth",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())

	// reduce height
	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"third",
		"50% (3/6)",
	})
	compare(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(8)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_ChangeContent(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
		{Content: "third"},
		{Content: "fourth"},
		{Content: "fifth"},
		{Content: "sixth"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
		"third",
		"50% (3/6)",
	})
	compare(t, expectedView, vp.View())

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"fourth",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())

	// remove content
	vp.SetContent([]RenderableString{})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
	})
	compare(t, expectedView, vp.View())

	// re-add content
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_StringToHighlight(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetStringToHighlight("second")
	vp.HighlightStyle = renderer().NewStyle().Foreground(red)
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
		{Content: "second"},
		{Content: "third"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first",
		"\x1b[38;2;255;0;0msecond\x1b[0m",
		"\x1b[38;2;255;0;0msecond\x1b[0m",
		"75% (3/4)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_StringToHighlightAnsi(t *testing.T) {
	w, h := 20, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetContent([]RenderableString{
		{Content: "line \x1b[38;2;255;0;0mred\x1b[0m e again"},
	})
	vp.SetStringToHighlight("e")
	vp.HighlightStyle = selectionStyle
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"lin\x1b[38;2;0;0;255me\x1b[0m \x1b[38;2;255;0;0mr\x1b[0m\x1b[38;2;0;0;255me\x1b[0m\x1b[38;2;255;0;0md\x1b[0m \x1b[38;2;0;0;255me\x1b[0m again",
	})
	compare(t, expectedView, vp.View())

	// should not highlight the ansi escape codes themselves
	vp.SetStringToHighlight("38")
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"line \x1b[38;2;255;0;0mred\x1b[0m e again",
	})
	compare(t, expectedView, vp.View())
}

// # SELECTION ENABLED, WRAP OFF

func TestViewport_SelectionOn_WrapOff_Empty(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetSelectionEnabled(true)
	expectedView := pad(vp.width, vp.height, []string{})
	compare(t, expectedView, vp.View())
	vp.SetHeader([]string{"header"})
	expectedView = pad(vp.width, vp.height, []string{"header"})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_SmolDimensions(t *testing.T) {
	w, h := 0, 0
	vp := newViewport(w, h)
	vp.SetSelectionEnabled(true)
	vp.SetHeader([]string{"header"})
	vp.SetContent([]RenderableString{{Content: "hi"}})
	expectedView := pad(vp.width, vp.height, []string{""})
	compare(t, expectedView, vp.View())

	vp.SetWidth(1)
	vp.SetHeight(1)
	expectedView = pad(vp.width, vp.height, []string{"."})
	compare(t, expectedView, vp.View())

	vp.SetWidth(2)
	vp.SetHeight(2)
	expectedView = pad(vp.width, vp.height, []string{"..", ""})
	compare(t, expectedView, vp.View())

	vp.SetWidth(3)
	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{
		"...",
		"\x1b[38;2;0;0;255mhi\x1b[0m",
		"...",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_Basic(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "first line"},
		{Content: renderer().NewStyle().Foreground(red).Render("second") + " line"},
		{Content: renderer().NewStyle().Foreground(red).Render("a really really long line")},
		{Content: renderer().NewStyle().Foreground(red).Render("a") + " really really long line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[0m",
		"\x1b[38;2;255;0;0msecond\x1b[0m line",
		"\x1b[38;2;255;0;0ma really rea...\x1b[0m",
		"\x1b[38;2;255;0;0ma\x1b[0m really rea...",
		"25% (1/4)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_GetConfigs(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
	})
	if selectionEnabled := vp.GetSelectionEnabled(); !selectionEnabled {
		t.Errorf("expected selection to be enabled, got %v", selectionEnabled)
	}
	if wrapText := vp.GetWrapText(); wrapText {
		t.Errorf("expected text wrapping to be disabled, got %v", wrapText)
	}
	if selectedItemIdx := vp.GetSelectedItemIdx(); selectedItemIdx != 0 {
		t.Errorf("expected selected item index to be 0, got %v", selectedItemIdx)
	}
	vp, _ = vp.Update(downKeyMsg)
	if selectedItemIdx := vp.GetSelectedItemIdx(); selectedItemIdx != 1 {
		t.Errorf("expected selected item index to be 1, got %v", selectedItemIdx)
	}
	if selectedItem := vp.GetSelectedItem(); selectedItem.Render() != "second" {
		t.Errorf("got unexpected selected item: %v", selectedItem)
	}
}

func TestViewport_SelectionOn_WrapOff_ShowFooter(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "first line"},
		{Content: renderer().NewStyle().Foreground(red).Render("second") + " line"},
		{Content: renderer().NewStyle().Foreground(red).Render("a really really long line")},
		{Content: renderer().NewStyle().Foreground(red).Render("a") + " really really long line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[0m",
		"\x1b[38;2;255;0;0msecond\x1b[0m line",
		"\x1b[38;2;255;0;0ma really rea...\x1b[0m",
		"25% (1/4)",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[0m",
		"\x1b[38;2;255;0;0msecond\x1b[0m line",
		"\x1b[38;2;255;0;0ma really rea...\x1b[0m",
		"\x1b[38;2;255;0;0ma\x1b[0m really rea...",
		"25% (1/4)",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(7)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[0m",
		"\x1b[38;2;255;0;0msecond\x1b[0m line",
		"\x1b[38;2;255;0;0ma really rea...\x1b[0m",
		"\x1b[38;2;255;0;0ma\x1b[0m really rea...",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_FooterStyle(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.FooterStyle = renderer().NewStyle().Foreground(red)
	vp.SetContent([]RenderableString{
		{Content: "1"},
		{Content: "2"},
		{Content: "3"},
		{Content: "4"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255m1\x1b[0m",
		"2",
		"3",
		"\x1b[38;2;255;0;0m25% (1/4)\x1b[0m",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_FooterDisabled(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "first line"},
		{Content: "second line"},
		{Content: "third line"},
		{Content: "fourth line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[0m",
		"second line",
		"third line",
		"25% (1/4)",
	})
	compare(t, expectedView, vp.View())

	vp.SetFooterVisible(false)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[0m",
		"second line",
		"third line",
		"fourth line",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_SpaceAround(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "    first line     "},
		{Content: "          first line          "},
		{Content: "               first line               "},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255m    first li...\x1b[0m",
		"          fi...",
		"            ...",
		"33% (1/3)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_MultiHeader(t *testing.T) {
	w, h := 15, 2
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header1", "header2"})
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "line1"},
		{Content: "line2"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header1",
		"header2",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(4)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"\x1b[38;2;0;0;255mline1\x1b[0m",
		"50% (1/2)",
	})
	compare(t, expectedView, vp.View())

	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"\x1b[38;2;0;0;255mline2\x1b[0m",
		"100% (2/2)",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(5)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line1",
		"\x1b[38;2;0;0;255mline2\x1b[0m",
		"100% (2/2)",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line1",
		"\x1b[38;2;0;0;255mline2\x1b[0m",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_OverflowLine(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"long header overflows"})
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "123456789012345"},
		{Content: "1234567890123456"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"long header ...",
		"\x1b[38;2;0;0;255m123456789012345\x1b[0m",
		"123456789012...",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_OverflowHeight(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "123456789012345"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255m123456789012345\x1b[0m",
		"123456789012...",
		"123456789012...",
		"123456789012...",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_Scrolling(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent := func() {
		vp.SetContent([]RenderableString{
			{Content: "first"},
			{Content: "second"},
			{Content: "third"},
			{Content: "fourth"},
			{Content: "fifth"},
			{Content: "sixth"},
		})
	}
	validate := func(expectedView string) {
		// set content multiple times to confirm no side effects of doing it
		compare(t, expectedView, vp.View())
		setContent()
		compare(t, expectedView, vp.View())
	}
	setContent()
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
		"second",
		"third",
		"fourth",
		"16% (1/6)",
	})
	validate(expectedView)

	// scrolling up past top is no-op
	vp, _ = vp.Update(upKeyMsg)
	validate(expectedView)

	// scrolling down by one
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"\x1b[38;2;0;0;255msecond\x1b[0m",
		"third",
		"fourth",
		"33% (2/6)",
	})
	validate(expectedView)

	// scrolling to bottom
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"third",
		"fourth",
		"fifth",
		"\x1b[38;2;0;0;255msixth\x1b[0m",
		"100% (6/6)",
	})
	validate(expectedView)

	// scrolling down past bottom when at bottom is no-op
	vp, _ = vp.Update(downKeyMsg)
	validate(expectedView)
}

func TestViewport_SelectionOn_WrapOff_ScrollToItem(t *testing.T) {
	w, h := 15, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
		{Content: "third"},
		{Content: "fourth"},
		{Content: "fifth"},
		{Content: "sixth"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
		"second",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())

	// attempting to scroll so selection out of view is no-op
	vp.ScrollSoItemIdxInView(5)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
		"second",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())

	// move selection down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"\x1b[38;2;0;0;255msecond\x1b[0m",
		"33% (2/6)",
	})
	compare(t, expectedView, vp.View())

	// scroll so third item in view
	vp.ScrollSoItemIdxInView(2)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255msecond\x1b[0m",
		"third",
		"33% (2/6)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_BulkScrolling(t *testing.T) {
	w, h := 15, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
		{Content: "third"},
		{Content: "fourth"},
		{Content: "fifth"},
		{Content: "sixth"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
		"second",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthird\x1b[0m",
		"fourth",
		"50% (3/6)",
	})
	compare(t, expectedView, vp.View())

	// half page down
	vp, _ = vp.Update(halfPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfourth\x1b[0m",
		"fifth",
		"66% (4/6)",
	})
	compare(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"fifth",
		"\x1b[38;2;0;0;255msixth\x1b[0m",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"third",
		"\x1b[38;2;0;0;255mfourth\x1b[0m",
		"66% (4/6)",
	})
	compare(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"second",
		"\x1b[38;2;0;0;255mthird\x1b[0m",
		"50% (3/6)",
	})
	compare(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"\x1b[38;2;0;0;255msecond\x1b[0m",
		"33% (2/6)",
	})
	compare(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
		"second",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())

	// go to bottom
	vp, _ = vp.Update(goToBottomKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"fifth",
		"\x1b[38;2;0;0;255msixth\x1b[0m",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())

	// go to top
	vp, _ = vp.Update(goToTopKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
		"second",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_Panning(t *testing.T) {
	w, h := 10, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header long"})
	vp.SetSelectionEnabled(true)
	setContent := func() {
		vp.SetContent([]RenderableString{
			{Content: "first line that is fairly long"},
			{Content: "second line that is even much longer than the first"},
			{Content: "third line that is fairly long"},
			{Content: "fourth"},
			{Content: "fifth line that is fairly long"},
			{Content: "sixth"},
		})
	}
	validate := func(expectedView string) {
		// set content multiple times to confirm no side effects of doing it
		compare(t, expectedView, vp.View())
		setContent()
		compare(t, expectedView, vp.View())
	}
	setContent()
	expectedView := pad(vp.width, vp.height, []string{
		"header ...",
		"\x1b[38;2;0;0;255mfirst l...\x1b[0m",
		"second ...",
		"third l...",
		"fourth",
		"16% (1/6)",
	})
	validate(expectedView)

	// pan right
	vp.safelySetXOffset(5)
	expectedView = pad(vp.width, vp.height, []string{
		"header ...",
		"\x1b[38;2;0;0;255m...ne t...\x1b[0m",
		"...ine ...",
		"...ne t...",
		".",
		"16% (1/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header ...",
		"...ne t...",
		"\x1b[38;2;0;0;255m...ine ...\x1b[0m",
		"...ne t...",
		".",
		"33% (2/6)",
	})
	validate(expectedView)

	// pan all the way right
	vp.safelySetXOffset(41)
	expectedView = pad(vp.width, vp.height, []string{
		"header ...",
		"...",
		"\x1b[38;2;0;0;255m...e first\x1b[0m",
		"...",
		"...",
		"33% (2/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header ...",
		"...",
		"...e first",
		"\x1b[38;2;0;0;255m...\x1b[0m",
		"...",
		"50% (3/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header ...",
		"...",
		"...e first",
		"...",
		"\x1b[38;2;0;0;255m...\x1b[0m",
		"66% (4/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header ...",
		"...e first",
		"...",
		"...",
		"\x1b[38;2;0;0;255m...\x1b[0m",
		"83% (5/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header ...",
		"...ly long",
		"...",
		"...ly long",
		"\x1b[38;2;0;0;255m...\x1b[0m",
		"100% (6/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header ...",
		"...ly long",
		"...",
		"\x1b[38;2;0;0;255m...ly long\x1b[0m",
		"...",
		"83% (5/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header ...",
		"...ly long",
		"\x1b[38;2;0;0;255m...\x1b[0m",
		"...ly long",
		"...",
		"66% (4/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header ...",
		"\x1b[38;2;0;0;255m...ly long\x1b[0m",
		"...",
		"...ly long",
		"...",
		"50% (3/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header ...",
		"\x1b[38;2;0;0;255m...n mu...\x1b[0m",
		"...ly long",
		"...",
		"...ly long",
		"33% (2/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header ...",
		"\x1b[38;2;0;0;255m...ly long\x1b[0m",
		"...n mu...",
		"...ly long",
		"...",
		"16% (1/6)",
	})
	validate(expectedView)

	// set shorter content
	vp.SetContent([]RenderableString{
		{Content: "the first one"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header ...",
		"\x1b[38;2;0;0;255m...rst one\x1b[0m",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_MaintainSelection(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetMaintainSelection(true)
	vp.SetContent([]RenderableString{
		{Content: "sixth"},
		{Content: "seventh"},
		{Content: "eighth"},
		{Content: "ninth"},
		{Content: "tenth"},
		{Content: "eleventh"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255msixth\x1b[0m",
		"seventh",
		"eighth",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())

	// selection down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"sixth",
		"\x1b[38;2;0;0;255mseventh\x1b[0m",
		"eighth",
		"33% (2/6)",
	})
	compare(t, expectedView, vp.View())

	// add content above
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
		{Content: "third"},
		{Content: "fourth"},
		{Content: "fifth"},
		{Content: "sixth"},
		{Content: "seventh"},
		{Content: "eighth"},
		{Content: "ninth"},
		{Content: "tenth"},
		{Content: "eleventh"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"sixth",
		"\x1b[38;2;0;0;255mseventh\x1b[0m",
		"eighth",
		"63% (7/11)",
	})
	compare(t, expectedView, vp.View())

	// add content below
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
		{Content: "third"},
		{Content: "fourth"},
		{Content: "fifth"},
		{Content: "sixth"},
		{Content: "seventh"},
		{Content: "eighth"},
		{Content: "ninth"},
		{Content: "tenth"},
		{Content: "eleventh"},
		{Content: "twelfth"},
		{Content: "thirteenth"},
		{Content: "fourteenth"},
		{Content: "fifteenth"},
		{Content: "sixteenth"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"sixth",
		"\x1b[38;2;0;0;255mseventh\x1b[0m",
		"eighth",
		"43% (7/16)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_StickyTop(t *testing.T) {
	w, h := 15, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	// stickyness should override maintain selection
	vp.SetMaintainSelection(true)
	vp.SetTopSticky(true)
	vp.SetContent([]RenderableString{
		{Content: "first"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
	})
	compare(t, expectedView, vp.View())

	// add content
	vp.SetContent([]RenderableString{
		{Content: "second"},
		{Content: "first"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255msecond\x1b[0m",
		"first",
		"50% (1/2)",
	})
	compare(t, expectedView, vp.View())

	// de-activate by moving selection down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"second",
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
		"100% (2/2)",
	})
	compare(t, expectedView, vp.View())

	// add content
	vp.SetContent([]RenderableString{
		{Content: "second"},
		{Content: "first"},
		{Content: "third"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"second",
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
		"66% (2/3)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_StickyBottom(t *testing.T) {
	w, h := 15, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	// stickyness should override maintain selection
	vp.SetMaintainSelection(true)
	vp.SetBottomSticky(true)
	vp.SetContent([]RenderableString{
		{Content: "first"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
	})
	compare(t, expectedView, vp.View())

	// add content
	vp.SetContent([]RenderableString{
		{Content: "second"},
		{Content: "first"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"second",
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
		"100% (2/2)",
	})
	compare(t, expectedView, vp.View())

	// de-activate by moving selection up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255msecond\x1b[0m",
		"first",
		"50% (1/2)",
	})
	compare(t, expectedView, vp.View())

	// add content
	vp.SetContent([]RenderableString{
		{Content: "second"},
		{Content: "first"},
		{Content: "third"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255msecond\x1b[0m",
		"first",
		"33% (1/3)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_StickyBottomOverflowHeight(t *testing.T) {
	w, h := 15, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	// stickyness should override maintain selection
	vp.SetMaintainSelection(true)
	vp.SetBottomSticky(true)

	// test covers case where first set content to empty, then overflow height
	vp.SetContent([]RenderableString{})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
	})
	compare(t, expectedView, vp.View())

	vp.SetContent([]RenderableString{
		{Content: "second"},
		{Content: "first"},
		{Content: "third"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"\x1b[38;2;0;0;255mthird\x1b[0m",
		"100% (3/3)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_StickyTopBottom(t *testing.T) {
	w, h := 15, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	// stickyness should override maintain selection
	vp.SetMaintainSelection(true)
	vp.SetTopSticky(true)
	vp.SetBottomSticky(true)
	vp.SetContent([]RenderableString{
		{Content: "first"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
	})
	compare(t, expectedView, vp.View())

	// add content, top sticky wins out arbitrarily when both set
	vp.SetContent([]RenderableString{
		{Content: "second"},
		{Content: "first"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255msecond\x1b[0m",
		"first",
		"50% (1/2)",
	})
	compare(t, expectedView, vp.View())

	// selection to bottom
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"second",
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
		"100% (2/2)",
	})
	compare(t, expectedView, vp.View())

	// add content
	vp.SetContent([]RenderableString{
		{Content: "second"},
		{Content: "first"},
		{Content: "third"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"\x1b[38;2;0;0;255mthird\x1b[0m",
		"100% (3/3)",
	})
	compare(t, expectedView, vp.View())

	// de-activate by moving selection up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
		"third",
		"66% (2/3)",
	})
	compare(t, expectedView, vp.View())

	// add content
	vp.SetContent([]RenderableString{
		{Content: "second"},
		{Content: "first"},
		{Content: "third"},
		{Content: "fourth"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
		"third",
		"50% (2/4)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_RemoveLogsWhenSelectionBottom(t *testing.T) {
	w, h := 10, 3
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)

	// add content
	vp.SetContent([]RenderableString{
		{Content: "second"},
		{Content: "first"},
		{Content: "third"},
		{Content: "fourth"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255msecond\x1b[0m",
		"25% (1/4)",
	})
	compare(t, expectedView, vp.View())

	// selection to bottom
	vp.SetSelectedItemIdx(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfourth\x1b[0m",
		"100% (4/4)",
	})
	compare(t, expectedView, vp.View())

	// remove content
	vp.SetContent([]RenderableString{
		{Content: "second"},
		{Content: "first"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
		"100% (2/2)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_ChangeHeight(t *testing.T) {
	w, h := 10, 3
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
		{Content: "third"},
		{Content: "fourth"},
		{Content: "fifth"},
		{Content: "sixth"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(8)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())

	// move selection to third line
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
		"\x1b[38;2;0;0;255mthird\x1b[0m",
		"fourth",
		"fifth",
		"sixth",
		"50% (3/6)",
	})
	compare(t, expectedView, vp.View())

	// reduce height
	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthird\x1b[0m",
		"50% (3/6)",
	})
	compare(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(8)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
		"\x1b[38;2;0;0;255mthird\x1b[0m",
		"fourth",
		"fifth",
		"sixth",
		"50% (3/6)",
	})
	compare(t, expectedView, vp.View())

	// move selection to last line
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"\x1b[38;2;0;0;255msixth\x1b[0m",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())

	// reduce height
	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255msixth\x1b[0m",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(8)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"\x1b[38;2;0;0;255msixth\x1b[0m",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_ChangeContent(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
		{Content: "third"},
		{Content: "fourth"},
		{Content: "fifth"},
		{Content: "sixth"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
		"second",
		"third",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())

	// move selection to bottom
	vp.SetSelectedItemIdx(5)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"fourth",
		"fifth",
		"\x1b[38;2;0;0;255msixth\x1b[0m",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())

	// remove content
	vp.SetContent([]RenderableString{
		{Content: "second"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255msecond\x1b[0m",
	})
	compare(t, expectedView, vp.View())

	// remove all content
	vp.SetContent([]RenderableString{})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
	})
	compare(t, expectedView, vp.View())

	// add content (maintain selection off)
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
		{Content: "third"},
		{Content: "fourth"},
		{Content: "fifth"},
		{Content: "sixth"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
		"second",
		"third",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_StringToHighlight(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetStringToHighlight("second")
	vp.HighlightStyle = renderer().NewStyle().Foreground(green)
	vp.HighlightStyleIfSelected = renderer().NewStyle().Foreground(red)
	vp.SetContent([]RenderableString{
		{Content: "the first line"},
		{Content: "the second line"},
		{Content: "the second line"},
		{Content: "the fourth line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first line\x1b[0m",
		"the \x1b[38;2;0;255;0msecond\x1b[0m line",
		"the \x1b[38;2;0;255;0msecond\x1b[0m line",
		"25% (1/4)",
	})
	compare(t, expectedView, vp.View())

	vp.SetStringToHighlight("first")
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe \x1b[0m\x1b[38;2;255;0;0mfirst\x1b[0m\x1b[38;2;0;0;255m line\x1b[0m",
		"the second line",
		"the second line",
		"25% (1/4)",
	})
	compare(t, expectedView, vp.View())

	vp.SetContent([]RenderableString{
		{Content: "first line"},
		{Content: "second line"},
		{Content: "second line"},
		{Content: "fourth line"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;255;0;0mfirst\x1b[0m\x1b[38;2;0;0;255m line\x1b[0m",
		"second line",
		"second line",
		"25% (1/4)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_AnsiOnSelection(t *testing.T) {
	w, h := 20, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "line with \x1b[38;2;255;0;0mred\x1b[0m text"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mline with \x1b[0m\x1b[38;2;255;0;0mred\x1b[0m\x1b[38;2;0;0;255m text\x1b[0m",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_SelectionEmpty(t *testing.T) {
	w, h := 20, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: ""},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255m \x1b[0m",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_ExtraSlash(t *testing.T) {
	w, h := 25, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "|2024|\x1b[38;2;0mfl..lq\x1b[0m/\x1b[38;2;0mflask-3\x1b[0m|"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255m|2024|\x1b[0m\x1b[38;2;0mfl..lq\x1b[0m\x1b[38;2;0;0;255m/\x1b[0m\x1b[38;2;0mflask-3\x1b[0m\x1b[38;2;0;0;255m|\x1b[0m",
	})
	compare(t, expectedView, vp.View())
}

// # SELECTION DISABLED, WRAP ON

func TestViewport_SelectionOff_WrapOn_Empty(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	expectedView := pad(vp.width, vp.height, []string{})
	compare(t, expectedView, vp.View())
	vp.SetHeader([]string{"header"})
	expectedView = pad(vp.width, vp.height, []string{"header"})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_SmolDimensions(t *testing.T) {
	w, h := 0, 0
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	vp.SetHeader([]string{"header"})
	vp.SetContent([]RenderableString{{Content: "hi"}})
	expectedView := pad(vp.width, vp.height, []string{""})
	compare(t, expectedView, vp.View())

	vp.SetWidth(1)
	vp.SetHeight(1)
	expectedView = pad(vp.width, vp.height, []string{"h"})
	compare(t, expectedView, vp.View())

	vp.SetWidth(2)
	vp.SetHeight(2)
	expectedView = pad(vp.width, vp.height, []string{"he", "ad"})
	compare(t, expectedView, vp.View())

	vp.SetWidth(3)
	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{"hea", "der", ""})
	compare(t, expectedView, vp.View())

	vp.SetWidth(4)
	vp.SetHeight(4)
	expectedView = pad(vp.width, vp.height, []string{"head", "er", "hi", "1..."})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_Basic(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "first line"},
		{Content: renderer().NewStyle().Foreground(red).Render("second") + " line"},
		{Content: renderer().NewStyle().Foreground(red).Render("a really really long line")},
		{Content: renderer().NewStyle().Foreground(red).Render("a") + " really really long line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond\x1b[0m line",
		"\x1b[38;2;255;0;0ma really really\x1b[0m",
		"\x1b[38;2;255;0;0m long line\x1b[0m",
		"75% (3/4)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_GetConfigs(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
	})
	if selectionEnabled := vp.GetSelectionEnabled(); selectionEnabled {
		t.Errorf("expected selection to be disabled, got %v", selectionEnabled)
	}
	if wrapText := vp.GetWrapText(); !wrapText {
		t.Errorf("expected text wrapping to be enabled, got %v", wrapText)
	}
	if selectedItemIdx := vp.GetSelectedItemIdx(); selectedItemIdx != 0 {
		t.Errorf("expected selected item index to be 0, got %v", selectedItemIdx)
	}
	if selectedItem := vp.GetSelectedItem(); selectedItem != nil {
		t.Errorf("expected selected item to be nil, got %v", selectedItem)
	}
}

func TestViewport_SelectionOff_WrapOn_ShowFooter(t *testing.T) {
	w, h := 15, 7
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "first line"},
		{Content: renderer().NewStyle().Foreground(red).Render("second") + " line"},
		{Content: renderer().NewStyle().Foreground(red).Render("a really really long line")},
		{Content: renderer().NewStyle().Foreground(red).Render("a") + " really really long line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond\x1b[0m line",
		"\x1b[38;2;255;0;0ma really really\x1b[0m",
		"\x1b[38;2;255;0;0m long line\x1b[0m",
		"\x1b[38;2;255;0;0ma\x1b[0m really really",
		"99% (4/4)",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(8)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond\x1b[0m line",
		"\x1b[38;2;255;0;0ma really really\x1b[0m",
		"\x1b[38;2;255;0;0m long line\x1b[0m",
		"\x1b[38;2;255;0;0ma\x1b[0m really really",
		"\x1b[38;2;255;0;0m\x1b[0m long line",
		"100% (4/4)",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(9)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond\x1b[0m line",
		"\x1b[38;2;255;0;0ma really really\x1b[0m",
		"\x1b[38;2;255;0;0m long line\x1b[0m",
		"\x1b[38;2;255;0;0ma\x1b[0m really really",
		"\x1b[38;2;255;0;0m\x1b[0m long line",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_FooterStyle(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.FooterStyle = renderer().NewStyle().Foreground(red)
	vp.SetContent([]RenderableString{
		{Content: "1"},
		{Content: "2"},
		{Content: "3"},
		{Content: "4"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"1",
		"2",
		"3",
		"\x1b[38;2;255;0;0m75% (3/4)\x1b[0m",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_FooterDisabled(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "first line"},
		{Content: "second line"},
		{Content: "third line"},
		{Content: "fourth line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"second line",
		"third line",
		"75% (3/4)",
	})
	compare(t, expectedView, vp.View())

	vp.SetFooterVisible(false)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"second line",
		"third line",
		"fourth line",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_SpaceAround(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "    first line     "},
		{Content: "          first line          "},
		{Content: "               first line               "},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"    first line ",
		"          first",
		" line",
		"66% (2/3)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_MultiHeader(t *testing.T) {
	w, h := 15, 2
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header1", "header2"})
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "line1"},
		{Content: "line2"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header1",
		"header2",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(4)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line1",
		"50% (1/2)",
	})
	compare(t, expectedView, vp.View())

	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line2",
		"100% (2/2)",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(5)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line1",
		"line2",
		"100% (2/2)",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line1",
		"line2",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_OverflowLine(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"long header overflows"})
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "123456789012345"},
		{Content: "1234567890123456"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"long header ove",
		"rflows",
		"123456789012345",
		"123456789012345",
		"6",
		"100% (2/2)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_OverflowHeight(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "123456789012345"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"123456789012345",
		"123456789012345",
		"6",
		"123456789012345",
		"50% (3/6)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_Scrolling(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent := func() {
		vp.SetContent([]RenderableString{
			{Content: "first"},
			{Content: "second"},
			{Content: "third"},
			{Content: "fourth"},
			{Content: "fifth"},
			{Content: "sixth"},
		})
	}
	validate := func(expectedView string) {
		// set content multiple times to confirm no side effects of doing it
		compare(t, expectedView, vp.View())
		setContent()
		compare(t, expectedView, vp.View())
	}
	setContent()
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
		"third",
		"fourth",
		"66% (4/6)",
	})
	validate(expectedView)

	// scrolling up past top is no-op
	vp, _ = vp.Update(upKeyMsg)
	validate(expectedView)

	// scrolling down by one
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"second",
		"third",
		"fourth",
		"fifth",
		"83% (5/6)",
	})
	validate(expectedView)

	// scrolling down by one again
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"third",
		"fourth",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	validate(expectedView)

	// scrolling down past bottom when at bottom is no-op
	vp, _ = vp.Update(downKeyMsg)
	validate(expectedView)
}

func TestViewport_SelectionOff_WrapOn_ScrollToItem(t *testing.T) {
	w, h := 10, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "the first line"},
		{Content: "the second line"},
		{Content: "the third line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"the first",
		"line",
		"33% (1/3)",
	})
	compare(t, expectedView, vp.View())

	// scroll so last item in view
	vp.ScrollSoItemIdxInView(2)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the third",
		"line",
		"100% (3/3)",
	})
	compare(t, expectedView, vp.View())

	// scroll so second item in view
	vp.ScrollSoItemIdxInView(1)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the second",
		" line",
		"66% (2/3)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_BulkScrolling(t *testing.T) {
	w, h := 10, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "the first line"},
		{Content: "the second line"},
		{Content: "the third line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"the first",
		"line",
		"33% (1/3)",
	})
	compare(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the second",
		" line",
		"66% (2/3)",
	})
	compare(t, expectedView, vp.View())

	// half page down
	vp, _ = vp.Update(halfPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		" line",
		"the third ",
		"99% (3/3)",
	})
	compare(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the third ",
		"line",
		"100% (3/3)",
	})
	compare(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the second",
		" line",
		"66% (2/3)",
	})
	compare(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"line",
		"the second",
		"66% (2/3)",
	})
	compare(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the first",
		"line",
		"33% (1/3)",
	})
	compare(t, expectedView, vp.View())

	// go to bottom
	vp, _ = vp.Update(goToBottomKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the third ",
		"line",
		"100% (3/3)",
	})
	compare(t, expectedView, vp.View())

	// go to top
	vp, _ = vp.Update(goToTopKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the first",
		"line",
		"33% (1/3)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_Panning(t *testing.T) {
	w, h := 10, 7
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header long"})
	vp.SetWrapText(true)
	setContent := func() {
		vp.SetContent([]RenderableString{
			{Content: "first line that is fairly long"},
			{Content: "second line that is even much longer than the first"},
			{Content: "third line that is fairly long"},
			{Content: "fourth"},
			{Content: "fifth line that is fairly long"},
			{Content: "sixth"},
		})
	}
	validate := func(expectedView string) {
		// set content multiple times to confirm no side effects of doing it
		compare(t, expectedView, vp.View())
		setContent()
		compare(t, expectedView, vp.View())
	}
	setContent()
	expectedView := pad(vp.width, vp.height, []string{
		"header lon",
		"g",
		"first line",
		" that is f",
		"airly long",
		"second lin",
		"33% (2/6)",
	})
	validate(expectedView)

	// pan right
	vp.safelySetXOffset(5)
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header lon",
		"g",
		" that is f",
		"airly long",
		"second lin",
		"e that is ",
		"33% (2/6)",
	})
	validate(expectedView)

	// pan all the way right
	vp.safelySetXOffset(41)
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header lon",
		"g",
		"airly long",
		"second lin",
		"e that is",
		"even much",
		"33% (2/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header lon",
		"g",
		"second lin",
		"e that is ",
		"even much ",
		"longer tha",
		"33% (2/6)",
	})
	validate(expectedView)
}

func TestViewport_SelectionOff_WrapOn_ChangeHeight(t *testing.T) {
	w, h := 10, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "the first line"},
		{Content: "the second line"},
		{Content: "the third line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"the first",
		"line",
		"33% (1/3)",
	})
	compare(t, expectedView, vp.View())

	// scroll down to bottom
	vp, _ = vp.Update(fullPgDownKeyMsg)
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the third",
		"line",
		"100% (3/3)",
	})
	compare(t, expectedView, vp.View())

	// reduce height
	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the third",
		"99% (3/3)",
	})
	compare(t, expectedView, vp.View())

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"line",
		"100% (3/3)",
	})
	compare(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(8)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the first",
		"line",
		"the second",
		" line",
		"the third",
		"line",
		"100% (3/3)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_ChangeContent(t *testing.T) {
	w, h := 10, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "the first line"},
		{Content: "the second line"},
		{Content: "the third line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"the first",
		"line",
		"33% (1/3)",
	})
	compare(t, expectedView, vp.View())

	// scroll down to bottom
	vp, _ = vp.Update(fullPgDownKeyMsg)
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the third",
		"line",
		"100% (3/3)",
	})
	compare(t, expectedView, vp.View())

	// remove content
	vp.SetContent([]RenderableString{
		{Content: "the first line"},
		{Content: "the second line"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the second",
		" line",
		"100% (2/2)",
	})
	compare(t, expectedView, vp.View())

	// add content
	vp.SetContent([]RenderableString{
		{Content: "the first line"},
		{Content: "the second line"},
		{Content: "the third line"},
		{Content: "the fourth line"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the second",
		" line",
		"50% (2/4)",
	})
	compare(t, expectedView, vp.View())

	// remove all content
	vp.SetContent([]RenderableString{})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_StringToHighlight(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetStringToHighlight("second")
	vp.HighlightStyle = renderer().NewStyle().Foreground(red)
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
		{Content: "second"},
		{Content: "third"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first",
		"\x1b[38;2;255;0;0msecond\x1b[0m",
		"\x1b[38;2;255;0;0msecond\x1b[0m",
		"75% (3/4)",
	})
	compare(t, expectedView, vp.View())

	vp.SetContent([]RenderableString{
		{Content: "averylongwordthatwraps"},
	})
	vp.SetStringToHighlight("wraps")
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"averylongw\x1b[38;2;255;0;0m\x1b[0m", // no-op artifact of applyAnsi, maybe fix later
		"ordthat\x1b[38;2;255;0;0mwra\x1b[0m",
		"\x1b[38;2;255;0;0mps\x1b[0m",
		"100% (1/1)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_StringToHighlightAnsi(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "line \x1b[38;2;255;0;0mred\x1b[0m e again"},
	})
	vp.SetStringToHighlight("e")
	vp.HighlightStyle = selectionStyle
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"lin\x1b[38;2;0;0;255me\x1b[0m \x1b[38;2;255;0;0mr\x1b[0m\x1b[38;2;0;0;255me\x1b[0m\x1b[38;2;255;0;0md\x1b[0m \x1b[38;2;0;0;255me\x1b[0m",
		"\x1b[38;2;0;0;255m\x1b[0m\x1b[38;2;255;0;0m\x1b[0m\x1b[38;2;0;0;255m\x1b[0m\x1b[38;2;255;0;0m\x1b[0m\x1b[38;2;0;0;255m\x1b[0m again",
	})
	compare(t, expectedView, vp.View())

	// should not highlight the ansi escape codes themselves
	vp.SetStringToHighlight("38")
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"line \x1b[38;2;255;0;0mred\x1b[0m e",
		"\x1b[38;2;255;0;0m\x1b[0m again",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_SuperLongWrappedLine(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "smol"},
		{Content: strings.Repeat("12345678", 1000000)},
		{Content: "smol"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"smol",
		"1234567812",
		"3456781234",
		"66% (2/3)",
	})
	compare(t, expectedView, vp.View())

	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"1234567812",
		"3456781234",
		"5678123456",
		"66% (2/3)",
	})
	compare(t, expectedView, vp.View())

	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"3456781234",
		"5678123456",
		"7812345678",
		"66% (2/3)",
	})
	compare(t, expectedView, vp.View())

	vp, _ = vp.Update(goToBottomKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"5678123456",
		"7812345678",
		"smol",
		"100% (3/3)",
	})
	compare(t, expectedView, vp.View())
}

// # SELECTION ENABLED, WRAP ON

func TestViewport_SelectionOn_WrapOn_Empty(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	expectedView := pad(vp.width, vp.height, []string{})
	compare(t, expectedView, vp.View())
	vp.SetHeader([]string{"header"})
	expectedView = pad(vp.width, vp.height, []string{"header"})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_SmolDimensions(t *testing.T) {
	w, h := 0, 0
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	vp.SetHeader([]string{"header"})
	vp.SetContent([]RenderableString{{Content: "hi"}})
	expectedView := pad(vp.width, vp.height, []string{""})
	compare(t, expectedView, vp.View())

	vp.SetWidth(1)
	vp.SetHeight(1)
	expectedView = pad(vp.width, vp.height, []string{"h"})
	compare(t, expectedView, vp.View())

	vp.SetWidth(2)
	vp.SetHeight(2)
	expectedView = pad(vp.width, vp.height, []string{"he", "ad"})
	compare(t, expectedView, vp.View())

	vp.SetWidth(3)
	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{"hea", "der", ""})
	compare(t, expectedView, vp.View())

	vp.SetWidth(4)
	vp.SetHeight(4)
	expectedView = pad(vp.width, vp.height, []string{
		"head",
		"er",
		"\x1b[38;2;0;0;255mhi\x1b[0m",
		"1...",
	})
	compare(t, expectedView, vp.View())

}

func TestViewport_SelectionOn_WrapOn_Basic(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "first line"},
		{Content: renderer().NewStyle().Foreground(red).Render("second") + " line"},
		{Content: renderer().NewStyle().Foreground(red).Render("a really really long line")},
		{Content: renderer().NewStyle().Foreground(red).Render("a") + " really really long line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[0m",
		"\x1b[38;2;255;0;0msecond\x1b[0m line",
		"\x1b[38;2;255;0;0ma really really\x1b[0m",
		"\x1b[38;2;255;0;0m long line\x1b[0m",
		"25% (1/4)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_GetConfigs(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
	})
	if selectionEnabled := vp.GetSelectionEnabled(); !selectionEnabled {
		t.Errorf("expected selection to be enabled, got %v", selectionEnabled)
	}
	if wrapText := vp.GetWrapText(); !wrapText {
		t.Errorf("expected text wrapping to be enabled, got %v", wrapText)
	}
	if selectedItemIdx := vp.GetSelectedItemIdx(); selectedItemIdx != 0 {
		t.Errorf("expected selected item index to be 0, got %v", selectedItemIdx)
	}
	vp, _ = vp.Update(downKeyMsg)
	if selectedItemIdx := vp.GetSelectedItemIdx(); selectedItemIdx != 1 {
		t.Errorf("expected selected item index to be 1, got %v", selectedItemIdx)
	}
	if selectedItem := vp.GetSelectedItem(); selectedItem.Render() != "second" {
		t.Errorf("got unexpected selected item: %v", selectedItem)
	}
}

func TestViewport_SelectionOn_WrapOn_ShowFooter(t *testing.T) {
	w, h := 15, 7
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "first line"},
		{Content: renderer().NewStyle().Foreground(red).Render("second") + " line"},
		{Content: renderer().NewStyle().Foreground(red).Render("a really really long line")},
		{Content: renderer().NewStyle().Foreground(red).Render("a") + " really really long line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[0m",
		"\x1b[38;2;255;0;0msecond\x1b[0m line",
		"\x1b[38;2;255;0;0ma really really\x1b[0m",
		"\x1b[38;2;255;0;0m long line\x1b[0m",
		"\x1b[38;2;255;0;0ma\x1b[0m really really",
		"25% (1/4)",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(8)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[0m",
		"\x1b[38;2;255;0;0msecond\x1b[0m line",
		"\x1b[38;2;255;0;0ma really really\x1b[0m",
		"\x1b[38;2;255;0;0m long line\x1b[0m",
		"\x1b[38;2;255;0;0ma\x1b[0m really really",
		"\x1b[38;2;255;0;0m\x1b[0m long line",
		"25% (1/4)",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(9)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[0m",
		"\x1b[38;2;255;0;0msecond\x1b[0m line",
		"\x1b[38;2;255;0;0ma really really\x1b[0m",
		"\x1b[38;2;255;0;0m long line\x1b[0m",
		"\x1b[38;2;255;0;0ma\x1b[0m really really",
		"\x1b[38;2;255;0;0m\x1b[0m long line",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_FooterStyle(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	vp.FooterStyle = renderer().NewStyle().Foreground(red)
	vp.SetContent([]RenderableString{
		{Content: "1"},
		{Content: "2"},
		{Content: "3"},
		{Content: "4"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255m1\x1b[0m",
		"2",
		"3",
		"\x1b[38;2;255;0;0m25% (1/4)\x1b[0m",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_FooterDisabled(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "first line"},
		{Content: "second line"},
		{Content: "third line"},
		{Content: "fourth line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[0m",
		"second line",
		"third line",
		"25% (1/4)",
	})
	compare(t, expectedView, vp.View())

	vp.SetFooterVisible(false)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[0m",
		"second line",
		"third line",
		"fourth line",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_SpaceAround(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "    first line     "},
		{Content: "          first line          "},
		{Content: "               first line               "},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255m    first line\x1b[0m",
		"          first",
		" line",
		"33% (1/3)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_MultiHeader(t *testing.T) {
	w, h := 15, 2
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header1", "header2"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "line1"},
		{Content: "line2"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header1",
		"header2",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(4)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"\x1b[38;2;0;0;255mline1\x1b[0m",
		"50% (1/2)",
	})
	compare(t, expectedView, vp.View())

	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"\x1b[38;2;0;0;255mline2\x1b[0m",
		"100% (2/2)",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(5)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line1",
		"\x1b[38;2;0;0;255mline2\x1b[0m",
		"100% (2/2)",
	})
	compare(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line1",
		"\x1b[38;2;0;0;255mline2\x1b[0m",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_OverflowLine(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"long header overflows"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "123456789012345"},
		{Content: "1234567890123456"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"long header ove",
		"rflows",
		"\x1b[38;2;0;0;255m123456789012345\x1b[0m",
		"123456789012345",
		"6",
		"50% (1/2)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_OverflowHeight(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "123456789012345"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
	})
	vp.SetSelectedItemIdx(1)
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"123456789012345",
		"\x1b[38;2;0;0;255m123456789012345\x1b[0m",
		"\x1b[38;2;0;0;255m6\x1b[0m",
		"123456789012345",
		"33% (2/6)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_Scrolling(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent := func() {
		vp.SetContent([]RenderableString{
			{Content: "first"},
			{Content: "second"},
			{Content: "third"},
			{Content: "fourth"},
			{Content: "fifth"},
			{Content: "sixth"},
		})
	}
	validate := func(expectedView string) {
		// set content multiple times to confirm no side effects of doing it
		compare(t, expectedView, vp.View())
		setContent()
		compare(t, expectedView, vp.View())
	}
	setContent()
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
		"second",
		"third",
		"fourth",
		"16% (1/6)",
	})
	validate(expectedView)

	// scrolling up past top is no-op
	vp, _ = vp.Update(upKeyMsg)
	validate(expectedView)

	// scrolling down by one
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"\x1b[38;2;0;0;255msecond\x1b[0m",
		"third",
		"fourth",
		"33% (2/6)",
	})
	validate(expectedView)

	// scrolling down by one again
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
		"\x1b[38;2;0;0;255mthird\x1b[0m",
		"fourth",
		"50% (3/6)",
	})
	validate(expectedView)

	// scroll to bottom
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"third",
		"fourth",
		"fifth",
		"\x1b[38;2;0;0;255msixth\x1b[0m",
		"100% (6/6)",
	})
	validate(expectedView)

	// scrolling down past bottom when at bottom is no-op
	vp, _ = vp.Update(downKeyMsg)
	validate(expectedView)
}

func TestViewport_SelectionOn_WrapOn_ScrollToItem(t *testing.T) {
	w, h := 10, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "the first line"},
		{Content: "the second line"},
		{Content: "the third line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"the second",
		" line",
		"33% (1/3)",
	})
	compare(t, expectedView, vp.View())

	// attempting to scroll so selection out of view is no-op
	vp.ScrollSoItemIdxInView(2)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"the second",
		" line",
		"33% (1/3)",
	})
	compare(t, expectedView, vp.View())

	// move selection down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the first",
		"line",
		"\x1b[38;2;0;0;255mthe second\x1b[0m",
		"\x1b[38;2;0;0;255m line\x1b[0m",
		"66% (2/3)",
	})
	compare(t, expectedView, vp.View())

	// scroll so third item in view
	vp.ScrollSoItemIdxInView(2)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe second\x1b[0m",
		"\x1b[38;2;0;0;255m line\x1b[0m",
		"the third",
		"line",
		"66% (2/3)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_BulkScrolling(t *testing.T) {
	w, h := 10, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "the first line"},
		{Content: "the second line"},
		{Content: "the third line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"33% (1/3)",
	})
	compare(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe second\x1b[0m",
		"\x1b[38;2;0;0;255m line\x1b[0m",
		"66% (2/3)",
	})
	compare(t, expectedView, vp.View())

	// half page down
	vp, _ = vp.Update(halfPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe third \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"100% (3/3)",
	})
	compare(t, expectedView, vp.View())

	// half page down
	vp, _ = vp.Update(halfPgDownKeyMsg)
	compare(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe second\x1b[0m",
		"\x1b[38;2;0;0;255m line\x1b[0m",
		"66% (2/3)",
	})
	compare(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"33% (1/3)",
	})
	compare(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	compare(t, expectedView, vp.View())

	// go to bottom
	vp, _ = vp.Update(goToBottomKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe third \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"100% (3/3)",
	})
	compare(t, expectedView, vp.View())

	// go to top
	vp, _ = vp.Update(goToTopKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"33% (1/3)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_Panning(t *testing.T) {
	w, h := 10, 7
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header long"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent := func() {
		vp.SetContent([]RenderableString{
			{Content: "first line that is fairly long"},
			{Content: "second line that is even much longer than the first"},
			{Content: "third line that is fairly long as well"},
			{Content: "fourth kinda long"},
			{Content: "fifth kinda long too"},
			{Content: "sixth"},
		})
	}
	validate := func(expectedView string) {
		// set content multiple times to confirm no side effects of doing it
		compare(t, expectedView, vp.View())
		setContent()
		compare(t, expectedView, vp.View())
	}
	setContent()
	expectedView := pad(vp.width, vp.height, []string{
		"header lon",
		"g",
		"\x1b[38;2;0;0;255mfirst line\x1b[0m",
		"\x1b[38;2;0;0;255m that is f\x1b[0m",
		"\x1b[38;2;0;0;255mairly long\x1b[0m",
		"second lin",
		"16% (1/6)",
	})
	validate(expectedView)

	// pan right
	vp.safelySetXOffset(5)
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header lon",
		"g",
		"\x1b[38;2;0;0;255msecond lin\x1b[0m",
		"\x1b[38;2;0;0;255me that is \x1b[0m",
		"\x1b[38;2;0;0;255meven much \x1b[0m",
		"\x1b[38;2;0;0;255mlonger tha\x1b[0m",
		"33% (2/6)",
	})
	validate(expectedView)

	// pan all the way right
	vp.safelySetXOffset(41)
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header lon",
		"g",
		"\x1b[38;2;0;0;255mthird line\x1b[0m",
		"\x1b[38;2;0;0;255m that is f\x1b[0m",
		"\x1b[38;2;0;0;255mairly long\x1b[0m",
		"\x1b[38;2;0;0;255m as well\x1b[0m",
		"50% (3/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header lon",
		"g",
		"airly long",
		" as well",
		"\x1b[38;2;0;0;255mfourth kin\x1b[0m",
		"\x1b[38;2;0;0;255mda long\x1b[0m",
		"66% (4/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header lon",
		"g",
		"fourth kin",
		"da long",
		"\x1b[38;2;0;0;255mfifth kind\x1b[0m",
		"\x1b[38;2;0;0;255ma long too\x1b[0m",
		"83% (5/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header lon",
		"g",
		"da long",
		"fifth kind",
		"a long too",
		"\x1b[38;2;0;0;255msixth\x1b[0m",
		"100% (6/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header lon",
		"g",
		"da long",
		"\x1b[38;2;0;0;255mfifth kind\x1b[0m",
		"\x1b[38;2;0;0;255ma long too\x1b[0m",
		"sixth",
		"83% (5/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header lon",
		"g",
		"\x1b[38;2;0;0;255mfourth kin\x1b[0m",
		"\x1b[38;2;0;0;255mda long\x1b[0m",
		"fifth kind",
		"a long too",
		"66% (4/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header lon",
		"g",
		"\x1b[38;2;0;0;255mthird line\x1b[0m",
		"\x1b[38;2;0;0;255m that is f\x1b[0m",
		"\x1b[38;2;0;0;255mairly long\x1b[0m",
		"\x1b[38;2;0;0;255m as well\x1b[0m",
		"50% (3/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header lon",
		"g",
		"\x1b[38;2;0;0;255msecond lin\x1b[0m",
		"\x1b[38;2;0;0;255me that is \x1b[0m",
		"\x1b[38;2;0;0;255meven much \x1b[0m",
		"\x1b[38;2;0;0;255mlonger tha\x1b[0m",
		"33% (2/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header lon",
		"g",
		"\x1b[38;2;0;0;255mfirst line\x1b[0m",
		"\x1b[38;2;0;0;255m that is f\x1b[0m",
		"\x1b[38;2;0;0;255mairly long\x1b[0m",
		"second lin",
		"16% (1/6)",
	})
	validate(expectedView)
}

func TestViewport_SelectionOn_WrapOn_MaintainSelection(t *testing.T) {
	w, h := 10, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	vp.SetMaintainSelection(true)
	vp.SetContent([]RenderableString{
		{Content: "sixth item"},
		{Content: "seventh item"},
		{Content: "eighth item"},
		{Content: "ninth item"},
		{Content: "tenth item"},
		{Content: "eleventh item"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255msixth item\x1b[0m",
		"seventh it",
		"em",
		"eighth ite",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())

	// selection down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"sixth item",
		"\x1b[38;2;0;0;255mseventh it\x1b[0m",
		"\x1b[38;2;0;0;255mem\x1b[0m",
		"eighth ite",
		"33% (2/6)",
	})
	compare(t, expectedView, vp.View())

	// add content above
	vp.SetContent([]RenderableString{
		{Content: "first item"},
		{Content: "second item"},
		{Content: "third item"},
		{Content: "fourth item"},
		{Content: "fifth item"},
		{Content: "sixth item"},
		{Content: "seventh item"},
		{Content: "eighth item"},
		{Content: "ninth item"},
		{Content: "tenth item"},
		{Content: "eleventh item"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"sixth item",
		"\x1b[38;2;0;0;255mseventh it\x1b[0m",
		"\x1b[38;2;0;0;255mem\x1b[0m",
		"eighth ite",
		"63% (7/11)",
	})
	compare(t, expectedView, vp.View())

	// add content below
	vp.SetContent([]RenderableString{
		{Content: "first item"},
		{Content: "second item"},
		{Content: "third item"},
		{Content: "fourth item"},
		{Content: "fifth item"},
		{Content: "sixth item"},
		{Content: "seventh item"},
		{Content: "eighth item"},
		{Content: "ninth item"},
		{Content: "tenth item"},
		{Content: "eleventh item"},
		{Content: "twelfth item"},
		{Content: "thirteenth item"},
		{Content: "fourteenth item"},
		{Content: "fifteenth item"},
		{Content: "sixteenth item"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"sixth item",
		"\x1b[38;2;0;0;255mseventh it\x1b[0m",
		"\x1b[38;2;0;0;255mem\x1b[0m",
		"eighth ite",
		"43% (7/16)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_StickyTop(t *testing.T) {
	w, h := 10, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	// stickyness should override maintain selection
	vp.SetMaintainSelection(true)
	vp.SetTopSticky(true)
	vp.SetContent([]RenderableString{
		{Content: "the first line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"100% (1/1)",
	})
	compare(t, expectedView, vp.View())

	// add content
	vp.SetContent([]RenderableString{
		{Content: "the second line"},
		{Content: "the first line"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe second\x1b[0m",
		"\x1b[38;2;0;0;255m line\x1b[0m",
		"50% (1/2)",
	})
	compare(t, expectedView, vp.View())

	// de-activate by moving selection down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"100% (2/2)",
	})
	compare(t, expectedView, vp.View())

	// add content
	vp.SetContent([]RenderableString{
		{Content: "the second line"},
		{Content: "the first line"},
		{Content: "the third line"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"66% (2/3)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_StickyBottom(t *testing.T) {
	w, h := 10, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	// stickyness should override maintain selection
	vp.SetMaintainSelection(true)
	vp.SetBottomSticky(true)
	vp.SetContent([]RenderableString{
		{Content: "the first line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"100% (1/1)",
	})
	compare(t, expectedView, vp.View())

	// add content
	vp.SetContent([]RenderableString{
		{Content: "the second line"},
		{Content: "the first line"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"100% (2/2)",
	})
	compare(t, expectedView, vp.View())

	// de-activate by moving selection up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe second\x1b[0m",
		"\x1b[38;2;0;0;255m line\x1b[0m",
		"50% (1/2)",
	})
	compare(t, expectedView, vp.View())

	// add content
	vp.SetContent([]RenderableString{
		{Content: "the second line"},
		{Content: "the first line"},
		{Content: "the third line"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe second\x1b[0m",
		"\x1b[38;2;0;0;255m line\x1b[0m",
		"33% (1/3)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_StickyBottomOverflowHeight(t *testing.T) {
	w, h := 10, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	// stickyness should override maintain selection
	vp.SetMaintainSelection(true)
	vp.SetBottomSticky(true)

	// test covers case where first set content to empty, then overflow height
	vp.SetContent([]RenderableString{})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
	})
	compare(t, expectedView, vp.View())

	vp.SetContent([]RenderableString{
		{Content: "the second line"},
		{Content: "the first line"},
		{Content: "the third line"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe third \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"100% (3/3)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_StickyTopBottom(t *testing.T) {
	w, h := 10, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	// stickyness should override maintain selection
	vp.SetMaintainSelection(true)
	vp.SetTopSticky(true)
	vp.SetBottomSticky(true)
	vp.SetContent([]RenderableString{
		{Content: "the first line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"100% (1/1)",
	})
	compare(t, expectedView, vp.View())

	// add content, top sticky wins out arbitrarily when both set
	vp.SetContent([]RenderableString{
		{Content: "the second line"},
		{Content: "the first line"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe second\x1b[0m",
		"\x1b[38;2;0;0;255m line\x1b[0m",
		"50% (1/2)",
	})
	compare(t, expectedView, vp.View())

	// selection to bottom
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"100% (2/2)",
	})
	compare(t, expectedView, vp.View())

	// add content
	vp.SetContent([]RenderableString{
		{Content: "the second line"},
		{Content: "the first line"},
		{Content: "the third line"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe third \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"100% (3/3)",
	})
	compare(t, expectedView, vp.View())

	// de-activate by moving selection up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"66% (2/3)",
	})
	compare(t, expectedView, vp.View())

	// add content
	vp.SetContent([]RenderableString{
		{Content: "the second line"},
		{Content: "the first line"},
		{Content: "the third line"},
		{Content: "the fourth line"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"50% (2/4)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_RemoveLogsWhenSelectionBottom(t *testing.T) {
	w, h := 10, 3
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)

	// add content
	vp.SetContent([]RenderableString{
		{Content: "the second line"},
		{Content: "the first line"},
		{Content: "the third line"},
		{Content: "the fourth line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe second\x1b[0m",
		"25% (1/4)",
	})
	compare(t, expectedView, vp.View())

	// selection to bottom
	vp.SetSelectedItemIdx(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe fourth\x1b[0m",
		"100% (4/4)",
	})
	compare(t, expectedView, vp.View())

	// remove content
	vp.SetContent([]RenderableString{
		{Content: "the second line"},
		{Content: "the first line"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[0m",
		"100% (2/2)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_ChangeHeight(t *testing.T) {
	w, h := 10, 3
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "the first line"},
		{Content: "the second line"},
		{Content: "the third line"},
		{Content: "the fourth line"},
		{Content: "the fifth line"},
		{Content: "the sixth line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[0m",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(6)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"the second",
		" line",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())

	// move selection to third line
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the second",
		" line",
		"\x1b[38;2;0;0;255mthe third \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"50% (3/6)",
	})
	compare(t, expectedView, vp.View())

	// reduce height
	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe third \x1b[0m",
		"50% (3/6)",
	})
	compare(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(8)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe third \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"the fourth",
		" line",
		"the fifth ",
		"line",
		"50% (3/6)",
	})
	compare(t, expectedView, vp.View())

	// move selection to last line
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the fourth",
		" line",
		"the fifth ",
		"line",
		"\x1b[38;2;0;0;255mthe sixth \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())

	// reduce height
	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe sixth \x1b[0m",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_ChangeContent(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "the first line"},
		{Content: "the second line"},
		{Content: "the third line"},
		{Content: "the fourth line"},
		{Content: "the fifth line"},
		{Content: "the sixth line"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"the second",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())

	// move selection to bottom
	vp.SetSelectedItemIdx(5)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"line",
		"\x1b[38;2;0;0;255mthe sixth \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())

	// remove content
	vp.SetContent([]RenderableString{
		{Content: "the second line"},
		{Content: "the third line"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		" line",
		"\x1b[38;2;0;0;255mthe third \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"100% (2/2)",
	})
	compare(t, expectedView, vp.View())

	// remove all content
	vp.SetContent([]RenderableString{})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
	})
	compare(t, expectedView, vp.View())

	// add content
	vp.SetContent([]RenderableString{
		{Content: "the first line"},
		{Content: "the second line"},
		{Content: "the third line"},
		{Content: "the fourth line"},
		{Content: "the fifth line"},
		{Content: "the sixth line"},
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[0m",
		"\x1b[38;2;0;0;255mline\x1b[0m",
		"the second",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_StringToHighlight(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	vp.SetStringToHighlight("second")
	vp.HighlightStyle = renderer().NewStyle().Foreground(green)
	vp.HighlightStyleIfSelected = renderer().NewStyle().Foreground(red)
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
		{Content: "second"},
		{Content: "third"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
		"\x1b[38;2;0;255;0msecond\x1b[0m",
		"\x1b[38;2;0;255;0msecond\x1b[0m",
		"25% (1/4)",
	})
	compare(t, expectedView, vp.View())

	vp.SetStringToHighlight("first")
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;255;0;0mfirst\x1b[0m",
		"second",
		"second",
		"25% (1/4)",
	})
	compare(t, expectedView, vp.View())

	vp.SetContent([]RenderableString{
		{Content: "averylongwordthatwrapsover"},
	})
	vp.SetStringToHighlight("wraps")
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255maverylongw\x1b[0m\x1b[38;2;255;0;0m\x1b[0m",
		"\x1b[38;2;0;0;255mordthat\x1b[0m\x1b[38;2;255;0;0mwra\x1b[0m",
		"\x1b[38;2;255;0;0mps\x1b[0m\x1b[38;2;0;0;255mover\x1b[0m",
		"100% (1/1)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_AnsiOnSelection(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "line with some \x1b[38;2;255;0;0mred\x1b[0m text"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mline with \x1b[0m\x1b[38;2;255;0;0m\x1b[0m",
		"\x1b[38;2;0;0;255msome \x1b[0m\x1b[38;2;255;0;0mred\x1b[0m\x1b[38;2;0;0;255m t\x1b[0m",
		"\x1b[38;2;255;0;0m\x1b[0m\x1b[38;2;0;0;255mext\x1b[0m",
		"100% (1/1)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_SelectionEmpty(t *testing.T) {
	w, h := 20, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: ""},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255m \x1b[0m",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_ExtraSlash(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "|2024|\x1b[38;2;0mfl..lq\x1b[0m/\x1b[38;2;0mflask-3\x1b[0m|"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255m|2024|\x1b[0m\x1b[38;2;0mfl..\x1b[0m\x1b[38;2;0m\x1b[0m",
		"\x1b[38;2;0mlq\x1b[0m\x1b[38;2;0;0;255m/\x1b[0m\x1b[38;2;0mflask-3\x1b[0m",
		"\x1b[38;2;0m\x1b[0m\x1b[38;2;0m\x1b[0m\x1b[38;2;0;0;255m|\x1b[0m",
		"100% (1/1)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_SuperLongWrappedLine(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "smol"},
		{Content: strings.Repeat("12345678", 1000000)},
		{Content: "smol"},
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255msmol\x1b[0m",
		"1234567812",
		"3456781234",
		"33% (1/3)",
	})
	compare(t, expectedView, vp.View())

	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255m1234567812\x1b[0m",
		"\x1b[38;2;0;0;255m3456781234\x1b[0m",
		"\x1b[38;2;0;0;255m5678123456\x1b[0m",
		"66% (2/3)",
	})
	compare(t, expectedView, vp.View())

	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"5678123456",
		"7812345678",
		"\x1b[38;2;0;0;255msmol\x1b[0m",
		"100% (3/3)",
	})
	compare(t, expectedView, vp.View())
}

// # OTHER

func TestViewport_SelectionOn_ToggleWrap_PreserveSelection(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "first line that is fairly long"},
		{Content: "second line that is even much longer than the first"},
		{Content: "third line that is fairly long"},
		{Content: "fourth"},
		{Content: "fifth line that is fairly long"},
		{Content: "sixth"},
	})

	// wrap off, selection on first line
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line t...\x1b[0m",
		"second line ...",
		"third line t...",
		"fourth",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())

	// move selection to third line
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first line t...",
		"second line ...",
		"\x1b[38;2;0;0;255mthird line t...\x1b[0m",
		"fourth",
		"50% (3/6)",
	})
	compare(t, expectedView, vp.View())

	// toggle wrap on
	vp.SetWrapText(true)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"longer than the",
		" first",
		"\x1b[38;2;0;0;255mthird line that\x1b[0m",
		"\x1b[38;2;0;0;255m is fairly long\x1b[0m",
		"50% (3/6)",
	})
	compare(t, expectedView, vp.View())

	// toggle wrap off
	vp.SetWrapText(false)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first line t...",
		"second line ...",
		"\x1b[38;2;0;0;255mthird line t...\x1b[0m",
		"fourth",
		"50% (3/6)",
	})
	compare(t, expectedView, vp.View())

	// move selection to last line
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"third line t...",
		"fourth",
		"fifth line t...",
		"\x1b[38;2;0;0;255msixth\x1b[0m",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())

	// toggle wrap on
	vp.SetWrapText(true)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"fourth",
		"fifth line that",
		" is fairly long",
		"\x1b[38;2;0;0;255msixth\x1b[0m",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())

	// toggle wrap off
	vp.SetWrapText(false)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"third line t...",
		"fourth",
		"fifth line t...",
		"\x1b[38;2;0;0;255msixth\x1b[0m",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_ToggleWrap_PreserveSelectionInView(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "a really really really really really really really really really really really really long preamble"},
		{Content: "first line that is fairly long"},
		{Content: "second line that is even much longer than the first"},
		{Content: "third line that is fairly long"},
	})
	vp.SetSelectedItemIdx(3)
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"a really rea...",
		"first line t...",
		"second line ...",
		"\x1b[38;2;0;0;255mthird line t...\x1b[0m",
		"100% (4/4)",
	})
	compare(t, expectedView, vp.View())

	// toggle wrap, full wrapped selection should remain in view
	vp.SetWrapText(true)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"longer than the",
		" first",
		"\x1b[38;2;0;0;255mthird line that\x1b[0m",
		"\x1b[38;2;0;0;255m is fairly long\x1b[0m",
		"100% (4/4)",
	})
	compare(t, expectedView, vp.View())

	// toggle wrap
	vp.SetWrapText(false)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"a really rea...",
		"first line t...",
		"second line ...",
		"\x1b[38;2;0;0;255mthird line t...\x1b[0m",
		"100% (4/4)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_ToggleWrap_ScrollInBounds(t *testing.T) {
	w, h := 10, 7
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "the first line"},
		{Content: "the second line"},
		{Content: "the third line"},
		{Content: "the fourth line"},
		{Content: "the fifth line"},
		{Content: "the sixth line"},
	})

	// scroll to bottom with selection at top of that view
	vp.SetSelectedItemIdx(5)
	vp, _ = vp.Update(upKeyMsg)
	vp, _ = vp.Update(upKeyMsg)
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe fourth\x1b[0m",
		"\x1b[38;2;0;0;255m line\x1b[0m",
		"the fifth ",
		"line",
		"the sixth ",
		"66% (4/6)",
	})
	compare(t, expectedView, vp.View())

	// toggle wrap
	vp.SetWrapText(false)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the sec...",
		"the thi...",
		"\x1b[38;2;0;0;255mthe fou...\x1b[0m",
		"the fif...",
		"the six...",
		"66% (4/6)",
	})
	compare(t, expectedView, vp.View())
}
