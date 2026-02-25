package page

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/filter"
	"github.com/robinovitch61/kl/internal/help"
	"github.com/robinovitch61/kl/internal/k8s/entity"
	"github.com/robinovitch61/kl/internal/keymap"
	"github.com/robinovitch61/kl/internal/style"
	"github.com/robinovitch61/viewport/filterableviewport"
	"github.com/robinovitch61/viewport/viewport"
)

type EntityPage struct {
	filterableViewport *filterableviewport.Model[entity.Entity]
	entityTree         entity.Tree
	keyMap             keymap.KeyMap
	styles             style.Styles
	focused            bool
	viewWhenEmpty      string
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

	vp := viewport.New[entity.Entity](width, height,
		viewport.WithKeyMap[entity.Entity](viewport.KeyMap{
			PageDown:     keyMap.PageDown,
			PageUp:       keyMap.PageUp,
			HalfPageUp:   keyMap.HalfPageUp,
			HalfPageDown: keyMap.HalfPageDown,
			Up:           keyMap.Up,
			Down:         keyMap.Down,
			Left:         keyMap.Left,
			Right:        keyMap.Right,
			Top:          keyMap.Top,
			Bottom:       keyMap.Bottom,
		}),
	)
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(false)

	fvp := filterableviewport.New(vp,
		filterableviewport.WithKeyMap[entity.Entity](filterableviewport.KeyMap{
			FilterKey:                  keyMap.Filter,
			RegexFilterKey:             keyMap.FilterRegex,
			CaseInsensitiveFilterKey:   keyMap.FilterCaseInsensitive,
			ApplyFilterKey:             keyMap.Enter,
			CancelFilterKey:            keyMap.Clear,
			ToggleMatchingItemsOnlyKey: keyMap.Context,
			NextMatchKey:               keyMap.FilterNextRow,
			PrevMatchKey:               keyMap.FilterPrevRow,
		}),
		filterableviewport.WithMatchingItemsOnly[entity.Entity](false),
		filterableviewport.WithCanToggleMatchingItemsOnly[entity.Entity](false),
		filterableviewport.WithEmptyText[entity.Entity]("'/', 'r', or 'i' to filter"),
		filterableviewport.WithFilterLinePosition[entity.Entity](filterableviewport.FilterLineTop),
		filterableviewport.WithFilterLinePrefix[entity.Entity]("(S)election"),
		filterableviewport.WithStyles[entity.Entity](filterableviewport.Styles{
			Match: filterableviewport.MatchStyles{
				Focused:   styles.Inverse,
				Unfocused: styles.AltInverse,
			},
		}),
		filterableviewport.WithAdjustObjectsForFilter(func(filterText string, isRegex bool) []entity.Entity {
			f := makeFilterFromText(filterText, isRegex, keyMap)
			entityTree.UpdatePrettyPrintPrefixes(f)
			return entityTree.GetVisibleEntities(f)
		}),
	)

	fvp.SetObjects(entityTree.GetEntities())

	p := EntityPage{
		filterableViewport: fvp,
		entityTree:         entityTree,
		keyMap:             keyMap,
		styles:             styles,
		viewWhenEmpty:      viewWhenEmpty,
	}
	p.updateStyles()

	return p
}

func (p EntityPage) Update(msg tea.Msg) (GenericPage, tea.Cmd) {
	dev.DebugUpdateMsg("EntityPage", msg)
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !p.HighjackingInput() {
			if key.Matches(msg, p.keyMap.Wrap) {
				p.filterableViewport.SetWrapText(!p.filterableViewport.GetWrapText())
				return p, nil
			}
		}
	}

	p.filterableViewport, cmd = p.filterableViewport.Update(msg)

	return p, cmd
}

func (p EntityPage) View() string {
	if len(p.entityTree.GetEntities()) == 0 {
		viewWhenEmpty := p.viewWhenEmpty
		if p.focused {
			viewWhenEmpty = p.styles.Blue.Render(viewWhenEmpty)
		}
		return viewWhenEmpty
	}
	return p.filterableViewport.View()
}

