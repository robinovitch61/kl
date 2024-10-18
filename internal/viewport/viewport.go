package viewport

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/kl/internal/dev"
	"regexp"
	"strings"
)

var (
	ansiPattern = regexp.MustCompile("\x1b\\[[0-9;]*m")
)

// Terminology:
// - allItems: an item to be rendered in the viewport
// - line: a row in the terminal
// - visible: in the vertical sense, a line is visible if it is within the viewport
// - truncated: in the horizontal sense, a line is truncated if it is too long to fit in the viewport
//
// wrap disabled:
//                           allItems index   line index
// this is the first line    0               0
// this is the second line   1               1
//
// wrap disabled, line overflow:
//                           allItems index   line index
// this is the first...      0               0
// this is the secon...      1               1
//
// wrap enabled:
//               allItems index   line index
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
	// these lines will wrap and be horizontally scrollable similar to other rendered allItems
	header []string

	// allItems is the complete list of items to be rendered in the viewport
	allItems []T

	// numContentLines is the number of lines of shown between the header and footer
	// TODO: make this a function, not state
	numContentLines int

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

	// maintainSelection is true if the viewport should try to maintain the current selection when allItems is added or removed
	maintainSelection bool

	// selectedItemIdx is the index of allItems of the current selection (only relevant when selectionEnabled is true)
	selectedItemIdx int

	// width is the width of the entire viewport in terminal columns
	width int

	// height is the height of the entire viewport in terminal rows
	height int

	// topItemIdx is the allItems index of the topmost visible viewport line
	topItemIdx int

	// topItemLineOffset is the number of lines scrolled out of view of the topmost visible line, when wrapped
	topItemLineOffset int

	// xOffset is the number of columns scrolled right when rendered lines overflow the viewport and wrapText is false
	xOffset int
}

// New creates a new viewport model with reasonable defaults
func New[T RenderableComparable](width, height int) (m Model[T]) {
	m.setWidthHeight(width, height)

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
				m.scrollUp(1)
			}

		case key.Matches(msg, m.KeyMap.Down):
			if m.selectionEnabled {
				m.selectedContentIdxDown(1)
			} else {
				m.scrollDown(1)
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
			m.scrollUp(m.numContentLines / 2)
			if m.selectionEnabled {
				m.selectedContentIdxUp(offset)
			}

		case key.Matches(msg, m.KeyMap.HalfPageDown):
			offset := max(1, m.getNumVisibleItems()/2)
			m.scrollDown(m.numContentLines / 2)
			if m.selectionEnabled {
				m.selectedContentIdxDown(offset)
			}

		case key.Matches(msg, m.KeyMap.PageUp):
			offset := m.getNumVisibleItems()
			m.scrollUp(m.numContentLines)
			if m.selectionEnabled {
				m.selectedContentIdxUp(offset)
			}

		case key.Matches(msg, m.KeyMap.PageDown):
			offset := m.getNumVisibleItems()
			m.scrollDown(m.numContentLines)
			if m.selectionEnabled {
				m.selectedContentIdxDown(offset)
			}

		case key.Matches(msg, m.KeyMap.Top):
			if m.selectionEnabled {
				m.SetSelectedContentIdx(0)
			} else {
				m.topItemIdx = 0
				m.topItemLineOffset = 0
			}

		case key.Matches(msg, m.KeyMap.Bottom):
			if m.selectionEnabled {
				m.selectedContentIdxDown(len(m.allItems))
			} else {
				m.scrollDown(len(m.allItems))
			}
		}
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// View renders the viewport
func (m Model[T]) View() string {
	var viewString string

	visibleHeaderLines := m.getVisibleHeaderLines()
	for i := range visibleHeaderLines {
		viewString += m.truncateNoXOffset(visibleHeaderLines[i]) + "\n"
	}

	//hasStringToHighlight := stringWidth(m.stringToHighlight) != 0

	// get the lines to show based on the topItemIdx and topItemLineOffset
	visibleContentLines := m.getVisibleContentLines()
	truncatedVisibleContentLines := make([]string, len(visibleContentLines.lines))
	for i := range visibleContentLines.lines {
		truncatedVisibleContentLines[i] = m.truncate(visibleContentLines.lines[i])
	}
	//fmt.Println(fmt.Sprintf("%q", visibleContentLines))

	// add selection style
	if m.selectionEnabled {
		for i := range truncatedVisibleContentLines {
			if visibleContentLines.itemIndexes[i] == m.selectedItemIdx {
				if truncatedVisibleContentLines[i] == "" {
					truncatedVisibleContentLines[i] = " " // ensure selection is visible even if content empty
				}
				truncatedVisibleContentLines[i] = m.SelectedContentStyle.Render(truncatedVisibleContentLines[i])
			}
		}
	}

	for i := range truncatedVisibleContentLines {
		viewString += truncatedVisibleContentLines[i] + "\n"
	}

	// TODO: use len(visibleContentLines.lines)?
	nVisibleLines := len(strings.Split(viewString, "\n"))
	if visibleContentLines.showFooter {
		// pad so footer shows up at bottom
		padCount := max(0, m.numContentLines-nVisibleLines-1) // 1 for footer itself
		viewString += strings.Repeat("\n", padCount)
		viewString += m.getTruncatedFooterLine()
	}

	// TODO: rm
	println(fmt.Sprintf("LEO numLines=%d, height=%d, numContentLines=%d, numHeaderLines=%d, numContentLines=%d", len(strings.Split(viewString, "\n")), m.height, m.numContentLines, len(visibleHeaderLines), len(truncatedVisibleContentLines)))

	return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(viewString)
}

