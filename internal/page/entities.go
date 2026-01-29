package page

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/robinovitch61/bubbleo/filterableviewport"
	"github.com/robinovitch61/bubbleo/viewport"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/help"
	"github.com/robinovitch61/kl/internal/k8s/entity"
	"github.com/robinovitch61/kl/internal/keymap"
	"github.com/robinovitch61/kl/internal/style"
)

type EntityPage struct {
	filterableViewport *filterableviewport.Model[entity.Entity]
	viewport           *viewport.Model[entity.Entity]
	entityTree         entity.Tree
	keyMap             keymap.KeyMap
	styles             style.Styles
	focused            bool
}

// assert EntityPage implements GenericPage
var _ GenericPage = EntityPage{}

func NewEntitiesPage(
	keyMap keymap.KeyMap,
	width, height int,
	entityTree entity.Tree,
	styles style.Styles,
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

	vp := viewport.New[entity.Entity](
		width,
		height-1, // -1 for filter line
		viewport.WithSelectionEnabled[entity.Entity](true),
		viewport.WithWrapText[entity.Entity](false),
		viewport.WithFooterEnabled[entity.Entity](true),
	)

	// Set up selection comparator to maintain selection when content changes
	vp.SetSelectionComparator(func(a, b entity.Entity) bool {
		return a.EqualTo(b)
	})

	fvp := filterableviewport.New[entity.Entity](
		vp,
		filterableviewport.WithPrefixText[entity.Entity]("(S)election"),
		filterableviewport.WithEmptyText[entity.Entity]("'/' or 'r' to filter"),
		filterableviewport.WithMatchingItemsOnly[entity.Entity](true),
		filterableviewport.WithCanToggleMatchingItemsOnly[entity.Entity](false),
	)

	// Set header to show subscription info when viewport is empty
	if len(entityTree.GetEntities()) == 0 {
		vp.SetHeader(strings.Split(viewWhenEmpty, "\n"))
	}

	// Update tree prefixes before setting objects
	entityTree.UpdatePrettyPrintPrefixes()
	fvp.SetObjects(entityTree.GetEntities())

	return EntityPage{
		filterableViewport: fvp,
		viewport:           vp,
		entityTree:         entityTree,
		keyMap:             keyMap,
		styles:             styles,
	}
}

func (p EntityPage) Update(msg tea.Msg) (GenericPage, tea.Cmd) {
	dev.DebugUpdateMsg("EntityPage", msg)
	var cmd tea.Cmd

	p.filterableViewport, cmd = p.filterableViewport.Update(msg)
	return p, cmd
}

func (p EntityPage) View() string {
	return p.filterableViewport.View()
}

func (p EntityPage) HighjackingInput() bool {
	return p.filterableViewport.IsCapturingInput()
}

func (p EntityPage) ContentForFile() []string {
	var content []string
	for _, e := range p.entityTree.GetEntities() {
		content = append(content, e.Repr())
	}
	return content
}

func (p EntityPage) HasAppliedFilter() bool {
	return p.filterableViewport.FilterFocused()
}

func (p EntityPage) ToggleShowContext() GenericPage {
	// In bubbleo, this is handled by the 'o' key (ToggleMatchingItemsOnlyKey)
	return p
}

func (p EntityPage) WithDimensions(width, height int) GenericPage {
	p.filterableViewport.SetWidth(width)
	p.filterableViewport.SetHeight(height)
	return p
}

func (p EntityPage) WithFocus() GenericPage {
	p.focused = true
	return p
}

func (p EntityPage) WithBlur() GenericPage {
	p.focused = false
	return p
}

func (p EntityPage) WithStyles(styles style.Styles) GenericPage {
	p.styles = styles
	return p
}

func (p EntityPage) Help() string {
	return help.MakeHelp(p.keyMap, p.styles.InverseUnderline)
}

func (p EntityPage) WithEntityTree(entityTree entity.Tree) EntityPage {
	p.entityTree = entityTree
	p.entityTree.UpdatePrettyPrintPrefixes()
	entities := p.entityTree.GetEntities()
	p.filterableViewport.SetObjects(entities)
	// Clear header once we have entities
	if len(entities) > 0 {
		p.viewport.SetHeader(nil)
	}
	return p
}

func (p EntityPage) WithMaintainSelection(maintainSelection bool) EntityPage {
	if maintainSelection {
		p.viewport.SetSelectionComparator(func(a, b entity.Entity) bool {
			return a.EqualTo(b)
		})
	} else {
		p.viewport.SetSelectionComparator(nil)
	}
	return p
}

func (p EntityPage) GetSelectionActions() (entity.Entity, map[entity.Entity]bool) {
	selectedEntity := p.viewport.GetSelectedItem()
	if selectedEntity == nil {
		return entity.Entity{}, nil
	}
	return *selectedEntity, p.entityTree.GetSelectionActions(*selectedEntity)
}
