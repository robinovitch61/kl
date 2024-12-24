package linebuffer

import (
	"fmt"
	"github.com/robinovitch61/kl/internal/constants"
	"github.com/robinovitch61/kl/internal/dev"
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
		if len(l.lineRunes) <= width {
			return l.line
		}
		return string(l.lineRunes[:width])
	}

	lenPlainText := len(l.plainTextRunes)

	if lenPlainText == 0 || xOffset >= lenPlainText {
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
			return ""
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
		return reapplyANSI(l.line, visible, l.byteOffsets[start], l.ansiCodeIndexes)
	}
	return visible
}

func reapplyANSI(original, truncated string, truncByteOffset int, ansiCodeIndexes [][]int) string {
	var result []byte
	var lenAnsiAdded int
	isReset := true
	truncatedBytes := []byte(truncated)

	for i := 0; i < len(truncatedBytes); {
		// collect all ansi codes that should be applied immediately before the current runes
		var ansisToAdd []string
		for len(ansiCodeIndexes) > 0 {
			candidateAnsi := ansiCodeIndexes[0]
			codeStart, codeEnd := candidateAnsi[0], candidateAnsi[1]
			originalIdx := truncByteOffset + i + lenAnsiAdded
			if codeStart <= originalIdx {
				code := original[codeStart:codeEnd]
				isReset = code == "\x1b[m"
				ansisToAdd = append(ansisToAdd, code)
				lenAnsiAdded += codeEnd - codeStart
				ansiCodeIndexes = ansiCodeIndexes[1:]
			} else {
				break
			}
		}

		// if there's just a bunch of reset sequences, compress it to one
		allReset := len(ansisToAdd) > 0
		for _, ansi := range ansisToAdd {
			if ansi != "\x1b[m" {
				allReset = false
				break
			}
		}
		if allReset {
			ansisToAdd = []string{"\x1b[m"}
		}

		// if the last sequence in a set of more than one is a reset, no point adding any of them
		redundant := len(ansisToAdd) > 1 && ansisToAdd[len(ansisToAdd)-1] == "\x1b[m"
		if !redundant {
			for _, ansi := range ansisToAdd {
				result = append(result, ansi...)
			}
		}

		// add the bytes of the current rune
		_, size := utf8.DecodeRune(truncatedBytes[i:])
		result = append(result, truncatedBytes[i:i+size]...)
		i += size
	}

	if !isReset {
		result = append(result, "\x1b[m"...)
	}

	return string(result)
}

func initByteOffsets(runes []rune) []int {
	offsets := make([]int, len(runes)+1)
	currentOffset := 0
	for i, r := range runes {
		offsets[i] = currentOffset
		runeLen := utf8.RuneLen(r)
		if runeLen == -1 {
			// invalid utf-8 value, assume 1 byte
			dev.Debug(fmt.Sprintf("invalid utf-8 value: %v", r))
			runeLen = 1
		}
		currentOffset += runeLen
	}
	offsets[len(runes)] = currentOffset
	return offsets
}
