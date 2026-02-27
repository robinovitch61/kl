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
	"github.com/robinovitch61/viewport/viewport/item"
)

type LogTimestamps struct {
	Short string
	Full  string
}

type Log struct {
	Timestamp   time.Time
	Timestamps  LogTimestamps
	Container   container.Container
	ContentItem item.SingleItem
	PrettyItems []item.SingleItem // pretty-printed JSON lines, nil if not valid JSON or single item
}

// FormatJSON attempts to pretty-print JSON input. Returns the input as-is if not valid JSON.
func FormatJSON(input string) []string {
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
			logContent = sanitizeTerminalSequences(logContent)

			localTime := parsedTime.Local()

			// precompute LogData here as logs come in as logs are immutable and instantiating new items is expensive
			contentItem := item.NewItem(logContent)
			var prettyItems []item.SingleItem
			if lines := FormatJSON(logContent); len(lines) > 1 {
				prettyItems = make([]item.SingleItem, len(lines))
				for i, line := range lines {
					prettyItems[i] = item.NewItem(line)
				}
			}

			ls.LogChan <- Log{
				Timestamp: parsedTime,
				Timestamps: LogTimestamps{
					Short: localTime.Format(time.TimeOnly),
					Full:  localTime.Format("2006-01-02T15:04:05.000Z07:00"),
				},
				Container:   ls.Container,
				ContentItem: contentItem,
				PrettyItems: prettyItems,
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

// sanitizeTerminalSequences removes terminal control sequences that could affect
// the terminal (cursor movement, screen clearing, etc.) while preserving ANSI
// styling sequences (SGR: colors, bold, underline, etc.).
func sanitizeTerminalSequences(s string) string {
	var buf strings.Builder
	buf.Grow(len(s))
	i := 0
	for i < len(s) {
		b := s[i]

		if b == '\x1b' {
			if i+1 >= len(s) {
				// lone ESC at end
				i++
				continue
			}
			next := s[i+1]

			// CSI sequence: ESC [
			if next == '[' {
				j := i + 2
				// skip parameter bytes (0x30-0x3F)
				for j < len(s) && s[j] >= 0x30 && s[j] <= 0x3F {
					j++
				}
				// skip intermediate bytes (0x20-0x2F)
				for j < len(s) && s[j] >= 0x20 && s[j] <= 0x2F {
					j++
				}
				// final byte (0x40-0x7E)
				if j < len(s) && s[j] >= 0x40 && s[j] <= 0x7E {
					if s[j] == 'm' {
						// SGR sequence - keep styling
						buf.WriteString(s[i : j+1])
					}
					i = j + 1
					continue
				}
				// incomplete CSI - skip ESC
				i++
				continue
			}

			// OSC sequence: ESC ] ... (BEL or ST)
			if next == ']' {
				j := i + 2
				for j < len(s) {
					if s[j] == '\x07' {
						j++
						break
					}
					if s[j] == '\x1b' && j+1 < len(s) && s[j+1] == '\\' {
						j += 2
						break
					}
					j++
				}
				i = j
				continue
			}

			// DCS, PM, APC sequences: ESC P, ESC ^, ESC _
			if next == 'P' || next == '^' || next == '_' {
				j := i + 2
				for j < len(s) {
					if s[j] == '\x1b' && j+1 < len(s) && s[j+1] == '\\' {
						j += 2
						break
					}
					j++
				}
				i = j
				continue
			}

			// other single-char escape sequences (e.g. ESC c = reset) - skip
			i += 2
			continue
		}

		// remove C0 control chars except tab (already converted to spaces above)
		if b < 0x20 && b != '\t' {
			i++
			continue
		}
		// remove DEL
		if b == 0x7F {
			i++
			continue
		}

		buf.WriteByte(b)
		i++
	}
	return buf.String()
}

func (ls LogScanner) Cancel() {
	if ls.cancel != nil {
		ls.cancel()
	}
}

func (ls LogScanner) Equals(other LogScanner) bool {
	return ls.uuid == other.uuid
}
