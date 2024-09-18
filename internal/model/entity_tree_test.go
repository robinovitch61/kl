package model

import (
	"fmt"
	"github.com/robinovitch61/kl/internal/filter"
	"github.com/robinovitch61/kl/internal/keymap"
	"strings"
	"testing"
)

var (
	cluster1 = Entity{
		Container: Container{Cluster: "cluster1"},
		IsCluster: true,
	}
	cluster2 = Entity{
		Container: Container{Cluster: "cluster2"},
		IsCluster: true,
	}
	namespace1 = Entity{
		Container:   Container{Cluster: "cluster1", Namespace: "namespace1"},
		IsNamespace: true,
	}
	namespace2 = Entity{
		Container:   Container{Cluster: "cluster2", Namespace: "namespace2"},
		IsNamespace: true,
	}
	deployment1 = Entity{
		Container:    Container{Cluster: "cluster1", Namespace: "namespace1", Deployment: "deployment1"},
		IsDeployment: true,
	}
	deployment2 = Entity{
		Container:    Container{Cluster: "cluster2", Namespace: "namespace2", Deployment: "deployment2"},
		IsDeployment: true,
	}
	pod1 = Entity{
		Container: Container{Cluster: "cluster1", Namespace: "namespace1", Deployment: "deployment1", Pod: "pod1"},
		IsPod:     true,
	}
	pod2 = Entity{
		Container: Container{Cluster: "cluster2", Namespace: "namespace2", Deployment: "deployment2", Pod: "pod2"},
		IsPod:     true,
	}
	container1Cluster1 = Entity{
		Container: Container{Cluster: "cluster1", Namespace: "namespace1", Deployment: "deployment1", Pod: "pod1", Name: "container1"},
	}
	container2Cluster1 = Entity{
		Container: Container{Cluster: "cluster1", Namespace: "namespace1", Deployment: "deployment1", Pod: "pod1", Name: "container2"},
	}
	container3Cluster1 = Entity{
		Container: Container{Cluster: "cluster1", Namespace: "namespace1", Deployment: "deployment1", Pod: "pod1", Name: "container3"},
	}
	container1Cluster2 = Entity{
		Container: Container{Cluster: "cluster2", Namespace: "namespace2", Deployment: "deployment2", Pod: "pod2", Name: "container1"},
	}
	emptyFilter           = newFilter("", false)
	container1Filter      = newFilter("container1", false)
	container2RegexFilter = newFilter("containe.2", true)
	cluster1RegexFilter   = newFilter("cluste.1", true)
	cluster2Filter        = newFilter("cluster2", false)
)

func newTree() EntityTree {
	allContextNameSpaces := []ClusterNamespaces{
		{
			Cluster:    "cluster1",
			Namespaces: []string{"namespace1", "namespace2"},
		},
		{
			Cluster:    "cluster2",
			Namespaces: []string{"namespace1", "namespace2"},
		},
	}
	return NewEntityTree(allContextNameSpaces)
}

func TestEntityTreeImpl_AddOrReplaceContainer(t *testing.T) {
	tree := newTree()

	tree.AddOrReplace(container1Cluster1)

	entities := tree.GetEntities()
	expected := []Entity{cluster1, namespace1, deployment1, pod1, container1Cluster1}

	if !entitiesEqual(entities, expected) {
		t.Errorf("GetEntities():\n%v\nWant\n%v", formatEntities(entities), formatEntities(expected))
	}
}

func TestEntityTreeImpl_AddOrReplaceContainers(t *testing.T) {
	tree := newTree()

	tree.AddOrReplace(container1Cluster1)
	tree.AddOrReplace(container1Cluster2)

	entities := tree.GetEntities()
	expected := []Entity{cluster1, namespace1, deployment1, pod1, container1Cluster1, cluster2, namespace2, deployment2, pod2, container1Cluster2}

	if !entitiesEqual(entities, expected) {
		t.Errorf("GetEntities() = %v, want %v", entities, expected)
	}
}

