package tree

import (
	"testing"
	"time"

	"github.com/robinovitch61/kl/internal/domain"
)

func TestNewTree(t *testing.T) {
	tree := NewTree()
	nodes := tree.Nodes()
	if len(nodes) != 0 {
		t.Errorf("new tree should have 0 nodes, got %d", len(nodes))
	}
}

func TestTreeUpdate(t *testing.T) {
	containers := []domain.SelectableContainer{
		{
			Container: domain.Container{
				ID: domain.ContainerID{
					Cluster:   "prod",
					Namespace: "default",
					Pod:       "api-pod-123",
					Container: "api",
				},
				OwnerName: "api-deployment",
				OwnerType: "Deployment",
				StartedAt: time.Now(),
				IsRunning: true,
			},
			State: domain.StateInactive,
		},
		{
			Container: domain.Container{
				ID: domain.ContainerID{
					Cluster:   "prod",
					Namespace: "default",
					Pod:       "api-pod-123",
					Container: "sidecar",
				},
				OwnerName: "api-deployment",
				OwnerType: "Deployment",
				StartedAt: time.Now(),
				IsRunning: true,
			},
			State: domain.StateInactive,
		},
	}

	tree := NewTree().Update(containers)
	nodes := tree.Nodes()

	// Should have: cluster, namespace, owner, pod, container1, container2
	// Exact count depends on implementation
	if len(nodes) == 0 {
		t.Error("tree should have nodes after update")
	}

	// Find container nodes
	containerCount := 0
	for _, n := range nodes {
		if n.IsContainer() {
			containerCount++
		}
	}
	if containerCount != 2 {
		t.Errorf("expected 2 container nodes, got %d", containerCount)
	}
}

func TestTreeToggleSelection(t *testing.T) {
	containers := []domain.SelectableContainer{
		{
			Container: domain.Container{
				ID: domain.ContainerID{
					Cluster:   "prod",
					Namespace: "default",
					Pod:       "api-pod",
					Container: "api",
				},
				IsRunning: true,
			},
			State: domain.StateInactive,
		},
	}

	tree := NewTree().Update(containers)
	nodes := tree.Nodes()

	// Find the container node index
	containerIdx := -1
	for i, n := range nodes {
		if n.IsContainer() {
			containerIdx = i
			break
		}
	}

	if containerIdx == -1 {
		t.Fatal("no container node found")
	}

	// Toggle selection
	newTree, changes := tree.ToggleSelection(containerIdx)
	if len(changes) == 0 {
		t.Error("expected state change after toggle")
	}

	// Verify state changed
	if len(changes) > 0 {
		if changes[0].FromState != domain.StateInactive {
			t.Errorf("expected FromState Inactive, got %v", changes[0].FromState)
		}
		// Should transition to WantScanning or ScannerStarting depending on IsRunning
	}

	// Check that node is now selected
	newNodes := newTree.Nodes()
	for _, n := range newNodes {
		if n.IsContainer() {
			c := n.Container()
			if c != nil && c.State == domain.StateInactive {
				t.Error("container should not be Inactive after selection")
			}
		}
	}
}

func TestTreeDeselectAll(t *testing.T) {
	containers := []domain.SelectableContainer{
		{
			Container: domain.Container{
				ID: domain.ContainerID{
					Cluster:   "prod",
					Namespace: "default",
					Pod:       "pod1",
					Container: "c1",
				},
			},
			State: domain.StateScanning,
		},
		{
			Container: domain.Container{
				ID: domain.ContainerID{
					Cluster:   "prod",
					Namespace: "default",
					Pod:       "pod2",
					Container: "c2",
				},
			},
			State: domain.StateScanning,
		},
	}

	tree := NewTree().Update(containers)
	newTree, changes := tree.DeselectAll()

	if len(changes) != 2 {
		t.Errorf("expected 2 state changes, got %d", len(changes))
	}

	// All containers should be deselected
	for _, n := range newTree.Nodes() {
		if n.IsContainer() {
			c := n.Container()
			if c != nil && c.State != domain.StateInactive && c.State != domain.StateScannerStopping {
				t.Errorf("expected container to be Inactive or ScannerStopping, got %v", c.State)
			}
		}
	}
}

func TestNodeGetItem(t *testing.T) {
	node := Node{
		kind:  KindCluster,
		label: "prod-cluster",
	}

	item := node.GetItem()
	content := item.Content()
	if content != "prod-cluster" {
		t.Errorf("expected content 'prod-cluster', got %q", content)
	}
}

func TestNodeIsContainer(t *testing.T) {
	clusterNode := Node{kind: KindCluster}
	containerNode := Node{kind: KindContainer}

	if clusterNode.IsContainer() {
		t.Error("cluster node should not be container")
	}
	if !containerNode.IsContainer() {
		t.Error("container node should be container")
	}
}