// SetContent sets the allItems, the selectable set of lines in the viewport
func (m *Model[T]) SetContent(content []T) {
	var stayAtTop, stayAtBottom bool
	var prevSelection T
	if m.selectionEnabled {
		//dev.Debug(fmt.Sprintf("LEO len(allItems)=%d, selectedItemIdx=%d, bottomSelectionsticky=%t, topSelectionSticky=%t", len(m.allItems), m.selectedItemIdx, m.bottomSelectionSticky, m.topSelectionSticky))
		if m.topSelectionSticky && len(m.allItems) > 0 && m.selectedItemIdx == 0 {
			//dev.Debug("LEO stay at top")
			stayAtTop = true
		} else if m.bottomSelectionSticky && (len(m.allItems) == 0 || (m.selectedItemIdx == len(m.allItems)-1)) {
			//dev.Debug("LEO stay at bottom")
			stayAtBottom = true
		} else if m.maintainSelection && 0 <= m.selectedItemIdx && m.selectedItemIdx < len(m.allItems) {
			//dev.Debug("LEO maintain selection")
			prevSelection = m.allItems[m.selectedItemIdx]
		}
	}

	m.allItems = content
	// ensure topItemIdx and topItemLineOffset are valid given new content
	m.safelySetTopItemIdxAndOffset(m.topItemIdx, m.topItemLineOffset)
	// num content lines depends on the footer existing, which depends on the items
	m.updateNumContentLines()

	if m.selectionEnabled {
		// stay at top, bottom, or maintain previous selection if desired
		if stayAtTop {
			m.selectedItemIdx = 0
		} else if stayAtBottom {
			// TODO: no test catches the max(0, ...) here, write one!
			m.selectedItemIdx = max(0, len(m.allItems)-1)
		} else if m.maintainSelection {
			// TODO: could flag when content is sorted & comparable and use binary search instead
			for i := range m.allItems {
				if m.allItems[i].Equals(prevSelection) {
					m.selectedItemIdx = i
					break
				}
			}
		} else {
			m.selectedItemIdx = clampValMinMax(m.selectedItemIdx, 0, len(m.allItems)-1)
		}
		m.scrollSoSelectionInView()
	}
}

// SetTopSticky sets whether selection should stay at top when new allItems added and selection is at the top
func (m *Model[T]) SetTopSticky(topSticky bool) {
	m.topSelectionSticky = topSticky
}

// SetBottomSticky sets whether selection should stay at bottom when new allItems added and selection is at the bottom
func (m *Model[T]) SetBottomSticky(bottomSticky bool) {
	m.bottomSelectionSticky = bottomSticky
}

// SetSelectionEnabled sets whether the viewport allows line selection
func (m *Model[T]) SetSelectionEnabled(selectionEnabled bool) {
	m.selectionEnabled = selectionEnabled
}

// SetMaintainSelection sets whether the viewport should try to maintain the current selection when allItems changes
func (m *Model[T]) SetMaintainSelection(maintainSelection bool) {
	m.maintainSelection = maintainSelection
}

// GetSelectionEnabled returns whether the viewport allows line selection
func (m Model[T]) GetSelectionEnabled() bool {
	return m.selectionEnabled
}

