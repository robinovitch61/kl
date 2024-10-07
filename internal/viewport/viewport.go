package viewport

import (
	"fmt"
	"github.com/robinovitch61/kl/internal/dev"
	"regexp"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	ansiPattern = regexp.MustCompile("\x1b\\[[0-9;]*m")
)

// Terminology:
// - content: an item to be rendered in the viewport
// - line: a row in the terminal
//
// wrap disabled:
//                           content index   line index
// this is the first line    0               0
// this is the second line   1               1
//
// wrap enabled:
//               content index   line index
// this is the   0               1
// first line    0               2
// this is the   1               3
// second line   1               4
//

// Model represents a viewport component
type Model[T RenderableComparable] struct {
	// KeyMap is the keymap for the viewport
	KeyMap KeyMap

	// styles
	FooterStyle          lipgloss.Style
	HighlightStyle       lipgloss.Style
	SelectedContentStyle lipgloss.Style

	// header is the fixed header lines at the top of the viewport
	// these lines will wrap and be horizontally scrollable similar to other rendered content
	header []string

	// content is the complete list of items to be rendered in the viewport
	content []T

	// contentHeight is the number of lines of content shown given the current header and footer
	contentHeight int

	// lineContinuationIndicator is the string to use to indicate that a line has been truncated from the left or right
	lineContinuationIndicator string

	// selectionEnabled is true if the viewport allows individual line selection
	selectionEnabled bool

	// wrapText is true if the viewport wraps text rather than showing that a line is truncated/horizontally scrollable
	wrapText bool

	// stringToHighlight is a string to highlight in the viewport wherever it shows up
	stringToHighlight string

	// topSelectionSticky is true when selection should remain at the top until user manually scrolls down
	topSelectionSticky bool

	// bottomSelectionSticky is true when selection should remain at the bottom until user manually scrolls up
	bottomSelectionSticky bool

	// maintainSelection is true if the viewport should try to maintain the current selection when content is added or removed
	maintainSelection bool

	// selectedContentIdx is the index of content of the currently selected item when selectionEnabled is true
	selectedContentIdx int

	// width is the width of the entire viewport in terminal columns
	width int

	// height is the height of the entire viewport in terminal rows
	height int

	// topContentIdx is the content index of the topmost visible line
	topContentIdx int

	// topContentLineOffset is the number of lines from the top of the topmost visible line, when wrapped
	topContentLineOffset int

	// xOffset is the number of columns scrolled right when content lines overflow the viewport and wrapText is false
	xOffset int
}

// New creates a new viewport model with reasonable defaults
func New[T RenderableComparable](width, height int) (m Model[T]) {
	m.width, m.height = width, height
	m.updateContentHeight()

	m.selectionEnabled = false
	m.wrapText = false

	m.KeyMap = DefaultKeyMap()
	m.lineContinuationIndicator = "..."
	return m
}

