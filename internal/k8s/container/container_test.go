package container

import (
	"testing"
	"time"
)

func TestContainerDeltaSet(t *testing.T) {
	t.Run("Add and Size", func(t *testing.T) {
		cds := &ContainerDeltaSet{}
		if cds.Size() != 0 {
			t.Errorf("Expected initial size to be 0, got %d", cds.Size())
		}

		delta1 := ContainerDelta{
			Time: time.Now(),
			Container: Container{
				Cluster: "cluster1",
				Name:    "container1",
			},
		}
		cds.Add(delta1)
		if cds.Size() != 1 {
			t.Errorf("Expected size to be 1 after adding one delta, got %d", cds.Size())
		}

		delta2 := ContainerDelta{
			Time: time.Now().Add(time.Hour),
			Container: Container{
				Cluster: "cluster1",
				Name:    "container2",
			},
		}
		cds.Add(delta2)
		if cds.Size() != 2 {
			t.Errorf("Expected size to be 2 after adding two deltas, got %d", cds.Size())
		}
	})

	t.Run("OrderedDeltas", func(t *testing.T) {
		cds := &ContainerDeltaSet{}
		now := time.Now()

		delta1 := ContainerDelta{
			Time: now,
			Container: Container{
				Cluster: "cluster1",
				Name:    "container1",
			},
		}
		delta2 := ContainerDelta{
			Time: now.Add(time.Hour),
			Container: Container{
				Cluster: "cluster1",
				Name:    "container2",
			},
		}
		delta3 := ContainerDelta{
			Time: now.Add(30 * time.Minute),
			Container: Container{
				Cluster: "cluster1",
				Name:    "container3",
			},
		}
		delta4 := ContainerDelta{
			Time: now.Add(30 * time.Minute),
			Container: Container{
				Cluster: "cluster1",
				Name:    "container4",
			},
		}

		cds.Add(delta1)
		cds.Add(delta2)
		cds.Add(delta3)
		cds.Add(delta4)

		orderedDeltas := cds.OrderedDeltas()
		if len(orderedDeltas) != 4 {
			t.Errorf("Expected 4 deltas, got %d", len(orderedDeltas))
		}

		if !deltasEqual(orderedDeltas[0], delta1) {
			t.Errorf("Expected first delta to be %v, got %v", delta1, orderedDeltas[0])
		}
		if !deltasEqual(orderedDeltas[1], delta3) {
			t.Errorf("Expected second delta to be %v, got %v", delta3, orderedDeltas[1])
		}
		if !deltasEqual(orderedDeltas[2], delta4) {
			t.Errorf("Expected third delta to be %v, got %v", delta4, orderedDeltas[2])
		}
		if !deltasEqual(orderedDeltas[3], delta2) {
			t.Errorf("Expected fourth delta to be %v, got %v", delta2, orderedDeltas[3])
		}
	})

	t.Run("Empty ContainerDeltaSet", func(t *testing.T) {
		cds := &ContainerDeltaSet{}
		if cds.Size() != 0 {
			t.Errorf("Expected size of empty set to be 0, got %d", cds.Size())
		}
		if len(cds.OrderedDeltas()) != 0 {
			t.Errorf("Expected empty set to have no deltas, got %d", len(cds.OrderedDeltas()))
		}
	})
}

func deltasEqual(a, b ContainerDelta) bool {
	return a.Time.Equal(b.Time) &&
		a.Container.Equals(b.Container) &&
		a.ToDelete == b.ToDelete
}
