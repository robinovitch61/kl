package internal

import (
	"github.com/robinovitch61/kl/internal/domain"
	"k8s.io/apimachinery/pkg/labels"
)

type Config struct {
	AllNamespaces    bool
	ContainerLimit   int
	Contexts         []string
	Descending       bool
	IgnoreOwnerTypes []string
	KubeConfigPath   string
	LogsView         bool
	LogFilter        domain.LogFilter
	Matchers         domain.Matchers
	Namespaces       []string
	Selector         labels.Selector
	SinceTime        domain.SinceTime
	Version          string
}
