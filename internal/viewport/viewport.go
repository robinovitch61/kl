package viewport

import (
	"fmt"
	"github.com/robinovitch61/kl/internal/dev"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents a viewport component
type Model[T RenderableComparable] struct {
	KeyMap                    KeyMap
	LineContinuationIndicator string
	BackgroundStyle           lipgloss.Style
	HeaderStyle               lipgloss.Style
	SelectedContentStyle      lipgloss.Style
	HighlightStyle            lipgloss.Style
	HighlightStyleIfSelected  lipgloss.Style
	ContentStyle              lipgloss.Style
	FooterStyle               lipgloss.Style

	header                       []string
	wrappedHeader                []string
	content                      []T
	wrappedContent               []string
	lenLineContinuationIndicator int

	// if topSticky/bottomSticky is true and selection is enabled, selection remains at top/bottom until scroll down/up
	topSticky    bool
	bottomSticky bool

	// maintainSelection is true if the viewport should try to maintain the current selection when content changes
	maintainSelection bool

	// wrappedContentIdxToContentIdx maps the item at an index of wrappedContent to the index of content it is associated with (many wrappedContent indexes -> one content index)
	wrappedContentIdxToContentIdx map[int]int

	// contentIdxToFirstWrappedContentIdx maps the item at an index of content to the first index of wrappedContent it is associated with (index of content -> first index of wrappedContent)
	contentIdxToFirstWrappedContentIdx map[int]int

	// contentIdxToHeight maps the item at an index of content to its wrapped height in terminal rows
	contentIdxToHeight map[int]int

	// selectedContentIdx is the index of content of the currently selected item when selectionEnabled is true
	selectedContentIdx int
	stringToHighlight  string
	selectionEnabled   bool
	wrapText           bool

	// width is the width of the entire viewport in terminal columns
	width int
	// height is the height of the entire viewport in terminal rows
	height int
	// contentHeight is the height of the viewport in terminal rows, excluding the header and footer
	contentHeight int
	// maxLineLength is the maximum line length in terminal characters across header and visible content
	maxLineLength int

	// yOffset is the index of the first row shown on screen - wrappedContent[yOffset] if wrapText, otherwise content[yOffset]
	yOffset int
	// xOffset is the number of columns scrolled right when content lines overflow the viewport and wrapText is false
	xOffset int
}

// New creates a new viewport model with reasonable defaults
func New[T RenderableComparable](width, height int) (m Model[T]) {
	m.setWidthAndHeight(width, height)
	m.updateContentHeight()

	m.selectionEnabled = false
	m.wrapText = false

	m.KeyMap = DefaultKeyMap()
	m.LineContinuationIndicator = "..."
	m.lenLineContinuationIndicator = stringWidth(m.LineContinuationIndicator)
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
				m.selectedContentIdxUp(m.yOffset + m.contentHeight)
			} else {
				m.viewUp(m.yOffset + m.contentHeight)
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
	visibleLines := m.getVisibleLines()

	for _, headerLine := range header {
		headerViewLine := m.getVisiblePartOfLine(headerLine)
		addLineToViewString(m.HeaderStyle.Render(headerViewLine))
	}

	hasNoHighlight := stringWidth(m.stringToHighlight) == 0
	for idx, line := range visibleLines {
		contentIdx := m.getContentIdx(m.yOffset + idx)
		isSelected := m.selectionEnabled && contentIdx == m.selectedContentIdx

		lineStyle := m.ContentStyle
		if isSelected {
			lineStyle = m.SelectedContentStyle
		}
		contentViewLine := m.getVisiblePartOfLine(line)

		if isSelected && contentViewLine == "" {
			contentViewLine = " "
		}

		if hasNoHighlight {
			addLineToViewString(lineStyle.Render(contentViewLine))
		} else {
			// this splitting and rejoining of styled content is expensive and causes increased flickering,
			// so only do it if something is actually highlighted
			highlightStyle := m.HighlightStyle
			if isSelected {
				highlightStyle = m.HighlightStyleIfSelected
			}
			lineChunks := strings.Split(contentViewLine, m.stringToHighlight)
			var styledChunks []string
			for _, chunk := range lineChunks {
				styledChunks = append(styledChunks, lineStyle.Render(chunk))
			}
			addLineToViewString(strings.Join(styledChunks, highlightStyle.Render(m.stringToHighlight)))
		}
	}

	if footerHeight > 0 {
		// pad so footer shows up at bottom
		padCount := max(0, m.contentHeight-len(visibleLines)-footerHeight)
		viewString += strings.Repeat("\n", padCount)
		viewString += footerString
	}
	renderedViewString := m.BackgroundStyle.Width(m.width).Height(m.height).Render(viewString)

	return renderedViewString
}

// SetContent sets the content, the selectable set of lines in the viewport
func (m *Model[T]) SetContent(content []T) {
	// TODO: clean this up
	dev.Debug("Setting viewport content")
	defer dev.Debug("Done setting viewport content")

	var initialTopPadding int
	initialSelection := m.GetSelectedContent()
	attemptMaintainSelection := m.maintainSelection && m.selectionEnabled && initialSelection != nil
	if attemptMaintainSelection {
		initialTopPadding = m.numLinesBetweenSelectionAndTop()
	}

	stayAtTop := false
	if m.topSticky && m.selectionEnabled && m.selectedContentIdx == 0 {
		stayAtTop = true
	}
	stayAtBottom := false
	if m.bottomSticky && m.selectionEnabled && m.selectedContentIdx == m.maxContentIdx() {
		stayAtBottom = true
	}

	m.content = content
	if m.wrapText {
		// ok to skip because this is updated when wrap is initially enabled
		m.updateWrappedContent()
	}
	m.updateForHeaderAndContent()

	// fix any sort of potential selection issues
	if m.selectedContentIdx < 0 {
		m.selectedContentIdx = 0
	} else if m.selectedContentIdx > m.maxContentIdx() {
		m.selectedContentIdx = m.maxContentIdx()
	}

	// stay at top or bottom if desired
	if stayAtTop {
		m.selectedContentIdx = 0
	} else if stayAtBottom {
		m.selectedContentIdx = m.maxContentIdx()
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
	m.topSticky = topSticky
}

// SetBottomSticky sets whether selection should stay at bottom when new content added and selection is at the bottom
func (m *Model[T]) SetBottomSticky(bottomSticky bool) {
	m.bottomSticky = bottomSticky
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
	m.setWidthAndHeight(width, m.height)
	m.updateForHeaderAndContent()
}

// GetWidth returns the viewport's width
func (m Model[T]) GetWidth() int {
	return m.width
}

// SetHeight sets the viewport's height, including header and footer
func (m *Model[T]) SetHeight(height int) {
	m.setWidthAndHeight(m.width, height)
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

	if maxSelectedIdx := m.maxContentIdx(); n > maxSelectedIdx {
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

func (m *Model[T]) updateWrappedHeader() {
	var allWrappedHeader []string
	for _, line := range m.header {
		wrappedLinesForLine := m.getWrappedLines(line)
		allWrappedHeader = append(allWrappedHeader, wrappedLinesForLine...)
	}
	m.wrappedHeader = allWrappedHeader
}

func (m *Model[T]) updateWrappedContent() {
	var allWrappedContent []string
	wrappedContentIdxToContentIdx := make(map[int]int)
	contentIdxToFirstWrappedContentIdx := make(map[int]int)
	contentIdxToHeight := make(map[int]int)

	var wrappedContentIdx int
	for contentIdx, item := range m.content {
		line := item.Render()
		wrappedLinesForLine := m.getWrappedLines(line)
		contentIdxToHeight[contentIdx] = len(wrappedLinesForLine)
		for _, wrappedLine := range wrappedLinesForLine {
			allWrappedContent = append(allWrappedContent, wrappedLine)

			wrappedContentIdxToContentIdx[wrappedContentIdx] = contentIdx
			if _, exists := contentIdxToFirstWrappedContentIdx[contentIdx]; !exists {
				contentIdxToFirstWrappedContentIdx[contentIdx] = wrappedContentIdx
			}

			wrappedContentIdx++
		}
	}
	m.wrappedContent = allWrappedContent
	m.wrappedContentIdxToContentIdx = wrappedContentIdxToContentIdx
	m.contentIdxToFirstWrappedContentIdx = contentIdxToFirstWrappedContentIdx
	m.contentIdxToHeight = contentIdxToHeight
}

func (m *Model[T]) updateForHeaderAndContent() {
	m.updateContentHeight()
	m.ensureViewContainsSelection()
	m.updateMaxVisibleLineLength()
}

func (m *Model[T]) updateMaxVisibleLineLength() {
	m.maxLineLength = 0
	header, content := m.getHeader(), m.getVisibleLines()
	for _, line := range append(header, content...) {
		if lineLength := stringWidth(line); lineLength > m.maxLineLength {
			m.maxLineLength = lineLength
		}
	}
}

func (m *Model[T]) setWidthAndHeight(width, height int) {
	m.width, m.height = width, height
	m.updateWrappedHeader()
	m.updateWrappedContent()
}

func (m *Model[T]) ensureViewContainsSelection() {
	currentLineIdx := m.getCurrentLineIdx()
	lastVisibleLineIdx := m.lastVisibleLineIdx()
	offScreenRowCount := currentLineIdx - lastVisibleLineIdx
	if offScreenRowCount >= 0 || m.lastContentItemSelected() {
		heightOffset := m.contentIdxToHeight[m.selectedContentIdx] - 1
		if !m.wrapText {
			heightOffset = 0
		}
		m.viewDown(offScreenRowCount + heightOffset)
	} else if currentLineIdx < m.yOffset {
		m.viewUp(m.yOffset - currentLineIdx)
	}

	if maxYOffset := m.maxYOffset(); m.yOffset > maxYOffset {
		m.setYOffset(maxYOffset)
	}
}

func (m *Model[T]) updateContentHeight() {
	_, footerHeight := m.getFooter()
	contentHeight := m.height - len(m.getHeader()) - footerHeight
	m.contentHeight = max(0, contentHeight)
}

func (m *Model[T]) setYOffset(n int) {
	if maxYOffset := m.maxYOffset(); n > maxYOffset {
		m.yOffset = maxYOffset
	} else {
		m.yOffset = max(0, n)
	}
	m.updateMaxVisibleLineLength()
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
	if m.wrapText {
		return m.wrappedHeader
	}
	return m.header
}

func (m Model[T]) getContentStrings(start, end int) []string {
	if m.wrapText {
		if end == -1 {
			end = len(m.wrappedContent)
		}
		return m.wrappedContent[start:end]
	}

	var contentStrings []string
	if end == -1 {
		end = len(m.content)
	}
	for _, item := range m.content[start:end] {
		contentStrings = append(contentStrings, item.Render())
	}
	return contentStrings
}

func (m Model[T]) getLenContentStrings() int {
	if m.wrapText {
		return len(m.wrappedContent)
	}
	return len(m.content)
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

func (m Model[T]) maxContentIdx() int {
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

// getVisibleLines retrieves the visible content based on the yOffset and contentHeight
func (m Model[T]) getVisibleLines() []string {
	maxVisibleLineIdx := m.maxVisibleLineIdx()
	start := max(0, min(maxVisibleLineIdx, m.yOffset))
	end := start + m.contentHeight
	if end > maxVisibleLineIdx {
		return m.getContentStrings(start, -1)
	}
	return m.getContentStrings(start, end)
}

func (m Model[T]) getVisiblePartOfLine(line string) string {
	rightTrimmedLineLength := stringWidth(strings.TrimRight(line, " "))
	end := min(stringWidth(line), m.xOffset+m.width)
	start := min(end, m.xOffset)
	line = line[start:end]
	if m.xOffset+m.width < rightTrimmedLineLength {
		truncate := max(0, stringWidth(line)-m.lenLineContinuationIndicator)
		line = line[:truncate] + m.LineContinuationIndicator
	}
	if m.xOffset > 0 {
		line = m.LineContinuationIndicator + line[min(stringWidth(line), m.lenLineContinuationIndicator):]
	}
	return line
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

func (m Model[T]) getWrappedLines(line string) []string {
	if stringWidth(line) < m.width {
		return []string{line}
	}
	line = strings.TrimRight(line, " ")
	return splitLineIntoSizedChunks(line, m.width)
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
		renderedFooterString := m.FooterStyle.Copy().MaxWidth(m.width).Render(footerString)
		footerHeight := lipgloss.Height(renderedFooterString)
		return renderedFooterString, footerHeight
	}
	return "", 0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func percent(a, b int) int {
	return int(float32(a) / float32(b) * 100)
}

func splitLineIntoSizedChunks(line string, chunkSize int) []string {
	var wrappedLines []string
	for {
		lineWidth := stringWidth(line)
		if lineWidth == 0 {
			break
		}

		width := chunkSize
		if lineWidth < chunkSize {
			width = lineWidth
		}

		wrappedLines = append(wrappedLines, line[0:width])
		line = line[width:]
	}
	return wrappedLines
}

// stringWidth is a function in case in the future something like utf8.RuneCountInString or lipgloss.Width is better
func stringWidth(s string) int {
	// NOTE: lipgloss.Width is significantly less performant than len
	return len(s)
}
