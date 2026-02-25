package client

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/robinovitch61/kl/internal/k8s/container"
)

func newTestListener() ContainerListener {
	ctx, cancel := context.WithCancel(context.Background())
	deltaChan := make(chan container.ContainerDelta, 100)
	return ContainerListener{
		Cluster:            "test-cluster",
		Namespace:          "test-namespace",
		containerDeltaChan: deltaChan,
		ctx:                ctx,
		Stop:               cancel,
	}
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
	listener := newTestListener()
	defer listener.Stop()

	go func() {
		listener.containerDeltaChan <- newTestDelta("container-1")
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
	listener := newTestListener()
	defer listener.Stop()

	listener.containerDeltaChan <- newTestDelta("container-1")
	listener.containerDeltaChan <- newTestDelta("container-2")
	listener.containerDeltaChan <- newTestDelta("container-3")

	deltaSet, err := listener.NextDeltaSet(time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deltaSet.Size() != 3 {
		t.Fatalf("expected 3 deltas, got %d", deltaSet.Size())
	}
}

func TestNextDeltaSet_ReturnsErrorOnStop(t *testing.T) {
	listener := newTestListener()

	go func() {
		listener.Stop()
	}()

	_, err := listener.NextDeltaSet(5 * time.Second)
	if err == nil {
		t.Fatal("expected error when listener is stopped, got nil")
	}
}

func TestSendOnStoppedListenerDoesNotPanic(t *testing.T) {
	listener := newTestListener()
	listener.Stop()

	// Simulate what the informer event handlers do: try to send on deltaChan
	// after context is cancelled. This should not panic.
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			delta := newTestDelta("container")
			select {
			case listener.containerDeltaChan <- delta:
			case <-listener.ctx.Done():
				return
			}
		}()
	}
	wg.Wait()
}

func TestStopIsIdempotent(t *testing.T) {
	listener := newTestListener()
	// calling Stop multiple times should not panic (unlike close(chan))
	listener.Stop()
	listener.Stop()
	listener.Stop()
}
