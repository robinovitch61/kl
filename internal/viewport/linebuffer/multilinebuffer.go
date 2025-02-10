package linebuffer

import (
	"github.com/charmbracelet/lipgloss/v2"
)

// MultiLineBuffer implements LineBufferer by wrapping multiple LineBuffers without extra memory allocation
type MultiLineBuffer struct {
	buffers []*LineBuffer
	//currentBufferIdx int // tracks which buffer we're currently reading from
	totalWidth int // cached total width across all buffers
}

// type assertion that MultiLineBuffer implements LineBufferer
var _ LineBufferer = MultiLineBuffer{}

// type assertion that *MultiLineBuffer implements LineBufferer
var _ LineBufferer = (*MultiLineBuffer)(nil)

func NewMulti(buffers ...*LineBuffer) *MultiLineBuffer {
	if len(buffers) == 0 {
		return &MultiLineBuffer{}
	}

	totalWidth := 0
	for _, buf := range buffers {
		totalWidth += buf.Width()
	}

	return &MultiLineBuffer{
		buffers:    buffers,
		totalWidth: totalWidth,
	}
}

func (m MultiLineBuffer) Width() int {
	return m.totalWidth
}

// TODO LEO: don't use this for e.g. search, instead inject filter into a Matches() bool method here
// TODO LEO: or store the concatenated string on init?
func (m MultiLineBuffer) Content() string {
	totalLen := 0
	for _, buf := range m.buffers {
		totalLen += len(buf.Content())
	}

	result := make([]byte, 0, totalLen)
	for _, buf := range m.buffers {
		result = append(result, buf.Content()...)
	}
	return string(result)
}

func (m MultiLineBuffer) Take(
	startWidth, takeWidth int,
	continuation, toHighlight string,
	highlightStyle lipgloss.Style,
) (string, int) {
	//if len(m.buffers) == 0 || width == 0 {
	//	return ""
	//}
	//
	//var result string
	//remainingWidth := width
	//
	//for remainingWidth > 0 && m.currentBufferIdx < len(m.buffers) {
	//	currentBuffer := m.buffers[m.currentBufferIdx]
	//	chunk := currentBuffer.Take(remainingWidth, continuation, toHighlight, highlightStyle)
	//
	//	if chunk == "" {
	//		if m.currentBufferIdx < len(m.buffers)-1 {
	//			m.currentBufferIdx++
	//			continue
	//		}
	//		break
	//	}
	//
	//	result += chunk
	//	remainingWidth -= lipgloss.Width(chunk)
	//
	//	if remainingWidth > 0 && m.currentBufferIdx < len(m.buffers)-1 {
	//		m.currentBufferIdx++
	//	}
	//}
	//
	//return result
	return "", 0
}

func (m MultiLineBuffer) WrappedLines(
	width int,
	maxLinesEachEnd int,
	toHighlight string,
	toHighlightStyle lipgloss.Style,
) []string {
	//if width <= 0 {
	//	return []string{}
	//}
	//
	//if maxLinesEachEnd <= 0 {
	//	maxLinesEachEnd = -1
	//}
	//
	//// preserve empty lines
	//if m.Content() == "" {
	//	return []string{""}
	//}
	//
	//var res []string
	//totalLines := m.totalLines(width)
	//
	//m.SeekToWidth(0)
	//if maxLinesEachEnd > 0 && totalLines > maxLinesEachEnd*2 {
	//	for nLines := 0; nLines < maxLinesEachEnd; nLines++ {
	//		res = append(res, m.Take(width, "", toHighlight, toHighlightStyle))
	//	}
	//
	//	m.seekToLine(totalLines-maxLinesEachEnd, width)
	//	for nLines := 0; nLines < maxLinesEachEnd; nLines++ {
	//		res = append(res, m.Take(width, "", toHighlight, toHighlightStyle))
	//	}
	//} else {
	//	for nLines := 0; nLines < totalLines; nLines++ {
	//		res = append(res, m.Take(width, "", toHighlight, toHighlightStyle))
	//	}
	//}
	//
	//m.SeekToWidth(0)
	//return res
	return []string{}
}

func (m MultiLineBuffer) Repr() string {
	v := "Multi("
	for i := range m.buffers {
		if i > 0 {
			v += ", "
		}
		v += m.buffers[i].Repr()
	}
	v += ")"
	return v
}

//func (m MultiLineBuffer) totalLines(width int) int {
//	if width == 0 {
//		return 0
//	}
//
//	total := 0
//	for _, buf := range m.buffers {
//		bufferWidth := int(buf.fullWidth())
//		if bufferWidth > 0 {
//			total += (bufferWidth + width - 1) / width
//		}
//	}
//	return total
//}

//func (m *MultiLineBuffer) seekToLine(line int, width int) {
//if line <= 0 {
//	m.SeekToWidth(0)
//	return
//}
//
//// find which buffer contains our target line
//remainingLines := line
//for i, buf := range m.buffers {
//	bufferLines := (int(buf.fullWidth()) + width - 1) / width
//	if remainingLines < bufferLines {
//		m.currentBufferIdx = i
//		m.buffers[i].seekToLine(remainingLines, width)
//		return
//	}
//	remainingLines -= bufferLines
//}
//
//if len(m.buffers) > 0 {
//	m.currentBufferIdx = len(m.buffers) - 1
//	lastBuf := m.buffers[m.currentBufferIdx]
//	lastBuf.seekToLine((int(lastBuf.fullWidth())+width-1)/width, width)
//}
//}
