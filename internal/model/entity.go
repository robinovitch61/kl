package model

import (
	"fmt"
	"github.com/robinovitch61/kl/internal/constants"
	"github.com/robinovitch61/kl/internal/util"
	"time"
)

// Entity represents a renderable & selectable kubernetes entity (cluster, namespace, deployment, pod, or container)
type Entity struct {
	Container                                   Container
	IsCluster, IsNamespace, IsDeployment, IsPod bool
	LogScanner                                  *LogScanner
	LogScannerPending                           bool
	Terminated                                  bool
	Prefix                                      string
}

func (e Entity) Render() string {
	prefix := "[ ]"
	if e.LogScannerPending {
		prefix = "[.]"
	} else if e.IsSelected() {
		prefix = "[x]"
	}

	if e.IsCluster {
		return e.Prefix + e.Container.Cluster
	} else if e.IsNamespace {
		return e.Prefix + e.Container.Namespace
	} else if e.IsDeployment {
		return e.Prefix + e.Container.Deployment
	} else if e.IsPod {
		return e.Prefix + e.Container.Pod
	} else {
		containerRepr := e.Prefix + prefix + " " + e.Container.Name + " (" + e.Container.Status.State.String()
		if !e.Container.Status.RunningSince.IsZero() {
			containerRepr += " for " + util.TimeSince(e.Container.Status.RunningSince)
		}
		if e.Container.Status.RunningSince.After(time.Now().Add(-constants.NewContainerThreshold)) {
			containerRepr += " - NEW"
		}
		containerRepr += ")"
		return containerRepr
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

// IsSelected corresponds anything in the tree that is visually selected, i.e. has [x]
// in code, it corresponds to having a non-nil LogScanner
// if an entity isn't selected, there should be no logs displayed for it
func (e Entity) IsSelected() bool {
	return e.LogScanner != nil
}

func (e Entity) IsContainer() bool {
	return !e.IsCluster && !e.IsNamespace && !e.IsDeployment && !e.IsPod
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

func (e Entity) IsChildContainerOfDeployment(deployment Entity) bool {
	return e.IsContainer() && e.Container.inDeploymentOf(deployment.Container)
}

func (e Entity) IsChildContainerOfPod(pod Entity) bool {
	return e.IsContainer() && e.Container.inPodOf(pod.Container)
}

func (e Entity) Type() string {
	if e.IsCluster {
		return "cluster"
	} else if e.IsNamespace {
		return "namespace"
	} else if e.IsDeployment {
		return "deployment"
	} else if e.IsPod {
		return "pod"
	} else {
		return "container"
	}
}
