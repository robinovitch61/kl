package viewport

import (
	"github.com/google/go-cmp/cmp"
	"strings"
	"testing"
)

// pad pads the given lines to the given width and height.
// for example, pad(5, 4, []string{"a", "b", "c"}) will be padded to:
// "a    "
// "b    "
// "c    "
// "     "
// as a single string
func pad(width, height int, lines []string) string {
	var res []string
	for _, line := range lines {
		resLine := line
		numSpaces := width - len(line)
		resLine += strings.Repeat(" ", numSpaces)
		res = append(res, resLine)
	}
	numEmptyLines := height - len(lines)
	for i := 0; i < numEmptyLines; i++ {
		res = append(res, strings.Repeat(" ", width))
	}
	return strings.Join(res, "\n")
}

func TestPad(t *testing.T) {
	width, height := 5, 4
	lines := []string{"a", "b", "c"}
	expected := `a    
b    
c    
     `
	if diff := cmp.Diff(expected, pad(width, height, lines)); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}
}

func newViewport(width, height int) Model[RenderableString] {
	return New[RenderableString](width, height)
}

func TestViewport_OneLine(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetContent([]RenderableString{
		{Content: "first line"},
		{Content: "second line"},
	})
	expectedView := pad(w, h, []string{
		"first line",
		"second line",
	})
	if diff := cmp.Diff(expectedView, vp.View()); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestViewport_OverflowLineNoWrap(t *testing.T) {
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
	if diff := cmp.Diff(expectedView, vp.View()); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestViewport_OverflowLineWrap(t *testing.T) {
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
	if diff := cmp.Diff(expectedView, vp.View()); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestViewport_OverflowHeightNoWrap(t *testing.T) {
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
	if diff := cmp.Diff(expectedView, vp.View()); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}
}
