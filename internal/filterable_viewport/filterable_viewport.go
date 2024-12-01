package filterable_viewport

import (
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/filter"
	"github.com/robinovitch61/kl/internal/keymap"
	"github.com/robinovitch61/kl/internal/style"
	"github.com/robinovitch61/kl/internal/viewport"
	"strings"
)

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
	styles            style.Styles
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
	styles style.Styles,
) FilterableViewport[T] {
	f := filter.New(km)
	f.SetFilteringWithContext(filterWithContext)

	var vp = viewport.New[T](width, height)

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
		styles:            styles,
	}
	fv.updateViewportStyles()
	fv.updateViewportHeader()
	return fv
}

func (fv FilterableViewport[T]) Update(msg tea.Msg) (FilterableViewport[T], tea.Cmd) {
	dev.DebugUpdateMsg("FilterableViewport", msg)
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	// any updates to the filter should reflect in the viewport header
	defer func() {
		fv.updateViewportHeader()
	}()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// clearing the filter is always available regardless of filter focus
		if key.Matches(msg, fv.keyMap.Clear) {
			fv.clearFilter()
			return fv, nil
		}

		if fv.Filter.Focused() {
			if key.Matches(msg, fv.keyMap.Enter) {
				// done editing
				fv.Filter.Blur()
				fv.Filter.UpdateLabelAndSuffix()
				fv.updateViewportStyles()
			}
		} else {
			// if not editing filter, pass through to viewport
			*fv.viewport, cmd = fv.viewport.Update(msg)
			cmds = append(cmds, cmd)

			// handle next match/prev match
			if key.Matches(msg, fv.Filter.KeyMap.FilterNextRow) || key.Matches(msg, fv.Filter.KeyMap.FilterPrevRow) {
				// if not filtering with context, or no filter text, ignore
				if !fv.Filter.FilteringWithContext || !fv.Filter.HasFilterText() {
					return fv, nil
				}
				if key.Matches(msg, fv.Filter.KeyMap.FilterNextRow) {
					fv.Filter.IncrementFilteredSelectionNum()
				} else if key.Matches(msg, fv.Filter.KeyMap.FilterPrevRow) {
					fv.Filter.DecrementFilteredSelectionNum()
				}
				if fv.Filter.HasContextualMatches() {
					fv.scrollViewportToItemIdx(fv.Filter.GetContextualMatchIdx())
				}
			}

			// focus filter and start editing
			if key.Matches(msg, fv.keyMap.Filter) || key.Matches(msg, fv.keyMap.FilterRegex) {
				prevIsRegex := fv.Filter.IsRegex()
				newIsRegex := key.Matches(msg, fv.keyMap.FilterRegex)
				fv.Filter.SetIsRegex(newIsRegex)
				fv.Filter.Focus()
				fv.updateViewportStyles()

				// if the filter type has changed, update the visible rows
				if prevIsRegex != newIsRegex {
					fv.updateVisibleRows()
				}

				return fv, textinput.Blink
			}

			// wrap text
			if key.Matches(msg, fv.keyMap.Wrap) {
				fv.viewport.SetWrapText(!fv.viewport.GetWrapText())
				return fv, nil
			}
		}

		prevFilterString := fv.Filter.Value()

		fv.Filter, cmd = fv.Filter.Update(msg)
		cmds = append(cmds, cmd)

		if fv.Filter.Value() != prevFilterString {
			fv.viewport.SetStringToHighlight(fv.Filter.Value())
			fv.updateVisibleRows()
			fv.Filter.UpdateLabelAndSuffix()

			// if filtering with context, reset the match number and scroll to the first match
			if fv.Filter.FilteringWithContext {
				fv.Filter.ResetContextualFilterMatchNum()
				fv.scrollViewportToItemIdx(fv.Filter.GetContextualMatchIdx())
			}
		}

		return fv, tea.Batch(cmds...)
	}

	fv.Filter, cmd = fv.Filter.Update(msg)
	cmds = append(cmds, cmd)
	return fv, tea.Batch(cmds...)
}

func (fv FilterableViewport[T]) View() string {
	var viewportView string
	if len(fv.allRows) == 0 {
		whenEmpty := fv.whenEmpty
		if fv.focused {
			whenEmpty = fv.styles.Blue.Render(whenEmpty)
		}
		viewportView = whenEmpty
	} else {
		viewportView = fv.viewport.View()
	}
	return viewportView
}

func (fv FilterableViewport[T]) HighjackingInput() bool {
	return fv.Filter.Focused()
}

func (fv FilterableViewport[T]) WithDimensions(width, height int) FilterableViewport[T] {
	fv.viewport.SetWidth(width)
	fv.viewport.SetHeight(height)
	return fv
}

func (fv FilterableViewport[T]) GetSelection() *T {
	return fv.viewport.GetSelectedItem()
}

func (fv FilterableViewport[T]) GetSelectionIdx() int {
	return fv.viewport.GetSelectedItemIdx()
}

func (fv FilterableViewport[T]) SetSelectedContentIdx(idx int) {
	fv.viewport.SetSelectedItemIdx(idx)
}

