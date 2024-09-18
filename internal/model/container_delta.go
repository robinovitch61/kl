package model

import (
	"github.com/emirpasic/gods/trees/redblacktree"
	"time"
)

type ContainerDelta struct {
	Time      time.Time
	Container Container
	ToDelete  bool
	Selected  bool
}

// ContainerDeltaSet sorts ContainerDeltas by time, Container ID ascending
type ContainerDeltaSet struct {
	allDeltas *redblacktree.Tree
}

func (cds *ContainerDeltaSet) Add(delta ContainerDelta) {
	if cds.allDeltas == nil {
		cds.allDeltas = redblacktree.NewWithStringComparator()
	}
	key := delta.Time.String() + delta.Container.ID()
	cds.allDeltas.Put(key, delta)
}

func (cds ContainerDeltaSet) OrderedDeltas() []ContainerDelta {
	if cds.allDeltas == nil {
		return nil
	}
	var deltas []ContainerDelta
	for _, v := range cds.allDeltas.Values() {
		deltas = append(deltas, v.(ContainerDelta))
	}
	return deltas
}

func (cds *ContainerDeltaSet) Size() int {
	if cds.allDeltas == nil {
		return 0
	}
	return cds.allDeltas.Size()
}