func TestEntityTreeImpl_AddOrReplaceUpdate(t *testing.T) {
	tree := newTree()
	tree.AddOrReplace(container1Cluster1)

	updated := container1Cluster1
	updated.LogScannerPending = true
	tree.AddOrReplace(updated)

	entities := tree.GetEntities()

	if len(entities) != 5 || !entities[4].EqualTo(updated) {
		t.Errorf("Updated entity not found or incorrect: got %v, want %v", entities[4], updated)
	}
}

func TestEntityTreeImpl_GetVisibleEntities(t *testing.T) {
	tree := newTree()
	tree.AddOrReplace(container1Cluster1)
	tree.AddOrReplace(container2Cluster1)
	tree.AddOrReplace(container1Cluster2)

	tests := []struct {
		name   string
		filter filter.Model
		want   []Entity
	}{
		{
			name:   "No filter",
			filter: emptyFilter,
			want:   []Entity{cluster1, namespace1, deployment1, pod1, container1Cluster1, container2Cluster1, cluster2, namespace2, deployment2, pod2, container1Cluster2},
		},
		{
			name:   "Filter matches container1",
			filter: container1Filter,
			want:   []Entity{cluster1, namespace1, deployment1, pod1, container1Cluster1, cluster2, namespace2, deployment2, pod2, container1Cluster2},
		},
		{
			name:   "Filter regex matches container2",
			filter: container2RegexFilter,
			want:   []Entity{cluster1, namespace1, deployment1, pod1, container2Cluster1},
		},
		{
			name:   "Filter matching cluster2 shows all children",
			filter: cluster2Filter,
			want:   []Entity{cluster2, namespace2, deployment2, pod2, container1Cluster2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tree.GetVisibleEntities(tt.filter)
			if !entitiesEqual(got, tt.want) {
				t.Errorf(
					"GetVisibleEntities() mismatch:\nGot:\n%s\nWant:\n%s",
					formatEntities(got),
					formatEntities(tt.want),
				)
			}
		})
	}
}

func TestEntityTreeImpl_GetClusterNamespaces_SetExplicitly(t *testing.T) {
	tree := newTree()
	tree.AddOrReplace(container1Cluster1)
	tree.AddOrReplace(container1Cluster2)

	got := tree.GetClusterNamespaces()
	expected := []ClusterNamespaces{
		{
			Cluster:    "cluster1",
			Namespaces: []string{"namespace1", "namespace2"},
		},
		{
			Cluster:    "cluster2",
			Namespaces: []string{"namespace1", "namespace2"},
		},
	}
	if len(got) != len(expected) {
		t.Errorf("GetClusterNamespaces() = %v, want %v", got, expected)
	}
	for i := range got {
		if got[i].Cluster != expected[i].Cluster {
			t.Errorf("GetClusterNamespaces() = %v, want %v", got, expected)
		}
		if len(got[i].Namespaces) != len(expected[i].Namespaces) {
			t.Errorf("GetClusterNamespaces() = %v, want %v", got, expected)
		}
		for j := range got[i].Namespaces {
			if got[i].Namespaces[j] != expected[i].Namespaces[j] {
				t.Errorf("GetClusterNamespaces() = %v, want %v", got, expected)
			}
		}
	}
}

