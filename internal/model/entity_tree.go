package model

import (
	"fmt"
	"github.com/robinovitch61/kl/internal/filter"
	"github.com/robinovitch61/kl/internal/util"
	"sort"
	"strings"
)

// EntityTree is a tree of entities with hierarchy Cluster > Namespace > PodOwner > Pod > Container
// can contain multiple clusters
type EntityTree interface {
	// AddOrReplace adds or updates an entity in the tree
	// if a Container entity is added that doesn't have the relevant parents in the tree already,
	// the parents are also added
	// if a Container entity is added in a namespace that the tree isn't aware of yet, the namespace will
	// be inserted in alphabetical order
	AddOrReplace(entity Entity)

	// GetEntities returns all entities in the tree
	// within a cluster and namespace, sorted by pod owner, pod, and container
	GetEntities() []Entity

	// GetVisibleEntities returns all entities that match the filter, any of their children match the filter,
	// or any of their parents match the filter. Returns in same order as GetEntities
	GetVisibleEntities(filter filter.Model) []Entity

	// GetClusterNamespaces returns all cluster namespaces
	GetClusterNamespaces() []ClusterNamespaces

	// AnyScannerStarting returns true if any entity in the tree is in the ScannerStarting state
	AnyScannerStarting() bool

	// IsVisibleGivenFilter returns true if the entity or any of its children or parents match the filter
	IsVisibleGivenFilter(entity Entity, filter filter.Model) bool

	// GetContainerEntities returns all entities that are containers in the tree
	GetContainerEntities() []Entity

	// Remove removes an entity from the tree. If it is the last entity in a parent, the parent is also removed,
	// all the way up the tree
	Remove(entity Entity)

	// GetSelectionActions returns a map of container Entity's to a boolean indicating if they
	// should be activated or deactivated based on the current selection
	// If an Entity doesn't match the filter, it is not included in the response
	// An Entity's current state is considered, e.g. no Inactive entities will be deactivated again
	// If the selection is a container, only the selection is returned
	// If it is a cluster, namespace, pod owner, or pod, all children containers are considered
	GetSelectionActions(selectedEntity Entity, filter filter.Model) map[Entity]bool

	// GetEntity gets an entity by its Container
	GetEntity(container Container) *Entity

	// UpdatePrettyPrintPrefixes updates the Prefix field of all entities in the tree
	// such that the tree renders as a nice visual tree given the current filter
	UpdatePrettyPrintPrefixes(filter filter.Model)

	// ContainerToShortName returns a function mapping a container to its short name
	// Short names are unique identifiers given all the other containers in the tree
	ContainerToShortName(minCharsEachSide int) func(Container) (PageLogContainerName, error)
}

type entityNode struct {
	entity   Entity
	children map[string]*entityNode
}

type isVisibleCache struct {
	filter string
	cache  map[string]bool
}

func newIsVisibleCache(filter filter.Model) isVisibleCache {
	return isVisibleCache{
		filter: filter.Value(),
		cache:  make(map[string]bool),
	}
}

func (c isVisibleCache) ValidFor(filter filter.Model) bool {
	return c.cache != nil && filter.Value() == c.filter
}

func (c isVisibleCache) Contains(e Entity) (bool, bool) {
	v, ok := c.cache[e.Container.ID()]
	return v, ok
}

func (c isVisibleCache) SetAndReturn(e Entity, v bool) bool {
	c.cache[e.Container.ID()] = v
	return v
}

type entityTreeImpl struct {
	allClusterNamespaces []ClusterNamespaces
	root                 map[string]*entityNode
	isVisibleCache       isVisibleCache
}

func NewEntityTree(allClusterNamespaces []ClusterNamespaces) EntityTree {
	return &entityTreeImpl{
		allClusterNamespaces: allClusterNamespaces,
		root:                 make(map[string]*entityNode),
	}
}

