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
	"github.com/robinovitch61/kl/internal/viewport/linebuffer"
)

type LogTimestamps struct {
	Short string
	Full  string
}

type Log struct {
	Timestamp  time.Time
	Timestamps LogTimestamps
	LineBuffer linebuffer.LineBuffer
	Container  container.Container
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

			// precompute LogData here as logs come in as logs are immutable. Having the LogData up front helps
			// to minimize expensive/repeated re-computation later, particularly making new line buffers
			newLog := Log{
				Timestamp: parsedTime,
				Timestamps: LogTimestamps{
					Short: localTime.Format(time.TimeOnly),
					Full:  localTime.Format("2006-01-02T15:04:05.000Z07:00"),
				},
				LineBuffer: linebuffer.New(logContent),
				Container:  ls.Container,
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
