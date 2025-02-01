package viewport

import (
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/kl/internal/viewport/linebuffer"
	"strings"
)

func percent(a, b int) int {
	return int(float32(a) / float32(b) * 100)
}

// pad is a test helper function that pads the given lines to the given width and height.
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
		numSpaces := width - lipgloss.Width(line)
		if numSpaces > 0 {
			resLine += strings.Repeat(" ", numSpaces)
		}
		res = append(res, resLine)
	}
	numEmptyLines := height - len(lines)
	for i := 0; i < numEmptyLines; i++ {
		res = append(res, strings.Repeat(" ", width))
	}
	return strings.Join(res, "\n")
}

func safeSliceUpToIdx[T any](s []T, i int) []T {
	if i > len(s) {
		return s
	}
	if i < 0 {
		return []T{}
	}
	return s[:i]
}

func safeSliceFromIdx(s []string, i int) []string {
	if i < 0 {
		return s
	}
	if i > len(s) {
		return []string{}
	}
	return s[i:]
}

func clampValMinMax(v, minimum, maximum int) int {
	return max(minimum, min(maximum, v))
}

func setContent(vp *Model[RenderableString], content []string) {
	renderableStrings := make([]RenderableString, len(content))
	for i := range content {
		lb := linebuffer.New(content[i])
		renderableStrings[i] = RenderableString{LineBuffer: &lb}
	}
	vp.SetContent(renderableStrings)
}
