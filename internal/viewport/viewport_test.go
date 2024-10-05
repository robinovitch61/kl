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
	testKeyMap = DefaultKeyMap()
	downKeyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(testKeyMap.Down.Keys()[0])}
	upKeyMsg   = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(testKeyMap.Up.Keys()[0])}
	red        = lipgloss.Color("#ff0000")
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

// # SELECTION DISABLED, WRAP ON

//func TestViewport_OverflowLineWrap(t *testing.T) {
//	w, h := 15, 5
//	vp := newViewport(w, h)
//	vp.SetWrapText(true)
//	vp.SetContent([]RenderableString{
//		{Content: "123456789012345"},
//		{Content: "1234567890123456"},
//	})
//	expectedView := pad(w, h, []string{
//		"123456789012345",
//		"123456789012345",
//		"6",
//	})
//	compare(t, expectedView, vp.View())
//}

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

//func TestViewport_SelectionOff_WrapOff_OverflowHeight(t *testing.T) {
//	w, h := 15, 5
//	vp := newViewport(w, h)
//	vp.SetWrapText(true)
//	vp.SetContent([]RenderableString{
//		{Content: "123456789012345"},
//		{Content: "1234567890123456"},
//		{Content: "1234567890123456"},
//		{Content: "1234567890123456"},
//		{Content: "1234567890123456"},
//		{Content: "1234567890123456"},
//	})
//	expectedView := pad(w, h, []string{
//		"123456789012345",
//		"123456789012...",
//		"123456789012...",
//		"123456789012...",
//		"66% (4/6)",
//	})
//	compare(t, expectedView, vp.View())
//}
//
//func TestViewport_SelectionOff_WrapOn_Scrolling(t *testing.T) {
//	w, h := 15, 5
//	vp := newViewport(w, h)
//	vp.SetWrapText(true)
//	vp.SetContent([]RenderableString{
//		{Content: "first"},
//		{Content: "second"},
//		{Content: "third"},
//		{Content: "fourth"},
//		{Content: "fifth"},
//		{Content: "sixth"},
//	})
//	expectedView := pad(w, h, []string{
//		"first",
//		"second",
//		"third",
//		"fourth",
//		"66% (4/6)",
//	})
//	compare(t, expectedView, vp.View())
//
//	// scrolling up past top is no-op
//	vp, _ = vp.Update(upKeyMsg)
//	compare(t, expectedView, vp.View())
//
//	// scrolling down by one
//	vp, _ = vp.Update(downKeyMsg)
//	expectedView = pad(w, h, []string{
//		"second",
//		"third",
//		"fourth",
//		"fifth",
//		"83% (5/6)",
//	})
//	compare(t, expectedView, vp.View())
//
//	// scrolling down by one again
//	vp, _ = vp.Update(downKeyMsg)
//	expectedView = pad(w, h, []string{
//		"third",
//		"fourth",
//		"fifth",
//		"sixth",
//		"100% (6/6)",
//	})
//	compare(t, expectedView, vp.View())
//
//	// scrolling down past bottom is no-op
//	vp, _ = vp.Update(downKeyMsg)
//	compare(t, expectedView, vp.View())
//}
//
//func TestViewport_SelectionOff_WrapOff_Panning(t *testing.T) {
//	w, h := 10, 5
//	vp := newViewport(w, h)
//	vp.SetWrapText(true)
//	vp.SetContent([]RenderableString{
//		{Content: "first line that is fairly long"},
//		{Content: "second line that is even much longer than the first"},
//		{Content: "third line that is fairly long"},
//		{Content: "fourth"},
//		{Content: "fifth line that is fairly long"},
//		{Content: "sixth"},
//	})
//	expectedView := pad(w, h, []string{
//		"first l...",
//		"second ...",
//		"third l...",
//		"fourth",
//		"66% (4/6)",
//	})
//	compare(t, expectedView, vp.View())
//
//	// pan right
//	vp.setXOffset(5)
//	expectedView = pad(w, h, []string{
//		"...ne t...",
//		"...ine ...",
//		"...ne t...",
//		".",
//		"66% (4/6)",
//	})
//	compare(t, expectedView, vp.View())
//
//	// scroll down
//	vp, _ = vp.Update(downKeyMsg)
//	expectedView = pad(w, h, []string{
//		"...ine ...",
//		"...ne t...",
//		".",
//		"...ne t...",
//		"83% (5/6)",
//	})
//	compare(t, expectedView, vp.View())
//
//	// pan all the way right
//	vp.setXOffset(41)
//	expectedView = pad(w, h, []string{
//		"...e first",
//		"",
//		"",
//		"",
//		"83% (5/6)",
//	})
//	compare(t, expectedView, vp.View())
//
//	// scroll down
//	// TODO LEO: should these be ...?
//	vp, _ = vp.Update(downKeyMsg)
//	expectedView = pad(w, h, []string{
//		"",
//		"",
//		"",
//		"",
//		"100% (6/6)",
//	})
//	compare(t, expectedView, vp.View())
//}

// TODO:
// transitioning between wrap/no wrap should preserve selection + position relative to top
// adding new content should preserve selected line when maintain selection enabled
// test with a bunch of spaces at end of line(s)
