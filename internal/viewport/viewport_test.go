package viewport

import (
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/kl/internal/util"
	"strings"
	"testing"
	"time"
)

var (
	downKeyMsg       = tea.KeyPressMsg{Code: 'j', Text: "j"}
	halfPgDownKeyMsg = tea.KeyPressMsg{Code: 'd', Text: "d"}
	fullPgDownKeyMsg = tea.KeyPressMsg{Code: 'f', Text: "f"}
	upKeyMsg         = tea.KeyPressMsg{Code: 'k', Text: "k"}
	halfPgUpKeyMsg   = tea.KeyPressMsg{Code: 'u', Text: "u"}
	fullPgUpKeyMsg   = tea.KeyPressMsg{Code: 'b', Text: "b"}
	goToTopKeyMsg    = tea.KeyPressMsg{Code: 'g', Text: "g"}
	goToBottomKeyMsg = tea.KeyPressMsg{Code: 'g', Text: "g", Mod: tea.ModShift}
	red              = lipgloss.Color("#ff0000")
	blue             = lipgloss.Color("#0000ff")
	green            = lipgloss.Color("#00ff00")
	selectionStyle   = lipgloss.NewStyle().Foreground(blue)
)

func newViewport(width, height int) Model[RenderableString] {
	km := KeyMap{
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "f", "ctrl+f"),
			key.WithHelp("f", "pgdn"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "b", "ctrl+b"),
			key.WithHelp("b", "pgup"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("u", "ctrl+u"),
			key.WithHelp("u", "½ page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("d", "ctrl+d"),
			key.WithHelp("d", "½ page down"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "scroll up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "scroll down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("←", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("→", "right"),
		),
		Top: key.NewBinding(
			key.WithKeys("g", "ctrl+g"),
			key.WithHelp("g", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("shift+g"),
			key.WithHelp("G", "bottom"),
		),
	}

	vp := New[RenderableString](width, height, km)
	vp.SelectedItemStyle = selectionStyle
	return vp
}

// # SELECTION DISABLED, WRAP OFF

func TestViewport_SelectionOff_WrapOff_Empty(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	expectedView := pad(vp.width, vp.height, []string{})
	util.CmpStr(t, expectedView, vp.View())
	vp.SetHeader([]string{"header"})
	expectedView = pad(vp.width, vp.height, []string{"header"})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_SmolDimensions(t *testing.T) {
	w, h := 0, 0
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(&vp, []string{"hi"})
	expectedView := pad(vp.width, vp.height, []string{""})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(1)
	vp.SetHeight(1)
	expectedView = pad(vp.width, vp.height, []string{"."})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(2)
	vp.SetHeight(2)
	expectedView = pad(vp.width, vp.height, []string{"..", ""})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(3)
	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{"...", "hi", "..."})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_Basic(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(&vp, []string{
		"first line",
		lipgloss.NewStyle().Foreground(red).Render("second") + " line",
		lipgloss.NewStyle().Foreground(red).Render("a really really long line"),
		lipgloss.NewStyle().Foreground(red).Render("a") + " really really long line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond\x1b[m line",
		"\x1b[38;2;255;0;0ma really rea...\x1b[m",
		"\x1b[38;2;255;0;0ma\x1b[m really rea...",
		"100% (4/4)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_GetConfigs(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(&vp, []string{
		"first",
		"second",
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
	setContent(&vp, []string{
		"first line",
		lipgloss.NewStyle().Foreground(red).Render("second") + " line",
		lipgloss.NewStyle().Foreground(red).Render("a really really long line"),
		lipgloss.NewStyle().Foreground(red).Render("a") + " really really long line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond\x1b[m line",
		"\x1b[38;2;255;0;0ma really rea...\x1b[m",
		"75% (3/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond\x1b[m line",
		"\x1b[38;2;255;0;0ma really rea...\x1b[m",
		"\x1b[38;2;255;0;0ma\x1b[m really rea...",
		"100% (4/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(7)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond\x1b[m line",
		"\x1b[38;2;255;0;0ma really rea...\x1b[m",
		"\x1b[38;2;255;0;0ma\x1b[m really rea...",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_FooterStyle(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.FooterStyle = lipgloss.NewStyle().Foreground(red)
	setContent(&vp, []string{
		"1",
		"2",
		"3",
		"4",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"1",
		"2",
		"3",
		"\x1b[38;2;255;0;0m75% (3/4)\x1b[m",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_FooterDisabled(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(&vp, []string{
		"first line",
		"second line",
		"third line",
		"fourth line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"second line",
		"third line",
		"75% (3/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetFooterEnabled(false)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"second line",
		"third line",
		"fourth line",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_SpaceAround(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(&vp, []string{
		"    first line     ",
		"          first line          ",
		"               first line               ",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"    first li...",
		"          fi...",
		"            ...",
		"100% (3/3)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_MultiHeader(t *testing.T) {
	w, h := 15, 2
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header1", "header2"})
	setContent(&vp, []string{
		"line1",
		"line2",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header1",
		"header2",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(4)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line1",
		"50% (1/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line2",
		"100% (2/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(5)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line1",
		"line2",
		"100% (2/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line1",
		"line2",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_OverflowLine(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"long header overflows"})
	setContent(&vp, []string{
		"123456789012345",
		"1234567890123456",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"long header ...",
		"123456789012345",
		"123456789012...",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_OverflowHeight(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(&vp, []string{
		"123456789012345",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"123456789012345",
		"123456789012...",
		"123456789012...",
		"123456789012...",
		"66% (4/6)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_Scrolling(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	doSetContent := func() {
		setContent(&vp, []string{
			"first",
			"second",
			"third",
			"fourth",
			"fifth",
			"sixth",
		})
	}
	validate := func(expectedView string) {
		// set content multiple times to confirm no side effects of doing it
		util.CmpStr(t, expectedView, vp.View())
		doSetContent()
		util.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
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
	setContent(&vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
		"33% (2/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// scroll so last item in view
	vp.ScrollSoItemIdxInView(5)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// scroll so second item in view
	vp.ScrollSoItemIdxInView(1)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"second",
		"third",
		"50% (3/6)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_BulkScrolling(t *testing.T) {
	w, h := 15, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(&vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
		"33% (2/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"third",
		"fourth",
		"66% (4/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// half page down
	vp, _ = vp.Update(halfPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"fourth",
		"fifth",
		"83% (5/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"third",
		"fourth",
		"66% (4/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"second",
		"third",
		"50% (3/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
		"33% (2/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// go to bottom
	vp, _ = vp.Update(goToBottomKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// go to top
	vp, _ = vp.Update(goToTopKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
		"33% (2/6)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_Panning(t *testing.T) {
	w, h := 10, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header long"})
	doSetContent := func() {
		setContent(&vp, []string{
			"first line that is fairly long",
			"second line that is even much longer than the first",
			"third line that is fairly long",
			"fourth",
			"fifth line that is fairly long",
			"sixth",
		})
	}
	validate := func(expectedView string) {
		// set content multiple times to confirm no side effects of doing it
		util.CmpStr(t, expectedView, vp.View())
		doSetContent()
		util.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
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
	setContent(&vp, []string{
		"the first one",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header ...",
		"...rst one",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_ChangeHeight(t *testing.T) {
	w, h := 10, 3
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(&vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first",
		"16% (1/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

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
	util.CmpStr(t, expectedView, vp.View())

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
	util.CmpStr(t, expectedView, vp.View())

	// reduce height
	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"third",
		"50% (3/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

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
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_ChangeContent(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(&vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
		"third",
		"50% (3/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

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
	util.CmpStr(t, expectedView, vp.View())

	// remove content
	setContent(&vp, []string{})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
	})
	util.CmpStr(t, expectedView, vp.View())

	// re-add content
	setContent(&vp, []string{
		"first",
		"second",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_StringToHighlight(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetStringToHighlight("second")
	vp.HighlightStyle = lipgloss.NewStyle().Foreground(red)
	setContent(&vp, []string{
		"first",
		"second",
		"second",
		"third",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first",
		"\x1b[38;2;255;0;0msecond\x1b[m",
		"\x1b[38;2;255;0;0msecond\x1b[m",
		"75% (3/4)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_StringToHighlightManyMatches(t *testing.T) {
	runTest := func(t *testing.T) {
		w, h := 10, 5
		vp := newViewport(w, h)
		vp.SetHeader([]string{"header"})
		setContent(&vp, []string{
			strings.Repeat("r", 100000),
		})
		vp.SetStringToHighlight("r")
		vp.HighlightStyle = lipgloss.NewStyle().Foreground(green)
		vp.HighlightStyleIfSelected = lipgloss.NewStyle().Foreground(red)
		expectedView := pad(vp.width, vp.height, []string{
			"header",
			strings.Repeat("\x1b[38;2;0;255;0mr\x1b[m", 7) + strings.Repeat(".", 3),
		})
		util.CmpStr(t, expectedView, vp.View())
	}
	util.RunWithTimeout(t, runTest, 20*time.Millisecond)
}

func TestViewport_SelectionOff_WrapOff_StringToHighlightAnsi(t *testing.T) {
	w, h := 20, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(&vp, []string{
		"line \x1b[38;2;255;0;0mred\x1b[m e again",
	})
	vp.SetStringToHighlight("e")
	vp.HighlightStyle = selectionStyle
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"lin\x1b[38;2;0;0;255me\x1b[m \x1b[38;2;255;0;0mr\x1b[m\x1b[38;2;0;0;255me\x1b[m\x1b[38;2;255;0;0md\x1b[m \x1b[38;2;0;0;255me\x1b[m again",
	})
	util.CmpStr(t, expectedView, vp.View())

	// should not highlight the ansi escape codes themselves
	vp.SetStringToHighlight("38")
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"line \x1b[38;2;255;0;0mred\x1b[m e again",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_StringToHighlightAnsiUnicode(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b) = 6w, 11b
	vp.SetHeader([]string{"A💖中é"})
	setContent(&vp, []string{
		"A💖中é",
		"A💖中éA💖中é",
	})
	vp.SetStringToHighlight("中é")
	vp.HighlightStyle = selectionStyle
	expectedView := pad(vp.width, vp.height, []string{
		"A💖中é",
		"A💖\x1b[38;2;0;0;255m中é\x1b[m",
		"A💖\x1b[38;2;0;0;255m中é\x1b[m...",
	})
	util.CmpStr(t, expectedView, vp.View())
}

// # SELECTION ENABLED, WRAP OFF

func TestViewport_SelectionOn_WrapOff_Empty(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetSelectionEnabled(true)
	expectedView := pad(vp.width, vp.height, []string{})
	util.CmpStr(t, expectedView, vp.View())
	vp.SetHeader([]string{"header"})
	expectedView = pad(vp.width, vp.height, []string{"header"})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_SmolDimensions(t *testing.T) {
	w, h := 0, 0
	vp := newViewport(w, h)
	vp.SetSelectionEnabled(true)
	vp.SetHeader([]string{"header"})
	setContent(&vp, []string{"hi"})
	expectedView := pad(vp.width, vp.height, []string{""})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(1)
	vp.SetHeight(1)
	expectedView = pad(vp.width, vp.height, []string{"."})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(2)
	vp.SetHeight(2)
	expectedView = pad(vp.width, vp.height, []string{"..", ""})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(3)
	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{
		"...",
		"\x1b[38;2;0;0;255mhi\x1b[m",
		"...",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_Basic(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"first line",
		lipgloss.NewStyle().Foreground(red).Render("second") + " line",
		lipgloss.NewStyle().Foreground(red).Render("a really really long line"),
		lipgloss.NewStyle().Foreground(red).Render("a") + " really really long line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[m",
		"\x1b[38;2;255;0;0msecond\x1b[m line",
		"\x1b[38;2;255;0;0ma really rea...\x1b[m",
		"\x1b[38;2;255;0;0ma\x1b[m really rea...",
		"25% (1/4)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_GetConfigs(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"first",
		"second",
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
	if selectedItem := vp.GetSelectedItem(); selectedItem.Render().Content() != "second" {
		t.Errorf("got unexpected selected item: %v", selectedItem)
	}
}

func TestViewport_SelectionOn_WrapOff_ShowFooter(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"first line",
		lipgloss.NewStyle().Foreground(red).Render("second") + " line",
		lipgloss.NewStyle().Foreground(red).Render("a really really long line"),
		lipgloss.NewStyle().Foreground(red).Render("a") + " really really long line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[m",
		"\x1b[38;2;255;0;0msecond\x1b[m line",
		"\x1b[38;2;255;0;0ma really rea...\x1b[m",
		"25% (1/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[m",
		"\x1b[38;2;255;0;0msecond\x1b[m line",
		"\x1b[38;2;255;0;0ma really rea...\x1b[m",
		"\x1b[38;2;255;0;0ma\x1b[m really rea...",
		"25% (1/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(7)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[m",
		"\x1b[38;2;255;0;0msecond\x1b[m line",
		"\x1b[38;2;255;0;0ma really rea...\x1b[m",
		"\x1b[38;2;255;0;0ma\x1b[m really rea...",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_FooterStyle(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.FooterStyle = lipgloss.NewStyle().Foreground(red)
	setContent(&vp, []string{
		"1",
		"2",
		"3",
		"4",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255m1\x1b[m",
		"2",
		"3",
		"\x1b[38;2;255;0;0m25% (1/4)\x1b[m",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_FooterDisabled(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"first line",
		"second line",
		"third line",
		"fourth line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[m",
		"second line",
		"third line",
		"25% (1/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetFooterEnabled(false)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[m",
		"second line",
		"third line",
		"fourth line",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_SpaceAround(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"    first line     ",
		"          first line          ",
		"               first line               ",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255m    first li...\x1b[m",
		"          fi...",
		"            ...",
		"33% (1/3)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_MultiHeader(t *testing.T) {
	w, h := 15, 2
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header1", "header2"})
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"line1",
		"line2",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header1",
		"header2",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(4)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"\x1b[38;2;0;0;255mline1\x1b[m",
		"50% (1/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"\x1b[38;2;0;0;255mline2\x1b[m",
		"100% (2/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(5)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line1",
		"\x1b[38;2;0;0;255mline2\x1b[m",
		"100% (2/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line1",
		"\x1b[38;2;0;0;255mline2\x1b[m",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_OverflowLine(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"long header overflows"})
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"123456789012345",
		"1234567890123456",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"long header ...",
		"\x1b[38;2;0;0;255m123456789012345\x1b[m",
		"123456789012...",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_OverflowHeight(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"123456789012345",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255m123456789012345\x1b[m",
		"123456789012...",
		"123456789012...",
		"123456789012...",
		"16% (1/6)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_Scrolling(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	doSetContent := func() {
		setContent(&vp, []string{
			"first",
			"second",
			"third",
			"fourth",
			"fifth",
			"sixth",
		})
	}
	validate := func(expectedView string) {
		// set content multiple times to confirm no side effects of doing it
		util.CmpStr(t, expectedView, vp.View())
		doSetContent()
		util.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[m",
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
		"\x1b[38;2;0;0;255msecond\x1b[m",
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
		"\x1b[38;2;0;0;255msixth\x1b[m",
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
	setContent(&vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[m",
		"second",
		"16% (1/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// attempting to scroll so selection out of view is no-op
	vp.ScrollSoItemIdxInView(5)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[m",
		"second",
		"16% (1/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// move selection down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"\x1b[38;2;0;0;255msecond\x1b[m",
		"33% (2/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// scroll so third item in view
	vp.ScrollSoItemIdxInView(2)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255msecond\x1b[m",
		"third",
		"33% (2/6)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_BulkScrolling(t *testing.T) {
	w, h := 15, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[m",
		"second",
		"16% (1/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthird\x1b[m",
		"fourth",
		"50% (3/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// half page down
	vp, _ = vp.Update(halfPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfourth\x1b[m",
		"fifth",
		"66% (4/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"fifth",
		"\x1b[38;2;0;0;255msixth\x1b[m",
		"100% (6/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"third",
		"\x1b[38;2;0;0;255mfourth\x1b[m",
		"66% (4/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"second",
		"\x1b[38;2;0;0;255mthird\x1b[m",
		"50% (3/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"\x1b[38;2;0;0;255msecond\x1b[m",
		"33% (2/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[m",
		"second",
		"16% (1/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// go to bottom
	vp, _ = vp.Update(goToBottomKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"fifth",
		"\x1b[38;2;0;0;255msixth\x1b[m",
		"100% (6/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// go to top
	vp, _ = vp.Update(goToTopKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[m",
		"second",
		"16% (1/6)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_Panning(t *testing.T) {
	w, h := 10, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header long"})
	vp.SetSelectionEnabled(true)
	doSetContent := func() {
		setContent(&vp, []string{
			"first line that is fairly long",
			"second line that is even much longer than the first",
			"third line that is fairly long",
			"fourth",
			"fifth line that is fairly long",
			"sixth",
		})
	}
	validate := func(expectedView string) {
		// set content multiple times to confirm no side effects of doing it
		util.CmpStr(t, expectedView, vp.View())
		doSetContent()
		util.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
	expectedView := pad(vp.width, vp.height, []string{
		"header ...",
		"\x1b[38;2;0;0;255mfirst l...\x1b[m",
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
		"\x1b[38;2;0;0;255m...ne t...\x1b[m",
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
		"\x1b[38;2;0;0;255m...ine ...\x1b[m",
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
		"\x1b[38;2;0;0;255m...e first\x1b[m",
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
		"\x1b[38;2;0;0;255m...\x1b[m",
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
		"\x1b[38;2;0;0;255m...\x1b[m",
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
		"\x1b[38;2;0;0;255m...\x1b[m",
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
		"\x1b[38;2;0;0;255m...\x1b[m",
		"100% (6/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header ...",
		"...ly long",
		"...",
		"\x1b[38;2;0;0;255m...ly long\x1b[m",
		"...",
		"83% (5/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header ...",
		"...ly long",
		"\x1b[38;2;0;0;255m...\x1b[m",
		"...ly long",
		"...",
		"66% (4/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header ...",
		"\x1b[38;2;0;0;255m...ly long\x1b[m",
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
		"\x1b[38;2;0;0;255m...n mu...\x1b[m",
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
		"\x1b[38;2;0;0;255m...ly long\x1b[m",
		"...n mu...",
		"...ly long",
		"...",
		"16% (1/6)",
	})
	validate(expectedView)

	// set shorter content
	setContent(&vp, []string{
		"the first one",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header ...",
		"\x1b[38;2;0;0;255m...rst one\x1b[m",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_MaintainSelection(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetMaintainSelection(true)
	setContent(&vp, []string{
		"sixth",
		"seventh",
		"eighth",
		"ninth",
		"tenth",
		"eleventh",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255msixth\x1b[m",
		"seventh",
		"eighth",
		"16% (1/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// selection down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"sixth",
		"\x1b[38;2;0;0;255mseventh\x1b[m",
		"eighth",
		"33% (2/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// add content above
	setContent(&vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
		"seventh",
		"eighth",
		"ninth",
		"tenth",
		"eleventh",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"sixth",
		"\x1b[38;2;0;0;255mseventh\x1b[m",
		"eighth",
		"63% (7/11)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// add content below
	setContent(&vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
		"seventh",
		"eighth",
		"ninth",
		"tenth",
		"eleventh",
		"twelfth",
		"thirteenth",
		"fourteenth",
		"fifteenth",
		"sixteenth",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"sixth",
		"\x1b[38;2;0;0;255mseventh\x1b[m",
		"eighth",
		"43% (7/16)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_StickyTop(t *testing.T) {
	w, h := 15, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	// stickyness should override maintain selection
	vp.SetMaintainSelection(true)
	vp.SetTopSticky(true)
	setContent(&vp, []string{
		"first",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[m",
	})
	util.CmpStr(t, expectedView, vp.View())

	// add content
	setContent(&vp, []string{
		"second",
		"first",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255msecond\x1b[m",
		"first",
		"50% (1/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// de-activate by moving selection down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"second",
		"\x1b[38;2;0;0;255mfirst\x1b[m",
		"100% (2/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// add content
	setContent(&vp, []string{
		"second",
		"first",
		"third",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"second",
		"\x1b[38;2;0;0;255mfirst\x1b[m",
		"66% (2/3)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_StickyBottom(t *testing.T) {
	w, h := 15, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	// stickyness should override maintain selection
	vp.SetMaintainSelection(true)
	vp.SetBottomSticky(true)
	setContent(&vp, []string{
		"first",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[m",
	})
	util.CmpStr(t, expectedView, vp.View())

	// add content
	setContent(&vp, []string{
		"second",
		"first",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"second",
		"\x1b[38;2;0;0;255mfirst\x1b[m",
		"100% (2/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// de-activate by moving selection up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255msecond\x1b[m",
		"first",
		"50% (1/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// add content
	setContent(&vp, []string{
		"second",
		"first",
		"third",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255msecond\x1b[m",
		"first",
		"33% (1/3)",
	})
	util.CmpStr(t, expectedView, vp.View())
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
	setContent(&vp, []string{})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
	})
	util.CmpStr(t, expectedView, vp.View())

	setContent(&vp, []string{
		"second",
		"first",
		"third",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"\x1b[38;2;0;0;255mthird\x1b[m",
		"100% (3/3)",
	})
	util.CmpStr(t, expectedView, vp.View())
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
	setContent(&vp, []string{
		"first",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[m",
	})
	util.CmpStr(t, expectedView, vp.View())

	// add content, top sticky wins out arbitrarily when both set
	setContent(&vp, []string{
		"second",
		"first",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255msecond\x1b[m",
		"first",
		"50% (1/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// selection to bottom
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"second",
		"\x1b[38;2;0;0;255mfirst\x1b[m",
		"100% (2/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// add content
	setContent(&vp, []string{
		"second",
		"first",
		"third",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"\x1b[38;2;0;0;255mthird\x1b[m",
		"100% (3/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// de-activate by moving selection up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[m",
		"third",
		"66% (2/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// add content
	setContent(&vp, []string{
		"second",
		"first",
		"third",
		"fourth",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[m",
		"third",
		"50% (2/4)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_RemoveLogsWhenSelectionBottom(t *testing.T) {
	w, h := 10, 3
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)

	// add content
	setContent(&vp, []string{
		"second",
		"first",
		"third",
		"fourth",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255msecond\x1b[m",
		"25% (1/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// selection to bottom
	vp.SetSelectedItemIdx(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfourth\x1b[m",
		"100% (4/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// remove content
	setContent(&vp, []string{
		"second",
		"first",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[m",
		"100% (2/2)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_ChangeHeight(t *testing.T) {
	w, h := 10, 3
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[m",
		"16% (1/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(8)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[m",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
		"16% (1/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// move selection to third line
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
		"\x1b[38;2;0;0;255mthird\x1b[m",
		"fourth",
		"fifth",
		"sixth",
		"50% (3/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// reduce height
	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthird\x1b[m",
		"50% (3/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(8)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
		"\x1b[38;2;0;0;255mthird\x1b[m",
		"fourth",
		"fifth",
		"sixth",
		"50% (3/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

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
		"\x1b[38;2;0;0;255msixth\x1b[m",
		"100% (6/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// reduce height
	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255msixth\x1b[m",
		"100% (6/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(8)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"\x1b[38;2;0;0;255msixth\x1b[m",
		"100% (6/6)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_ChangeContent(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[m",
		"second",
		"third",
		"16% (1/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// move selection to bottom
	vp.SetSelectedItemIdx(5)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"fourth",
		"fifth",
		"\x1b[38;2;0;0;255msixth\x1b[m",
		"100% (6/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// remove content
	setContent(&vp, []string{
		"second",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255msecond\x1b[m",
	})
	util.CmpStr(t, expectedView, vp.View())

	// remove all content
	setContent(&vp, []string{})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
	})
	util.CmpStr(t, expectedView, vp.View())

	// add content (maintain selection off)
	setContent(&vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[m",
		"second",
		"third",
		"16% (1/6)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_StringToHighlight(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetStringToHighlight("second")
	vp.HighlightStyle = lipgloss.NewStyle().Foreground(green)
	vp.HighlightStyleIfSelected = lipgloss.NewStyle().Foreground(red)
	setContent(&vp, []string{
		"the first line",
		"the second line",
		"the second line",
		"the fourth line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first line\x1b[m",
		"the \x1b[38;2;0;255;0msecond\x1b[m line",
		"the \x1b[38;2;0;255;0msecond\x1b[m line",
		"25% (1/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetStringToHighlight("first")
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe \x1b[m\x1b[38;2;255;0;0mfirst\x1b[m\x1b[38;2;0;0;255m line\x1b[m",
		"the second line",
		"the second line",
		"25% (1/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	setContent(&vp, []string{
		"first line",
		"second line",
		"second line",
		"fourth line",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;255;0;0mfirst\x1b[m\x1b[38;2;0;0;255m line\x1b[m",
		"second line",
		"second line",
		"25% (1/4)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_StringToHighlightManyMatches(t *testing.T) {
	runTest := func(t *testing.T) {
		w, h := 10, 5
		vp := newViewport(w, h)
		vp.SetHeader([]string{"header"})
		vp.SetSelectionEnabled(true)
		setContent(&vp, []string{
			strings.Repeat("r", 100000),
		})
		vp.SetStringToHighlight("r")
		vp.HighlightStyle = lipgloss.NewStyle().Foreground(green)
		vp.HighlightStyleIfSelected = lipgloss.NewStyle().Foreground(red)
		expectedView := pad(vp.width, vp.height, []string{
			"header",
			strings.Repeat("\x1b[38;2;255;0;0mr\x1b[m", 7) + "\x1b[38;2;0;0;255m" + strings.Repeat(".", 3) + "\x1b[m",
		})
		util.CmpStr(t, expectedView, vp.View())
	}
	util.RunWithTimeout(t, runTest, 10*time.Millisecond)
}

func TestViewport_SelectionOn_WrapOff_AnsiOnSelection(t *testing.T) {
	w, h := 20, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"line with \x1b[38;2;255;0;0mred\x1b[m text",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mline with \x1b[m\x1b[38;2;255;0;0mred\x1b[m\x1b[38;2;0;0;255m text\x1b[m",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_SelectionEmpty(t *testing.T) {
	w, h := 20, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255m \x1b[m",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_ExtraSlash(t *testing.T) {
	w, h := 25, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"|2024|\x1b[38;2;0mfl..lq\x1b[m/\x1b[38;2;0mflask-3\x1b[m|",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255m|2024|\x1b[m\x1b[38;2;0mfl..lq\x1b[m\x1b[38;2;0;0;255m/\x1b[m\x1b[38;2;0mflask-3\x1b[m\x1b[38;2;0;0;255m|\x1b[m",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_StringToHighlightAnsiUnicode(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b) = 6w, 11b
	vp.SetHeader([]string{"A💖中é"})
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"A💖中é",
		"A💖中éA💖中é",
	})
	vp.SetStringToHighlight("中é")
	vp.HighlightStyle = lipgloss.NewStyle().Foreground(green)
	vp.HighlightStyleIfSelected = lipgloss.NewStyle().Foreground(red)
	expectedView := pad(vp.width, vp.height, []string{
		"A💖中é",
		"\x1b[38;2;0;0;255mA💖\x1b[m\x1b[38;2;255;0;0m中é\x1b[m",
		"A💖\x1b[38;2;0;255;0m中é\x1b[m...",
	})
	util.CmpStr(t, expectedView, vp.View())
}

// # SELECTION DISABLED, WRAP ON

func TestViewport_SelectionOff_WrapOn_Empty(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	expectedView := pad(vp.width, vp.height, []string{})
	util.CmpStr(t, expectedView, vp.View())
	vp.SetHeader([]string{"header"})
	expectedView = pad(vp.width, vp.height, []string{"header"})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_SmolDimensions(t *testing.T) {
	w, h := 0, 0
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	vp.SetHeader([]string{"header"})
	setContent(&vp, []string{"hi"})
	expectedView := pad(vp.width, vp.height, []string{""})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(1)
	vp.SetHeight(1)
	expectedView = pad(vp.width, vp.height, []string{"h"})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(2)
	vp.SetHeight(2)
	expectedView = pad(vp.width, vp.height, []string{"he", "ad"})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(3)
	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{"hea", "der", ""})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(4)
	vp.SetHeight(4)
	expectedView = pad(vp.width, vp.height, []string{"head", "er", "hi", "1..."})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_Basic(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(&vp, []string{
		"first line",
		lipgloss.NewStyle().Foreground(red).Render("second") + " line",
		lipgloss.NewStyle().Foreground(red).Render("a really really long line"),
		lipgloss.NewStyle().Foreground(red).Render("a") + " really really long line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond\x1b[m line",
		"\x1b[38;2;255;0;0ma really really\x1b[m",
		"\x1b[38;2;255;0;0m long line\x1b[m",
		"75% (3/4)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_GetConfigs(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(&vp, []string{
		"first",
		"second",
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
	setContent(&vp, []string{
		"first line",
		lipgloss.NewStyle().Foreground(red).Render("second") + " line",
		lipgloss.NewStyle().Foreground(red).Render("a really really long line"),
		lipgloss.NewStyle().Foreground(red).Render("a") + " really really long line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond\x1b[m line",
		"\x1b[38;2;255;0;0ma really really\x1b[m",
		"\x1b[38;2;255;0;0m long line\x1b[m",
		"\x1b[38;2;255;0;0ma\x1b[m really really",
		"99% (4/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(8)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond\x1b[m line",
		"\x1b[38;2;255;0;0ma really really\x1b[m",
		"\x1b[38;2;255;0;0m long line\x1b[m",
		"\x1b[38;2;255;0;0ma\x1b[m really really",
		" long line",
		"100% (4/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(9)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond\x1b[m line",
		"\x1b[38;2;255;0;0ma really really\x1b[m",
		"\x1b[38;2;255;0;0m long line\x1b[m",
		"\x1b[38;2;255;0;0ma\x1b[m really really",
		" long line",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_FooterStyle(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.FooterStyle = lipgloss.NewStyle().Foreground(red)
	setContent(&vp, []string{
		"1",
		"2",
		"3",
		"4",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"1",
		"2",
		"3",
		"\x1b[38;2;255;0;0m75% (3/4)\x1b[m",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_FooterDisabled(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(&vp, []string{
		"first line",
		"second line",
		"third line",
		"fourth line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"second line",
		"third line",
		"75% (3/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetFooterEnabled(false)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"second line",
		"third line",
		"fourth line",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_SpaceAround(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(&vp, []string{
		"    first line     ",
		"          first line          ",
		"               first line               ",
	})
	// trailing space is not trimmed
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"    first line ",
		"",
		"          first",
		"66% (2/3)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_MultiHeader(t *testing.T) {
	w, h := 15, 2
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header1", "header2"})
	vp.SetWrapText(true)
	setContent(&vp, []string{
		"line1",
		"line2",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header1",
		"header2",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(4)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line1",
		"50% (1/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line2",
		"100% (2/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(5)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line1",
		"line2",
		"100% (2/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line1",
		"line2",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_OverflowLine(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"long header overflows"})
	vp.SetWrapText(true)
	setContent(&vp, []string{
		"123456789012345",
		"1234567890123456",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"long header ove",
		"rflows",
		"123456789012345",
		"123456789012345",
		"6",
		"100% (2/2)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_OverflowHeight(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(&vp, []string{
		"123456789012345",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"123456789012345",
		"123456789012345",
		"6",
		"123456789012345",
		"50% (3/6)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_Scrolling(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	doSetContent := func() {
		setContent(&vp, []string{
			"first",
			"second",
			"third",
			"fourth",
			"fifth",
			"sixth",
		})
	}
	validate := func(expectedView string) {
		// set content multiple times to confirm no side effects of doing it
		util.CmpStr(t, expectedView, vp.View())
		doSetContent()
		util.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
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
	setContent(&vp, []string{
		"the first line",
		"the second line",
		"the third line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"the first",
		"line",
		"33% (1/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// scroll so last item in view
	vp.ScrollSoItemIdxInView(2)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the third",
		"line",
		"100% (3/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// scroll so second item in view
	vp.ScrollSoItemIdxInView(1)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the second",
		" line",
		"66% (2/3)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_BulkScrolling(t *testing.T) {
	w, h := 10, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(&vp, []string{
		"the first line",
		"the second line",
		"the third line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"the first",
		"line",
		"33% (1/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the second",
		" line",
		"66% (2/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// half page down
	vp, _ = vp.Update(halfPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		" line",
		"the third ",
		"99% (3/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the third ",
		"line",
		"100% (3/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the second",
		" line",
		"66% (2/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"line",
		"the second",
		"66% (2/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the first",
		"line",
		"33% (1/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// go to bottom
	vp, _ = vp.Update(goToBottomKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the third ",
		"line",
		"100% (3/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// go to top
	vp, _ = vp.Update(goToTopKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the first",
		"line",
		"33% (1/3)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_Panning(t *testing.T) {
	w, h := 10, 7
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header long"})
	vp.SetWrapText(true)
	doSetContent := func() {
		setContent(&vp, []string{
			"first line that is fairly long",
			"second line that is even much longer than the first",
			"third line that is fairly long",
			"fourth",
			"fifth line that is fairly long",
			"sixth",
		})
	}
	validate := func(expectedView string) {
		// set content multiple times to confirm no side effects of doing it
		util.CmpStr(t, expectedView, vp.View())
		doSetContent()
		util.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
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
	setContent(&vp, []string{
		"the first line",
		"the second line",
		"the third line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"the first",
		"line",
		"33% (1/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// scroll down to bottom
	vp, _ = vp.Update(fullPgDownKeyMsg)
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the third",
		"line",
		"100% (3/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// reduce height
	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the third",
		"99% (3/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"line",
		"100% (3/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

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
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_ChangeContent(t *testing.T) {
	w, h := 10, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(&vp, []string{
		"the first line",
		"the second line",
		"the third line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"the first",
		"line",
		"33% (1/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// scroll down to bottom
	vp, _ = vp.Update(fullPgDownKeyMsg)
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the third",
		"line",
		"100% (3/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// remove content
	setContent(&vp, []string{
		"the first line",
		"the second line",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the second",
		" line",
		"100% (2/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// add content
	setContent(&vp, []string{
		"the first line",
		"the second line",
		"the third line",
		"the fourth line",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the second",
		" line",
		"50% (2/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// remove all content
	setContent(&vp, []string{})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_StringToHighlight(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetStringToHighlight("second")
	vp.HighlightStyle = lipgloss.NewStyle().Foreground(red)
	setContent(&vp, []string{
		"first",
		"second",
		"second",
		"third",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first",
		"\x1b[38;2;255;0;0msecond\x1b[m",
		"\x1b[38;2;255;0;0msecond\x1b[m",
		"75% (3/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	setContent(&vp, []string{
		"averylongwordthatwraps",
	})
	vp.SetStringToHighlight("wraps")
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"averylongw",
		"ordthat\x1b[38;2;255;0;0mwra\x1b[m",
		"\x1b[38;2;255;0;0mps\x1b[m",
		"100% (1/1)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_StringToHighlightManyMatches(t *testing.T) {
	runTest := func(t *testing.T) {
		w, h := 10, 5
		vp := newViewport(w, h)
		vp.SetHeader([]string{"header"})
		vp.SetWrapText(true)
		setContent(&vp, []string{
			strings.Repeat("r", 100000),
		})
		vp.SetStringToHighlight("r")
		vp.HighlightStyle = lipgloss.NewStyle().Foreground(green)
		vp.HighlightStyleIfSelected = lipgloss.NewStyle().Foreground(red)
		expectedView := pad(vp.width, vp.height, []string{
			"header",
			strings.Repeat("\x1b[38;2;0;255;0mr\x1b[m", 10),
			strings.Repeat("\x1b[38;2;0;255;0mr\x1b[m", 10),
			strings.Repeat("\x1b[38;2;0;255;0mr\x1b[m", 10),
			"99% (1/1)",
		})
		util.CmpStr(t, expectedView, vp.View())
	}
	util.RunWithTimeout(t, runTest, 10*time.Millisecond)
}

func TestViewport_SelectionOff_WrapOn_StringToHighlightAnsi(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(&vp, []string{
		"line \x1b[38;2;255;0;0mred\x1b[m e again",
	})
	vp.SetStringToHighlight("e")
	vp.HighlightStyle = selectionStyle
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"lin\x1b[38;2;0;0;255me\x1b[m \x1b[38;2;255;0;0mr\x1b[m\x1b[38;2;0;0;255me\x1b[m\x1b[38;2;255;0;0md\x1b[m \x1b[38;2;0;0;255me\x1b[m",
		" again",
	})
	util.CmpStr(t, expectedView, vp.View())

	// should not highlight the ansi escape codes themselves
	vp.SetStringToHighlight("38")
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"line \x1b[38;2;255;0;0mred\x1b[m e",
		" again",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_SuperLongWrappedLine(t *testing.T) {
	runTest := func(t *testing.T) {
		w, h := 10, 5
		vp := newViewport(w, h)
		vp.SetHeader([]string{"header"})
		vp.SetWrapText(true)
		setContent(&vp, []string{
			"smol",
			strings.Repeat("12345678", 1000000),
			"smol",
		})
		expectedView := pad(vp.width, vp.height, []string{
			"header",
			"smol",
			"1234567812",
			"3456781234",
			"66% (2/3)",
		})
		util.CmpStr(t, expectedView, vp.View())

		vp, _ = vp.Update(downKeyMsg)
		expectedView = pad(vp.width, vp.height, []string{
			"header",
			"1234567812",
			"3456781234",
			"5678123456",
			"66% (2/3)",
		})
		util.CmpStr(t, expectedView, vp.View())

		vp, _ = vp.Update(downKeyMsg)
		expectedView = pad(vp.width, vp.height, []string{
			"header",
			"3456781234",
			"5678123456",
			"7812345678",
			"66% (2/3)",
		})
		util.CmpStr(t, expectedView, vp.View())

		vp, _ = vp.Update(goToBottomKeyMsg)
		expectedView = pad(vp.width, vp.height, []string{
			"header",
			"5678123456",
			"7812345678",
			"smol",
			"100% (3/3)",
		})
		util.CmpStr(t, expectedView, vp.View())
	}
	util.RunWithTimeout(t, runTest, 500*time.Millisecond)
}

func TestViewport_SelectionOff_WrapOn_StringToHighlightAnsiUnicode(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b) = 6w, 11b
	vp.SetHeader([]string{"A💖中é"})
	vp.SetWrapText(true)
	setContent(&vp, []string{
		"A💖中é",
		"A💖中éA💖中é",
	})
	vp.SetStringToHighlight("中é")
	vp.HighlightStyle = lipgloss.NewStyle().Foreground(green)
	vp.HighlightStyleIfSelected = lipgloss.NewStyle().Foreground(red)
	expectedView := pad(vp.width, vp.height, []string{
		"A💖中é",
		"A💖\x1b[38;2;0;255;0m中é\x1b[m",
		"A💖\x1b[38;2;0;255;0m中é\x1b[mA💖",
		"\x1b[38;2;0;255;0m中é\x1b[m",
		"100% (2/2)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

// # SELECTION ENABLED, WRAP ON

func TestViewport_SelectionOn_WrapOn_Empty(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	expectedView := pad(vp.width, vp.height, []string{})
	util.CmpStr(t, expectedView, vp.View())
	vp.SetHeader([]string{"header"})
	expectedView = pad(vp.width, vp.height, []string{"header"})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_SmolDimensions(t *testing.T) {
	w, h := 0, 0
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	vp.SetHeader([]string{"header"})
	setContent(&vp, []string{"hi"})
	expectedView := pad(vp.width, vp.height, []string{""})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(1)
	vp.SetHeight(1)
	expectedView = pad(vp.width, vp.height, []string{"h"})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(2)
	vp.SetHeight(2)
	expectedView = pad(vp.width, vp.height, []string{"he", "ad"})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(3)
	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{"hea", "der", ""})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(4)
	vp.SetHeight(4)
	expectedView = pad(vp.width, vp.height, []string{
		"head",
		"er",
		"\x1b[38;2;0;0;255mhi\x1b[m",
		"1...",
	})
	util.CmpStr(t, expectedView, vp.View())

}

func TestViewport_SelectionOn_WrapOn_Basic(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"first line",
		lipgloss.NewStyle().Foreground(red).Render("second") + " line",
		lipgloss.NewStyle().Foreground(red).Render("a really really long line"),
		lipgloss.NewStyle().Foreground(red).Render("a") + " really really long line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[m",
		"\x1b[38;2;255;0;0msecond\x1b[m line",
		"\x1b[38;2;255;0;0ma really really\x1b[m",
		"\x1b[38;2;255;0;0m long line\x1b[m",
		"25% (1/4)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_GetConfigs(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"first",
		"second",
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
	if selectedItem := vp.GetSelectedItem(); selectedItem.Render().Content() != "second" {
		t.Errorf("got unexpected selected item: %v", selectedItem)
	}
}

func TestViewport_SelectionOn_WrapOn_ShowFooter(t *testing.T) {
	w, h := 15, 7
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"first line",
		lipgloss.NewStyle().Foreground(red).Render("second") + " line",
		lipgloss.NewStyle().Foreground(red).Render("a really really long line"),
		lipgloss.NewStyle().Foreground(red).Render("a") + " really really long line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[m",
		"\x1b[38;2;255;0;0msecond\x1b[m line",
		"\x1b[38;2;255;0;0ma really really\x1b[m",
		"\x1b[38;2;255;0;0m long line\x1b[m",
		"\x1b[38;2;255;0;0ma\x1b[m really really",
		"25% (1/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(8)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[m",
		"\x1b[38;2;255;0;0msecond\x1b[m line",
		"\x1b[38;2;255;0;0ma really really\x1b[m",
		"\x1b[38;2;255;0;0m long line\x1b[m",
		"\x1b[38;2;255;0;0ma\x1b[m really really",
		" long line",
		"25% (1/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(9)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[m",
		"\x1b[38;2;255;0;0msecond\x1b[m line",
		"\x1b[38;2;255;0;0ma really really\x1b[m",
		"\x1b[38;2;255;0;0m long line\x1b[m",
		"\x1b[38;2;255;0;0ma\x1b[m really really",
		" long line",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_FooterStyle(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	vp.FooterStyle = lipgloss.NewStyle().Foreground(red)
	setContent(&vp, []string{
		"1",
		"2",
		"3",
		"4",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255m1\x1b[m",
		"2",
		"3",
		"\x1b[38;2;255;0;0m25% (1/4)\x1b[m",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_FooterDisabled(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	setContent(&vp, []string{
		"first line",
		"second line",
		"third line",
		"fourth line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[m",
		"second line",
		"third line",
		"25% (1/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetFooterEnabled(false)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line\x1b[m",
		"second line",
		"third line",
		"fourth line",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_SpaceAround(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	setContent(&vp, []string{
		"    first line     ",
		"          first line          ",
		"               first line               ",
	})
	// trailing space is not trimmed
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255m    first line \x1b[m",
		"\x1b[38;2;0;0;255m    \x1b[m",
		"          first",
		"33% (1/3)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_MultiHeader(t *testing.T) {
	w, h := 15, 2
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header1", "header2"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	setContent(&vp, []string{
		"line1",
		"line2",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header1",
		"header2",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(4)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"\x1b[38;2;0;0;255mline1\x1b[m",
		"50% (1/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"\x1b[38;2;0;0;255mline2\x1b[m",
		"100% (2/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(5)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line1",
		"\x1b[38;2;0;0;255mline2\x1b[m",
		"100% (2/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = pad(vp.width, vp.height, []string{
		"header1",
		"header2",
		"line1",
		"\x1b[38;2;0;0;255mline2\x1b[m",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_OverflowLine(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"long header overflows"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"123456789012345",
		"1234567890123456",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"long header ove",
		"rflows",
		"\x1b[38;2;0;0;255m123456789012345\x1b[m",
		"123456789012345",
		"6",
		"50% (1/2)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_OverflowHeight(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"123456789012345",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
	})
	vp.SetSelectedItemIdx(1)
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"123456789012345",
		"\x1b[38;2;0;0;255m123456789012345\x1b[m",
		"\x1b[38;2;0;0;255m6\x1b[m",
		"123456789012345",
		"33% (2/6)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_Scrolling(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	doSetContent := func() {
		setContent(&vp, []string{
			"first",
			"second",
			"third",
			"fourth",
			"fifth",
			"sixth",
		})
	}
	validate := func(expectedView string) {
		// set content multiple times to confirm no side effects of doing it
		util.CmpStr(t, expectedView, vp.View())
		doSetContent()
		util.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[m",
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
		"\x1b[38;2;0;0;255msecond\x1b[m",
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
		"\x1b[38;2;0;0;255mthird\x1b[m",
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
		"\x1b[38;2;0;0;255msixth\x1b[m",
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
	setContent(&vp, []string{
		"the first line",
		"the second line",
		"the third line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"the second",
		" line",
		"33% (1/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// attempting to scroll so selection out of view is no-op
	vp.ScrollSoItemIdxInView(2)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"the second",
		" line",
		"33% (1/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// move selection down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the first",
		"line",
		"\x1b[38;2;0;0;255mthe second\x1b[m",
		"\x1b[38;2;0;0;255m line\x1b[m",
		"66% (2/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// scroll so third item in view
	vp.ScrollSoItemIdxInView(2)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe second\x1b[m",
		"\x1b[38;2;0;0;255m line\x1b[m",
		"the third",
		"line",
		"66% (2/3)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_BulkScrolling(t *testing.T) {
	w, h := 10, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"the first line",
		"the second line",
		"the third line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"33% (1/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe second\x1b[m",
		"\x1b[38;2;0;0;255m line\x1b[m",
		"66% (2/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// half page down
	vp, _ = vp.Update(halfPgDownKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe third \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"100% (3/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// half page down
	vp, _ = vp.Update(halfPgDownKeyMsg)
	util.CmpStr(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe second\x1b[m",
		"\x1b[38;2;0;0;255m line\x1b[m",
		"66% (2/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"33% (1/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	util.CmpStr(t, expectedView, vp.View())

	// go to bottom
	vp, _ = vp.Update(goToBottomKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe third \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"100% (3/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// go to top
	vp, _ = vp.Update(goToTopKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"33% (1/3)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_Panning(t *testing.T) {
	w, h := 10, 7
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header long"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	doSetContent := func() {
		setContent(&vp, []string{
			"first line that is fairly long",
			"second line that is even much longer than the first",
			"third line that is fairly long as well",
			"fourth kinda long",
			"fifth kinda long too",
			"sixth",
		})
	}
	validate := func(expectedView string) {
		// set content multiple times to confirm no side effects of doing it
		util.CmpStr(t, expectedView, vp.View())
		doSetContent()
		util.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
	expectedView := pad(vp.width, vp.height, []string{
		"header lon",
		"g",
		"\x1b[38;2;0;0;255mfirst line\x1b[m",
		"\x1b[38;2;0;0;255m that is f\x1b[m",
		"\x1b[38;2;0;0;255mairly long\x1b[m",
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
		"\x1b[38;2;0;0;255msecond lin\x1b[m",
		"\x1b[38;2;0;0;255me that is \x1b[m",
		"\x1b[38;2;0;0;255meven much \x1b[m",
		"\x1b[38;2;0;0;255mlonger tha\x1b[m",
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
		"\x1b[38;2;0;0;255mthird line\x1b[m",
		"\x1b[38;2;0;0;255m that is f\x1b[m",
		"\x1b[38;2;0;0;255mairly long\x1b[m",
		"\x1b[38;2;0;0;255m as well\x1b[m",
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
		"\x1b[38;2;0;0;255mfourth kin\x1b[m",
		"\x1b[38;2;0;0;255mda long\x1b[m",
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
		"\x1b[38;2;0;0;255mfifth kind\x1b[m",
		"\x1b[38;2;0;0;255ma long too\x1b[m",
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
		"\x1b[38;2;0;0;255msixth\x1b[m",
		"100% (6/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header lon",
		"g",
		"da long",
		"\x1b[38;2;0;0;255mfifth kind\x1b[m",
		"\x1b[38;2;0;0;255ma long too\x1b[m",
		"sixth",
		"83% (5/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header lon",
		"g",
		"\x1b[38;2;0;0;255mfourth kin\x1b[m",
		"\x1b[38;2;0;0;255mda long\x1b[m",
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
		"\x1b[38;2;0;0;255mthird line\x1b[m",
		"\x1b[38;2;0;0;255m that is f\x1b[m",
		"\x1b[38;2;0;0;255mairly long\x1b[m",
		"\x1b[38;2;0;0;255m as well\x1b[m",
		"50% (3/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header lon",
		"g",
		"\x1b[38;2;0;0;255msecond lin\x1b[m",
		"\x1b[38;2;0;0;255me that is \x1b[m",
		"\x1b[38;2;0;0;255meven much \x1b[m",
		"\x1b[38;2;0;0;255mlonger tha\x1b[m",
		"33% (2/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header lon",
		"g",
		"\x1b[38;2;0;0;255mfirst line\x1b[m",
		"\x1b[38;2;0;0;255m that is f\x1b[m",
		"\x1b[38;2;0;0;255mairly long\x1b[m",
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
	setContent(&vp, []string{
		"sixth item",
		"seventh item",
		"eighth item",
		"ninth item",
		"tenth item",
		"eleventh item",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255msixth item\x1b[m",
		"seventh it",
		"em",
		"eighth ite",
		"16% (1/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// selection down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"sixth item",
		"\x1b[38;2;0;0;255mseventh it\x1b[m",
		"\x1b[38;2;0;0;255mem\x1b[m",
		"eighth ite",
		"33% (2/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// add content above
	setContent(&vp, []string{
		"first item",
		"second item",
		"third item",
		"fourth item",
		"fifth item",
		"sixth item",
		"seventh item",
		"eighth item",
		"ninth item",
		"tenth item",
		"eleventh item",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"sixth item",
		"\x1b[38;2;0;0;255mseventh it\x1b[m",
		"\x1b[38;2;0;0;255mem\x1b[m",
		"eighth ite",
		"63% (7/11)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// add content below
	setContent(&vp, []string{
		"first item",
		"second item",
		"third item",
		"fourth item",
		"fifth item",
		"sixth item",
		"seventh item",
		"eighth item",
		"ninth item",
		"tenth item",
		"eleventh item",
		"twelfth item",
		"thirteenth item",
		"fourteenth item",
		"fifteenth item",
		"sixteenth item",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"sixth item",
		"\x1b[38;2;0;0;255mseventh it\x1b[m",
		"\x1b[38;2;0;0;255mem\x1b[m",
		"eighth ite",
		"43% (7/16)",
	})
	util.CmpStr(t, expectedView, vp.View())
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
	setContent(&vp, []string{
		"the first line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"100% (1/1)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// add content
	setContent(&vp, []string{
		"the second line",
		"the first line",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe second\x1b[m",
		"\x1b[38;2;0;0;255m line\x1b[m",
		"50% (1/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// de-activate by moving selection down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"100% (2/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// add content
	setContent(&vp, []string{
		"the second line",
		"the first line",
		"the third line",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"66% (2/3)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_StickyBottom(t *testing.T) {
	w, h := 10, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	// stickyness should override maintain selection
	vp.SetMaintainSelection(true)
	vp.SetBottomSticky(true)
	setContent(&vp, []string{
		"the first line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
	})
	util.CmpStr(t, expectedView, vp.View())

	// add content
	setContent(&vp, []string{
		"the second line",
		"the first line",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the second",
		" line",
		"\x1b[38;2;0;0;255mthe first \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"100% (2/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// add longer content at bottom
	setContent(&vp, []string{
		"the second line",
		"the first line",
		"a very long line that wraps a lot",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255ma very lon\x1b[m",
		"\x1b[38;2;0;0;255mg line tha\x1b[m",
		"\x1b[38;2;0;0;255mt wraps a \x1b[m",
		"\x1b[38;2;0;0;255mlot\x1b[m",
		"100% (3/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// de-activate by moving selection up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"a very lon",
		"g line tha",
		"66% (2/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// add content
	setContent(&vp, []string{
		"the second line",
		"the first line",
		"a very long line that wraps a lot",
		"the third line",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"a very lon",
		"g line tha",
		"50% (2/4)",
	})
	util.CmpStr(t, expectedView, vp.View())
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
	setContent(&vp, []string{})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
	})
	util.CmpStr(t, expectedView, vp.View())

	setContent(&vp, []string{
		"the second line",
		"the first line",
		"the third line",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe third \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"100% (3/3)",
	})
	util.CmpStr(t, expectedView, vp.View())
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
	setContent(&vp, []string{
		"the first line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"100% (1/1)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// add content, top sticky wins out arbitrarily when both set
	setContent(&vp, []string{
		"the second line",
		"the first line",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe second\x1b[m",
		"\x1b[38;2;0;0;255m line\x1b[m",
		"50% (1/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// selection to bottom
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"100% (2/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// add content
	setContent(&vp, []string{
		"the second line",
		"the first line",
		"the third line",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe third \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"100% (3/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// de-activate by moving selection up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"66% (2/3)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// add content
	setContent(&vp, []string{
		"the second line",
		"the first line",
		"the third line",
		"the fourth line",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"50% (2/4)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_StickyBottomLongLine(t *testing.T) {
	w, h := 10, 10
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	// stickyness should override maintain selection
	vp.SetMaintainSelection(true)
	vp.SetBottomSticky(true)
	setContent(&vp, []string{
		"first line",
		"next line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"\x1b[38;2;0;0;255mnext line\x1b[m",
	})
	util.CmpStr(t, expectedView, vp.View())

	setContent(&vp, []string{
		"first line",
		"next line",
		"a very long line at the bottom that wraps many times",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first line",
		"next line",
		"\x1b[38;2;0;0;255ma very lon\x1b[m",
		"\x1b[38;2;0;0;255mg line at \x1b[m",
		"\x1b[38;2;0;0;255mthe bottom\x1b[m",
		"\x1b[38;2;0;0;255m that wrap\x1b[m",
		"\x1b[38;2;0;0;255ms many tim\x1b[m",
		"\x1b[38;2;0;0;255mes\x1b[m",
		"100% (3/3)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_RemoveLogsWhenSelectionBottom(t *testing.T) {
	w, h := 10, 3
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)

	// add content
	setContent(&vp, []string{
		"the second line",
		"the first line",
		"the third line",
		"the fourth line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe second\x1b[m",
		"25% (1/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// selection to bottom
	vp.SetSelectedItemIdx(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe fourth\x1b[m",
		"100% (4/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// remove content
	setContent(&vp, []string{
		"the second line",
		"the first line",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[m",
		"100% (2/2)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_ChangeHeight(t *testing.T) {
	w, h := 10, 3
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"the first line",
		"the second line",
		"the third line",
		"the fourth line",
		"the fifth line",
		"the sixth line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[m",
		"16% (1/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(6)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"the second",
		" line",
		"16% (1/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// move selection to third line
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the second",
		" line",
		"\x1b[38;2;0;0;255mthe third \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"50% (3/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// reduce height
	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe third \x1b[m",
		"50% (3/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(8)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe third \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"the fourth",
		" line",
		"the fifth ",
		"line",
		"50% (3/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

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
		"\x1b[38;2;0;0;255mthe sixth \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"100% (6/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// reduce height
	vp.SetHeight(3)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe sixth \x1b[m",
		"100% (6/6)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_ChangeContent(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"the first line",
		"the second line",
		"the third line",
		"the fourth line",
		"the fifth line",
		"the sixth line",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"the second",
		"16% (1/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// move selection to bottom
	vp.SetSelectedItemIdx(5)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"line",
		"\x1b[38;2;0;0;255mthe sixth \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"100% (6/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// remove content
	setContent(&vp, []string{
		"the second line",
		"the third line",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		" line",
		"\x1b[38;2;0;0;255mthe third \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"100% (2/2)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// remove all content
	setContent(&vp, []string{})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
	})
	util.CmpStr(t, expectedView, vp.View())

	// add content
	setContent(&vp, []string{
		"the first line",
		"the second line",
		"the third line",
		"the fourth line",
		"the fifth line",
		"the sixth line",
	})
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe first \x1b[m",
		"\x1b[38;2;0;0;255mline\x1b[m",
		"the second",
		"16% (1/6)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_StringToHighlight(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	vp.SetStringToHighlight("second")
	vp.HighlightStyle = lipgloss.NewStyle().Foreground(green)
	vp.HighlightStyleIfSelected = lipgloss.NewStyle().Foreground(red)
	setContent(&vp, []string{
		"first",
		"second",
		"second",
		"third",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst\x1b[m",
		"\x1b[38;2;0;255;0msecond\x1b[m",
		"\x1b[38;2;0;255;0msecond\x1b[m",
		"25% (1/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	vp.SetStringToHighlight("first")
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;255;0;0mfirst\x1b[m",
		"second",
		"second",
		"25% (1/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	setContent(&vp, []string{
		"averylongwordthatwrapsover",
	})
	vp.SetStringToHighlight("wraps")
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255maverylongw\x1b[m",
		"\x1b[38;2;0;0;255mordthat\x1b[m\x1b[38;2;255;0;0mwra\x1b[m",
		"\x1b[38;2;255;0;0mps\x1b[m\x1b[38;2;0;0;255mover\x1b[m",
		"100% (1/1)",
	})
	util.CmpStr(t, expectedView, vp.View())

	setContent(&vp, []string{
		"a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line ",
	})
	vp.SetStringToHighlight("l")
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255ma super \x1b[m\x1b[38;2;255;0;0ml\x1b[m\x1b[38;2;0;0;255mo\x1b[m",
		"\x1b[38;2;0;0;255mng \x1b[m\x1b[38;2;255;0;0ml\x1b[m\x1b[38;2;0;0;255mine a \x1b[m",
		"\x1b[38;2;0;0;255msuper \x1b[m\x1b[38;2;255;0;0ml\x1b[m\x1b[38;2;0;0;255mong\x1b[m",
		"100% (1/1)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_StringToHighlightManyMatches(t *testing.T) {
	runTest := func(t *testing.T) {
		w, h := 10, 5
		vp := newViewport(w, h)
		vp.SetHeader([]string{"header"})
		vp.SetSelectionEnabled(true)
		vp.SetWrapText(true)
		setContent(&vp, []string{
			strings.Repeat("r", 100000),
		})
		vp.SetStringToHighlight("r")
		vp.HighlightStyle = lipgloss.NewStyle().Foreground(green)
		vp.HighlightStyleIfSelected = lipgloss.NewStyle().Foreground(red)
		expectedView := pad(vp.width, vp.height, []string{
			"header",
			strings.Repeat("\x1b[38;2;255;0;0mr\x1b[m", 10),
			strings.Repeat("\x1b[38;2;255;0;0mr\x1b[m", 10),
			strings.Repeat("\x1b[38;2;255;0;0mr\x1b[m", 10),
			"100% (1/1)",
		})
		util.CmpStr(t, expectedView, vp.View())
	}
	util.RunWithTimeout(t, runTest, 10*time.Millisecond)
}

func TestViewport_SelectionOn_WrapOn_AnsiOnSelection(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	setContent(&vp, []string{
		"line with some \x1b[38;2;255;0;0mred\x1b[m text",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mline with \x1b[m",
		"\x1b[38;2;0;0;255msome \x1b[m\x1b[38;2;255;0;0mred\x1b[m\x1b[38;2;0;0;255m t\x1b[m",
		"\x1b[38;2;0;0;255mext\x1b[m",
		"100% (1/1)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_SelectionEmpty(t *testing.T) {
	w, h := 20, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	setContent(&vp, []string{
		"",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255m \x1b[m",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_ExtraSlash(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	setContent(&vp, []string{
		"|2024|\x1b[38;2;0mfl..lq\x1b[m/\x1b[38;2;0mflask-3\x1b[m|",
	})
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255m|2024|\x1b[m\x1b[38;2;0mfl..\x1b[m",
		"\x1b[38;2;0mlq\x1b[m\x1b[38;2;0;0;255m/\x1b[m\x1b[38;2;0mflask-3\x1b[m",
		"\x1b[38;2;0;0;255m|\x1b[m",
		"100% (1/1)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_SuperLongWrappedLine(t *testing.T) {
	runTest := func(t *testing.T) {
		w, h := 10, 5
		vp := newViewport(w, h)
		vp.SetHeader([]string{"header"})
		vp.SetSelectionEnabled(true)
		vp.SetWrapText(true)
		setContent(&vp, []string{
			"smol",
			strings.Repeat("12345678", 1000000),
			"smol",
		})
		expectedView := pad(vp.width, vp.height, []string{
			"header",
			"\x1b[38;2;0;0;255msmol\x1b[m",
			"1234567812",
			"3456781234",
			"33% (1/3)",
		})
		util.CmpStr(t, expectedView, vp.View())

		vp, _ = vp.Update(downKeyMsg)
		expectedView = pad(vp.width, vp.height, []string{
			"header",
			"\x1b[38;2;0;0;255m1234567812\x1b[m",
			"\x1b[38;2;0;0;255m3456781234\x1b[m",
			"\x1b[38;2;0;0;255m5678123456\x1b[m",
			"66% (2/3)",
		})
		util.CmpStr(t, expectedView, vp.View())

		vp, _ = vp.Update(downKeyMsg)
		expectedView = pad(vp.width, vp.height, []string{
			"header",
			"5678123456",
			"7812345678",
			"\x1b[38;2;0;0;255msmol\x1b[m",
			"100% (3/3)",
		})
		util.CmpStr(t, expectedView, vp.View())
	}
	util.RunWithTimeout(t, runTest, 500*time.Millisecond)
}

func TestViewport_SelectionOn_WrapOn_StringToHighlightAnsiUnicode(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b) = 6w, 11b
	vp.SetHeader([]string{"A💖中é"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	setContent(&vp, []string{
		"A💖中é",
		"A💖中éA💖中é",
	})
	vp.SetStringToHighlight("中é")
	vp.HighlightStyle = lipgloss.NewStyle().Foreground(green)
	vp.HighlightStyleIfSelected = lipgloss.NewStyle().Foreground(red)
	expectedView := pad(vp.width, vp.height, []string{
		"A💖中é",
		"\x1b[38;2;0;0;255mA💖\x1b[m\x1b[38;2;255;0;0m中é\x1b[m",
		"A💖\x1b[38;2;0;255;0m中é\x1b[mA💖",
		"\x1b[38;2;0;255;0m中é\x1b[m",
		"50% (1/2)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

// # OTHER

func TestViewport_SelectionOn_ToggleWrap_PreserveSelection(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"first line that is fairly long",
		"second line that is even much longer than the first",
		"third line that is fairly long",
		"fourth",
		"fifth line that is fairly long",
		"sixth",
	})

	// wrap off, selection on first line
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line t...\x1b[m",
		"second line ...",
		"third line t...",
		"fourth",
		"16% (1/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// move selection to third line
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first line t...",
		"second line ...",
		"\x1b[38;2;0;0;255mthird line t...\x1b[m",
		"fourth",
		"50% (3/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// toggle wrap on
	vp.SetWrapText(true)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"longer than the",
		" first",
		"\x1b[38;2;0;0;255mthird line that\x1b[m",
		"\x1b[38;2;0;0;255m is fairly long\x1b[m",
		"50% (3/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// toggle wrap off
	vp.SetWrapText(false)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"first line t...",
		"second line ...",
		"\x1b[38;2;0;0;255mthird line t...\x1b[m",
		"fourth",
		"50% (3/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// move selection to last line
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"third line t...",
		"fourth",
		"fifth line t...",
		"\x1b[38;2;0;0;255msixth\x1b[m",
		"100% (6/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// toggle wrap on
	vp.SetWrapText(true)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"fourth",
		"fifth line that",
		" is fairly long",
		"\x1b[38;2;0;0;255msixth\x1b[m",
		"100% (6/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// toggle wrap off
	vp.SetWrapText(false)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"third line t...",
		"fourth",
		"fifth line t...",
		"\x1b[38;2;0;0;255msixth\x1b[m",
		"100% (6/6)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_ToggleWrap_PreserveSelectionInView(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"a really really really really really really really really really really really really long preamble",
		"first line that is fairly long",
		"second line that is even much longer than the first",
		"third line that is fairly long",
	})
	vp.SetSelectedItemIdx(3)
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"a really rea...",
		"first line t...",
		"second line ...",
		"\x1b[38;2;0;0;255mthird line t...\x1b[m",
		"100% (4/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// toggle wrap, full wrapped selection should remain in view
	vp.SetWrapText(true)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"longer than the",
		" first",
		"\x1b[38;2;0;0;255mthird line that\x1b[m",
		"\x1b[38;2;0;0;255m is fairly long\x1b[m",
		"100% (4/4)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// toggle wrap
	vp.SetWrapText(false)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"a really rea...",
		"first line t...",
		"second line ...",
		"\x1b[38;2;0;0;255mthird line t...\x1b[m",
		"100% (4/4)",
	})
	util.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_ToggleWrap_ScrollInBounds(t *testing.T) {
	w, h := 10, 7
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent(&vp, []string{
		"the first line",
		"the second line",
		"the third line",
		"the fourth line",
		"the fifth line",
		"the sixth line",
	})

	// scroll to bottom with selection at top of that view
	vp.SetSelectedItemIdx(5)
	vp, _ = vp.Update(upKeyMsg)
	vp, _ = vp.Update(upKeyMsg)
	expectedView := pad(vp.width, vp.height, []string{
		"header",
		"\x1b[38;2;0;0;255mthe fourth\x1b[m",
		"\x1b[38;2;0;0;255m line\x1b[m",
		"the fifth ",
		"line",
		"the sixth ",
		"66% (4/6)",
	})
	util.CmpStr(t, expectedView, vp.View())

	// toggle wrap
	vp.SetWrapText(false)
	expectedView = pad(vp.width, vp.height, []string{
		"header",
		"the sec...",
		"the thi...",
		"\x1b[38;2;0;0;255mthe fou...\x1b[m",
		"the fif...",
		"the six...",
		"66% (4/6)",
	})
	util.CmpStr(t, expectedView, vp.View())
}