func (p EntityPage) HighjackingInput() bool {
	return p.filterableViewport.IsCapturingInput()
}

func (p EntityPage) ContentForFile() []string {
	var content []string
	for _, l := range p.getVisibleEntities() {
		content = append(content, l.Repr())
	}
	return content
}

func (p EntityPage) HasAppliedFilter() bool {
	return p.filterableViewport.GetFilterText() != ""
}

func (p EntityPage) ToggleShowContext() GenericPage {
	// EntityPage doesn't support show context toggle
	return p
}

func (p EntityPage) WithDimensions(width, height int) GenericPage {
	p.filterableViewport.SetWidth(width)
	p.filterableViewport.SetHeight(height)
	return p
}

func (p EntityPage) WithFocus() GenericPage {
	p.focused = true
	p.updateStyles()
	return p
}

func (p EntityPage) WithBlur() GenericPage {
	p.focused = false
	p.updateStyles()
	return p
}

func (p EntityPage) WithStyles(styles style.Styles) GenericPage {
	p.styles = styles
	p.updateStyles()
	// Update filterableviewport match styles
	p.filterableViewport.SetFilterableViewportStyles(filterableviewport.Styles{
		Match: filterableviewport.MatchStyles{
			Focused:   styles.Inverse,
			Unfocused: styles.AltInverse,
		},
	})
	return p
}

func (p EntityPage) Help() string {
	return help.MakeHelp(p.keyMap, p.styles.InverseUnderline)
}

func (p EntityPage) WithEntityTree(entityTree entity.Tree) EntityPage {
	p.entityTree = entityTree
	f := p.getCurrentFilter()
	p.entityTree.UpdatePrettyPrintPrefixes(f)
	p.filterableViewport.SetObjects(p.entityTree.GetVisibleEntities(f))
	p.filterableViewport.SetSelectionComparator(func(a, b entity.Entity) bool {
		return a.EqualTo(b)
	})
	p.filterableViewport.SetAdjustObjectsForFilter(func(filterText string, isRegex bool) []entity.Entity {
		f := makeFilterFromText(filterText, isRegex, p.keyMap)
		entityTree.UpdatePrettyPrintPrefixes(f)
		return entityTree.GetVisibleEntities(f)
	})
	return p
}

func (p EntityPage) WithMaintainSelection(maintainSelection bool) EntityPage {
	if maintainSelection {
		p.filterableViewport.SetSelectionComparator(func(a, b entity.Entity) bool {
			return a.EqualTo(b)
		})
	} else {
		p.filterableViewport.SetSelectionComparator(nil)
	}
	return p
}

func (p EntityPage) GetSelectionActions() (entity.Entity, map[entity.Entity]bool) {
	selectedEntity := p.filterableViewport.GetSelectedItem()
	if selectedEntity == nil {
		return entity.Entity{}, nil
	}
	return *selectedEntity, p.entityTree.GetSelectionActions(*selectedEntity, p.getCurrentFilter())
}

func (p EntityPage) getVisibleEntities() []entity.Entity {
	return p.entityTree.GetVisibleEntities(p.getCurrentFilter())
}

func (p EntityPage) getCurrentFilter() filter.Model {
	return makeFilterFromText(p.filterableViewport.GetFilterText(), p.filterableViewport.IsRegexMode(), p.keyMap)
}

func (p *EntityPage) updateStyles() {
	p.filterableViewport.SetViewportStyles(viewportStylesForFocus(p.focused, p.styles))

	prefix := "(S)election"
	if p.focused {
		prefix = p.styles.Blue.Render(prefix)
	}
	p.filterableViewport.SetFilterLinePrefix(prefix)
}

// makeFilterFromText creates a filter.Model from filter text and regex mode
func makeFilterFromText(filterText string, isRegex bool, keyMap keymap.KeyMap) filter.Model {
	return filter.NewFromText(filterText, isRegex, keyMap)
}
