package client_test

import (
	"context"
	"testing"
	"time"

	"github.com/robinovitch61/kl/internal/k8s/client"
	"github.com/robinovitch61/kl/internal/k8s/container"
)

func newTestListener() (client.ContainerListener, chan container.ContainerDelta) {
	ctx, cancel := context.WithCancel(context.Background())
	deltaChan := make(chan container.ContainerDelta, 100)
	listener := client.NewContainerListener(ctx, "test-cluster", "test-namespace", deltaChan, cancel)
	return listener, deltaChan
}

func newTestDelta(name string) container.ContainerDelta {
	return container.ContainerDelta{
		Time: time.Now(),
		Container: container.Container{
			Cluster:   "test-cluster",
			Namespace: "test-namespace",
			PodOwner:  "test-owner",
			Pod:       "test-pod",
			Name:      name,
		},
	}
}

func TestNextDeltaSet_BlocksUntilFirstDelta(t *testing.T) {
	listener, deltaChan := newTestListener()
	defer listener.Stop()

	go func() {
		deltaChan <- newTestDelta("container-1")
	}()

	deltaSet, err := listener.NextDeltaSet(time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deltaSet.Size() != 1 {
		t.Fatalf("expected 1 delta, got %d", deltaSet.Size())
	}
}

func TestNextDeltaSet_BatchesDeltas(t *testing.T) {
	listener, deltaChan := newTestListener()
	defer listener.Stop()

	deltaChan <- newTestDelta("container-1")
	deltaChan <- newTestDelta("container-2")
	deltaChan <- newTestDelta("container-3")

	deltaSet, err := listener.NextDeltaSet(time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deltaSet.Size() != 3 {
		t.Fatalf("expected 3 deltas, got %d", deltaSet.Size())
	}
}

func TestNextDeltaSet_ReturnsErrorOnStop(t *testing.T) {
	listener, _ := newTestListener()

	go func() {
		listener.Stop()
	}()

	_, err := listener.NextDeltaSet(5 * time.Second)
	if err == nil {
		t.Fatal("expected error when listener is stopped, got nil")
	}
}

func TestStopIsIdempotent(t *testing.T) {
	listener, _ := newTestListener()
	// calling Stop multiple times should not panic (unlike close(chan))
	listener.Stop()
	listener.Stop()
	listener.Stop()
}
