package filter

import (
	"github.com/robinovitch61/viewport/filterableviewport"
)

// Model is a lightweight filter matcher backed by a viewport FilterMode.
type Model struct {
	value string
	mode  filterableviewport.FilterMode
}

// New creates a filter from text and a FilterMode.
func New(text string, mode filterableviewport.FilterMode) Model {
	return Model{value: text, mode: mode}
}

func (m Model) Matches(s string) bool {
	return m.mode.Matches(m.value, s)
}

func (m Model) Value() string {
	return m.value
}