// Update processes messages and updates the model
func (m Model[T]) Update(msg tea.Msg) (Model[T], tea.Cmd) {
	dev.DebugMsg("Viewport", msg)
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Up):
			if m.selectionEnabled {
				m.selectedContentIdxUp(1)
			} else {
				m.viewUp(1)
			}

		case key.Matches(msg, m.KeyMap.Down):
			if m.selectionEnabled {
				m.selectedContentIdxDown(1)
			} else {
				m.viewDown(1)
			}

		case key.Matches(msg, m.KeyMap.Left):
			if !m.wrapText {
				m.viewLeft(m.width / 4)
			}

		case key.Matches(msg, m.KeyMap.Right):
			if !m.wrapText {
				m.viewRight(m.width / 4)
			}

		case key.Matches(msg, m.KeyMap.HalfPageUp):
			offset := max(1, m.getNumVisibleItems()/2)
			m.viewUp(m.contentHeight / 2)
			if m.selectionEnabled {
				m.selectedContentIdxUp(offset)
			}

		case key.Matches(msg, m.KeyMap.HalfPageDown):
			offset := max(1, m.getNumVisibleItems()/2)
			m.viewDown(m.contentHeight / 2)
			if m.selectionEnabled {
				m.selectedContentIdxDown(offset)
			}

		case key.Matches(msg, m.KeyMap.PageUp):
			offset := m.getNumVisibleItems()
			m.viewUp(m.contentHeight)
			if m.selectionEnabled {
				m.selectedContentIdxUp(offset)
			}

		case key.Matches(msg, m.KeyMap.PageDown):
			offset := m.getNumVisibleItems()
			m.viewDown(m.contentHeight)
			if m.selectionEnabled {
				m.selectedContentIdxDown(offset)
			}

		case key.Matches(msg, m.KeyMap.Top):
			if m.selectionEnabled {
				m.SetSelectedContentIdx(0)
			} else {
				m.topContentIdx = 0
				m.topContentLineOffset = 0
			}

		case key.Matches(msg, m.KeyMap.Bottom):
			if m.selectionEnabled {
				m.selectedContentIdxDown(m.maxVisibleLineIdx())
			} else {
				m.viewDown(m.maxVisibleLineIdx())
			}
		}
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// View renders the viewport
func (m Model[T]) View() string {
	var viewString string

	footerString, footerHeight := m.getFooter()

	addLineToViewString := func(line string) {
		viewString += line + "\n"
	}

	header := m.getHeader()
	for _, headerLine := range header {
		visiblePart := m.getVisible(headerLine)
		addLineToViewString(fmt.Sprintf("%s\n", visiblePart))
	}

	hasStringToHighlight := stringWidth(m.stringToHighlight) != 0

	// get the lines to show based on the topContentIdx and topContentLineOffset
	// TODO

	// get the visible part of each line given the xOffset
	// TODO
	//visibleLines := m.getVisibleContentLines()
	//for idx, line := range visibleLines {
	//	contentIdx := m.getContentIdx(m.yOffset + idx)
	//	isSelected := m.selectionEnabled && contentIdx == m.selectedContentIdx
	//
	//	lineStyle := m.contentStyle
	//	if isSelected {
	//		lineStyle = m.SelectedContentStyle
	//	}
	//	visiblePartOfLine := m.getVisible(line)
	//
	//	if isSelected && visiblePartOfLine == "" {
	//		visiblePartOfLine = " "
	//	}
	//
	//	if hasStringToHighlight {
	//		// this splitting and rejoining of styled content is expensive and causes increased flickering,
	//		// so only do it if something is actually highlighted
	//		highlightStyle := m.HighlightStyle
	//		if isSelected {
	//			highlightStyle = m.highlightStyleIfSelected
	//		}
	//		lineChunks := strings.Split(visiblePartOfLine, m.stringToHighlight)
	//		var styledChunks []string
	//		for _, chunk := range lineChunks {
	//			styledChunks = append(styledChunks, lineStyle.Render(chunk))
	//		}
	//		addLineToViewString(strings.Join(styledChunks, highlightStyle.Render(m.stringToHighlight)))
	//	} else {
	//		addLineToViewString(lineStyle.Render(visiblePartOfLine))
	//	}
	//}

	if footerHeight > 0 {
		// pad so footer shows up at bottom
		padCount := max(0, m.contentHeight-len(visibleLines)-footerHeight)
		viewString += strings.Repeat("\n", padCount)
		viewString += footerString
	}
	renderedViewString := m.backgroundStyle.Width(m.width).Height(m.height).Render(viewString)

	return renderedViewString
}

// SetContent sets the content, the selectable set of lines in the viewport
func (m *Model[T]) SetContent(content []T) {
	dev.Debug("Setting viewport content")
	defer dev.Debug("Done setting viewport content")

	var initialTopPadding int
	initialSelection := m.GetSelectedContent()
	attemptMaintainSelection := m.maintainSelection && m.selectionEnabled && initialSelection != nil
	if attemptMaintainSelection {
		initialTopPadding = m.numLinesBetweenSelectionAndTop()
	}

	var stayAtTop, stayAtBottom bool
	if m.topSelectionSticky && m.selectionEnabled && m.selectedContentIdx == 0 {
		stayAtTop = true
	}
	if m.bottomSelectionSticky && m.selectionEnabled && m.selectedContentIdx == m.finalContentIdx() {
		stayAtBottom = true
	}

	m.content = content
	if m.wrapText {
		// ok to skip if no wrap because this is called when wrap is initially (re)enabled
		m.updateWrappedContent()
	}
	m.updateForHeaderAndContent()

	// fix any sort of potential selection issues
	if m.selectedContentIdx < 0 {
		m.selectedContentIdx = 0
	} else if finalContentIdx := m.finalContentIdx(); m.selectedContentIdx > finalContentIdx {
		m.selectedContentIdx = finalContentIdx
	}

	// stay at top, bottom, or maintain previous selection if desired
	if stayAtTop {
		m.selectedContentIdx = 0
	} else if stayAtBottom {
		m.selectedContentIdx = m.finalContentIdx()
	} else if attemptMaintainSelection {
		newSelectedContent := m.GetSelectedContent()
		if newSelectedContent != nil && !(*newSelectedContent).Equals(*initialSelection) {
			// try to keep the same item selected
			// TODO: for e.g. logs page which is sorted, can flag the sort order to the viewport and use that to
			//  find the item with binary search to make this more performant in the future
			for i := range m.content {
				if (*initialSelection).Equals(m.content[i]) {
					m.selectedContentIdx = i
					break
				}
			}
			newTopPadding := m.numLinesBetweenSelectionAndTop()
			if newTopPadding != initialTopPadding {
				m.viewDown(newTopPadding - initialTopPadding)
			}
		}
	}

	m.ensureViewContainsSelection()
}

