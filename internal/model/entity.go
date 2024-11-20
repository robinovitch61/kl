package model

import (
	"fmt"
	"github.com/robinovitch61/kl/internal/constants"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/util"
	"time"
)

// Entity represents a renderable & selectable kubernetes entity (cluster, namespace, pod owner, pod, or container)
type Entity struct {
	Container                                 Container
	IsCluster, IsNamespace, IsPodOwner, IsPod bool
	LogScanner                                *LogScanner
	Prefix                                    string
	State                                     EntityState
}

func (e Entity) Render() string {
	if e.IsCluster {
		return e.Prefix + e.Container.Cluster
	} else if e.IsNamespace {
		return e.Prefix + e.Container.Namespace
	} else if e.IsPodOwner {
		res := e.Prefix + e.Container.PodOwner
		if e.Container.PodOwnerMetadata.OwnerType != "" {
			res += " <" + e.Container.PodOwnerMetadata.OwnerType + ">"
		}
		return res
	} else if e.IsPod {
		return e.Prefix + e.Container.Pod
	} else {
		// for containers
		res := e.Prefix + e.State.StatusIndicator() + " " + e.Container.Name + " (" + e.Container.Status.State.String()

		// running container with started at time, show "for X time"
		if e.Container.Status.State == ContainerRunning && !e.Container.Status.StartedAt.IsZero() {
			res += " for " + util.TimeSince(e.Container.Status.StartedAt)
		}

		// terminated containers with terminated at time
		if e.Container.Status.State == ContainerTerminated && !e.Container.Status.TerminatedAt.IsZero() {
			if e.Container.Status.StartedAt.IsZero() {
				// terminated container with just terminated at time, show "for X time"
				res += " for " + util.TimeSince(e.Container.Status.TerminatedAt)
			} else {
				// terminated container with started at and terminated at time, show "for X time, ran X time"
				res += " for " + util.TimeSince(e.Container.Status.TerminatedAt) + ", ran " + util.FormatDuration(e.Container.Status.TerminatedAt.Sub(e.Container.Status.StartedAt))
			}

			if e.Container.Status.TerminatedFor != "" {
				res += ": " + e.Container.Status.TerminatedFor
			}
		}

		// waiting container with waiting for reason, show "waiting for X"
		if e.Container.Status.State == ContainerWaiting && e.Container.Status.WaitingFor != "" {
			res += ": " + e.Container.Status.WaitingFor
		}

		// add "NEW" to newly started containers
		if e.Container.Status.State == ContainerRunning && e.Container.Status.StartedAt.After(time.Now().Add(-constants.NewContainerThreshold)) {
			res += " - NEW"
		}

		res += ")"
		return res
	}
}

func (e Entity) Equals(other interface{}) bool {
	otherEntity, ok := other.(Entity)
	if !ok {
		return false
	}
	return e.EqualTo(otherEntity)
}

func (e Entity) EqualTo(other Entity) bool {
	return e.Container.ID() == other.Container.ID()
}

func (e Entity) IsContainer() bool {
	return !e.IsCluster && !e.IsNamespace && !e.IsPodOwner && !e.IsPod
}

func (e Entity) AssertIsContainer() error {
	if !e.IsContainer() {
		return fmt.Errorf("entity is not a container: %s", e.Container.HumanReadable())
	}
	return nil
}

func (e Entity) IsChildContainerOfCluster(cluster Entity) bool {
	return e.IsContainer() && e.Container.inClusterOf(cluster.Container)
}

func (e Entity) IsChildContainerOfNamespace(namespace Entity) bool {
	return e.IsContainer() && e.Container.inNamespaceOf(namespace.Container)
}

func (e Entity) IsChildContainerOfPodOwner(podOwner Entity) bool {
	return e.IsContainer() && e.Container.inPodOwnerOf(podOwner.Container)
}

func (e Entity) IsChildContainerOfPod(pod Entity) bool {
	return e.IsContainer() && e.Container.inPodOf(pod.Container)
}

func (e Entity) Type() string {
	if e.IsCluster {
		return "cluster"
	} else if e.IsNamespace {
		return "namespace"
	} else if e.IsPodOwner {
		return "podOwner"
	} else if e.IsPod {
		return "pod"
	} else {
		return "container"
	}
}

func (e Entity) Activate(tree EntityTree) (Entity, EntityTree, []EntityAction) {
	dev.Debug(fmt.Sprintf("Activate %v starts %v with container %v", e.Container.HumanReadable(), e.State, e.Container.Status.State))
	defer func() {
		dev.Debug(fmt.Sprintf("Activate %v ends %v", e.Container.HumanReadable(), e.State))
	}()
	switch e.State {
	case Inactive:
		if e.Container.Status.State == ContainerWaiting {
			e.State = WantScanning
			tree.AddOrReplace(e)
			return e, tree, []EntityAction{}
		} else {
			e.State = ScannerStarting
			tree.AddOrReplace(e)
			return e, tree, []EntityAction{StartScanner}
		}
	default:
		panic(fmt.Sprintf("Activate called for entity in %v state", e.State))
	}
}

func (e Entity) Deactivate(tree EntityTree) (Entity, EntityTree, []EntityAction) {
	dev.Debug(fmt.Sprintf("Deactivate %v starts %v", e.Container.HumanReadable(), e.State))
	defer func() {
		dev.Debug(fmt.Sprintf("Deactivate %v ends %v", e.Container.HumanReadable(), e.State))
	}()
	switch e.State {
	case WantScanning:
		e.State = Inactive
		tree.AddOrReplace(e)
		return e, tree, []EntityAction{}
	case Scanning:
		e.State = ScannerStopping
		tree.AddOrReplace(e)
		return e, tree, []EntityAction{StopScanner}
	case Deleted:
		return e, tree, []EntityAction{RemoveEntity}
	default:
		panic(fmt.Sprintf("Deactivate called for entity in %v state", e.State))
	}
}

