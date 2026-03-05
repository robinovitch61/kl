package container_test

import (
	"testing"
	"time"

	"github.com/robinovitch61/kl/internal/k8s/container"
)

func TestContainerDeltaSetAdd(t *testing.T) {
	cds := container.ContainerDeltaSet{}

	if cds.Size() != 0 {
		t.Errorf("Expected size 0 for empty set, got %d", cds.Size())
	}

	now := time.Now()

	delta1 := container.ContainerDelta{
		Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Container: container.Container{
			Cluster: "cluster1", Namespace: "ns1", PodOwner: "dep1", Pod: "pod1", Name: "container1",
			Status: container.ContainerStatus{
				State:     container.ContainerRunning,
				StartedAt: now,
			},
		},
		ToDelete: false,
	}
	cds.Add(delta1)

	if cds.Size() != 1 {
		t.Errorf("Expected size 1, got %d", cds.Size())
	}

	delta2 := container.ContainerDelta{
		Time: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		Container: container.Container{
			Cluster: "cluster1", Namespace: "ns1", PodOwner: "dep1", Pod: "pod1", Name: "container2",
			Status: container.ContainerStatus{
				State:     container.ContainerTerminated,
				StartedAt: now.Add(-1 * time.Hour),
			},
		},
		ToDelete: true,
	}
	cds.Add(delta2)

	if cds.Size() != 2 {
		t.Errorf("Expected size 2, got %d", cds.Size())
	}
}

func TestContainerDeltaSetOrderedDeltas(t *testing.T) {
	cds := container.ContainerDeltaSet{}

	now := time.Now()

	delta1 := container.ContainerDelta{
		Time: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		Container: container.Container{
			Cluster: "cluster1", Namespace: "ns1", PodOwner: "dep1", Pod: "pod1", Name: "container1",
			Status: container.ContainerStatus{
				State:     container.ContainerRunning,
				StartedAt: now,
			},
		},
		ToDelete: false,
	}
	delta2 := container.ContainerDelta{
		Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Container: container.Container{
			Cluster: "cluster1", Namespace: "ns1", PodOwner: "dep1", Pod: "pod1", Name: "container1",
			Status: container.ContainerStatus{
				State:     container.ContainerUnknown,
				StartedAt: time.Time{},
			},
		},
		ToDelete: false,
	}
	delta3 := container.ContainerDelta{
		Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Container: container.Container{
			Cluster: "cluster1", Namespace: "ns1", PodOwner: "dep1", Pod: "pod1", Name: "container2",
			Status: container.ContainerStatus{
				State:     container.ContainerWaiting,
				StartedAt: time.Time{},
			},
		},
		ToDelete: true,
	}

	cds.Add(delta1)
	cds.Add(delta2)
	cds.Add(delta3)

	orderedDeltas := cds.OrderedDeltas()

	if len(orderedDeltas) != 3 {
		t.Errorf("Expected 3 deltas, got %d", len(orderedDeltas))
	}

	if orderedDeltas[0].Container.Name != "container1" || orderedDeltas[0].Time != delta2.Time {
		t.Errorf("First delta not ordered correctly")
	}
	if orderedDeltas[1].Container.Name != "container2" || orderedDeltas[1].Time != delta3.Time {
		t.Errorf("Second delta not ordered correctly")
	}
	if orderedDeltas[2].Container.Name != "container1" || orderedDeltas[2].Time != delta1.Time {
		t.Errorf("Third delta not ordered correctly")
	}
}

func TestContainerDeltaSetEmptyOrderedDeltas(t *testing.T) {
	cds := container.ContainerDeltaSet{}

	orderedDeltas := cds.OrderedDeltas()

	if orderedDeltas != nil {
		t.Errorf("Expected nil for empty set, got %v", orderedDeltas)
	}
}
