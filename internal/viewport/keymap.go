package viewport

import "github.com/charmbracelet/bubbles/v2/key"

// KeyMap contains viewport key bindings
type KeyMap struct {
	PageDown     key.Binding
	PageUp       key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
	Up           key.Binding
	Down         key.Binding
	Left         key.Binding
	Right        key.Binding
	Top          key.Binding
	Bottom       key.Binding
}