// SetWrapText sets whether the viewport wraps text
func (m *Model[T]) SetWrapText(wrapText bool) {
	var initialNumLinesAboveSelection int
	if m.selectionEnabled {
		initialNumLinesAboveSelection = m.getNumLinesAboveSelection()
	}
	m.wrapText = wrapText
	m.topItemLineOffset = 0
	m.xOffset = 0
	m.updateNumContentLines()
	m.scrollSoSelectionInView()
	if m.selectionEnabled {
		newNumLinesAboveSelection := m.getNumLinesAboveSelection()
		m.scrollUp(initialNumLinesAboveSelection - newNumLinesAboveSelection)
		m.scrollSoSelectionInView()
	}
}

// GetWrapText returns whether the viewport wraps text
func (m Model[T]) GetWrapText() bool {
	return m.wrapText
}

// SetWidth sets the viewport's width
func (m *Model[T]) SetWidth(width int) {
	m.setWidthHeight(width, m.height)
}

// GetWidth returns the viewport's width
func (m Model[T]) GetWidth() int {
	return m.width
}

// SetHeight sets the viewport's height, including header and footer
func (m *Model[T]) SetHeight(height int) {
	m.setWidthHeight(m.width, height)
}

// GetHeight returns the viewport's height
func (m Model[T]) GetHeight() int {
	return m.height
}

// SetSelectedContentIdx sets the selected context index. Automatically puts selection in view as necessary
func (m *Model[T]) SetSelectedContentIdx(n int) {
	if !m.selectionEnabled || m.numContentLines == 0 {
		return
	}
	m.selectedItemIdx = clampValMinMax(n, 0, len(m.allItems)-1)
	m.scrollSoSelectionInView()
}

// GetSelectedContentIdx returns the currently selected allItems index
func (m Model[T]) GetSelectedContentIdx() int {
	return m.selectedItemIdx
}

// GetSelectedContent returns the currently selected allItems
func (m Model[T]) GetSelectedContent() *T {
	if m.selectedItemIdx >= len(m.allItems) || m.selectedItemIdx < 0 {
		return nil
	}
	return &m.allItems[m.selectedItemIdx]
}

// SetStringToHighlight sets a string to highlight in the viewport
func (m *Model[T]) SetStringToHighlight(h string) {
	m.stringToHighlight = h
}

// SetHeader sets the header, an unselectable set of lines at the top of the viewport
func (m *Model[T]) SetHeader(header []string) {
	m.header = header
	m.updateNumContentLines()
}

// ScrollToTop scrolls the viewport to the top
func (m *Model[T]) ScrollToTop() {
	m.selectedContentIdxUp(m.selectedItemIdx)
	m.scrollUp(m.selectedItemIdx)
}

func (m Model[T]) maxLineWidth() int {
	maxLineWidth := 0
	headerLines := m.getVisibleHeaderLines()
	visibleContentLines := m.getVisibleContentLines()
	allVisibleLines := append(headerLines, visibleContentLines.lines...)
	if visibleContentLines.showFooter {
		allVisibleLines = append(allVisibleLines, m.getTruncatedFooterLine())
	}
	for i := range allVisibleLines {
		if w := stringWidth(allVisibleLines[i]); w > maxLineWidth {
			maxLineWidth = w
		}
	}
	return maxLineWidth
}

func (m *Model[T]) setXOffset(n int) {
	maxXOffset := m.maxLineWidth() - m.width
	m.xOffset = max(0, min(maxXOffset, n))
}

func (m *Model[T]) setWidthHeight(width, height int) {
	m.width, m.height = max(0, width), max(0, height)
	m.updateNumContentLines()
}

func (m *Model[T]) safelySetTopItemIdxAndOffset(topItemIdx, topItemLineOffset int) {
	maxTopItemIdx, maxTopItemLineOffset := m.maxItemIdxAndMaxTopLineOffset()
	//println(maxTopItemIdx, maxTopItemLineOffset)
	//println(topItemIdx, topItemLineOffset)
	m.topItemIdx = clampValMinMax(topItemIdx, 0, maxTopItemIdx)
	m.topItemLineOffset = topItemLineOffset
	if m.topItemIdx == maxTopItemIdx {
		m.topItemLineOffset = clampValMinMax(topItemLineOffset, 0, maxTopItemLineOffset)
	}
}

