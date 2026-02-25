package filter

import (
	"charm.land/bubbles/v2/key"
)

type filterKeyMap struct {
	Forward       key.Binding
	Back          key.Binding
	Filter        key.Binding
	FilterRegex   key.Binding
	FilterPrevRow key.Binding
	FilterNextRow key.Binding
}
