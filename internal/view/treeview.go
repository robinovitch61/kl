package view

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/bubbleo/filterableviewport"
	"github.com/robinovitch61/bubbleo/viewport"
	"github.com/robinovitch61/kl/internal/tree"
)

// TreeView displays the hierarchical container tree
type TreeView struct {
	viewport *filterableviewport.Model[tree.Node]
	tree     tree.Tree
	width    int
	height   int
}

// NewTreeView creates a new tree view
func NewTreeView(width, height int) TreeView {
	vp := viewport.New[tree.Node](width, height)
	fv := filterableviewport.New[tree.Node](vp)
	return TreeView{
		viewport: fv,
		tree:     tree.NewTree(),
		width:    width,
		height:   height,
	}
}

// Update handles messages and returns updated view and commands
func (v TreeView) Update(msg tea.Msg) (TreeView, tea.Cmd) {
	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)
	return v, cmd
}

// View renders the tree view
func (v TreeView) View() string {
	return v.viewport.View()
}

// SetTree updates the tree
func (v TreeView) SetTree(t tree.Tree) TreeView {
	v.tree = t
	v.viewport.SetObjects(t.Nodes())
	return v
}

// SetSize updates dimensions
func (v TreeView) SetSize(width, height int) TreeView {
	v.width = width
	v.height = height
	v.viewport.SetWidth(width)
	v.viewport.SetHeight(height)
	return v
}

// SelectedNode returns the currently selected node
func (v TreeView) SelectedNode() *tree.Node {
	return v.viewport.GetSelectedItem()
}
