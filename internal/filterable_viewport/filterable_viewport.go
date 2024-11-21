package filterable_viewport

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/filter"
	"github.com/robinovitch61/kl/internal/keymap"
	"github.com/robinovitch61/kl/internal/style"
	"github.com/robinovitch61/kl/internal/viewport"
	"strings"
)

var focusedStyle = style.Blue

type FilterableViewport[T viewport.RenderableComparable] struct {
	Filter            filter.Model
	viewport          *viewport.Model[T]
	allRows           []T
	matchesFilter     func(T, filter.Model) bool
	keyMap            keymap.KeyMap
	filterWithContext bool
	whenEmpty         string
	topHeader         string
	focused           bool
}

func NewFilterableViewport[T viewport.RenderableComparable](
	topHeader string,
	filterWithContext bool,
	startSelectionEnabled bool,
	startWrapOn bool,
	km keymap.KeyMap,
	width, height int,
	allRows []T,
	matchesFilter func(T, filter.Model) bool,
	viewWhenEmpty string,
) FilterableViewport[T] {
	f := filter.New(km)
	f.SetFilteringWithContext(filterWithContext)

	var vp = viewport.New[T](width, height)
	vp.FooterStyle = style.Bold
	vp.SelectedItemStyle = style.Inverse
	vp.HighlightStyle = style.Inverse

	vp.SetSelectionEnabled(startSelectionEnabled)
	vp.SetWrapText(startWrapOn)

	fv := FilterableViewport[T]{
		Filter:            f,
		viewport:          &vp,
		allRows:           allRows,
		matchesFilter:     matchesFilter,
		keyMap:            km,
		filterWithContext: filterWithContext,
		whenEmpty:         viewWhenEmpty,
		topHeader:         topHeader,
	}
	fv.updateViewportHeader()
	return fv
}

func (p FilterableViewport[T]) Update(msg tea.Msg) (FilterableViewport[T], tea.Cmd) {
	dev.DebugUpdateMsg("FilterableViewport", msg)
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	// any updates to the filter should reflect in the viewport header
	defer func() {
		p.updateViewportHeader()
	}()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// clearing the filter is always available regardless of filter focus
		if key.Matches(msg, p.keyMap.Clear) {
			p.clearFilter()
			return p, nil
		}

		if p.Filter.Focused() {
			if key.Matches(msg, p.keyMap.Enter) {
				// done editing
				p.viewport.SelectedItemStyle = style.Inverse
				p.Filter.Blur()
				p.Filter.UpdateLabelAndSuffix()
			}
		} else {
			// if not editing filter, pass through to viewport
			*p.viewport, cmd = p.viewport.Update(msg)
			cmds = append(cmds, cmd)

			// handle next match/prev match
			if key.Matches(msg, p.Filter.KeyMap.FilterNextRow) || key.Matches(msg, p.Filter.KeyMap.FilterPrevRow) {
				// if not filtering with context, or no filter text, ignore
				if !p.Filter.FilteringWithContext || !p.Filter.HasFilterText() {
					return p, nil
				}
				if key.Matches(msg, p.Filter.KeyMap.FilterNextRow) {
					p.Filter.IncrementFilteredSelectionNum()
				} else if key.Matches(msg, p.Filter.KeyMap.FilterPrevRow) {
					p.Filter.DecrementFilteredSelectionNum()
				}
				if p.Filter.HasContextualMatches() {
					p.scrollViewportToItemIdx(p.Filter.GetContextualMatchIdx())
				}
			}

			// focus filter and start editing
			if key.Matches(msg, p.keyMap.Filter) || key.Matches(msg, p.keyMap.FilterRegex) {
				prevIsRegex := p.Filter.IsRegex()
				newIsRegex := key.Matches(msg, p.keyMap.FilterRegex)
				p.Filter.SetIsRegex(newIsRegex)
				p.Filter.Focus()

				// if the filter type has changed, update the visible rows
				if prevIsRegex != newIsRegex {
					p.updateVisibleRows()
				}

				// change the color of the selection
				p.viewport.SelectedItemStyle = style.AltInverse
				return p, textinput.Blink
			}

			// wrap text
			if key.Matches(msg, p.keyMap.Wrap) {
				p.viewport.SetWrapText(!p.viewport.GetWrapText())
				return p, nil
			}
		}

		prevFilterString := p.Filter.Value()

		p.Filter, cmd = p.Filter.Update(msg)
		cmds = append(cmds, cmd)

		if p.Filter.Value() != prevFilterString {
			p.viewport.SetStringToHighlight(p.Filter.Value())
			p.updateVisibleRows()
			p.Filter.UpdateLabelAndSuffix()

			// if filtering with context, reset the match number and scroll to the first match
			if p.Filter.FilteringWithContext {
				p.Filter.ResetContextualFilterMatchNum()
				p.scrollViewportToItemIdx(p.Filter.GetContextualMatchIdx())
			}
		}

		return p, tea.Batch(cmds...)
	}

	p.Filter, cmd = p.Filter.Update(msg)
	cmds = append(cmds, cmd)
	return p, tea.Batch(cmds...)
}