func TestEntityTreeImpl_GetClusterNamespaces_SetImplicitly(t *testing.T) {
	tree := NewEntityTree([]ClusterNamespaces{{Cluster: "cluster1"}, {Cluster: "cluster2"}})
	tree.AddOrReplace(container1Cluster1)
	tree.AddOrReplace(container1Cluster2)

	got := tree.GetClusterNamespaces()
	expected := []ClusterNamespaces{
		{
			Cluster:    "cluster1",
			Namespaces: []string{"namespace1"},
		},
		{
			Cluster:    "cluster2",
			Namespaces: []string{"namespace2"},
		},
	}
	if len(got) != len(expected) {
		t.Errorf("GetClusterNamespaces() = %v, want %v", got, expected)
	}
	for i := range got {
		if got[i].Cluster != expected[i].Cluster {
			t.Errorf("GetClusterNamespaces() = %v, want %v", got, expected)
		}
		if len(got[i].Namespaces) != len(expected[i].Namespaces) {
			t.Errorf("GetClusterNamespaces() = %v, want %v", got, expected)
		}
		for j := range got[i].Namespaces {
			if got[i].Namespaces[j] != expected[i].Namespaces[j] {
				t.Errorf("GetClusterNamespaces() = %v, want %v", got, expected)
			}
		}
	}
}

func TestEntityTreeImpl_AnyPendingContainers(t *testing.T) {
	tree := newTree()
	tree.AddOrReplace(container1Cluster1)
	tree.AddOrReplace(container1Cluster2)

	if tree.AnyPendingContainers() {
		t.Errorf("AnyPendingContainers() = true, want false")
	}

	pendingContainer := container3Cluster1
	pendingContainer.LogScannerPending = true
	tree.AddOrReplace(pendingContainer)

	if !tree.AnyPendingContainers() {
		t.Errorf("AnyPendingContainers() = false, want true")
	}
}

func TestEntityTreeImpl_IsVisibleGivenFilter(t *testing.T) {
	tree := newTree()
	tree.AddOrReplace(container1Cluster1)
	tree.AddOrReplace(container2Cluster1)
	tree.AddOrReplace(container1Cluster2)

	tests := []struct {
		name      string
		filter    filter.Model
		entity    Entity
		isVisible bool
	}{
		{
			name:      "No filter shows visible",
			filter:    emptyFilter,
			entity:    container1Cluster1,
			isVisible: true,
		},
		{
			name:      "Container1 filter shows container1 visible",
			filter:    container1Filter,
			entity:    container1Cluster1,
			isVisible: true,
		},
		{
			name:      "Regex container2 filter shows container2 visible",
			filter:    container2RegexFilter,
			entity:    container2Cluster1,
			isVisible: true,
		},
		{
			name:      "Container1 filter hides container2",
			filter:    container1Filter,
			entity:    container2Cluster1,
			isVisible: false,
		},
		{
			name:      "Container1 filter shows cluster1 visible",
			filter:    container1Filter,
			entity:    cluster1,
			isVisible: true,
		},
		{
			name:      "Cluster1 regex filter shows container1 visible",
			filter:    cluster1RegexFilter,
			entity:    container1Cluster1,
			isVisible: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tree.IsVisibleGivenFilter(tt.entity, tt.filter)
			if got != tt.isVisible {
				t.Errorf("IsVisibleGivenFilter() = %v, want %v", got, tt.isVisible)
			}
		})
	}
}

func TestEntityTreeImpl_GetContainerEntities(t *testing.T) {
	tree := newTree()
	tree.AddOrReplace(container1Cluster1)
	tree.AddOrReplace(container1Cluster2)

	got := tree.GetContainerEntities()
	expected := []Entity{container1Cluster1, container1Cluster2}
	if !entitiesEqual(got, expected) {
		t.Errorf(
			"GetContainerEntities() mismatch:\nGot:\n%s\nWant:\n%s",
			formatEntities(got),
			formatEntities(expected),
		)
	}
}

func TestRemoveAll(t *testing.T) {
	tree := newTree()

	tree.AddOrReplace(container1Cluster1)
	tree.Remove(container1Cluster1)

	entities := tree.GetEntities()

	if len(entities) != 0 {
		t.Errorf("Removed final entity, but still found %d entities", len(entities))
	}
}