// SetTopSticky sets whether selection should stay at top when new content added and selection is at the top
func (m *Model[T]) SetTopSticky(topSticky bool) {
	m.topSelectionSticky = topSticky
}

// SetBottomSticky sets whether selection should stay at bottom when new content added and selection is at the bottom
func (m *Model[T]) SetBottomSticky(bottomSticky bool) {
	m.bottomSelectionSticky = bottomSticky
}

// SetSelectionEnabled sets whether the viewport allows line selection
func (m *Model[T]) SetSelectionEnabled(selectionEnabled bool) {
	m.selectionEnabled = selectionEnabled
}

// SetMaintainSelection sets whether the viewport should try to maintain the current selection when content changes
func (m *Model[T]) SetMaintainSelection(maintainSelection bool) {
	m.maintainSelection = maintainSelection
}

// GetSelectionEnabled returns whether the viewport allows line selection
func (m Model[T]) GetSelectionEnabled() bool {
	return m.selectionEnabled
}

// SetWrapText sets whether the viewport wraps text
func (m *Model[T]) SetWrapText(wrapText bool) {
	initialTopPadding := m.numLinesBetweenSelectionAndTop()
	m.wrapText = wrapText
	m.updateWrappedContent()
	m.updateContentHeight()
	m.ResetHorizontalOffset()
	newTopPadding := m.numLinesBetweenSelectionAndTop()
	if newTopPadding != initialTopPadding {
		m.viewDown(newTopPadding - initialTopPadding)
	}
	m.ensureViewContainsSelection()
	m.updateMaxVisibleLineLength()
}

// GetWrapText returns whether the viewport wraps text
func (m Model[T]) GetWrapText() bool {
	return m.wrapText
}

// SetWidth sets the viewport's width
func (m *Model[T]) SetWidth(width int) {
	m.width = width
	m.updateForHeaderAndContent()
}

// GetWidth returns the viewport's width
func (m Model[T]) GetWidth() int {
	return m.width
}

// SetHeight sets the viewport's height, including header and footer
func (m *Model[T]) SetHeight(height int) {
	m.height = height
	m.updateForHeaderAndContent()
}

// GetHeight returns the viewport's height
func (m Model[T]) GetHeight() int {
	return m.height
}

// SetSelectedContentIdx sets the selected context index. Automatically puts selection in view as necessary
func (m *Model[T]) SetSelectedContentIdx(n int) {
	if m.contentHeight == 0 {
		return
	}

	if maxSelectedIdx := m.finalContentIdx(); n > maxSelectedIdx {
		m.selectedContentIdx = maxSelectedIdx
	} else {
		m.selectedContentIdx = max(0, n)
	}

	m.ensureViewContainsSelection()
}

// GetSelectedContentIdx returns the currently selected content index
func (m Model[T]) GetSelectedContentIdx() int {
	return m.selectedContentIdx
}

// GetSelectedContent returns the currently selected content
func (m Model[T]) GetSelectedContent() *T {
	if m.selectedContentIdx >= len(m.content) || m.selectedContentIdx < 0 {
		return nil
	}
	return &m.content[m.selectedContentIdx]
}

// SetStringToHighlight sets a string to highlight in the viewport
func (m *Model[T]) SetStringToHighlight(h string) {
	m.stringToHighlight = h
}

// SetHeader sets the header, an unselectable set of lines at the top of the viewport
func (m *Model[T]) SetHeader(header []string) {
	m.header = header
	m.updateWrappedHeader()
	m.updateForHeaderAndContent()
}

// ResetHorizontalOffset resets the horizontal offset to the leftmost position
func (m *Model[T]) ResetHorizontalOffset() {
	m.xOffset = 0
}

// ScrollToTop scrolls the viewport to the top
func (m *Model[T]) ScrollToTop() {
	m.selectedContentIdxUp(m.selectedContentIdx)
	m.viewUp(m.selectedContentIdx)
}

// ScrollToBottom scrolls the viewport to the bottom
func (m *Model[T]) ScrollToBottom() {
	m.selectedContentIdxDown(len(m.content))
	m.viewDown(len(m.content))
}

func (m *Model[T]) setXOffset(n int) {
	maxXOffset := m.maxLineLength - m.width
	m.xOffset = max(0, min(maxXOffset, n))
}

func (m *Model[T]) updateForHeaderAndContent() {
	m.updateContentHeight()
}

func (m *Model[T]) updateContentHeight() {
	_, footerHeight := m.getFooter()
	contentHeight := m.height - len(m.getHeader()) - footerHeight
	m.contentHeight = max(0, contentHeight)
}

func (m *Model[T]) selectedContentIdxDown(n int) {
	m.SetSelectedContentIdx(m.selectedContentIdx + n)
}

