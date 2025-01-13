package model

import (
	"bufio"
	"context"
	"fmt"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/google/uuid"
	"github.com/mattn/go-runewidth"
	"github.com/robinovitch61/kl/internal/constants"
	"github.com/robinovitch61/kl/internal/dev"
	"strings"
	"time"
	"unicode/utf8"
)

type Log struct {
	Timestamp time.Time
	Data      LogData
	Container Container
}

// LogData is a collection of data about a log
// since log contents are immutable, all fields of LogData should also not be mutated
type LogData struct {
	Content             string  // the log content itself
	Width               int     // width in terminal cells (not bytes or runes)
	LineRunes           []rune  // runes of line
	RuneIdxToByteOffset []int   // idx of lineRunes to byte offset. len(runeIdxToByteOffset) == len(lineRunes)
	LineNoAnsi          string  // line without ansi codes. utf-8 bytes
	LineNoAnsiRunes     []rune  // runes of lineNoAnsi. len(lineNoAnsiRunes) == len(lineNoAnsiWidths)
	LineNoAnsiWidths    []int   // terminal cell widths of lineNoAnsi. len(lineNoAnsiWidths) == len(lineNoAnsiRunes)
	LineNoAnsiCumWidths []int   // cumulative lineNoAnsiWidths
	AnsiCodeIndexes     [][]int // slice of startByte, endByte indexes of ansi codes in the line
}

type LogScanner struct {
	Container      Container
	LogChan        chan Log
	ErrChan        chan error
	cancel         context.CancelFunc
	uuid           string
	logLineScanner *bufio.Scanner
}

func NewLogScanner(container Container, scanner *bufio.Scanner, cancelK8sStream context.CancelFunc) LogScanner {
	return LogScanner{
		Container:      container,
		LogChan:        make(chan Log, 1), // this value doesn't seem to affect performance much
		ErrChan:        make(chan error, 1),
		cancel:         cancelK8sStream,
		uuid:           uuid.New().String(),
		logLineScanner: scanner,
	}
}

// StartReadingLogs starts a goroutine that reads logs from the scanner and sends them to the LogChan
func (ls LogScanner) StartReadingLogs() {
	go func() {
		for ls.logLineScanner != nil && ls.logLineScanner.Scan() {
			// logs are space-separated, so split on spaces
			vals := strings.Split(ls.logLineScanner.Text(), " ")

			// logs should have at least a timestamp and content - ignore ones that do not
			if len(vals) < 2 {
				dev.Debug(fmt.Sprintf("skipping log: %v", ls.logLineScanner.Text()))
				continue
			}

			// timestamps should be parseable as RFC3339 - ignore ones that are not
			parsedTime, err := time.Parse(time.RFC3339, vals[0])
			if err != nil {
				continue
			}

			logContent := strings.Join(vals[1:], " ")
			logContent = strings.ReplaceAll(logContent, "\t", "    ")

			// precompute LogData here as logs come in as logs are immutable. Having the LogData up front helps
			// to minimize expensive/repeated re-computation later
			newLog := Log{
				Timestamp: parsedTime,
				Data:      getLogData(logContent),
				Container: ls.Container,
			}
			ls.LogChan <- newLog
		}

		err := ls.logLineScanner.Err()
		errorExists := err != nil
		// if err is "context canceled", scanner was stopped by the user
		stoppedByUser := errorExists && err.Error() == "context canceled"

		if errorExists && !stoppedByUser {
			ls.ErrChan <- err
		}

		ls.Cancel()
		close(ls.LogChan)
		close(ls.ErrChan)
	}()
}

func (ls LogScanner) Cancel() {
	if ls.cancel != nil {
		ls.cancel()
	}
}

func (ls LogScanner) Equals(other LogScanner) bool {
	return ls.uuid == other.uuid
}

func getLogData(s string) LogData {
	ansiCodeIndexes := constants.AnsiRegex.FindAllStringIndex(s, -1)
	lineNoAnsi := stripAnsi(s)

	lineRunes := []rune(lineNoAnsi)
	runeIdxToByteOffset := initByteOffsets(lineRunes)

	lineNoAnsiRunes := []rune(lineNoAnsi)

	lineNoAnsiWidths := make([]int, len(lineNoAnsiRunes))
	lineNoAnsiCumWidths := make([]int, len(lineNoAnsiRunes))
	for i := range lineNoAnsiRunes {
		runeWidth := runewidth.RuneWidth(lineNoAnsiRunes[i])
		lineNoAnsiWidths[i] = runeWidth
		if i == 0 {
			lineNoAnsiCumWidths[i] = runeWidth
		} else {
			lineNoAnsiCumWidths[i] = lineNoAnsiCumWidths[i-1] + runeWidth
		}
	}

	return LogData{
		Content:             s,
		Width:               lipgloss.Width(s),
		LineRunes:           lineRunes,
		RuneIdxToByteOffset: runeIdxToByteOffset,
		LineNoAnsi:          lineNoAnsi,
		LineNoAnsiRunes:     lineNoAnsiRunes,
		LineNoAnsiWidths:    lineNoAnsiWidths,
		LineNoAnsiCumWidths: lineNoAnsiCumWidths,
		AnsiCodeIndexes:     ansiCodeIndexes,
	}
}

func stripAnsi(input string) string {
	return constants.AnsiRegex.ReplaceAllString(input, "")
}

func initByteOffsets(runes []rune) []int {
	offsets := make([]int, len(runes)+1)
	currentOffset := 0
	for i, r := range runes {
		offsets[i] = currentOffset
		runeLen := utf8.RuneLen(r)
		currentOffset += runeLen
	}
	offsets[len(runes)] = currentOffset
	return offsets
}