func (m *Model[T]) updateNumContentLines() {
	contentHeight := m.height - len(m.getVisibleHeaderLines())
	visibleContentLines := m.getVisibleContentLines()
	if visibleContentLines.showFooter {
		contentHeight-- // one for footer
	}
	m.numContentLines = max(0, contentHeight)
}

func (m *Model[T]) scrollSoSelectionInView() {
	if len(m.allItems) == 0 {
		m.safelySetTopItemIdxAndOffset(0, 0)
		return
	}

	numLinesInSelection := 1
	if m.wrapText {
		numLinesInSelection = len(wrap(m.allItems[m.selectedItemIdx].Render(), m.width))
	}

	numLinesOfSelectionInView := m.numLinesOfSelectionInView()
	if numLinesInSelection != numLinesOfSelectionInView {
		if m.topItemIdx < m.selectedItemIdx {
			//dev.Debug(fmt.Sprintf("LEO topItemIdx=%d, selectedItemIdx=%d, numLinesInSelection=%d, numLinesOfSelectionInView=%d", m.topItemIdx, m.selectedItemIdx, numLinesInSelection, numLinesOfSelectionInView))
			// if selection is below, scroll until it's fully in view at the bottom
			// first, put it at the top, intentionally not doing it in manner that protects the view going past the bottom
			m.topItemIdx, m.topItemLineOffset = m.selectedItemIdx, 0
			// then scroll up so that selection is at the bottom, unless it already takes up the whole screen
			m.scrollUp(max(0, m.numContentLines-numLinesInSelection))
		} else {
			// if selection above, scroll until it's fully in view at the top
			m.topItemIdx, m.topItemLineOffset = m.selectedItemIdx, 0
		}
	}
}

func (m *Model[T]) selectedContentIdxDown(n int) {
	m.SetSelectedContentIdx(m.selectedItemIdx + n)
}

func (m *Model[T]) selectedContentIdxUp(n int) {
	m.SetSelectedContentIdx(m.selectedItemIdx - n)
}

func (m *Model[T]) scrollDown(n int) {
	m.scrollByNLines(n)
}

func (m *Model[T]) scrollUp(n int) {
	m.scrollByNLines(-n)
}

func (m *Model[T]) viewLeft(n int) {
	m.setXOffset(m.xOffset - n)
}

func (m *Model[T]) viewRight(n int) {
	m.setXOffset(m.xOffset + n)
}

// scrollByNLines edits topItemIdx and topItemLineOffset to scroll the viewport by n lines (negative for up, positive for down)
func (m *Model[T]) scrollByNLines(n int) {
	if n == 0 {
		return
	}

	// scrolling down past bottom
	if n > 0 && m.isScrolledToBottom() {
		return
	}

	// scrolling up past top
	if n < 0 && m.topItemIdx == 0 && m.topItemLineOffset == 0 {
		return
	}

	newTopItemIdx, newTopItemLineOffset := m.topItemIdx, m.topItemLineOffset
	if !m.wrapText {
		newTopItemIdx = m.topItemIdx + n
	} else {
		// wrapped
		if n < 0 { // negative n, scrolling up
			// up
			if newTopItemLineOffset >= -n {
				// same item, just change offset
				newTopItemLineOffset += n
			} else {
				// take lines from items until scrolled up desired amount
				n += newTopItemLineOffset
				for n < 0 {
					newTopItemIdx -= 1
					if newTopItemIdx < 0 {
						// scrolled up past top - stay at top
						newTopItemIdx = 0
						newTopItemLineOffset = 0
						break
					}
					numLinesInTopItem := len(wrap(m.allItems[newTopItemIdx].Render(), m.width))
					for i := range numLinesInTopItem {
						n += 1
						if n == 0 {
							newTopItemLineOffset = numLinesInTopItem - (i + 1)
							break
						}
					}
				}
			}
		} else { // positive n, scrolling down
			numLinesInTopItem := len(wrap(m.allItems[newTopItemIdx].Render(), m.width))
			if newTopItemLineOffset+n < numLinesInTopItem {
				// same item, just change offset
				newTopItemLineOffset += n
			} else {
				// take lines from items until scrolled down desired amount
				n -= numLinesInTopItem - (newTopItemLineOffset + 1)
				for n > 0 {
					newTopItemIdx += 1
					if newTopItemIdx >= len(m.allItems) {
						newTopItemIdx = len(m.allItems) - 1
						break
					}
					numLinesInTopItem = len(wrap(m.allItems[newTopItemIdx].Render(), m.width))
					for i := range numLinesInTopItem {
						n -= 1
						if n == 0 {
							newTopItemLineOffset = i
							break
						}
					}
				}
			}
		}
	}
	m.safelySetTopItemIdxAndOffset(newTopItemIdx, newTopItemLineOffset)
}