func TestRemoveOne(t *testing.T) {
	tree := newTree()

	tree.AddOrReplace(container1Cluster1)
	tree.AddOrReplace(container1Cluster2)
	tree.Remove(container1Cluster1)

	entities := tree.GetEntities()

	expected := []Entity{cluster2, namespace2, deployment2, pod2, container1Cluster2}
	if !entitiesEqual(entities, expected) {
		t.Errorf(
			"GetVisibleEntities() mismatch:\nGot:\n%s\nWant:\n%s",
			formatEntities(entities),
			formatEntities(expected),
		)
	}
}

func TestEntityTreeImpl_GetSelectionActions(t *testing.T) {
	tree := newTree()
	selectedContainer1Cluster1 := container1Cluster1
	selectedContainer1Cluster1.LogScanner = &LogScanner{}
	deselectedButRunningContainer1Cluster2 := container1Cluster2
	deselectedButRunningContainer1Cluster2.Container.Status.State = ContainerRunning
	tree.AddOrReplace(selectedContainer1Cluster1)
	tree.AddOrReplace(container2Cluster1)
	tree.AddOrReplace(deselectedButRunningContainer1Cluster2)

	pendingContainer := container3Cluster1
	pendingContainer.LogScannerPending = true
	tree.AddOrReplace(pendingContainer)

	tests := []struct {
		name            string
		selectedEntity  Entity
		filter          filter.Model
		expectedActions int
	}{
		{
			name:            "Select container1Cluster1",
			selectedEntity:  selectedContainer1Cluster1,
			filter:          emptyFilter,
			expectedActions: 1,
		},
		{
			name:            "Select pod1",
			selectedEntity:  pod1,
			filter:          emptyFilter,
			expectedActions: 1,
		},
		{
			name:            "Select deployment1",
			selectedEntity:  deployment1,
			filter:          emptyFilter,
			expectedActions: 1,
		},
		{
			name:            "Select namespace1",
			selectedEntity:  namespace1,
			filter:          emptyFilter,
			expectedActions: 1,
		},
		{
			name:            "Select cluster1",
			selectedEntity:  cluster1,
			filter:          emptyFilter,
			expectedActions: 1,
		},
		{
			name:            "Select cluster1 with container1 filter",
			selectedEntity:  cluster1,
			filter:          container1Filter,
			expectedActions: 1,
		},
		{
			name:            "Select selectedContainer1Cluster1 with cluster1 regex filter",
			selectedEntity:  selectedContainer1Cluster1,
			filter:          cluster1RegexFilter,
			expectedActions: 1,
		},
		{
			name:            "Select cluster2",
			selectedEntity:  cluster2,
			filter:          emptyFilter,
			expectedActions: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actions := tree.GetSelectionActions(tt.selectedEntity, tt.filter)
			if len(actions) != tt.expectedActions {
				t.Errorf("Expected %d actions, got %d", tt.expectedActions, len(actions))
			}
			selectionContainsActive := false
			for entity := range actions {
				if entity.EqualTo(selectedContainer1Cluster1) {
					selectionContainsActive = true
					break
				}
			}
			for _, shouldActivate := range actions {
				if selectionContainsActive && shouldActivate {
					t.Errorf("Selection included already active container, but action was to activate")
				}
			}
		})
	}
}