func (et *entityTreeImpl) AddOrReplace(entity Entity) {
	et.isVisibleCache = isVisibleCache{}

	if !entity.IsContainer() {
		// for now keep this true, but leave the implementation such that we can remove this later
		panic("entity must be a container")
	}
	// if a container for a cluster is added that doesn't have the namespace in the tree yet, add it
	for i := range et.allClusterNamespaces {
		if et.allClusterNamespaces[i].Cluster == entity.Container.Cluster {
			alreadyExists := false
			for _, namespace := range et.allClusterNamespaces[i].Namespaces {
				if namespace == entity.Container.Namespace {
					alreadyExists = true
					break
				}
			}
			if !alreadyExists {
				et.allClusterNamespaces[i].Namespaces = append(et.allClusterNamespaces[i].Namespaces, entity.Container.Namespace)
				sort.Strings(et.allClusterNamespaces[i].Namespaces)
			}
			break
		}
	}
	if entity.IsCluster {
		et.addCluster(entity, true)
	} else if entity.IsNamespace {
		et.addNamespace(entity, true)
	} else if entity.IsPodOwner {
		et.addPodOwner(entity, true)
	} else if entity.IsPod {
		et.addOrReplacePod(entity, true)
	} else if entity.IsContainer() {
		et.addOrReplaceContainer(entity, true)
	} else {
		panic("unknown entity type")
	}
}

func (et *entityTreeImpl) addCluster(entity Entity, replace bool) {
	clusterId := entity.Container.Cluster
	if _, exists := et.root[clusterId]; !exists {
		et.root[clusterId] = &entityNode{
			entity:   entity,
			children: make(map[string]*entityNode),
		}
	} else if replace {
		et.root[clusterId].entity = entity
	}
}

func (et *entityTreeImpl) addNamespace(entity Entity, replace bool) {
	clusterID := entity.Container.Cluster
	namespaceID := entity.Container.Namespace
	et.addCluster(Entity{Container: Container{Cluster: clusterID}, IsCluster: true}, false)

	cluster := et.root[clusterID]
	if _, exists := cluster.children[namespaceID]; !exists {
		cluster.children[namespaceID] = &entityNode{
			entity:   entity,
			children: make(map[string]*entityNode),
		}
	} else if replace {
		cluster.children[namespaceID].entity = entity
	}
}

func (et *entityTreeImpl) addPodOwner(entity Entity, replace bool) {
	clusterID := entity.Container.Cluster
	namespaceID := entity.Container.Namespace
	podOwnerId := entity.Container.PodOwner
	et.addNamespace(
		Entity{Container: Container{Cluster: clusterID, Namespace: namespaceID}, IsNamespace: true},
		false,
	)

	namespace := et.root[clusterID].children[namespaceID]
	if _, exists := namespace.children[podOwnerId]; !exists {
		namespace.children[podOwnerId] = &entityNode{
			entity:   entity,
			children: make(map[string]*entityNode),
		}
	} else if replace {
		namespace.children[podOwnerId].entity = entity
	}
}

func (et *entityTreeImpl) addOrReplacePod(entity Entity, replace bool) {
	clusterID := entity.Container.Cluster
	namespaceID := entity.Container.Namespace
	podOwnerId := entity.Container.PodOwner
	podID := entity.Container.Pod
	et.addPodOwner(
		Entity{Container: Container{Cluster: clusterID, Namespace: namespaceID, PodOwner: podOwnerId, PodOwnerMetadata: entity.Container.PodOwnerMetadata}, IsPodOwner: true},
		false,
	)

	podOwner := et.root[clusterID].children[namespaceID].children[podOwnerId]
	if _, exists := podOwner.children[podID]; !exists {
		podOwner.children[podID] = &entityNode{
			entity:   entity,
			children: make(map[string]*entityNode),
		}
	} else if replace {
		podOwner.children[podID].entity = entity
	}
}

func (et *entityTreeImpl) addOrReplaceContainer(entity Entity, replace bool) {
	clusterID := entity.Container.Cluster
	namespaceID := entity.Container.Namespace
	podOwnerId := entity.Container.PodOwner
	podID := entity.Container.Pod
	containerID := entity.Container.Name
	et.addOrReplacePod(
		Entity{Container: Container{Cluster: clusterID, Namespace: namespaceID, PodOwner: podOwnerId, Pod: podID, PodOwnerMetadata: entity.Container.PodOwnerMetadata}, IsPod: true},
		false,
	)

	pod := et.root[clusterID].children[namespaceID].children[podOwnerId].children[podID]
	if _, exists := pod.children[containerID]; !exists {
		pod.children[containerID] = &entityNode{
			entity:   entity,
			children: nil,
		}
	} else if replace {
		pod.children[containerID].entity = entity
	}
}

