package internal

import (
	"github.com/robinovitch61/kl/internal/model"
)

type Config struct {
	AllNamespaces    bool
	ContainerLimit   int
	Contexts         string
	Descending       bool
	IgnoreOwnerTypes []string
	KubeConfigPath   string
	LogsView         bool
	Matchers         model.Matchers
	Namespaces       string
	SinceTime        model.SinceTime
	Version          string
}