func TestEntityTreeImpl_GetEntity(t *testing.T) {
	tree := newTree()
	tree.AddOrReplace(container1Cluster1)
	tree.AddOrReplace(container1Cluster2)

	tests := []struct {
		name string
		want Entity
	}{
		{
			name: "Get container1Cluster1",
			want: container1Cluster1,
		},
		{
			name: "Get container1Cluster2",
			want: container1Cluster2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tree.GetEntity(tt.want.Container)
			if !got.EqualTo(tt.want) {
				t.Errorf("GetEntity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEntityTreeImpl_UpdatePrettyPrintPrefixes_Simple(t *testing.T) {
	tree := newTree()
	tree.AddOrReplace(container1Cluster1)
	tree.AddOrReplace(container1Cluster2)

	tree.UpdatePrettyPrintPrefixes(cluster2Filter)

	entities := tree.GetEntities()
	expected := []Entity{cluster1, namespace1, deployment1, pod1, container1Cluster1, cluster2, namespace2, deployment2, pod2, container1Cluster2}

	if !entitiesEqual(entities, expected) {
		t.Errorf("GetEntities():\n%v\nWant\n%v", formatEntities(entities), formatEntities(expected))
	}
	// first 5 entities prefix should be ""
	for i := 0; i < 5; i++ {
		if entities[i].Prefix != "" {
			t.Errorf("Expected prefix to be empty, got %s", entities[i].Prefix)
		}
	}
	// last 5 entities prefix should be tree-like
	expectedPrefixes := []string{"", "  ", "  └─", "    └─", "      └─"}
	for i := 5; i < 10; i++ {
		if entities[i].Prefix != expectedPrefixes[i-5] {
			t.Errorf("Expected prefix to be %s, got %s", expectedPrefixes[i-5], entities[i].Prefix)
		}
	}
}

func TestEntityTreeImpl_UpdatePrettyPrintPrefixes_Multi(t *testing.T) {
	tree := newTree()
	tree.AddOrReplace(container1Cluster1)
	tree.AddOrReplace(container1Cluster2)
	tree.AddOrReplace(container2Cluster1)

	tree.UpdatePrettyPrintPrefixes(emptyFilter)

	entities := tree.GetEntities()
	expected := []Entity{cluster1, namespace1, deployment1, pod1, container1Cluster1, container2Cluster1, cluster2, namespace2, deployment2, pod2, container1Cluster2}

	if !entitiesEqual(entities, expected) {
		t.Errorf("GetEntities():\n%v\nWant\n%v", formatEntities(entities), formatEntities(expected))
	}
	// check all prefixes
	expectedPrefixes := []string{
		"",
		"  ",
		"  └─",
		"    └─",
		"      ├─",
		"      └─",
		"",
		"  ",
		"  └─",
		"    └─",
		"      └─",
	}
	for i := range entities {
		if entities[i].Prefix != expectedPrefixes[i] {
			t.Errorf("Expected prefix to be %s, got %s", expectedPrefixes[i], entities[i].Prefix)
		}
	}
}

func TestEntityTreeImpl_ContainerToShortName(t *testing.T) {
	tree := newTree()
	for _, e := range []Entity{container1Cluster1, container1Cluster2, container2Cluster1} {
		e.LogScanner = &LogScanner{}
		tree.AddOrReplace(e)
	}
	f := tree.ContainerToShortName(3)
	expected := map[Container]string{
		container1Cluster1.Container: "clu..er1/nam..ce1/dep..nt1/pod1/con..er1",
		container1Cluster2.Container: "clu..er2/nam..ce2/dep..nt2/pod2/con..er1",
		container2Cluster1.Container: "clu..er1/nam..ce1/dep..nt1/pod1/con..er2",
	}
	for c, short := range expected {
		n, err := f(c)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if n != short {
			t.Errorf("Expected short name %s, got %s", short, n)
		}
	}
	_, err := f(Container{Name: "doesntexist"})
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func entitiesEqual(a, b []Entity) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].EqualTo(b[i]) {
			return false
		}
	}
	return true
}

func formatEntities(entities []Entity) string {
	var result strings.Builder
	for i, e := range entities {
		result.WriteString(fmt.Sprintf("%d: {%s}\n", i, formatEntity(e)))
	}
	return result.String()
}

func formatEntity(e Entity) string {
	return fmt.Sprintf("%s, IsCluster: %t, IsNamespace: %t, IsDeployment: %t, IsPod: %t",
		e.Container.ID(),
		e.IsCluster,
		e.IsNamespace,
		e.IsDeployment,
		e.IsPod)
}

func newFilter(s string, isRegex bool) filter.Model {
	f := filter.New("", 0, keymap.KeyMap{})
	f.SetValue(s)
	f.SetIsRegex(isRegex)
	return f
}