func (et *entityTreeImpl) GetEntities() []Entity {
	var result []Entity

	for _, clusterNamespaces := range et.allClusterNamespaces {
		if cluster, ok := et.root[clusterNamespaces.Cluster]; ok {
			result = append(result, cluster.entity)

			for _, namespaceID := range clusterNamespaces.Namespaces {
				if namespace, ok := cluster.children[namespaceID]; ok {
					result = append(result, namespace.entity)

					podOwners := make([]string, 0, len(namespace.children))
					for podOwnerId := range namespace.children {
						podOwners = append(podOwners, podOwnerId)
					}
					sort.Strings(podOwners)

					for _, podOwnerId := range podOwners {
						podOwner := namespace.children[podOwnerId]
						result = append(result, podOwner.entity)

						pods := make([]string, 0, len(podOwner.children))
						for podID := range podOwner.children {
							pods = append(pods, podID)
						}
						sort.Strings(pods)

						for _, podID := range pods {
							pod := podOwner.children[podID]
							result = append(result, pod.entity)

							containers := make([]string, 0, len(pod.children))
							for containerID := range pod.children {
								containers = append(containers, containerID)
							}
							sort.Strings(containers)

							for _, containerID := range containers {
								container := pod.children[containerID]
								result = append(result, container.entity)
							}
						}
					}
				}
			}
		}
	}

	return result
}

func (et entityTreeImpl) GetVisibleEntities(filter filter.Model) []Entity {
	allEntities := et.GetEntities()
	var visibleEntities []Entity
	for _, entity := range allEntities {
		if et.IsVisibleGivenFilter(entity, filter) {
			visibleEntities = append(visibleEntities, entity)
		}
	}
	return visibleEntities
}

func (et entityTreeImpl) GetClusterNamespaces() []ClusterNamespaces {
	return et.allClusterNamespaces
}

func (et entityTreeImpl) AnyScannerStarting() bool {
	allEntities := et.GetEntities()
	for _, entity := range allEntities {
		if entity.State == ScannerStarting {
			return true
		}
	}
	return false
}

// IsVisibleGivenFilter tends to be called many times in a row with the same filter,
// so uses a filter-specific cache for performance
func (et *entityTreeImpl) IsVisibleGivenFilter(entity Entity, filter filter.Model) bool {
	if et.isVisibleCache.ValidFor(filter) {
		if v, ok := et.isVisibleCache.Contains(entity); ok {
			return v
		}
	} else {
		et.isVisibleCache = newIsVisibleCache(filter)
	}

	if filter.Matches(entity) {
		return et.isVisibleCache.SetAndReturn(entity, true)
	}

	node := et.findNode(entity)
	if node != nil {
		for _, child := range node.children {
			if et.IsVisibleGivenFilter(child.entity, filter) {
				return et.isVisibleCache.SetAndReturn(entity, true)
			}
		}
	}

	parent := et.getParentEntity(entity)
	for !parent.EqualTo(Entity{}) {
		if filter.Matches(parent) {
			return et.isVisibleCache.SetAndReturn(entity, true)
		}
		parent = et.getParentEntity(parent)
	}

	return et.isVisibleCache.SetAndReturn(entity, filter.Matches(parent))
}

func (et *entityTreeImpl) GetContainerEntities() []Entity {
	allEntities := et.GetEntities()
	var containers []Entity
	for _, entity := range allEntities {
		if entity.IsContainer() {
			containers = append(containers, entity)
		}
	}
	return containers
}

func (et *entityTreeImpl) Remove(entity Entity) {
	et.isVisibleCache = isVisibleCache{}

	path := []string{
		entity.Container.Cluster,
		entity.Container.Namespace,
		entity.Container.PodOwner,
		entity.Container.Pod,
		entity.Container.Name,
	}
	et.removeEntity(path, 0, et.root)
}

func (et *entityTreeImpl) removeEntity(path []string, depth int, current map[string]*entityNode) bool {
	if depth >= len(path) || path[depth] == "" {
		return false
	}

	id := path[depth]
	node, ok := current[id]
	if !ok {
		return false
	}

	if depth == len(path)-1 {
		delete(current, id)
		return len(current) == 0
	}

	isEmpty := et.removeEntity(path, depth+1, node.children)
	if isEmpty {
		if len(node.children) == 0 {
			delete(current, id)
			return len(current) == 0
		}
	}

	return false
}

