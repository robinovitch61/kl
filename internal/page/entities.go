package page

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/filterable_viewport"
	"github.com/robinovitch61/kl/internal/keymap"
	"github.com/robinovitch61/kl/internal/model"
	"strings"
)

type EntityPage struct {
	filterableViewport filterable_viewport.FilterableViewport[model.Entity]
	entityTree         model.EntityTree
	keyMap             keymap.KeyMap
}

// assert EntityPage implements GenericPage
var _ GenericPage = EntityPage{}

func NewEntitiesPage(
	keyMap keymap.KeyMap,
	width, height int,
	entityTree model.EntityTree,
) EntityPage {
	viewWhenEmptyLines := []string{"Subscribing to updates for:"}
	for _, cns := range entityTree.GetClusterNamespaces() {
		viewWhenEmptyLines = append(viewWhenEmptyLines, fmt.Sprintf("- Cluster %s", cns.Cluster))
		if len(cns.Namespaces) == 1 && cns.Namespaces[0] == "" {
			viewWhenEmptyLines = append(viewWhenEmptyLines, "  * All Namespaces")
		} else {
			for _, n := range cns.Namespaces {
				viewWhenEmptyLines = append(viewWhenEmptyLines, fmt.Sprintf("  * Namespace %s", n))
			}
		}
	}
	viewWhenEmpty := strings.Join(viewWhenEmptyLines, "\n")

	filterableViewport := filterable_viewport.NewFilterableViewport[model.Entity](
		"(S)election",
		false,
		true,
		false,
		keyMap,
		width,
		height,
		entityTree.GetEntities(),
		entityTree.IsVisibleGivenFilter,
		viewWhenEmpty,
	)
	return EntityPage{
		filterableViewport: filterableViewport,
		entityTree:         entityTree,
		keyMap:             keyMap,
	}
}

func (p EntityPage) Update(msg tea.Msg) (GenericPage, tea.Cmd) {
	dev.DebugUpdateMsg("EntityPage", msg)
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	prevFilterValue := p.filterableViewport.Filter.Value()
	p.filterableViewport, cmd = p.filterableViewport.Update(msg)
	cmds = append(cmds, cmd)

	// if filter has changed, also need to update the entity tree's prefixes
	if prevFilterValue != p.filterableViewport.Filter.Value() {
		p.entityTree.UpdatePrettyPrintPrefixes(p.filterableViewport.Filter)
		p.filterableViewport.SetAllRows(p.entityTree.GetEntities())
	}
	return p, tea.Batch(cmds...)
}

func (p EntityPage) View() string {
	return p.filterableViewport.View()
}

func (p EntityPage) HighjackingInput() bool {
	return p.filterableViewport.HighjackingInput()
}

func (p EntityPage) ContentForFile() []string {
	var content []string
	for _, l := range p.getVisibleEntities() {
		content = append(content, l.Render())
	}
	return content
}

func (p EntityPage) WithDimensions(width, height int) GenericPage {
	p.filterableViewport = p.filterableViewport.WithDimensions(width, height)
	return p
}

func (p EntityPage) WithFocus() GenericPage {
	p.filterableViewport.SetFocus(true, true)
	return p
}

func (p EntityPage) WithBlur() GenericPage {
	p.filterableViewport.SetFocus(false, false)
	return p
}

func (p EntityPage) Help() string {
	local := []key.Binding{
		keymap.WithDesc(p.keyMap.Enter, "select/deselect"),
		p.keyMap.TogglePause,
	}
	local = append(local, keymap.LookbackKeyBindings(p.keyMap)...)
	return makePageHelp(
		"Selection",
		keymap.GlobalKeyBindings(p.keyMap),
		local,
	)
}

func (p EntityPage) WithEntityTree(entityTree model.EntityTree) EntityPage {
	p.entityTree = entityTree
	p.entityTree.UpdatePrettyPrintPrefixes(p.filterableViewport.Filter)
	p.filterableViewport.SetAllRowsAndMatchesFilter(p.entityTree.GetEntities(), p.entityTree.IsVisibleGivenFilter)
	return p
}

func (p EntityPage) WithMaintainSelection(maintainSelection bool) EntityPage {
	p.filterableViewport.SetMaintainSelection(maintainSelection)
	return p
}

func (p EntityPage) GetSelectionActions() (model.Entity, map[model.Entity]bool) {
	selectedEntity := p.filterableViewport.GetSelection()
	if selectedEntity == nil {
		return model.Entity{}, nil
	}
	return *selectedEntity, p.entityTree.GetSelectionActions(*selectedEntity, p.filterableViewport.Filter)
}

func (p EntityPage) getVisibleEntities() []model.Entity {
	if p.filterableViewport.Filter.FilteringWithContext {
		return p.entityTree.GetEntities()
	}
	return p.entityTree.GetVisibleEntities(p.filterableViewport.Filter)
}
