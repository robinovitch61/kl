package model

import (
	"github.com/charmbracelet/lipgloss"
	"strings"
)

const idSeparator = "/"

type Container struct {
	Cluster, Namespace, PodOwner, Pod, Name string
	Status                                  ContainerStatus
	PodOwnerMetadata                        PodOwnerMetadata
}

func (c Container) ID() string {
	return strings.Join([]string{c.Cluster, c.Namespace, c.PodOwner, c.Pod, c.Name}, idSeparator)
}

func (c Container) IDWithoutContainerName() string {
	return strings.Join([]string{c.Cluster, c.Namespace, c.PodOwner, c.Pod}, idSeparator)
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

func (c Container) inPodOwnerOf(other Container) bool {
	return c.Cluster == other.Cluster && c.Namespace == other.Namespace && c.PodOwner == other.PodOwner
}

func (c Container) inPodOf(other Container) bool {
	return c.Cluster == other.Cluster && c.Namespace == other.Namespace && c.PodOwner == other.PodOwner && c.Pod == other.Pod
}

type ContainerColors struct {
	// the entire container's ID (full specification)
	ID lipgloss.Color
	// just the container name
	Name lipgloss.Color
}
