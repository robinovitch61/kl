package viewport

import (
	"fmt"
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/linebuffer"
	"regexp"
	"unicode"

	"strings"
	"time"
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

var surroundingAnsiRegex = regexp.MustCompile(`(\x1b\[[0-9;]*m.*?\x1b\[0?m)`)

// Model represents a viewport component
type Model[T RenderableComparable] struct {
	// KeyMap is the keymap for the viewport
	KeyMap KeyMap

	// styles
	FooterStyle              lipgloss.Style
	HighlightStyle           lipgloss.Style
	HighlightStyleIfSelected lipgloss.Style
	SelectedItemStyle        lipgloss.Style

	// header is the fixed header lines at the top of the viewport
	// these lines will wrap and be horizontally scrollable similar to other rendered allItems
	header []string

	// allItems is the complete list of items to be rendered in the viewport
	allItems []T

	// continuationIndicator is the string to use to indicate that a line has been truncated from the left or right
	continuationIndicator string

	// selectionEnabled is true if the viewport allows individual line selection
	selectionEnabled bool

	// footerVisible is true if the viewport will show the footer when it overflows
	footerVisible bool

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

	// lineBufferCache is a cache of linebuffers as calls to linebuffer.New are expensive and repeated multiple times
	// for the same s for each Update -> View cycle at the moment. It is faster to hash the inputs and store in a
	// map than rerun linebuffer.New multiple times
	lineBufferCache *expirable.LRU[string, *linebuffer.LineBuffer]
}

// New creates a new viewport model with reasonable defaults
func New[T RenderableComparable](width, height int) (m Model[T]) {
	m.setWidthHeight(width, height)

	m.selectionEnabled = false
	m.wrapText = false

	m.KeyMap = DefaultKeyMap()
	m.continuationIndicator = "..."
	m.footerVisible = true

	// the cache is purged at the start of each call to Update, as it's mostly helpful for the repeated calls to
	// linebuffer.New within a single Update -> View cycle. In addition, it's configured with LRU params in case
	// for some reason Update -> View isn't called for a while:
	// - size of 500 is an estimate for the max reasonable height of a terminal viewport
	// - 5 seconds expiry is a conservative upper bound for a single Update -> View cycle to complete (don't want to
	//   expire the cache midway through this operation)
	m.lineBufferCache = expirable.NewLRU[string, *linebuffer.LineBuffer](500, nil, 5*time.Second)

	return m
}

// Update processes messages and updates the model
func (m Model[T]) Update(msg tea.Msg) (Model[T], tea.Cmd) {
	dev.DebugUpdateMsg("Viewport", msg)

	m.lineBufferCache.Purge()

	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Up):
			if m.selectionEnabled {
				m.selectedItemIdxUp(1)
			} else {
				m.scrollUp(1)
			}

		case key.Matches(msg, m.KeyMap.Down):
			if m.selectionEnabled {
				m.selectedItemIdxDown(1)
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
			m.scrollUp(m.getNumContentLines() / 2)
			if m.selectionEnabled {
				m.selectedItemIdxUp(offset)
			}

		case key.Matches(msg, m.KeyMap.HalfPageDown):
			offset := max(1, m.getNumVisibleItems()/2)
			m.scrollDown(m.getNumContentLines() / 2)
			if m.selectionEnabled {
				m.selectedItemIdxDown(offset)
			}

		case key.Matches(msg, m.KeyMap.PageUp):
			offset := m.getNumVisibleItems()
			m.scrollUp(m.getNumContentLines())
			if m.selectionEnabled {
				m.selectedItemIdxUp(offset)
			}

		case key.Matches(msg, m.KeyMap.PageDown):
			offset := m.getNumVisibleItems()
			m.scrollDown(m.getNumContentLines())
			if m.selectionEnabled {
				m.selectedItemIdxDown(offset)
			}

		case key.Matches(msg, m.KeyMap.Top):
			if m.selectionEnabled {
				m.SetSelectedItemIdx(0)
			} else {
				m.topItemIdx = 0
				m.topItemLineOffset = 0
			}

		case key.Matches(msg, m.KeyMap.Bottom):
			if m.selectionEnabled {
				m.selectedItemIdxDown(len(m.allItems))
			} else {
				m.safelySetTopItemIdxAndOffset(m.maxItemIdxAndMaxTopLineOffset())
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
		lineBuffer := m.getLineBuffer(visibleHeaderLines[i], m.width)
		viewString += lineBuffer.PopLeft(m.continuationIndicator, "", lipgloss.NewStyle()) + "\n"
	}

	// get the lines to show based on the vertical scroll position (topItemIdx and topItemLineOffset)
	visibleContentLines := m.getVisibleContentLines()

	var truncated string
	truncatedVisibleContentLines := make([]string, len(visibleContentLines.lines))
	for i := range visibleContentLines.lines {
		if m.wrapText {
			truncated = visibleContentLines.lines[i]
		} else {
			lineBuffer := m.getLineBuffer(visibleContentLines.lines[i], m.width)
			lineBuffer.SeekToWidth(m.xOffset)
			truncated = lineBuffer.PopLeft(m.continuationIndicator, m.stringToHighlight, m.highlightStyle(visibleContentLines.itemIndexes[i]))
		}

		isSelection := m.selectionEnabled && visibleContentLines.itemIndexes[i] == m.selectedItemIdx
		if isSelection {
			truncated = m.styleSelection(truncated)
		}

		if !m.wrapText && m.xOffset > 0 && lipgloss.Width(truncated) == 0 && lipgloss.Width(visibleContentLines.lines[i]) > 0 {
			// if panned right past where line ends, show continuation indicator
			lineBuffer := m.getLineBuffer(m.getLineContinuationIndicator(), m.width)
			truncated = lineBuffer.PopLeft("", "", lipgloss.NewStyle())
			if isSelection {
				truncated = m.styleSelection(truncated)
			}
		}

		if isSelection && truncated == "" {
			// ensure selection is visible even if content empty
			truncated = m.styleSelection(" ")
		}

		truncatedVisibleContentLines[i] = truncated
	}

	for i := range truncatedVisibleContentLines {
		viewString += truncatedVisibleContentLines[i] + "\n"
	}

	nVisibleLines := len(visibleContentLines.lines)
	if visibleContentLines.showFooter {
		// pad so footer shows up at bottom
		padCount := max(0, m.getNumContentLines()-nVisibleLines-1) // 1 for footer itself
		viewString += strings.Repeat("\n", padCount)
		viewString += m.getTruncatedFooterLine(visibleContentLines)
	} else {
		viewString = strings.TrimSuffix(viewString, "\n")
	}

	return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(viewString)
}

// SetContent sets the allItems, the selectable set of lines in the viewport
func (m *Model[T]) SetContent(content []T) {
	var initialNumLinesAboveSelection int
	var stayAtTop, stayAtBottom bool
	var prevSelection T
	if m.selectionEnabled {
		if inView := m.selectionInViewInfo(); inView.numLinesSelectionInView > 0 {
			initialNumLinesAboveSelection = inView.numLinesAboveSelection
		}
		if m.topSelectionSticky && len(m.allItems) > 0 && m.selectedItemIdx == 0 {
			stayAtTop = true
		} else if m.bottomSelectionSticky && (len(m.allItems) == 0 || (m.selectedItemIdx == len(m.allItems)-1)) {
			stayAtBottom = true
		} else if m.maintainSelection && 0 <= m.selectedItemIdx && m.selectedItemIdx < len(m.allItems) {
			prevSelection = m.allItems[m.selectedItemIdx]
		}
	}

	m.allItems = content
	// ensure topItemIdx and topItemLineOffset are valid given new content
	m.safelySetTopItemIdxAndOffset(m.topItemIdx, m.topItemLineOffset)

	// ensure xOffset is valid given new content
	m.safelySetXOffset(m.xOffset)

	if m.selectionEnabled {
		if stayAtTop {
			m.selectedItemIdx = 0
		} else if stayAtBottom {
			m.selectedItemIdx = max(0, len(m.allItems)-1)
			m.scrollSoSelectionInView()
		} else if m.maintainSelection {
			// TODO: could flag when content is sorted & comparable and use binary search instead
			found := false
			for i := range m.allItems {
				if m.allItems[i].Equals(prevSelection) {
					m.selectedItemIdx = i
					found = true
					break
				}
			}
			if !found {
				m.selectedItemIdx = 0
			}
		}

		// when staying at bottom, just want to scroll so selection in view, which is done above
		if !stayAtBottom {
			m.selectedItemIdx = clampValMinMax(m.selectedItemIdx, 0, len(m.allItems)-1)
			m.scrollSoSelectionInView()
			if inView := m.selectionInViewInfo(); inView.numLinesSelectionInView > 0 {
				m.scrollUp(initialNumLinesAboveSelection - inView.numLinesAboveSelection)
			}
		}
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

// SetFooterVisible sets whether the viewport shows the footer when it overflows
func (m *Model[T]) SetFooterVisible(footerVisible bool) {
	m.footerVisible = footerVisible
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
		if inView := m.selectionInViewInfo(); inView.numLinesSelectionInView > 0 {
			initialNumLinesAboveSelection = inView.numLinesAboveSelection
		}
	}
	m.wrapText = wrapText
	m.topItemLineOffset = 0
	m.xOffset = 0
	if m.selectionEnabled {
		m.scrollSoSelectionInView()
		if inView := m.selectionInViewInfo(); inView.numLinesSelectionInView > 0 {
			m.scrollUp(initialNumLinesAboveSelection - inView.numLinesAboveSelection)
			m.scrollSoSelectionInView()
		}
	}
	m.safelySetTopItemIdxAndOffset(m.topItemIdx, m.topItemLineOffset)
}

// GetWrapText returns whether the viewport wraps text
func (m Model[T]) GetWrapText() bool {
	return m.wrapText
}

// SetWidth sets the viewport's width
func (m *Model[T]) SetWidth(width int) {
	m.setWidthHeight(width, m.height)
}

// SetHeight sets the viewport's height, including header and footer
func (m *Model[T]) SetHeight(height int) {
	m.setWidthHeight(m.width, height)
}

// SetSelectedItemIdx sets the selected context index. Automatically puts selection in view as necessary
func (m *Model[T]) SetSelectedItemIdx(selectedItemIdx int) {
	if !m.selectionEnabled || m.getNumContentLines() == 0 {
		return
	}
	m.selectedItemIdx = clampValMinMax(selectedItemIdx, 0, len(m.allItems)-1)
	m.scrollSoSelectionInView()
}

// GetSelectedItemIdx returns the currently selected item index
func (m Model[T]) GetSelectedItemIdx() int {
	if !m.selectionEnabled {
		return 0
	}
	return m.selectedItemIdx
}

// GetSelectedItem returns a pointer to the currently selected item
func (m Model[T]) GetSelectedItem() *T {
	if !m.selectionEnabled || m.selectedItemIdx >= len(m.allItems) || m.selectedItemIdx < 0 {
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
}

func (m *Model[T]) ScrollSoItemIdxInView(itemIdx int) {
	if len(m.allItems) == 0 {
		m.safelySetTopItemIdxAndOffset(0, 0)
		return
	}
	originalTopItemIdx, originalTopItemLineOffset := m.topItemIdx, m.topItemLineOffset

	numLinesInItem := 1
	if m.wrapText {
		numLinesInItem = m.numLinesForItem(itemIdx)
	}

	visibleLines := m.getVisibleContentLines()
	numItemLinesInView := 0
	for i := range visibleLines.itemIndexes {
		if visibleLines.itemIndexes[i] == itemIdx {
			numItemLinesInView++
		}
	}
	if numLinesInItem != numItemLinesInView {
		if m.topItemIdx < itemIdx {
			// if item is below, scroll until it's fully in view at the bottom
			m.topItemIdx, m.topItemLineOffset = itemIdx, 0
			// then scroll up so that item is at the bottom, unless it already takes up the whole screen
			m.scrollUp(max(0, m.getNumContentLines()-numLinesInItem))
		} else {
			// if item above, scroll until it's fully in view at the top
			m.topItemIdx, m.topItemLineOffset = itemIdx, 0
		}
	}

	if m.selectionEnabled {
		// if scrolled such that selection is fully out of view, undo it
		if m.selectionInViewInfo().numLinesSelectionInView == 0 {
			m.topItemIdx, m.topItemLineOffset = originalTopItemIdx, originalTopItemLineOffset
		}
	}
}

func (m Model[T]) maxLineWidth() int {
	maxLineWidth := 0
	headerLines := m.getVisibleHeaderLines()
	visibleContentLines := m.getVisibleContentLines()
	allVisibleLines := append(headerLines, visibleContentLines.lines...)
	if visibleContentLines.showFooter {
		allVisibleLines = append(allVisibleLines, m.getTruncatedFooterLine(visibleContentLines))
	}
	for i := range allVisibleLines {
		if w := lipgloss.Width(allVisibleLines[i]); w > maxLineWidth {
			maxLineWidth = w
		}
	}
	return maxLineWidth
}

func (m Model[T]) numLinesForItem(itemIdx int) int {
	if m.width == 0 {
		return 0
	}
	if len(m.allItems) == 0 || itemIdx < 0 || itemIdx >= len(m.allItems) {
		return 0
	}
	lb := m.getLineBuffer(m.allItems[itemIdx].Render(), m.width)
	return len(lb.WrappedLines(m.height, "", lipgloss.NewStyle()))
}

func (m *Model[T]) safelySetXOffset(n int) {
	maxXOffset := m.maxLineWidth() - m.width
	m.xOffset = max(0, min(maxXOffset, n))
}

func (m *Model[T]) setWidthHeight(width, height int) {
	m.width, m.height = max(0, width), max(0, height)
	if m.width == 0 || m.height == 0 {
		return
	}
	if m.selectionEnabled {
		m.scrollSoSelectionInView()
	}
	m.safelySetTopItemIdxAndOffset(m.topItemIdx, m.topItemLineOffset)
}

func (m *Model[T]) safelySetTopItemIdxAndOffset(topItemIdx, topItemLineOffset int) {
	maxTopItemIdx, maxTopItemLineOffset := m.maxItemIdxAndMaxTopLineOffset()
	m.topItemIdx = clampValMinMax(topItemIdx, 0, maxTopItemIdx)
	m.topItemLineOffset = topItemLineOffset
	if m.topItemIdx == maxTopItemIdx {
		m.topItemLineOffset = clampValMinMax(topItemLineOffset, 0, maxTopItemLineOffset)
	}
}

// getNumContentLines returns the number of lines of between the header and footer
func (m *Model[T]) getNumContentLines() int {
	contentHeight := m.height - len(m.getVisibleHeaderLines())
	visibleContentLines := m.getVisibleContentLines()
	if visibleContentLines.showFooter {
		contentHeight-- // one for footer
	}
	return max(0, contentHeight)
}

func (m *Model[T]) scrollSoSelectionInView() {
	if !m.selectionEnabled {
		panic("scrollSoSelectionInView called when selection is not enabled")
	}
	m.ScrollSoItemIdxInView(m.selectedItemIdx)
}

func (m *Model[T]) selectedItemIdxDown(n int) {
	m.SetSelectedItemIdx(m.selectedItemIdx + n)
}

func (m *Model[T]) selectedItemIdxUp(n int) {
	m.SetSelectedItemIdx(m.selectedItemIdx - n)
}

func (m *Model[T]) scrollDown(n int) {
	m.scrollByNLines(n)
}

func (m *Model[T]) scrollUp(n int) {
	m.scrollByNLines(-n)
}

func (m *Model[T]) viewLeft(n int) {
	m.safelySetXOffset(m.xOffset - n)
}

func (m *Model[T]) viewRight(n int) {
	m.safelySetXOffset(m.xOffset + n)
}

func (m *Model[T]) getLineBuffer(s string, width int) *linebuffer.LineBuffer {
	k := fmt.Sprintf("%s-%d", s, width)
	if lb, ok := m.lineBufferCache.Get(k); ok {
		lb.SeekToWidth(0)
		return lb
	}
	lb := linebuffer.New(s, width)
	m.lineBufferCache.Add(k, &lb)
	return &lb
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
					numLinesInTopItem := m.numLinesForItem(newTopItemIdx)
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
			numLinesInTopItem := m.numLinesForItem(newTopItemIdx)
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
					numLinesInTopItem = m.numLinesForItem(newTopItemIdx)
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
	m.safelySetXOffset(m.xOffset)
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
			lb := m.getLineBuffer(s, m.width)
			wrappedHeaderLines = append(wrappedHeaderLines, lb.WrappedLines(m.height, "", lipgloss.NewStyle())...)
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
	if m.width == 0 {
		return visibleContentLinesResult{lines: nil, itemIndexes: nil, showFooter: false}
	}
	if len(m.allItems) == 0 {
		return visibleContentLinesResult{lines: nil, itemIndexes: nil, showFooter: false}
	}

	var contentLines []string
	var itemIndexes []int

	numLinesAfterHeader := max(0, m.height-len(m.getVisibleHeaderLines()))

	addLine := func(l string, itemIndex int) bool {
		contentLines = append(contentLines, l)
		itemIndexes = append(itemIndexes, itemIndex)
		return len(contentLines) == numLinesAfterHeader
	}
	addLines := func(ls []string, itemIndex int) bool {
		for i := range ls {
			if addLine(ls[i], itemIndex) {
				return true
			}
		}
		return false
	}

	currItemIdx := clampValMinMax(m.topItemIdx, 0, len(m.allItems)-1)

	currItem := m.allItems[currItemIdx]
	done := numLinesAfterHeader == 0
	if done {
		return visibleContentLinesResult{lines: contentLines, itemIndexes: itemIndexes, showFooter: false}
	}

	if m.wrapText {
		lb := m.getLineBuffer(strings.TrimRightFunc(currItem.Render(), unicode.IsSpace), m.width)
		itemLines := lb.WrappedLines(m.height, m.stringToHighlight, m.highlightStyle(currItemIdx))
		offsetLines := safeSliceFromIdx(itemLines, m.topItemLineOffset)
		done = addLines(offsetLines, currItemIdx)

		for !done {
			currItemIdx += 1
			if currItemIdx >= len(m.allItems) {
				done = true
			} else {
				currItem = m.allItems[currItemIdx]
				lb = m.getLineBuffer(strings.TrimRightFunc(currItem.Render(), unicode.IsSpace), m.width)
				itemLines = lb.WrappedLines(m.height, m.stringToHighlight, m.highlightStyle(currItemIdx))
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

	scrolledToTop := m.topItemIdx == 0 && m.topItemLineOffset == 0
	showFooter := false
	if scrolledToTop && len(contentLines)+1 >= numLinesAfterHeader {
		// if seeing all the content on screen, show footer
		// if one blank line at bottom, still show footer
		// if two blank lines at bottom, do not show footer
		showFooter = true
	}
	if !scrolledToTop {
		// if scrolled at all, should be showing footer
		showFooter = true
	}

	if !m.footerVisible {
		showFooter = false
	}

	if showFooter {
		// num visible lines exceeds vertical space, leave one line for the footer
		contentLines = safeSliceUpToIdx(contentLines, numLinesAfterHeader-1)
		itemIndexes = safeSliceUpToIdx(itemIndexes, numLinesAfterHeader-1)
	}
	return visibleContentLinesResult{lines: contentLines, itemIndexes: itemIndexes, showFooter: showFooter}
}

func (m Model[T]) highlightStyle(itemIdx int) lipgloss.Style {
	if m.selectionEnabled && itemIdx == m.selectedItemIdx {
		return m.HighlightStyleIfSelected
	}
	return m.HighlightStyle
}

func (m Model[T]) getTruncatedFooterLine(visibleContentLines visibleContentLinesResult) string {
	numerator := m.selectedItemIdx + 1 // 0th line is 1st
	denominator := len(m.allItems)
	if !visibleContentLines.showFooter {
		panic("getTruncatedFooterLine called when footer should not be shown")
	}
	if len(visibleContentLines.lines) == 0 {
		return ""
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
	// use m.continuationIndicator regardless of wrapText

	footerBuffer := m.getLineBuffer(footerString, m.width)
	return m.FooterStyle.Render(footerBuffer.PopLeft(m.continuationIndicator, "", lipgloss.NewStyle()))
}

func (m Model[T]) getLineContinuationIndicator() string {
	if m.wrapText {
		return ""
	}
	return m.continuationIndicator
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

type selectionInViewInfoResult struct {
	numLinesSelectionInView int
	numLinesAboveSelection  int
}

func (m Model[T]) selectionInViewInfo() selectionInViewInfoResult {
	if !m.selectionEnabled {
		panic("selectionInViewInfo called when selection is disabled")
	}
	visibleContentLines := m.getVisibleContentLines()
	numLinesSelectionInView := 0
	numLinesAboveSelection := 0
	assignedNumLinesAboveSelection := false
	for i := range visibleContentLines.itemIndexes {
		if visibleContentLines.itemIndexes[i] == m.selectedItemIdx {
			if !assignedNumLinesAboveSelection {
				numLinesAboveSelection = i
				assignedNumLinesAboveSelection = true
			}
			numLinesSelectionInView++
		}
	}
	return selectionInViewInfoResult{
		numLinesSelectionInView: numLinesSelectionInView,
		numLinesAboveSelection:  numLinesAboveSelection,
	}
}

func (m Model[T]) maxItemIdxAndMaxTopLineOffset() (int, int) {
	lenAllItems := len(m.allItems)
	if lenAllItems == 0 {
		return 0, 0
	}
	if !m.wrapText {
		return max(0, lenAllItems-m.getNumContentLines()), 0
	}
	// wrapped
	maxTopItemIdx, maxTopItemLineOffset := lenAllItems-1, 0
	nLinesLastItem := m.numLinesForItem(lenAllItems - 1)
	if m.getNumContentLines() <= nLinesLastItem {
		// same item, just change offset
		maxTopItemLineOffset = nLinesLastItem - m.getNumContentLines()
	} else {
		// take lines from items until scrolled up desired amount
		n := m.getNumContentLines() - nLinesLastItem
		for n > 0 {
			maxTopItemIdx -= 1
			if maxTopItemIdx < 0 {
				// scrolled up past top - stay at top
				maxTopItemIdx = 0
				maxTopItemLineOffset = 0
				break
			}
			numLinesInTopItem := m.numLinesForItem(maxTopItemIdx)
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

func (m Model[T]) getNumVisibleItems() int {
	if !m.wrapText {
		return m.getNumContentLines()
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

func (m Model[T]) styleSelection(s string) string {
	split := surroundingAnsiRegex.Split(s, -1)
	matches := surroundingAnsiRegex.FindAllString(s, -1)
	var builder strings.Builder

	// Pre-allocate the builder's capacity based on the s string length
	// This is optional but can improve performance for longer strings
	builder.Grow(len(s))

	for i, section := range split {
		if section != "" {
			builder.WriteString(m.SelectedItemStyle.Render(section))
		}
		if i < len(split)-1 && i < len(matches) {
			builder.WriteString(matches[i])
		}
	}
	return builder.String()
}
