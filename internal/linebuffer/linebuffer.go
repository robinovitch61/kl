package linebuffer

import (
	"github.com/robinovitch61/kl/internal/constants"
	"strings"
	"unicode/utf8"
)

type LineBuffer struct {
	line                      string
	lineRunes                 []rune
	plainText                 string
	plainTextRunes            []rune
	byteOffsets               []int
	ansiCodeIndexes           [][]int
	lineContinuationIndicator string
}

func New(line, lineContinuationIndicator string) LineBuffer {
	ansiCodeIndexes := constants.AnsiRegex.FindAllStringIndex(line, -1)

	var plainText string
	if len(ansiCodeIndexes) > 0 {
		plainText = constants.AnsiRegex.ReplaceAllString(line, "")
	} else {
		plainText = line
	}
	plainTextRunes := []rune(plainText)

	var byteOffsets []int
	if len(ansiCodeIndexes) > 0 {
		byteOffsets = initByteOffsets(plainTextRunes)
	}

	return LineBuffer{
		line:                      line,
		lineRunes:                 []rune(line),
		plainText:                 plainText,
		plainTextRunes:            plainTextRunes,
		byteOffsets:               byteOffsets,
		ansiCodeIndexes:           ansiCodeIndexes,
		lineContinuationIndicator: lineContinuationIndicator,
	}
}

func (l LineBuffer) Truncate(xOffset, width int) string {
	if width <= 0 {
		return ""
	}
	if xOffset == 0 && len(l.lineContinuationIndicator) == 0 && !strings.Contains(l.line, "\x1b") {
		if len(l.line) <= width {
			return l.line
		}
		return string(l.lineRunes[:width])
	}

	lenPlainText := len(l.plainTextRunes)

	if lenPlainText == 0 || xOffset >= lenPlainText {
		if len(l.ansiCodeIndexes) > 0 {
			return l.reapplyANSI("", 0)
		}
		return ""
	}

	indicatorLen := utf8.RuneCountInString(l.lineContinuationIndicator)
	if width <= indicatorLen && lenPlainText > width {
		return l.lineContinuationIndicator[:width]
	}

	var b strings.Builder

	start := xOffset
	if start < 0 {
		start = 0
	}
	end := xOffset + width
	if end > lenPlainText {
		end = lenPlainText
	}
	if end <= start {
		if len(l.ansiCodeIndexes) > 0 {
			return l.reapplyANSI("", 0)
		}
		return ""
	}

	if start == 0 && end == lenPlainText && len(l.ansiCodeIndexes) == 0 {
		return l.line
	}

	visible := string(l.plainTextRunes[start:end])

	if indicatorLen > 0 {
		if end-start <= indicatorLen && lenPlainText > indicatorLen {
			return l.lineContinuationIndicator[:min(indicatorLen, end-start)]
		}
		visLen := utf8.RuneCountInString(visible)
		if xOffset > 0 && visLen > indicatorLen {
			b.WriteString(l.lineContinuationIndicator)
			b.WriteString(string([]rune(visible)[indicatorLen:]))
			visible = b.String()
		} else if xOffset > 0 {
			visible = l.lineContinuationIndicator
		}
		if end < lenPlainText && visLen > indicatorLen {
			b.Reset()
			b.WriteString(string([]rune(visible)[:visLen-indicatorLen]))
			b.WriteString(l.lineContinuationIndicator)
			visible = b.String()
		} else if end < lenPlainText {
			visible = l.lineContinuationIndicator
		}
	}

	if len(l.ansiCodeIndexes) > 0 {
		return l.reapplyANSI(visible, l.byteOffsets[start])
	}
	return visible
}

func (l LineBuffer) reapplyANSI(truncated string, startBytes int) string {
	var result []byte
	var lenAnsiAdded int
	truncatedBytes := []byte(truncated)

	ansiCodeIndexes := l.ansiCodeIndexes
	for i := 0; i < len(truncatedBytes); {
		for len(ansiCodeIndexes) > 0 {
			candidateAnsi := ansiCodeIndexes[0]
			codeStart, codeEnd := candidateAnsi[0], candidateAnsi[1]
			originalIdx := startBytes + i + lenAnsiAdded
			if codeStart <= originalIdx || codeEnd <= originalIdx {
				result = append(result, l.line[codeStart:codeEnd]...)
				lenAnsiAdded += codeEnd - codeStart
				ansiCodeIndexes = ansiCodeIndexes[1:]
			} else {
				break
			}
		}

		_, size := utf8.DecodeRune(truncatedBytes[i:])
		result = append(result, truncatedBytes[i:i+size]...)
		i += size
	}

	// add remaining ansi codes in order to end
	for _, codeIndexes := range ansiCodeIndexes {
		codeStart, codeEnd := codeIndexes[0], codeIndexes[1]
		result = append(result, l.line[codeStart:codeEnd]...)
	}

	return string(result)
}

func initByteOffsets(runes []rune) []int {
	offsets := make([]int, len(runes)+1)
	currentOffset := 0
	for i, r := range runes {
		offsets[i] = currentOffset
		currentOffset += utf8.RuneLen(r)
	}
	offsets[len(runes)] = currentOffset
	return offsets
}
