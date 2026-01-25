package tree

import (
	"github.com/robinovitch61/bubbleo/viewport/item"
	"github.com/robinovitch61/kl/internal/domain"
)

// Kind represents the type of node in the tree hierarchy
type Kind int

const (
	KindCluster Kind = iota
	KindNamespace
	KindOwner
	KindPod
	KindContainer
)

// Node is a single row in the tree
type Node struct {
	kind      Kind
	depth     int
	label     string
	container *domain.SelectableContainer
}

// GetItem implements viewport.Object
func (n Node) GetItem() item.Item {
	// TODO: implement
	return item.NewItem(n.label)
}

// IsContainer returns true if this node represents a container
func (n Node) IsContainer() bool {
	return n.kind == KindContainer
}

// Container returns the underlying container, or nil if not a container node
func (n Node) Container() *domain.SelectableContainer {
	return n.container
}
