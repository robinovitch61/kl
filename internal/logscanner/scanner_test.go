package logscanner

import (
	"testing"

	"github.com/robinovitch61/kl/internal/domain"
	"github.com/robinovitch61/kl/internal/k8s"
	"github.com/robinovitch61/kl/internal/tree"
)

func TestNewCoordinator(t *testing.T) {
	manager, _ := k8s.NewManager("~/.kube/config", []string{"test"})
	timeRange := domain.NewTimeRange(1) // 1 minute

	coord := NewCoordinator(manager, timeRange)

	if coord == nil {
		t.Fatal("expected non-nil coordinator")
	}
}

func TestHandleStateChange_StartScanning(t *testing.T) {
	manager, _ := k8s.NewManager("~/.kube/config", []string{"test"})
	coord := NewCoordinator(manager, domain.NewTimeRange(1))

	change := tree.StateChange{
		ContainerID: domain.ContainerID{
			Cluster:   "test",
			Namespace: "default",
			Pod:       "pod1",
			Container: "c1",
		},
		FromState: domain.StateInactive,
		ToState:   domain.StateScannerStarting,
	}

	cmd := coord.HandleStateChange(change)

	// TODO: once implemented, should return a command that starts scanning
	_ = cmd
}

func TestHandleStateChange_StopScanning(t *testing.T) {
	manager, _ := k8s.NewManager("~/.kube/config", []string{"test"})
	coord := NewCoordinator(manager, domain.NewTimeRange(1))

	change := tree.StateChange{
		ContainerID: domain.ContainerID{
			Cluster:   "test",
			Namespace: "default",
			Pod:       "pod1",
			Container: "c1",
		},
		FromState: domain.StateScanning,
		ToState:   domain.StateScannerStopping,
	}

	cmd := coord.HandleStateChange(change)

	// TODO: once implemented, should return a command that stops scanning
	_ = cmd
}

func TestSetTimeRange(t *testing.T) {
	manager, _ := k8s.NewManager("~/.kube/config", []string{"test"})
	coord := NewCoordinator(manager, domain.NewTimeRange(1))

	newRange := domain.NewTimeRange(5) // 1 hour
	cmd := coord.SetTimeRange(newRange)

	// TODO: once implemented, should return commands to restart all scanners
	_ = cmd
}

func TestShutdown(t *testing.T) {
	manager, _ := k8s.NewManager("~/.kube/config", []string{"test"})
	coord := NewCoordinator(manager, domain.NewTimeRange(1))

	cmd := coord.Shutdown()

	// TODO: once implemented, should return a command that stops all scanners
	_ = cmd
}

func TestLogBatchMsg(t *testing.T) {
	msg := LogBatchMsg{
		ContainerID: domain.ContainerID{Container: "test"},
		Logs:        []domain.Log{{Content: "hello"}},
	}

	if len(msg.Logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(msg.Logs))
	}
}

func TestScannerStoppedMsg(t *testing.T) {
	msg := ScannerStoppedMsg{
		ContainerID: domain.ContainerID{Container: "test"},
		Reason:      ReasonUserDeselected,
	}

	if msg.Reason != ReasonUserDeselected {
		t.Errorf("expected ReasonUserDeselected, got %v", msg.Reason)
	}
}