func (m *Model[T]) selectedContentIdxUp(n int) {
	m.SetSelectedContentIdx(m.selectedContentIdx - n)
}

func (m *Model[T]) viewDown(n int) {
	m.setYOffset(m.yOffset + n)
}

func (m *Model[T]) viewUp(n int) {
	m.setYOffset(m.yOffset - n)
}

func (m *Model[T]) viewLeft(n int) {
	m.setXOffset(m.xOffset - n)
}

func (m *Model[T]) viewRight(n int) {
	m.setXOffset(m.xOffset + n)
}

func (m Model[T]) getHeader() []string {
	if !m.wrapText {
		return m.header
	}
	var wrappedHeaderLines []string
	for _, s := range m.header {
		wrappedHeaderLines = append(wrappedHeaderLines, wrap(s, m.width)...)
	}
	return wrappedHeaderLines
}

// lastVisibleLineIdx returns the maximum visible line index
func (m Model[T]) lastVisibleLineIdx() int {
	return min(m.maxVisibleLineIdx(), m.yOffset+m.contentHeight-1)
}

// maxYOffset returns the maximum yOffset (the yOffset that shows the final screen)
func (m Model[T]) maxYOffset() int {
	if m.maxVisibleLineIdx() < m.contentHeight {
		return 0
	}
	return m.getLenContentStrings() - m.contentHeight
}

func (m *Model[T]) maxVisibleLineIdx() int {
	return m.getLenContentStrings() - 1
}

func (m Model[T]) finalContentIdx() int {
	return len(m.content) - 1
}

func (m Model[T]) numLinesBetweenSelectionAndTop() int {
	if !m.selectionEnabled {
		return 0
	}
	selection := m.GetSelectedContent()
	if selection == nil {
		return 0
	}
	if !m.wrapText {
		return m.selectedContentIdx - m.yOffset
	} else {
		// start at m.yOffset, and until m.content[m.wrappedContentIdxToContentIdx] == initialSelection, increment top padding
		for i := m.yOffset; i < len(m.wrappedContent); i++ {
			if m.content[m.wrappedContentIdxToContentIdx[i]].Equals(*selection) {
				return i - m.yOffset
			}
		}
	}
	return 0
}

// getVisibleContentLines retrieves the visible content based on the yOffset and contentHeight
func (m Model[T]) getVisibleContentLines() []string {
	maxVisibleLineIdx := m.maxVisibleLineIdx()
	start := max(0, min(maxVisibleLineIdx, m.yOffset))
	end := start + m.contentHeight
	if end > maxVisibleLineIdx {
		return m.getContentStrings(start, -1)
	}
	return m.getContentStrings(start, end)
}

func (m Model[T]) getVisible(line string) string {
	lineNoTrailingSpace := strings.TrimRightFunc(line, unicode.IsSpace)
	return getVisiblePartOfLine(lineNoTrailingSpace, m.xOffset, m.width, m.lineContinuationIndicator)
}

func (m Model[T]) getContentIdx(wrappedContentIdx int) int {
	if !m.wrapText {
		return wrappedContentIdx
	}
	return m.wrappedContentIdxToContentIdx[wrappedContentIdx]
}

func (m Model[T]) getCurrentLineIdx() int {
	if m.wrapText {
		return m.contentIdxToFirstWrappedContentIdx[m.selectedContentIdx]
	}
	return m.selectedContentIdx
}

func (m Model[T]) getNumVisibleItems() int {
	if !m.wrapText {
		return m.contentHeight
	}

	var itemCount int
	var rowCount int
	contentIdx := m.wrappedContentIdxToContentIdx[m.yOffset]
	for rowCount < m.contentHeight {
		if height, exists := m.contentIdxToHeight[contentIdx]; exists {
			rowCount += height
		} else {
			break
		}
		contentIdx++
		itemCount++
	}
	return itemCount
}

func (m Model[T]) lastContentItemSelected() bool {
	return m.selectedContentIdx == len(m.content)-1
}

func (m Model[T]) getFooter() (string, int) {
	numerator := m.selectedContentIdx + 1
	denominator := len(m.content)
	totalNumLines := m.getLenContentStrings()

	// if selection is disabled, percentage should show from the bottom of the visible content
	// such that panning the view to the bottom shows 100%
	if !m.selectionEnabled {
		numerator = m.yOffset + m.contentHeight
		denominator = totalNumLines
	}

	if totalNumLines >= m.height-len(m.getHeader()) {
		percentScrolled := percent(numerator, denominator)
		footerString := fmt.Sprintf("%d%% (%d/%d)", percentScrolled, numerator, denominator)
		renderedFooterString := m.FooterStyle.MaxWidth(m.width).Render(footerString)
		footerHeight := lipgloss.Height(renderedFooterString)
		return renderedFooterString, footerHeight
	}
	return "", 0
}
