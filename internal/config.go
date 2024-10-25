package internal

import (
	"github.com/robinovitch61/kl/internal/model"
	"k8s.io/apimachinery/pkg/labels"
)

type Config struct {
	AllNamespaces    bool
	ContainerLimit   int
	Contexts         string
	Descending       bool
	IgnoreOwnerTypes []string
	KubeConfigPath   string
	LogsView         bool
	LogFilter        model.LogFilter
	Matchers         model.Matchers
	Namespaces       string
	Selector         labels.Selector
	SinceTime        model.SinceTime
	Version          string
}
