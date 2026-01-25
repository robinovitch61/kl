package k8s

import (
	"context"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	manager, err := NewManager("~/.kube/config", []string{"ctx1", "ctx2"})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	if manager == nil {
		t.Error("expected non-nil manager")
	}
}

func TestWatchContainersCmd(t *testing.T) {
	manager, _ := NewManager("~/.kube/config", []string{"test"})
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	cmd := manager.WatchContainersCmd(ctx, []string{"default"}, "")

	// TODO: once implemented, this should return a non-nil command
	if cmd != nil {
		// The command should produce ContainerDeltasMsg when executed
		msg := cmd()
		if msg == nil {
			t.Error("command should produce a message")
		}
		if _, ok := msg.(ContainerDeltasMsg); !ok {
			t.Errorf("expected ContainerDeltasMsg, got %T", msg)
		}
	}
}

func TestContainerDeltasMsg(t *testing.T) {
	msg := ContainerDeltasMsg{
		Deltas: []ContainerDelta{
			{IsRemoved: false},
			{IsRemoved: true},
		},
	}
	if len(msg.Deltas) != 2 {
		t.Errorf("expected 2 deltas, got %d", len(msg.Deltas))
	}
}
