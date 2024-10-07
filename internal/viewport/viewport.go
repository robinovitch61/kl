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
// - allItems: an item to be rendered in the viewport
// - line: a row in the terminal
//
// wrap disabled:
//                           allItems index   line index
// this is the first line    0               0
// this is the second line   1               1
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

	// contentHeight is the number of lines of shown between the header and footer
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
				m.topItemIdx = 0
				m.topItemLineOffset = 0
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

	headerLines := m.getHeaderLines()
	for _, headerLine := range headerLines {
		visiblePart := m.getVisible(headerLine)
		addLineToViewString(visiblePart)
	}

	//hasStringToHighlight := stringWidth(m.stringToHighlight) != 0

	// get the lines to show based on the topItemIdx and topItemLineOffset
	var lines []string
	addLine := func(l string) bool {
		lines = append(lines, l)
		return len(lines) == m.contentHeight
	}
	currItemIdx := m.topItemIdx
	currItem := m.allItems[currItemIdx]
	done := m.contentHeight <= 0
	if m.wrapText {
		itemLines := wrap(currItem.Render(), m.width)
		truncLines := itemLines[m.topItemLineOffset:]
		for i := range truncLines {
			if done {
				break
			}
			done = addLine(truncLines[i])
		}

		for !done {
			currItemIdx += 1
			currItem = m.allItems[currItemIdx]
			itemLines = wrap(currItem.Render(), m.width)
			for i := range itemLines {
				if done {
					break
				}
				done = addLine(itemLines[i])
			}
			if done {
				break
			}
		}
	} else {
		for !done {
			currItemIdx += 1
			currItem = m.allItems[currItemIdx]
			done = addLine(currItem.Render())
		}
	}

	// get the visible part of each line given the xOffset
	var linesAccountingForWidth []string
	for i := range lines {
		linesAccountingForWidth = append(linesAccountingForWidth, m.getVisible(lines[i]))
	}

	// add selection style
	// TODO LEO

	nVisibleLines := len(strings.Split(viewString, "\n"))
	if footerHeight > 0 {
		// pad so footer shows up at bottom
		padCount := max(0, m.contentHeight-nVisibleLines-footerHeight)
		viewString += strings.Repeat("\n", padCount)
		viewString += footerString
	}
	return lipgloss.Place(m.width, m.height, 0, 0, viewString)
}

// SetContent sets the allItems, the selectable set of lines in the viewport
func (m *Model[T]) SetContent(content []T) {
	var stayAtTop, stayAtBottom bool
	if m.topSelectionSticky && m.selectionEnabled && m.selectedItemIdx == 0 {
		stayAtTop = true
	}
	if m.bottomSelectionSticky && m.selectionEnabled && m.selectedItemIdx == m.lastItemIdx() {
		stayAtBottom = true
	}

	m.allItems = content
	m.updateContentHeight()

	// fix any sort of potential selection issues
	if m.selectedItemIdx < 0 {
		m.selectedItemIdx = 0
	} else if finalContentIdx := m.lastItemIdx(); m.selectedItemIdx > finalContentIdx {
		m.selectedItemIdx = finalContentIdx
	}

	// stay at top, bottom, or maintain previous selection if desired
	if stayAtTop {
		m.selectedItemIdx = 0
	} else if stayAtBottom {
		m.selectedItemIdx = m.lastItemIdx()
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
	m.wrapText = wrapText
	m.updateContentHeight()
	m.resetHorizontalOffset()
}

// GetWrapText returns whether the viewport wraps text
func (m Model[T]) GetWrapText() bool {
	return m.wrapText
}

// SetWidth sets the viewport's width
func (m *Model[T]) SetWidth(width int) {
	m.width = width
	m.updateContentHeight()
}

// GetWidth returns the viewport's width
func (m Model[T]) GetWidth() int {
	return m.width
}

// SetHeight sets the viewport's height, including header and footer
func (m *Model[T]) SetHeight(height int) {
	m.height = height
	m.updateContentHeight()
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

	if maxSelectedIdx := m.lastItemIdx(); n > maxSelectedIdx {
		m.selectedItemIdx = maxSelectedIdx
	} else {
		m.selectedItemIdx = max(0, n)
	}
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
	m.updateContentHeight()
}

// resetHorizontalOffset resets the horizontal offset to the leftmost position
func (m *Model[T]) resetHorizontalOffset() {
	m.xOffset = 0
}

// ScrollToTop scrolls the viewport to the top
func (m *Model[T]) ScrollToTop() {
	m.selectedContentIdxUp(m.selectedItemIdx)
	m.viewUp(m.selectedItemIdx)
}

// ScrollToBottom scrolls the viewport to the bottom
func (m *Model[T]) ScrollToBottom() {
	m.selectedContentIdxDown(len(m.allItems))
	m.viewDown(len(m.allItems))
}

func (m *Model[T]) setXOffset(n int) {
	maxXOffset := m.maxLineLength - m.width
	m.xOffset = max(0, min(maxXOffset, n))
}

func (m *Model[T]) updateContentHeight() {
	_, footerHeight := m.getFooter()
	contentHeight := m.height - (len(m.getHeaderLines()) + footerHeight)
	m.contentHeight = max(0, contentHeight)
}

func (m *Model[T]) selectedContentIdxDown(n int) {
	m.SetSelectedContentIdx(m.selectedItemIdx + n)
}

func (m *Model[T]) selectedContentIdxUp(n int) {
	m.SetSelectedContentIdx(m.selectedItemIdx - n)
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

func (m Model[T]) getHeaderLines() []string {
	if !m.wrapText {
		return m.header
	}
	var wrappedHeaderLines []string
	for _, s := range m.header {
		wrappedHeaderLines = append(wrappedHeaderLines, wrap(s, m.width)...)
	}
	return wrappedHeaderLines
}

func (m *Model[T]) maxVisibleLineIdx() int {
	return m.getLenContentStrings() - 1
}

func (m Model[T]) lastItemIdx() int {
	return len(m.allItems) - 1
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
		return m.selectedItemIdx - m.yOffset
	} else {
		// start at m.yOffset, and until m.allItems[m.wrappedContentIdxToContentIdx] == initialSelection, increment top padding
		for i := m.yOffset; i < len(m.wrappedContent); i++ {
			if m.allItems[m.wrappedContentIdxToContentIdx[i]].Equals(*selection) {
				return i - m.yOffset
			}
		}
	}
	return 0
}

func (m Model[T]) getVisible(line string) string {
	lineNoTrailingSpace := strings.TrimRightFunc(line, unicode.IsSpace)
	return getVisiblePartOfLine(lineNoTrailingSpace, m.xOffset, m.width, m.lineContinuationIndicator)
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
	return m.selectedItemIdx == len(m.allItems)-1
}

func (m Model[T]) getFooter() (string, int) {
	numerator := m.selectedItemIdx + 1
	denominator := len(m.allItems)
	totalNumLines := m.getLenContentStrings()

	// if selection is disabled, percentage should show from the bottom of the visible allItems
	// such that panning the view to the bottom shows 100%
	if !m.selectionEnabled {
		numerator = m.yOffset + m.contentHeight
		denominator = totalNumLines
	}

	if totalNumLines >= m.height-len(m.getHeaderLines()) {
		percentScrolled := percent(numerator, denominator)
		footerString := fmt.Sprintf("%d%% (%d/%d)", percentScrolled, numerator, denominator)
		renderedFooterString := m.FooterStyle.MaxWidth(m.width).Render(footerString)
		footerHeight := lipgloss.Height(renderedFooterString)
		return renderedFooterString, footerHeight
	}
	return "", 0
}
