package internal

import (
	"github.com/robinovitch61/kl/internal/keymap"
	"github.com/robinovitch61/kl/internal/model"
)

type Config struct {
	KeyMap         keymap.KeyMap
	AllNamespaces  bool
	Descending     bool
	KubeConfigPath string
	LogsView       bool
	Contexts       string
	Namespaces     string
	SinceTime      model.SinceTime
	Selectors      model.Selectors
	Version        string
}
