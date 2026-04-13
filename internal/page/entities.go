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
	theme              style.Theme
	focused            bool
	viewWhenEmpty      string
}

// assert EntityPage implements GenericPage
var _ GenericPage = EntityPage{}

func NewEntitiesPage(
	keyMap keymap.KeyMap,
	width, height int,
	entityTree entity.Tree,
	theme style.Theme,
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
		viewport.WithSelectionEnabled[entity.Entity](true),
		viewport.WithWrapText[entity.Entity](false),
	)

	exactMode := filterableviewport.ExactFilterMode(keyMap.Filter)
	fuzzyMode := filterableviewport.FuzzyFilterMode(keyMap.FilterFuzzy)
	filterModes := []filterableviewport.FilterMode{exactMode, fuzzyMode}
	filterModesByName := map[filterableviewport.FilterModeName]filterableviewport.FilterMode{
		exactMode.Name: exactMode,
		fuzzyMode.Name: fuzzyMode,
	}

	fvp := filterableviewport.New(vp,
		filterableviewport.WithKeyMap[entity.Entity](filterableviewport.KeyMap{
			ApplyFilterKey:             keyMap.Enter,
			CancelFilterKey:            keyMap.Clear,
			ToggleMatchingItemsOnlyKey: keyMap.Context,
			NextMatchKey:               keyMap.FilterNextRow,
			PrevMatchKey:               keyMap.FilterPrevRow,
			SearchHistoryPrevKey:       keyMap.SearchHistoryPrev,
			SearchHistoryNextKey:       keyMap.SearchHistoryNext,
		}),
		filterableviewport.WithFilterModes[entity.Entity](filterModes),
		filterableviewport.WithMatchingItemsOnly[entity.Entity](false),
		filterableviewport.WithCanToggleMatchingItemsOnly[entity.Entity](false),
		filterableviewport.WithEmptyText[entity.Entity]("'/' or 'z' to filter"),
		filterableviewport.WithFilterLinePosition[entity.Entity](filterableviewport.FilterLineTop),
		filterableviewport.WithFilterLinePrefix[entity.Entity]("(S)election"),
		filterableviewport.WithStyles[entity.Entity](filterableviewport.Styles{
			Match: filterableviewport.MatchStyles{
				Focused:           theme.MatchFocused,
				FocusedIfSelected: theme.MatchFocusedIfSelected,
				Unfocused:         theme.MatchUnfocused,
			},
		}),
		filterableviewport.WithAdjustObjectsForFilter(func(filterText string, modeName filterableviewport.FilterModeName) []entity.Entity {
			mode, ok := filterModesByName[modeName]
			if !ok {
				mode = exactMode
			}
			f := filter.New(filterText, mode)
			entityTree.UpdatePrettyPrintPrefixes(f)
			return entityTree.GetVisibleEntities(f)
		}),
	)

	fvp.SetObjects(entityTree.GetEntities())

	p := EntityPage{
		filterableViewport: fvp,
		entityTree:         entityTree,
		keyMap:             keyMap,
		theme:              theme,
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
		if p.focused {
			return p.theme.FilterPrefixFocused.Render(p.viewWhenEmpty)
		}
		return p.viewWhenEmpty
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

func (p EntityPage) WithTheme(theme style.Theme) GenericPage {
	p.theme = theme
	p.updateStyles()
	p.filterableViewport.SetFilterableViewportStyles(filterableviewport.Styles{
		Match: filterableviewport.MatchStyles{
			Focused:           theme.MatchFocused,
			FocusedIfSelected: theme.MatchFocusedIfSelected,
			Unfocused:         theme.MatchUnfocused,
		},
	})
	return p
}

func (p EntityPage) Help() string {
	return help.MakeHelp(p.keyMap, p.theme.HelpKeyColumn)
}

func (p EntityPage) WithEntityTree(entityTree entity.Tree) EntityPage {
	p.entityTree = entityTree
	f := p.getCurrentFilter()
	p.entityTree.UpdatePrettyPrintPrefixes(f)
	p.filterableViewport.SetObjects(p.entityTree.GetVisibleEntities(f))
	filterModesByName := make(map[filterableviewport.FilterModeName]filterableviewport.FilterMode)
	for _, mode := range p.filterableViewport.FilterModes() {
		filterModesByName[mode.Name] = mode
	}
	defaultMode := p.filterableViewport.FilterModes()[0]
	p.filterableViewport.SetAdjustObjectsForFilter(func(filterText string, modeName filterableviewport.FilterModeName) []entity.Entity {
		mode, ok := filterModesByName[modeName]
		if !ok {
			mode = defaultMode
		}
		f := filter.New(filterText, mode)
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
	activeMode := p.filterableViewport.GetActiveFilterMode()
	if activeMode == nil {
		// no active filter mode — return a filter that matches everything
		return filter.New("", p.filterableViewport.FilterModes()[0])
	}
	return filter.New(p.filterableViewport.GetFilterText(), *activeMode)
}

func (p *EntityPage) updateStyles() {
	p.filterableViewport.SetViewportStyles(viewportStylesForFocus(p.focused, p.theme))

	prefix := "(S)election"
	if p.focused {
		prefix = p.theme.FilterPrefixFocused.Render(prefix)
	}
	p.filterableViewport.SetFilterLinePrefix(prefix)
}
