package view

import (
	"testing"

	"github.com/robinovitch61/kl/internal/tree"
)

func TestNewTreeView(t *testing.T) {
	tv := NewTreeView(80, 24)

	if tv.width != 80 {
		t.Errorf("expected width 80, got %d", tv.width)
	}
	if tv.height != 24 {
		t.Errorf("expected height 24, got %d", tv.height)
	}
}

func TestTreeView_SetTree(t *testing.T) {
	tv := NewTreeView(80, 24)

	tr := tree.NewTree()
	tv = tv.SetTree(tr)

	// Basic check that SetTree doesn't panic
	if tv.viewport == nil {
		t.Error("expected viewport to be initialized")
	}
}

func TestTreeView_SetSize(t *testing.T) {
	tv := NewTreeView(80, 24)

	tv = tv.SetSize(120, 40)

	if tv.width != 120 {
		t.Errorf("expected width 120, got %d", tv.width)
	}
	if tv.height != 40 {
		t.Errorf("expected height 40, got %d", tv.height)
	}
}

func TestTreeView_SelectedNode_Empty(t *testing.T) {
	tv := NewTreeView(80, 24)

	node := tv.SelectedNode()
	if node != nil {
		t.Error("expected nil for empty tree")
	}
}