// getVisibleHeaderLines returns the lines of header that are visible in the viewport
// header lines will take precedence over content and footer if there is not enough vertical height
func (m Model[T]) getVisibleHeaderLines() []string {
	if m.height == 0 {
		return nil
	}

	if !m.wrapText {
		return safeSliceUpToIdx(m.header, m.height)
	} else {
		// wrapped
		var wrappedHeaderLines []string
		for _, s := range m.header {
			wrappedHeaderLines = append(wrappedHeaderLines, wrap(s, m.width)...)
		}
		return safeSliceUpToIdx(wrappedHeaderLines, m.height)
	}
}

type visibleContentLinesResult struct {
	// lines is the untruncated visible lines, each corresponding to one terminal row
	lines []string
	// itemIndexes is the index of the item in allItems that corresponds to each line. len(itemIndexes) == len(lines)
	itemIndexes []int
	// showFooter is true if the footer should be shown due to the num visible lines exceeding the vertical space
	showFooter bool
}

// getVisibleContentLines returns the lines of content that are visible in the viewport given vertical scroll position
// and the content. It also returns the item index for each associated visible line and whether or not to show the footer
func (m Model[T]) getVisibleContentLines() visibleContentLinesResult {
	if len(m.allItems) == 0 {
		return visibleContentLinesResult{lines: nil, itemIndexes: nil, showFooter: false}
	}

	var lines []string
	var itemIndexes []int

	// numContentLines is the number of lines of content that can be shown in the viewport, excluding the header lines
	numContentLines := max(0, m.height-len(m.getVisibleHeaderLines()))

	// convenience functions that add lines to the lines slice and return true if reached numContentLines
	addLine := func(l string, itemIndex int) bool {
		lines = append(lines, l)
		itemIndexes = append(itemIndexes, itemIndex)
		return len(lines) == numContentLines
	}
	addLines := func(ls []string, itemIndex int) bool {
		for i := range ls {
			if addLine(ls[i], itemIndex) {
				return true
			}
		}
		return false
	}

	currItemIdx := m.topItemIdx
	if currItemIdx < 0 || currItemIdx >= len(m.allItems) {
		panic("invalid topItemIdx")
	}
	currItem := m.allItems[currItemIdx]
	done := numContentLines == 0
	if done {
		return visibleContentLinesResult{lines: lines, itemIndexes: itemIndexes, showFooter: false}
	}
	if m.wrapText {
		itemLines := wrap(currItem.Render(), m.width)
		offsetLines := safeSliceFromIdx(itemLines, m.topItemLineOffset)
		done = addLines(offsetLines, currItemIdx)

		for !done {
			currItemIdx += 1
			if currItemIdx >= len(m.allItems) {
				done = true
			} else {
				currItem = m.allItems[currItemIdx]
				itemLines = wrap(currItem.Render(), m.width)
				done = addLines(itemLines, currItemIdx)
			}
		}
	} else {
		done = addLine(currItem.Render(), currItemIdx)
		for !done {
			currItemIdx += 1
			if currItemIdx >= len(m.allItems) {
				done = true
			} else {
				currItem = m.allItems[currItemIdx]
				done = addLine(currItem.Render(), currItemIdx)
			}
		}
	}
	// we show the footer when either:
	// - len(lines) equals or exceeds numContentLines
	// - view is not at top, and len(lines)+1 == numContentLines
	linesMeetOrExceedContent := len(lines) >= numContentLines
	scrolledToTop := m.topItemIdx == 0 && m.topItemLineOffset == 0
	scrolledToBottom := len(lines)+1 == numContentLines
	showFooter := linesMeetOrExceedContent || (!scrolledToTop && scrolledToBottom)
	if showFooter {
		// num visible lines exceeds vertical space, leave one line for the footer
		lines = safeSliceUpToIdx(lines, numContentLines-1)
		itemIndexes = safeSliceUpToIdx(itemIndexes, numContentLines-1)
	}
	return visibleContentLinesResult{lines: lines, itemIndexes: itemIndexes, showFooter: showFooter}
}

