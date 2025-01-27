package linebuffer

import "github.com/charmbracelet/lipgloss/v2"

type LineBufferer interface {
	Width() int
	SeekToWidth(width int)
	PopLeft(
		width int,
		continuation, toHighlight string,
		highlightStyle lipgloss.Style,
	) string
	WrappedLines(
		width int,
		maxLinesEachEnd int,
		toHighlight string,
		toHighlightStyle lipgloss.Style,
	) []string
}
