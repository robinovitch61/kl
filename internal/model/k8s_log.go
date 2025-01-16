package model

import (
	"bufio"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/viewport/linebuffer"
	"strings"
	"time"
)

type Log struct {
	Timestamp  time.Time
	LineBuffer linebuffer.LineBuffer
	Container  Container
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
				Timestamp:  parsedTime,
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

func (ls LogScanner) Cancel() {
	if ls.cancel != nil {
		ls.cancel()
	}
}

func (ls LogScanner) Equals(other LogScanner) bool {
	return ls.uuid == other.uuid
}
