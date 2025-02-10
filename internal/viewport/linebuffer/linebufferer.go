package linebuffer

import "github.com/charmbracelet/lipgloss/v2"

type LineBufferer interface {
	Repr() string
	Width() int
	Content() string
	Take(
		startRuneIdx, takeWidth int,
		continuation, toHighlight string,
		highlightStyle lipgloss.Style,
	) (string, int)
	WrappedLines(
		width int,
		maxLinesEachEnd int,
		toHighlight string,
		toHighlightStyle lipgloss.Style,
	) []string
}
