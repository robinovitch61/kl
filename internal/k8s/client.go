package k8s

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/kl/internal/domain"
)

// ContainerDelta represents a container add/update/remove event
type ContainerDelta struct {
	Container domain.Container
	IsRemoved bool
}

// ContainerDeltasMsg contains batched container changes
type ContainerDeltasMsg struct {
	Deltas []ContainerDelta
}

// client wraps Kubernetes client for a single cluster (internal)
type client struct {
	contextName string
}

// newClient creates a client for the given context
func newClient(kubeconfig, contextName string) (*client, error) {
	// TODO: implement
	return &client{contextName: contextName}, nil
}

// Manager coordinates container watching across multiple clusters
type Manager struct {
	clients []*client
}

// NewManager creates a manager for multiple contexts
func NewManager(kubeconfig string, contexts []string) (*Manager, error) {
	// TODO: implement
	var clients []*client
	for _, ctx := range contexts {
		c, err := newClient(kubeconfig, ctx)
		if err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}
	return &Manager{clients: clients}, nil
}

// WatchContainersCmd returns a command that watches for container changes
func (m *Manager) WatchContainersCmd(ctx context.Context, namespaces []string, selector string) tea.Cmd {
	// TODO: implement
	return nil
}
