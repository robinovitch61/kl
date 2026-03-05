package k8s_log

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
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
	prettyItems    []item.SingleItem // pretty-printed JSON lines, nil if not valid JSON or single item
	prettyComputed bool
}

// GetPrettyItems returns the pretty-printed JSON lines for this log, computing
// and caching the result on first access. Returns nil if the content is not
// multi-line JSON.
func (l *Log) GetPrettyItems() []item.SingleItem {
	if !l.prettyComputed {
		if lines := PrettyPrintJSON(l.ContentItem.Content()); len(lines) > 1 {
			l.prettyItems = make([]item.SingleItem, len(lines))
			for i, line := range lines {
				l.prettyItems[i] = item.NewItem(line)
			}
		}
		l.prettyComputed = true
	}
	return l.prettyItems
}

// PrettyPrintJSON attempts to pretty-print JSON input. Returns the input as-is if not valid JSON.
func PrettyPrintJSON(input string) []string {
	var raw map[string]interface{}

	err := json.Unmarshal([]byte(input), &raw)
	if err != nil {
		return []string{input}
	}

	var prettyJSON bytes.Buffer
	encoder := json.NewEncoder(&prettyJSON)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(raw)
	if err != nil {
		return []string{input}
	}

	lines := strings.Split(prettyJSON.String(), "\n")

	// remove trailing empty line if exists
	if len(lines) > 1 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	var result []string
	for i := range lines {
		if strings.Contains(lines[i], "\\n") || strings.Contains(lines[i], "\\t") {
			lines[i] = strings.ReplaceAll(lines[i], "\\t", "    ")
			parts := strings.Split(lines[i], "\\n")
			result = append(result, parts...)
		} else {
			result = append(result, lines[i])
		}
	}
	return result
}

type LogScanner struct {
	Container      container.Container
	LogChan        chan Log
	ErrChan        chan error
	cancel         context.CancelFunc
	uuid           string
	logLineScanner *bufio.Scanner
}

func NewLogScanner(ct container.Container, scanner *bufio.Scanner, cancelK8sStream context.CancelFunc) LogScanner {
	return LogScanner{
		Container:      ct,
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
