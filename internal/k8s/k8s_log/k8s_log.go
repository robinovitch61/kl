package k8s_log

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/k8s/container"
	"github.com/robinovitch61/kl/internal/util"
	"github.com/robinovitch61/viewport/viewport/item"
)

type LogTimestamps struct {
	Short string
	Full  string
}

type Log struct {
	Timestamp      time.Time
	Timestamps     LogTimestamps
	Container      container.Container
	ContentItem    item.SingleItem
	colorize       func(string) string // optional JSON colorizer
	prettyItems    []item.SingleItem   // pretty-printed JSON lines, nil if not valid JSON or single item
	prettyComputed bool
}

// Colorize returns the colorize function, or nil if no colorization is configured.
func (l *Log) Colorize() func(string) string {
	return l.colorize
}

// GetPrettyItems returns the pretty-printed JSON lines for this log, computing
// and caching the result on first access. Returns nil if the content is not
// multi-line JSON.
func (l *Log) GetPrettyItems() []item.SingleItem {
	if !l.prettyComputed {
		if lines := util.PrettyPrintJSON(l.ContentItem.ContentNoAnsi(), l.colorize); len(lines) > 1 {
			l.prettyItems = make([]item.SingleItem, len(lines))
			for i, line := range lines {
				l.prettyItems[i] = item.NewItem(line)
			}
		}
		l.prettyComputed = true
	}
	return l.prettyItems
}

type LogScanner struct {
	Container      container.Container
	LogChan        chan Log
	ErrChan        chan error
	cancel         context.CancelFunc
	uuid           string
	logLineScanner *bufio.Scanner
	colorize       func(string) string
}

func NewLogScanner(ct container.Container, scanner *bufio.Scanner, cancelK8sStream context.CancelFunc, colorize func(string) string) LogScanner {
	return LogScanner{
		Container:      ct,
		LogChan:        make(chan Log, 1), // this value doesn't seem to affect performance much
		ErrChan:        make(chan error, 1),
		cancel:         cancelK8sStream,
		uuid:           uuid.New().String(),
		logLineScanner: scanner,
		colorize:       colorize,
	}
}

// StartReadingLogs starts a goroutine that reads logs from the scanner and sends them to the LogChan
func (ls LogScanner) StartReadingLogs() {
	go func() {
		for ls.logLineScanner != nil && ls.logLineScanner.Scan() {
			bs := ls.logLineScanner.Bytes()

			// parse everything before first space as timestamp
			firstSpace := bytes.IndexByte(bs, ' ')
			if firstSpace < 0 {
				dev.Debug(fmt.Sprintf("skipping log: %s", bs))
				continue
			}

			// timestamps should be parseable as RFC3339 - ignore ones that are not
			parsedTime, err := time.Parse(time.RFC3339, string(bs[:firstSpace]))
			if err != nil {
				dev.Debug(fmt.Sprintf("timestamp not parseable as RFC3339: %s", bs))
				continue
			}

			// content is everything after first space, trimmed
			logContent := string(bs[firstSpace+1:])
			logContent = strings.TrimRightFunc(logContent, unicode.IsSpace)
			logContent = strings.ReplaceAll(logContent, "\t", "    ")
			logContent = util.SanitizeTerminalSequences(logContent)

			if ls.colorize != nil {
				logContent = ls.colorize(logContent)
			}

			localTime := parsedTime.Local()

			contentItem := item.NewItem(logContent)

			ls.LogChan <- Log{
				Timestamp: parsedTime,
				Timestamps: LogTimestamps{
					Short: localTime.Format(time.TimeOnly),
					Full:  localTime.Format("2006-01-02T15:04:05.000Z07:00"),
				},
				Container:   ls.Container,
				ContentItem: contentItem,
				colorize:    ls.colorize,
			}
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