func (p FilterableViewport[T]) View() string {
	var viewportView string
	if len(p.allRows) == 0 {
		whenEmpty := p.whenEmpty
		if p.focused {
			whenEmpty = focusedStyle.Render(whenEmpty)
		}
		viewportView = whenEmpty
	} else {
		viewportView = p.viewport.View()
	}
	return viewportView
}

func (p FilterableViewport[T]) HighjackingInput() bool {
	return p.Filter.Focused()
}

func (p FilterableViewport[T]) WithDimensions(width, height int) FilterableViewport[T] {
	p.viewport.SetWidth(width)
	p.viewport.SetHeight(height)
	return p
}

func (p FilterableViewport[T]) GetSelection() *T {
	return p.viewport.GetSelectedItem()
}

func (p FilterableViewport[T]) GetSelectionIdx() int {
	return p.viewport.GetSelectedItemIdx()
}

func (p FilterableViewport[T]) SetSelectedContentIdx(idx int) {
	p.viewport.SetSelectedItemIdx(idx)
}

func (p *FilterableViewport[T]) SetTopHeader(topHeader string) {
	p.topHeader = topHeader
	p.updateViewportHeader()
}

func (p *FilterableViewport[T]) SetAllRows(allRows []T) {
	p.allRows = allRows
	p.updateVisibleRows()
}

func (p *FilterableViewport[T]) SetFocus(focused bool, selectionEnabled bool) {
	p.focused = focused
	p.viewport.SetSelectionEnabled(selectionEnabled)
	if focused {
		p.viewport.FooterStyle = style.Regular
	} else {
		p.viewport.FooterStyle = style.Alt
	}

	p.updateViewportHeader()
}

func (p *FilterableViewport[T]) SetAllRowsAndMatchesFilter(allRows []T, matchesFilter func(T, filter.Model) bool) {
	p.allRows = allRows
	p.matchesFilter = matchesFilter
	p.updateVisibleRows()
}

func (p *FilterableViewport[T]) SetTopSticky(topSticky bool) {
	p.viewport.SetTopSticky(topSticky)
}

func (p *FilterableViewport[T]) SetBottomSticky(bottomSticky bool) {
	p.viewport.SetBottomSticky(bottomSticky)
}

func (p *FilterableViewport[T]) SetMaintainSelection(maintainSelection bool) {
	p.viewport.SetMaintainSelection(maintainSelection)
}

func (p *FilterableViewport[T]) ToggleFilteringWithContext() {
	p.Filter.SetFilteringWithContext(!p.Filter.FilteringWithContext)
	p.updateVisibleRows()
}

func (p *FilterableViewport[T]) SetUpDownMovementWithShift() {
	upDownBindings := []*key.Binding{
		&p.viewport.KeyMap.Up,
		&p.viewport.KeyMap.Down,
		&p.viewport.KeyMap.PageUp,
		&p.viewport.KeyMap.PageDown,
		&p.viewport.KeyMap.HalfPageUp,
		&p.viewport.KeyMap.HalfPageDown,
	}
	for i := range upDownBindings {
		newKeys := upDownBindings[i].Keys()
		for j := range newKeys {
			if (newKeys[j] == "up" || newKeys[j] == "down") && !strings.Contains(newKeys[j], "shift") {
				newKeys[j] = "shift+" + newKeys[j]
			}
			if len(newKeys[j]) == 1 {
				newKeys[j] = strings.ToUpper(newKeys[j])
			}
		}
		upDownBindings[i].SetKeys(newKeys...)
	}
}

func (p *FilterableViewport[T]) updateVisibleRows() {
	dev.Debug("Updating visible rows")
	defer dev.Debug("Done updating visible rows")

	if p.Filter.FilteringWithContext && p.Filter.Value() != "" {
		var entityIndexesMatchingFilter []int
		for i := range p.allRows {
			if p.matchesFilter(p.allRows[i], p.Filter) {
				entityIndexesMatchingFilter = append(entityIndexesMatchingFilter, i)
			}
		}
		p.Filter.SetIndexesMatchingFilter(entityIndexesMatchingFilter)
		p.viewport.SetContent(p.allRows)
	} else if p.Filter.Value() != "" {
		var filtered []T
		for i := range p.allRows {
			if p.matchesFilter(p.allRows[i], p.Filter) {
				filtered = append(filtered, p.allRows[i])
			}
		}
		p.viewport.SetContent(filtered)
	} else {
		p.viewport.SetContent(p.allRows)
	}
}

func (p *FilterableViewport[T]) updateViewportHeader() {
	prefix := p.topHeader
	if p.focused {
		prefix = focusedStyle.Render(prefix)
	}
	p.viewport.SetHeader([]string{prefix + " " + p.Filter.View()})
}

func (p *FilterableViewport[T]) clearFilter() {
	p.Filter.BlurAndClear()
	p.viewport.SetStringToHighlight("")
	p.viewport.SelectedItemStyle = style.Inverse
	p.updateVisibleRows()
}

func (p *FilterableViewport[T]) scrollViewportToItemIdx(itemIdx int) {
	if p.viewport.GetSelectionEnabled() {
		p.viewport.SetSelectedItemIdx(itemIdx)
	} else {
		p.viewport.ScrollSoItemIdxInView(itemIdx)
	}
	p.Filter.UpdateLabelAndSuffix()
}