func (e Entity) Restart(tree EntityTree) (Entity, EntityTree, []EntityAction) {
	dev.Debug(fmt.Sprintf("Restart %v starts %v", e.Container.HumanReadable(), e.State))
	defer func() {
		dev.Debug(fmt.Sprintf("Restart %v ends %v", e.Container.HumanReadable(), e.State))
	}()
	switch e.State {
	case Scanning:
		e.State = ScannerStopping
		tree.AddOrReplace(e)
		// restarting scanners needs to happen in bulk, so the caller handles it outside the normal action flow
		return e, tree, []EntityAction{}
	default:
		panic(fmt.Sprintf("Restart called for entity in %v state", e.State))
	}
}

func (e Entity) Delete(tree EntityTree, delta ContainerDelta) (Entity, EntityTree, []EntityAction) {
	dev.Debug(fmt.Sprintf("Delete %v starts %v", e.Container.HumanReadable(), e.State))
	defer func() {
		dev.Debug(fmt.Sprintf("Delete %v ends %v", e.Container.HumanReadable(), e.State))
	}()
	switch e.State {
	case Inactive:
		tree.Remove(e)
		return e, tree, []EntityAction{}
	case WantScanning:
		tree.Remove(e)
		return e, tree, []EntityAction{}
	case ScannerStarting:
		tree.Remove(e)
		return e, tree, []EntityAction{StopScanner}
	case Scanning:
		e.State = Deleted
		e.Container.Status = delta.Container.Status
		tree.AddOrReplace(e)
		return e, tree, []EntityAction{StopScannerKeepLogs, MarkLogsTerminated}
	case ScannerStopping:
		tree.Remove(e)
		return e, tree, []EntityAction{StopScanner}
	default:
		panic(fmt.Sprintf("Delete called for entity in %v state", e.State))
	}
}

func (e Entity) Create(tree EntityTree, delta ContainerDelta) (Entity, EntityTree, []EntityAction) {
	defer func() {
		dev.Debug(fmt.Sprintf("CreateEntity %v ends %v", e.Container.HumanReadable(), delta.ToActivate))
	}()
	if delta.ToActivate {
		switch e.Container.Status.State {
		case ContainerUnknown, ContainerWaiting:
			e.State = WantScanning
			tree.AddOrReplace(e)
			return e, tree, []EntityAction{}
		case ContainerRunning, ContainerTerminated:
			e.State = ScannerStarting
			tree.AddOrReplace(e)
			return e, tree, []EntityAction{StartScanner}
		default:
			panic(fmt.Sprintf("Create called to activate with container in %v state", e.Container.Status.State))
		}
	} else {
		e.State = Inactive
		tree.AddOrReplace(e)
		return e, tree, []EntityAction{}
	}
}

type UpdateResult struct {
	Entity             Entity
	Tree               EntityTree
	StartScanner       bool
	MarkLogsTerminated bool
}

func (e Entity) Update(tree EntityTree, delta ContainerDelta) (Entity, EntityTree, []EntityAction) {
	e.Container = delta.Container
	tree.AddOrReplace(e)

	var actions []EntityAction
	if delta.Container.Status.State == ContainerTerminated {
		actions = append(actions, MarkLogsTerminated)
	}

	switch e.State {
	case WantScanning:
		containerState := e.Container.Status.State
		if containerState == ContainerRunning || containerState == ContainerTerminated {
			e.State = ScannerStarting
			tree.AddOrReplace(e)
			actions = append(actions, StartScanner)
		}
		return e, tree, actions
	default:
		// an entity in any other state has its container updated and remains in the same entity state
		return e, tree, actions
	}
}

func (e Entity) ScannerStarted(tree EntityTree, startErr error, scanner LogScanner) (Entity, EntityTree, []EntityAction) {
	dev.Debug(fmt.Sprintf("ScannerStarted %v starts %v", e.Container.HumanReadable(), e.State))
	defer func() {
		dev.Debug(fmt.Sprintf("ScannerStarted %v ends %v", e.Container.HumanReadable(), e.State))
	}()
	switch e.State {
	case ScannerStarting:
		e.Container.Status = scanner.Container.Status

		if startErr != nil {
			e.State = Inactive
			scanner.Cancel()
		} else {
			e.State = Scanning
			e.LogScanner = &scanner
		}

		tree.AddOrReplace(e)
		return e, tree, []EntityAction{}
	default:
		panic(fmt.Sprintf("ScannerStarted called for entity in %v state", e.State))
	}
}

func (e Entity) ScannerStopped(tree EntityTree) (Entity, EntityTree, []EntityAction) {
	dev.Debug(fmt.Sprintf("ScannerStopped %v starts %v", e.Container.HumanReadable(), e.State))
	defer func() {
		dev.Debug(fmt.Sprintf("ScannerStopped %v ends %v", e.Container.HumanReadable(), e.State))
	}()
	switch e.State {
	case ScannerStopping:
		e.State = Inactive
		e.LogScanner = nil
		tree.AddOrReplace(e)
		return e, tree, []EntityAction{}
	case Deleted:
		e.LogScanner = nil
		tree.AddOrReplace(e)
		return e, tree, []EntityAction{}
	default:
		panic(fmt.Sprintf("ScannerStopped called for entity in %v state", e.State))
	}
}
