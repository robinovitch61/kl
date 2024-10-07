package viewport

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-cmp/cmp"
	"github.com/muesli/termenv"
	"io"
	"testing"
)

var (
	testKeyMap     = DefaultKeyMap()
	downKeyMsg     = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(testKeyMap.Down.Keys()[0])}
	upKeyMsg       = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(testKeyMap.Up.Keys()[0])}
	red            = lipgloss.Color("#ff0000")
	blue           = lipgloss.Color("#0000ff")
	selectionStyle = renderer().NewStyle().Foreground(blue)
)

func renderer() *lipgloss.Renderer {
	r := lipgloss.NewRenderer(io.Discard)
	r.SetColorProfile(termenv.TrueColor)
	r.SetHasDarkBackground(true)
	return r
}

// # SELECTION DISABLED, WRAP OFF

func TestViewport_SelectionOff_WrapOff_Basic(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetContent([]RenderableString{
		{Content: "first line"},
		{Content: renderer().NewStyle().Foreground(red).Render("second") + " line"},
		{Content: renderer().NewStyle().Foreground(red).Render("a really really long line")},
		{Content: renderer().NewStyle().Foreground(red).Render("a") + " really really long line"},
	})
	expectedView := pad(w, h, []string{
		"first line",
		"\x1b[38;2;255;0;0msecond\x1b[0m line",
		"\x1b[38;2;255;0;0ma really rea...\x1b[0m",
		"\x1b[38;2;255;0;0ma\x1b[0m really rea...",
	})
	if diff := cmp.Diff(expectedView, vp.View()); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestViewport_SelectionOff_WrapOff_OverflowLine(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetContent([]RenderableString{
		{Content: "123456789012345"},
		{Content: "1234567890123456"},
	})
	expectedView := pad(w, h, []string{
		"123456789012345",
		"123456789012...",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_OverflowHeight(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetContent([]RenderableString{
		{Content: "123456789012345"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
	})
	expectedView := pad(w, h, []string{
		"123456789012345",
		"123456789012...",
		"123456789012...",
		"123456789012...",
		"66% (4/6)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_Scrolling(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
		{Content: "third"},
		{Content: "fourth"},
		{Content: "fifth"},
		{Content: "sixth"},
	})
	expectedView := pad(w, h, []string{
		"first",
		"second",
		"third",
		"fourth",
		"66% (4/6)",
	})
	compare(t, expectedView, vp.View())

	// scrolling up past top is no-op
	vp, _ = vp.Update(upKeyMsg)
	compare(t, expectedView, vp.View())

	// scrolling down by one
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(w, h, []string{
		"second",
		"third",
		"fourth",
		"fifth",
		"83% (5/6)",
	})
	compare(t, expectedView, vp.View())

	// scrolling down by one again
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(w, h, []string{
		"third",
		"fourth",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())

	// scrolling down past bottom is no-op
	vp, _ = vp.Update(downKeyMsg)
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_Panning(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetContent([]RenderableString{
		{Content: "first line that is fairly long"},
		{Content: "second line that is even much longer than the first"},
		{Content: "third line that is fairly long"},
		{Content: "fourth"},
		{Content: "fifth line that is fairly long"},
		{Content: "sixth"},
	})
	expectedView := pad(w, h, []string{
		"first l...",
		"second ...",
		"third l...",
		"fourth",
		"66% (4/6)",
	})
	compare(t, expectedView, vp.View())

	// pan right
	vp.setXOffset(5)
	expectedView = pad(w, h, []string{
		"...ne t...",
		"...ine ...",
		"...ne t...",
		".",
		"66% (4/6)",
	})
	compare(t, expectedView, vp.View())

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(w, h, []string{
		"...ine ...",
		"...ne t...",
		".",
		"...ne t...",
		"83% (5/6)",
	})
	compare(t, expectedView, vp.View())

	// pan all the way right
	vp.setXOffset(41)
	expectedView = pad(w, h, []string{
		"...e first",
		"",
		"",
		"",
		"83% (5/6)",
	})
	compare(t, expectedView, vp.View())

	// scroll down
	// TODO LEO: should these be ...?
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(w, h, []string{
		"",
		"",
		"",
		"",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())
}

// # SELECTION ENABLED, WRAP OFF

func TestViewport_SelectionOn_WrapOff_Basic(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "first line"},
		{Content: renderer().NewStyle().Foreground(red).Render("second") + " line"},
		{Content: renderer().NewStyle().Foreground(red).Render("a really really long line")},
		{Content: renderer().NewStyle().Foreground(red).Render("a") + " really really long line"},
	})
	expectedView := pad(w, h, []string{
		"\x1b[38;2;0;0;255mfirst line\x1b[0m",
		"\x1b[38;2;255;0;0msecond\x1b[0m line",
		"\x1b[38;2;255;0;0ma really rea...\x1b[0m",
		"\x1b[38;2;255;0;0ma\x1b[0m really rea...",
	})
	if diff := cmp.Diff(expectedView, vp.View()); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestViewport_SelectionOn_WrapOff_OverflowLine(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "123456789012345"},
		{Content: "1234567890123456"},
	})
	expectedView := pad(w, h, []string{
		"\x1b[38;2;0;0;255m123456789012345\x1b[0m",
		"123456789012...",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_OverflowHeight(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "123456789012345"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
	})
	expectedView := pad(w, h, []string{
		"\x1b[38;2;0;0;255m123456789012345\x1b[0m",
		"123456789012...",
		"123456789012...",
		"123456789012...",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_Scrolling(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
		{Content: "third"},
		{Content: "fourth"},
		{Content: "fifth"},
		{Content: "sixth"},
	})
	expectedView := pad(w, h, []string{
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
		"second",
		"third",
		"fourth",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())

	// scrolling up past top is no-op
	vp, _ = vp.Update(upKeyMsg)
	compare(t, expectedView, vp.View())

	// scrolling down by one
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(w, h, []string{
		"first",
		"\x1b[38;2;0;0;255msecond\x1b[0m",
		"third",
		"fourth",
		"33% (2/6)",
	})
	compare(t, expectedView, vp.View())

	// scrolling to bottom
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(w, h, []string{
		"third",
		"fourth",
		"fifth",
		"\x1b[38;2;0;0;255msixth\x1b[0m",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())

	// scrolling down past bottom is no-op
	vp, _ = vp.Update(downKeyMsg)
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_Panning(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "first line that is fairly long"},
		{Content: "second line that is even much longer than the first"},
		{Content: "third line that is fairly long"},
		{Content: "fourth"},
		{Content: "fifth line that is fairly long"},
		{Content: "sixth"},
	})
	expectedView := pad(w, h, []string{
		"\x1b[38;2;0;0;255mfirst l...\x1b[0m",
		"second ...",
		"third l...",
		"fourth",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())

	// pan right
	vp.setXOffset(5)
	expectedView = pad(w, h, []string{
		"\x1b[38;2;0;0;255m...ne t...\x1b[0m",
		"...ine ...",
		"...ne t...",
		".",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(w, h, []string{
		"...ne t...",
		"\x1b[38;2;0;0;255m...ine ...\x1b[0m",
		"...ne t...",
		".",
		"33% (2/6)",
	})
	compare(t, expectedView, vp.View())

	// pan all the way right
	vp.setXOffset(41)
	expectedView = pad(w, h, []string{
		"",
		"\x1b[38;2;0;0;255m...e first\x1b[0m",
		"",
		"",
		"33% (2/6)",
	})
	compare(t, expectedView, vp.View())

	// scroll down
	// TODO LEO: should these be ...?
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(w, h, []string{
		"",
		"...e first",
		"\x1b[38;2;0;0;255m \x1b[0m",
		"",
		"50% (3/6)",
	})
	compare(t, expectedView, vp.View())
}

// # SELECTION DISABLED, WRAP ON

func TestViewport_SelectionOff_WrapOn_Basic(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "first line"},
		{Content: renderer().NewStyle().Foreground(red).Render("second") + " line"},
		{Content: renderer().NewStyle().Foreground(red).Render("a really really long line")},
		{Content: renderer().NewStyle().Foreground(red).Render("a") + " really really long line"},
	})
	expectedView := pad(w, h, []string{
		"first line",
		"\x1b[38;2;255;0;0msecond\x1b[0m line",
		"\x1b[38;2;255;0;0ma really really\x1b[0m",
		"\x1b[38;2;255;0;0m long line\x1b[0m",
		"66% (4/6)",
	})
	if diff := cmp.Diff(expectedView, vp.View()); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestViewport_SelectionOff_WrapOn_OverflowLine(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "123456789012345"},
		{Content: "1234567890123456"},
	})
	expectedView := pad(w, h, []string{
		"123456789012345",
		"123456789012345",
		"6",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_OverflowHeight(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "123456789012345"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
		{Content: "1234567890123456"},
	})
	expectedView := pad(w, h, []string{
		"123456789012345",
		"123456789012345",
		"6",
		"123456789012345",
		"36% (4/11)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_Scrolling(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
		{Content: "third"},
		{Content: "fourth"},
		{Content: "fifth"},
		{Content: "sixth"},
	})
	expectedView := pad(w, h, []string{
		"first",
		"second",
		"third",
		"fourth",
		"66% (4/6)",
	})
	compare(t, expectedView, vp.View())

	// scrolling up past top is no-op
	vp, _ = vp.Update(upKeyMsg)
	compare(t, expectedView, vp.View())

	// scrolling down by one
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(w, h, []string{
		"second",
		"third",
		"fourth",
		"fifth",
		"83% (5/6)",
	})
	compare(t, expectedView, vp.View())

	// scrolling down by one again
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(w, h, []string{
		"third",
		"fourth",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())

	// scrolling down past bottom is no-op
	vp, _ = vp.Update(downKeyMsg)
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_Panning(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	vp.SetContent([]RenderableString{
		{Content: "first line that is fairly long"},
		{Content: "second line that is even much longer than the first"},
		{Content: "third line that is fairly long"},
		{Content: "fourth"},
		{Content: "fifth line that is fairly long"},
		{Content: "sixth"},
	})
	expectedView := pad(w, h, []string{
		"first line",
		" that is f",
		"airly long",
		"second lin",
		"23% (4/17)",
	})
	compare(t, expectedView, vp.View())

	// pan right
	vp.setXOffset(5)
	compare(t, expectedView, vp.View())

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(w, h, []string{
		" that is f",
		"airly long",
		"second lin",
		"e that is ",
		"29% (5/17)",
	})
	compare(t, expectedView, vp.View())

	// pan all the way right
	vp.setXOffset(41)
	compare(t, expectedView, vp.View())

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(w, h, []string{
		"airly long",
		"second lin",
		"e that is",
		"even much",
		"35% (6/17)",
	})
	compare(t, expectedView, vp.View())
}

// # SELECTION ENABLED, WRAP ON

func TestViewport_SelectionOn_WrapOn_Basic(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "first line"},
		{Content: renderer().NewStyle().Foreground(red).Render("second") + " line"},
		{Content: renderer().NewStyle().Foreground(red).Render("a really really long line")},
		{Content: renderer().NewStyle().Foreground(red).Render("a") + " really really long line"},
	})
	expectedView := pad(w, h, []string{
		"\x1b[38;2;0;0;255mfirst line\x1b[0m",
		"\x1b[38;2;255;0;0msecond\x1b[0m line",
		"\x1b[38;2;255;0;0ma really really\x1b[0m",
		"\x1b[38;2;255;0;0m long line\x1b[0m",
		"25% (1/4)",
	})
	if diff := cmp.Diff(expectedView, vp.View()); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestViewport_SelectionOn_WrapOn_OverflowLine(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "123456789012345"},
		{Content: "1234567890123456"},
	})
	expectedView := pad(w, h, []string{
		"\x1b[38;2;0;0;255m123456789012345\x1b[0m",
		"123456789012345",
		"6",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_OverflowHeight(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
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
	vp.SetSelectedContentIdx(1)
	expectedView := pad(w, h, []string{
		"123456789012345",
		"\x1b[38;2;0;0;255m123456789012345\x1b[0m",
		"\x1b[38;2;0;0;255m6\x1b[0m",
		"123456789012345",
		"33% (2/6)",
	})
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_Scrolling(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "first"},
		{Content: "second"},
		{Content: "third"},
		{Content: "fourth"},
		{Content: "fifth"},
		{Content: "sixth"},
	})
	expectedView := pad(w, h, []string{
		"\x1b[38;2;0;0;255mfirst\x1b[0m",
		"second",
		"third",
		"fourth",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())

	// scrolling up past top is no-op
	vp, _ = vp.Update(upKeyMsg)
	compare(t, expectedView, vp.View())

	// scrolling down by one
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(w, h, []string{
		"first",
		"\x1b[38;2;0;0;255msecond\x1b[0m",
		"third",
		"fourth",
		"33% (2/6)",
	})
	compare(t, expectedView, vp.View())

	// scrolling down by one again
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(w, h, []string{
		"first",
		"second",
		"\x1b[38;2;0;0;255mthird\x1b[0m",
		"fourth",
		"50% (3/6)",
	})
	compare(t, expectedView, vp.View())

	// scroll to bottom
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(w, h, []string{
		"third",
		"fourth",
		"fifth",
		"\x1b[38;2;0;0;255msixth\x1b[0m",
		"100% (6/6)",
	})
	compare(t, expectedView, vp.View())

	// scrolling past bottom is a no-op
	vp, _ = vp.Update(downKeyMsg)
	compare(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_Panning(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	vp.SetContent([]RenderableString{
		{Content: "first line that is fairly long"},
		{Content: "second line that is even much longer than the first"},
		{Content: "third line that is fairly long"},
		{Content: "fourth"},
		{Content: "fifth line that is fairly long"},
		{Content: "sixth"},
	})
	expectedView := pad(w, h, []string{
		"\x1b[38;2;0;0;255mfirst line\x1b[0m",
		"\x1b[38;2;0;0;255m that is f\x1b[0m",
		"\x1b[38;2;0;0;255mairly long\x1b[0m",
		"second lin",
		"16% (1/6)",
	})
	compare(t, expectedView, vp.View())

	// pan right
	vp.setXOffset(5)
	compare(t, expectedView, vp.View())

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(w, h, []string{
		"\x1b[38;2;0;0;255msecond lin\x1b[0m",
		"\x1b[38;2;0;0;255me that is \x1b[0m",
		"\x1b[38;2;0;0;255meven much \x1b[0m",
		"\x1b[38;2;0;0;255mlonger the\x1b[0m",
		"33% (2/6)",
	})
	compare(t, expectedView, vp.View())

	// pan all the way right
	vp.setXOffset(41)
	compare(t, expectedView, vp.View())

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(w, h, []string{
		"\x1b[38;2;0;0;255mthird line\x1b[0m",
		"\x1b[38;2;0;0;255mthat is fa\x1b[0m",
		"\x1b[38;2;0;0;255mirly long\x1b[0m",
		"fourth",
		"50% (3/6)",
	})
	compare(t, expectedView, vp.View())
}

// TODO:
// add header to all test cases
// transitioning between wrap/no wrap should preserve selection + position relative to top
// adding new content should preserve selected line when maintain selection enabled
// test with a bunch of spaces at end of line(s)
// test string to highlight
// zero & one width/height viewport in each case
