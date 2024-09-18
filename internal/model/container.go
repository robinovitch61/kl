package model

import (
	"strings"
)

const idSeparator = "/"

type Container struct {
	Cluster, Namespace, Deployment, Pod, Name string
	Status                                    ContainerStatus
}

func (c Container) ID() string {
	return strings.Join([]string{c.Cluster, c.Namespace, c.Deployment, c.Pod, c.Name}, idSeparator)
}

func (c Container) HumanReadable() string {
	entries := strings.Split(c.ID(), idSeparator)
	var nonEmptyEntries []string
	for _, entry := range entries {
		if strings.TrimSpace(entry) != "" {
			nonEmptyEntries = append(nonEmptyEntries, entry)
		}
	}
	return strings.Join(nonEmptyEntries, idSeparator)
}

func (c Container) Equals(other Container) bool {
	return c.ID() == other.ID()
}

func (c Container) inClusterOf(other Container) bool {
	return c.Cluster == other.Cluster
}

func (c Container) inNamespaceOf(other Container) bool {
	return c.Cluster == other.Cluster && c.Namespace == other.Namespace
}

func (c Container) inDeploymentOf(other Container) bool {
	return c.Cluster == other.Cluster && c.Namespace == other.Namespace && c.Deployment == other.Deployment
}

func (c Container) inPodOf(other Container) bool {
	return c.Cluster == other.Cluster && c.Namespace == other.Namespace && c.Deployment == other.Deployment && c.Pod == other.Pod
}
