package tree

import "github.com/robinovitch61/kl/internal/domain"

// StateChange represents a container state transition request
type StateChange struct {
	ContainerID domain.ContainerID
	FromState   domain.ContainerState
	ToState     domain.ContainerState
}

// Tree manages the hierarchical container structure
type Tree struct {
	nodes []Node
}

// NewTree creates a new empty tree
func NewTree() Tree {
	return Tree{}
}

// Update applies container changes and returns a new tree
func (t Tree) Update(containers []domain.SelectableContainer) Tree {
	// TODO: implement - build hierarchical tree from flat container list
	return Tree{}
}

// Nodes returns the flattened list for viewport display
func (t Tree) Nodes() []Node {
	return t.nodes
}

// ToggleSelection toggles the container at index and returns new tree + changes
func (t Tree) ToggleSelection(idx int) (Tree, []StateChange) {
	// TODO: implement
	return t, nil
}

// DeselectAll deselects all containers and returns new tree + changes
func (t Tree) DeselectAll() (Tree, []StateChange) {
	// TODO: implement
	return t, nil
}