func (fv *FilterableViewport[T]) SetTopHeader(topHeader string) {
	fv.topHeader = topHeader
	fv.updateViewportHeader()
}

func (fv *FilterableViewport[T]) SetAllRows(allRows []T) {
	fv.allRows = allRows
	fv.updateVisibleRows()
}

func (fv *FilterableViewport[T]) SetFocus(focused bool) {
	fv.focused = focused
	fv.updateViewportStyles()
	fv.updateViewportHeader()
}

func (fv *FilterableViewport[T]) SetAllRowsAndMatchesFilter(allRows []T, matchesFilter func(T, filter.Model) bool) {
	fv.allRows = allRows
	fv.matchesFilter = matchesFilter
	fv.updateVisibleRows()
}

func (fv *FilterableViewport[T]) SetTopSticky(topSticky bool) {
	fv.viewport.SetTopSticky(topSticky)
}

func (fv *FilterableViewport[T]) SetBottomSticky(bottomSticky bool) {
	fv.viewport.SetBottomSticky(bottomSticky)
}

func (fv *FilterableViewport[T]) SetMaintainSelection(maintainSelection bool) {
	fv.viewport.SetMaintainSelection(maintainSelection)
}

func (fv *FilterableViewport[T]) ToggleFilteringWithContext() {
	fv.Filter.SetFilteringWithContext(!fv.Filter.FilteringWithContext)
	fv.updateVisibleRows()
}

func (fv *FilterableViewport[T]) SetUpDownMovementWithShift() {
	upDownBindings := []*key.Binding{
		&fv.viewport.KeyMap.Up,
		&fv.viewport.KeyMap.Down,
		&fv.viewport.KeyMap.PageUp,
		&fv.viewport.KeyMap.PageDown,
		&fv.viewport.KeyMap.HalfPageUp,
		&fv.viewport.KeyMap.HalfPageDown,
	}
	for i := range upDownBindings {
		newKeys := upDownBindings[i].Keys()
		for j := range newKeys {
			if !strings.Contains(newKeys[j], "shift") {
				newKeys[j] = "shift+" + newKeys[j]
			}
		}
		upDownBindings[i].SetKeys(newKeys...)
	}
}

func (fv *FilterableViewport[T]) updateVisibleRows() {
	dev.Debug("Updating visible rows")
	defer dev.Debug("Done updating visible rows")

	if fv.Filter.FilteringWithContext && fv.Filter.Value() != "" {
		var entityIndexesMatchingFilter []int
		for i := range fv.allRows {
			if fv.matchesFilter(fv.allRows[i], fv.Filter) {
				entityIndexesMatchingFilter = append(entityIndexesMatchingFilter, i)
			}
		}
		fv.Filter.SetIndexesMatchingFilter(entityIndexesMatchingFilter)
		fv.viewport.SetContent(fv.allRows)
	} else if fv.Filter.Value() != "" {
		var filtered []T
		for i := range fv.allRows {
			if fv.matchesFilter(fv.allRows[i], fv.Filter) {
				filtered = append(filtered, fv.allRows[i])
			}
		}
		fv.viewport.SetContent(filtered)
	} else {
		fv.viewport.SetContent(fv.allRows)
	}
}

func (fv *FilterableViewport[T]) updateViewportHeader() {
	prefix := fv.topHeader
	if fv.focused {
		prefix = fv.styles.Blue.Render(prefix)
	}
	fv.viewport.SetHeader([]string{prefix + " " + fv.Filter.View()})
}

func (fv *FilterableViewport[T]) clearFilter() {
	fv.Filter.BlurAndClear()
	fv.viewport.SetStringToHighlight("")
	fv.updateViewportStyles()
	fv.updateVisibleRows()
}

func (fv *FilterableViewport[T]) scrollViewportToItemIdx(itemIdx int) {
	if fv.viewport.GetSelectionEnabled() {
		fv.viewport.SetSelectedItemIdx(itemIdx)
	} else {
		fv.viewport.ScrollSoItemIdxInView(itemIdx)
	}
	fv.Filter.UpdateLabelAndSuffix()
}

func (fv *FilterableViewport[T]) SetStyles(styles style.Styles) {
	fv.styles = styles
	fv.Filter.SetStyles(styles)
	fv.updateViewportStyles()
	fv.updateViewportHeader()
}

func (fv *FilterableViewport[T]) updateViewportStyles() {
	if fv.focused {
		fv.viewport.SelectedItemStyle = fv.styles.Inverse
		fv.viewport.FooterStyle = lipgloss.NewStyle()
	} else {
		fv.viewport.SelectedItemStyle = lipgloss.NewStyle()
		fv.viewport.FooterStyle = fv.styles.Alt
	}

	if fv.Filter.Focused() {
		fv.viewport.SelectedItemStyle = fv.styles.AltInverse
	}

	fv.viewport.HighlightStyle = fv.styles.Inverse
	fv.viewport.HighlightStyleIfSelected = fv.styles.Unset
}
