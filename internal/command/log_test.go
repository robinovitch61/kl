package command_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/robinovitch61/kl/internal/command"
	"github.com/robinovitch61/kl/internal/k8s/k8s_log"
)

func newTestLogScanner() k8s_log.LogScanner {
	return k8s_log.LogScanner{
		LogChan: make(chan k8s_log.Log, 10),
		ErrChan: make(chan error, 1),
	}
}

func newTestLog(ts time.Time) k8s_log.Log {
	return k8s_log.Log{Timestamp: ts}
}

func TestGetNextLogsCmd_ReceivesLogs(t *testing.T) {
	ls := newTestLogScanner()
	now := time.Now()
	ls.LogChan <- newTestLog(now)
	ls.LogChan <- newTestLog(now.Add(time.Second))

	cmd := command.GetNextLogsCmd(ls, 10*time.Millisecond)
	msg := cmd().(command.GetNewLogsMsg)

	if msg.Err != nil {
		t.Fatalf("unexpected error: %v", msg.Err)
	}
	if len(msg.NewLogs) != 2 {
		t.Fatalf("expected 2 logs, got %d", len(msg.NewLogs))
	}
	if msg.DoneScanning {
		t.Error("expected DoneScanning to be false")
	}
}

func TestGetNextLogsCmd_Error(t *testing.T) {
	ls := newTestLogScanner()
	now := time.Now()
	ls.LogChan <- newTestLog(now)
	ls.ErrChan <- fmt.Errorf("connection reset by peer")

	cmd := command.GetNextLogsCmd(ls, time.Second)
	msg := cmd().(command.GetNewLogsMsg)

	// the cmd may return the log batch first or the error first depending on select ordering;
	// if it returns logs first, call cmd again to get the error
	if msg.Err == nil {
		cmd = command.GetNextLogsCmd(ls, time.Second)
		msg = cmd().(command.GetNewLogsMsg)
	}

	if msg.Err == nil {
		t.Fatal("expected error")
	}
	if msg.Err.Error() != "connection reset by peer" {
		t.Errorf("expected 'connection reset by peer', got %q", msg.Err.Error())
	}
}

func TestGetNextLogsCmd_ChannelClosed(t *testing.T) {
	ls := newTestLogScanner()
	now := time.Now()
	ls.LogChan <- newTestLog(now)
	close(ls.LogChan)

	cmd := command.GetNextLogsCmd(ls, time.Second)
	msg := cmd().(command.GetNewLogsMsg)

	if msg.Err != nil {
		t.Fatalf("unexpected error: %v", msg.Err)
	}
	if !msg.DoneScanning {
		t.Error("expected DoneScanning to be true when channel closed")
	}
	if len(msg.NewLogs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(msg.NewLogs))
	}
}

func TestGetNextLogsCmd_ErrorWithNoLogs(t *testing.T) {
	ls := newTestLogScanner()
	ls.ErrChan <- fmt.Errorf("stream error")

	cmd := command.GetNextLogsCmd(ls, time.Second)
	msg := cmd().(command.GetNewLogsMsg)

	if msg.Err == nil {
		t.Fatal("expected error")
	}
	if len(msg.NewLogs) != 0 {
		t.Fatalf("expected 0 logs, got %d", len(msg.NewLogs))
	}
}