func (et *entityTreeImpl) GetSelectionActions(selectedEntity Entity, filter filter.Model) map[Entity]bool {
	actions := make(map[Entity]bool)

	if selectedEntity.IsContainer() && et.IsVisibleGivenFilter(selectedEntity, filter) {
		actions[selectedEntity] = selectedEntity.State.ActivatesWhenSelected()
	}

	et.traverseChildren(selectedEntity, filter, actions)

	deactivateAny := false
	for _, shouldActivate := range actions {
		if !shouldActivate {
			deactivateAny = true
			break
		}
	}
	for entity := range actions {
		if deactivateAny {
			actions[entity] = false
		} else {
			actions[entity] = true
		}
	}

	// exclude invalid states
	for entity, shouldActivate := range actions {
		switch entity.State {
		case Inactive:
			if !shouldActivate {
				delete(actions, entity)
			}
		case WantScanning:
			if shouldActivate {
				delete(actions, entity)
			}
		case Scanning:
			if shouldActivate {
				delete(actions, entity)
			}
		case Deleted:
			if shouldActivate {
				delete(actions, entity)
			}
		default:
			delete(actions, entity)
		}
	}

	return actions
}

func (et *entityTreeImpl) traverseChildren(entity Entity, filter filter.Model, actions map[Entity]bool) {
	node := et.findNode(entity)
	if node == nil {
		return
	}

	for _, child := range node.children {
		if child.entity.IsContainer() {
			if et.IsVisibleGivenFilter(child.entity, filter) {
				actions[child.entity] = child.entity.State.ActivatesWhenSelected()
			}
		} else {
			et.traverseChildren(child.entity, filter, actions)
		}
	}
}

func (et *entityTreeImpl) findNode(entity Entity) *entityNode {
	if entity.IsCluster {
		return et.root[entity.Container.Cluster]
	}

	parent := et.findNode(et.getParentEntity(entity))
	if parent == nil {
		return nil
	}

	if entity.IsNamespace {
		return parent.children[entity.Container.Namespace]
	} else if entity.IsPodOwner {
		return parent.children[entity.Container.PodOwner]
	} else if entity.IsPod {
		return parent.children[entity.Container.Pod]
	}

	return parent.children[entity.Container.Name]
}

func (et *entityTreeImpl) getParentEntity(entity Entity) Entity {
	if entity.IsNamespace {
		return Entity{Container: Container{Cluster: entity.Container.Cluster}, IsCluster: true}
	} else if entity.IsPodOwner {
		return Entity{Container: Container{Cluster: entity.Container.Cluster, Namespace: entity.Container.Namespace}, IsNamespace: true}
	} else if entity.IsPod {
		return Entity{Container: Container{Cluster: entity.Container.Cluster, Namespace: entity.Container.Namespace, PodOwner: entity.Container.PodOwner, PodOwnerMetadata: entity.Container.PodOwnerMetadata}, IsPodOwner: true}
	} else if entity.IsContainer() {
		return Entity{Container: Container{Cluster: entity.Container.Cluster, Namespace: entity.Container.Namespace, PodOwner: entity.Container.PodOwner, Pod: entity.Container.Pod, PodOwnerMetadata: entity.Container.PodOwnerMetadata}, IsPod: true}
	}
	return Entity{}
}

func (et *entityTreeImpl) GetEntity(container Container) *Entity {
	for _, cluster := range et.root {
		for _, namespace := range cluster.children {
			for _, podOwner := range namespace.children {
				for _, pod := range podOwner.children {
					for _, containerNode := range pod.children {
						if containerNode.entity.Container.Equals(container) {
							return &containerNode.entity
						}
					}
				}
			}
		}
	}
	return nil
}