func (m Model[T]) getTruncatedFooterLine() string {
	numerator := m.selectedItemIdx + 1 // 0th line is 1st
	denominator := len(m.allItems)
	visibleContentLines := m.getVisibleContentLines()
	if !visibleContentLines.showFooter {
		panic("getTruncatedFooterLine called when footer should not be shown")
	}

	// if selection is disabled, numerator should be item index of bottom visible line
	if !m.selectionEnabled {
		numerator = visibleContentLines.itemIndexes[len(visibleContentLines.itemIndexes)-1] + 1
		if m.wrapText && numerator == denominator && !m.isScrolledToBottom() {
			// if wrapped && bottom visible line is max item index, but actually not fully scrolled to bottom, show 99%
			return fmt.Sprintf("99%% (%d/%d)", numerator, denominator)
		}
	}

	percentScrolled := percent(numerator, denominator)
	footerString := fmt.Sprintf("%d%% (%d/%d)", percentScrolled, numerator, denominator)
	return m.FooterStyle.Render(truncateLine(footerString, 0, m.width, m.lineContinuationIndicator))
}

func (m Model[T]) isScrolledToBottom() bool {
	maxItemIdx, maxTopItemLineOffset := m.maxItemIdxAndMaxTopLineOffset()
	if m.topItemIdx > maxItemIdx {
		return true
	}
	if m.topItemIdx == maxItemIdx {
		return m.topItemLineOffset >= maxTopItemLineOffset
	}
	return false
}

func (m Model[T]) numLinesOfSelectionInView() int {
	visibleContentLines := m.getVisibleContentLines()
	res := 0
	for i := range visibleContentLines.itemIndexes {
		if visibleContentLines.itemIndexes[i] == m.selectedItemIdx {
			res++
		}
	}
	return res
}

func (m Model[T]) maxItemIdxAndMaxTopLineOffset() (int, int) {
	lenAllItems := len(m.allItems)
	if lenAllItems == 0 {
		return 0, 0
	}
	if !m.wrapText {
		return max(0, lenAllItems-m.numContentLines), 0
	}
	// wrapped
	maxTopItemIdx, maxTopItemLineOffset := lenAllItems-1, 0
	nLinesLastItem := len(wrap(m.allItems[lenAllItems-1].Render(), m.width))
	if m.numContentLines <= nLinesLastItem {
		// same item, just change offset
		maxTopItemLineOffset = nLinesLastItem - m.numContentLines
	} else {
		// take lines from items until scrolled up desired amount
		n := m.numContentLines - nLinesLastItem
		for n > 0 {
			maxTopItemIdx -= 1
			if maxTopItemIdx < 0 {
				// scrolled up past top - stay at top
				maxTopItemIdx = 0
				maxTopItemLineOffset = 0
				break
			}
			numLinesInTopItem := len(wrap(m.allItems[maxTopItemIdx].Render(), m.width))
			for i := range numLinesInTopItem {
				n -= 1
				if n == 0 {
					maxTopItemLineOffset = numLinesInTopItem - (i + 1)
					break
				}
			}
		}
	}
	return max(0, maxTopItemIdx), max(0, maxTopItemLineOffset)
}

func (m Model[T]) truncate(line string) string {
	//truncated := truncateLine(line, m.xOffset, m.width, m.lineContinuationIndicator)
	//fmt.Println(fmt.Sprintf("truncateLine(%q, %d, %d, %q) = %q", line, m.xOffset, m.width, m.lineContinuationIndicator, truncated))
	//return truncated
	return truncateLine(line, m.xOffset, m.width, m.lineContinuationIndicator)
}

func (m Model[T]) truncateNoXOffset(line string) string {
	return truncateLine(line, 0, m.width, m.lineContinuationIndicator)
}

func (m Model[T]) getNumVisibleItems() int {
	if !m.wrapText {
		return m.numContentLines
	} else {
		visibleContentLines := m.getVisibleContentLines()
		// return distinct number of items
		itemIndexSet := make(map[int]struct{})
		for _, i := range visibleContentLines.itemIndexes {
			itemIndexSet[i] = struct{}{}
		}
		return len(itemIndexSet)
	}
}

func (m Model[T]) getNumLinesAboveSelection() int {
	if !m.selectionEnabled {
		panic("getNumLinesAboveSelection called when selection is disabled")
	}
	visibleContentLines := m.getVisibleContentLines()
	for i := range visibleContentLines.itemIndexes {
		if visibleContentLines.itemIndexes[i] == m.selectedItemIdx {
			return i
		}
	}
	panic("selected item not found in visible content")
}