func (et *entityTreeImpl) UpdatePrettyPrintPrefixes(filter filter.Model) {
	// TODO: consider using lipgloss's tree functionality?
	et.isVisibleCache = isVisibleCache{}

	visibleEntities := et.GetVisibleEntities(filter)

	seenNamespace := false
	seenPodOwner := false
	seenPod := false
	seenContainer := false

	for i := len(visibleEntities) - 1; i >= 0; i-- {
		entity := visibleEntities[i]

		if entity.IsContainer() {
			suffix := "└─"
			if seenContainer {
				suffix = "├─"
			}
			podBar := " "
			if seenPod {
				podBar = "│"
			}
			if seenNamespace && seenPodOwner {
				entity.Prefix = "  │ " + podBar + " " + suffix
			} else if seenNamespace {
				entity.Prefix = "    " + podBar + " " + suffix
			} else if seenPodOwner {
				entity.Prefix = "  │ " + podBar + " " + suffix
			} else {
				entity.Prefix = "    " + podBar + " " + suffix
			}
			seenContainer = true
		} else if entity.IsPod {
			suffix := "└─"
			if seenPod {
				suffix = "├─"
			}
			if seenNamespace && seenPodOwner {
				entity.Prefix = "  │ " + suffix
			} else if seenNamespace {
				entity.Prefix = "    " + suffix
			} else if seenPodOwner {
				entity.Prefix = "  │ " + suffix
			} else {
				entity.Prefix = "    " + suffix
			}
			seenContainer = false
			seenPod = true
		} else if entity.IsPodOwner {
			if seenNamespace && seenPodOwner {
				entity.Prefix = "  ├─"
			} else if seenPodOwner {
				entity.Prefix = "  ├─"
			} else if seenNamespace {
				entity.Prefix = "  └─"
			} else {
				entity.Prefix = "  └─"
			}
			seenPod = false
			seenPodOwner = true
		} else if entity.IsNamespace {
			entity.Prefix = "  "
			seenNamespace = true
		} else if entity.IsCluster {
			seenNamespace = false
			seenPodOwner = false
		}

		visibleEntities[i] = entity
	}

	for _, entity := range visibleEntities {
		if node := et.findNode(entity); node != nil {
			node.entity.Prefix = entity.Prefix
		}
	}
}

func (et entityTreeImpl) ContainerToShortName(minCharsEachSide int) func(Container) (PageLogContainerName, error) {
	entities := et.GetContainerEntities()

	activeClusters := make(map[string]bool)
	activeNamespaces := make(map[string]bool)
	activePodOwners := make(map[string]bool)
	activePods := make(map[string]bool)
	for _, e := range entities {
		if !e.State.MayHaveLogs() {
			continue
		}
		activeClusters[e.Container.Cluster] = true
		activeNamespaces[e.Container.Namespace] = true
		activePodOwners[e.Container.PodOwner] = true
		activePods[e.Container.Pod] = true
	}

	shortNameFromCluster := util.GetUniqueShortNamesFromEdges(activeClusters, minCharsEachSide)
	shortNameFromNamespace := util.GetUniqueShortNamesFromEdges(activeNamespaces, minCharsEachSide)
	shortNameFromPodOwner := util.GetUniqueShortNamesFromEdges(activePodOwners, minCharsEachSide)
	shortNameFromPod := util.GetUniqueShortNamesFromEdges(activePods, minCharsEachSide)
	specFromContainerId := make(map[string]Container)
	for _, e := range entities {
		specFromContainerId[e.Container.ID()] = e.Container
	}

	return func(container Container) (PageLogContainerName, error) {
		c, ok := specFromContainerId[container.ID()]
		if !ok {
			return PageLogContainerName{}, fmt.Errorf("container not found when getting short name: %s", container.HumanReadable())
		}

		shortCluster := shortNameFromCluster[c.Cluster]
		if len(shortNameFromCluster) == 1 {
			shortCluster = ""
		}
		shortNamespace := shortNameFromNamespace[c.Namespace]
		if len(shortNameFromNamespace) == 1 {
			shortNamespace = ""
		}
		shortPodOwner := shortNameFromPodOwner[c.PodOwner]
		if len(shortNameFromPodOwner) == 1 {
			shortPodOwner = ""
		}
		shortPod := shortNameFromPod[c.Pod]
		if len(shortNameFromPod) == 1 {
			shortPod = ""
		}
		var toJoin []string
		for _, v := range []string{shortCluster, shortNamespace, shortPodOwner, shortPod} {
			if v != "" {
				toJoin = append(toJoin, v)
			}
		}
		name := PageLogContainerName{
			Prefix:        strings.Join(toJoin, "/"),
			ContainerName: container.Name,
		}
		return name, nil
	}
}
