package domain

import "time"

// ContainerID uniquely identifies a container across clusters
type ContainerID struct {
	Cluster   string
	Namespace string
	Pod       string
	Container string
}

// Container represents a Kubernetes container with metadata
type Container struct {
	ID           ContainerID
	OwnerName    string
	OwnerType    string
	StartedAt    time.Time
	IsRunning    bool
	IsTerminated bool
}

// ContainerState represents the scanning state machine
type ContainerState int

const (
	StateInactive ContainerState = iota
	StateWantScanning
	StateScannerStarting
	StateScanning
	StateScannerStopping
	StateDeleted
)

// SelectableContainer pairs a container with its selection state
type SelectableContainer struct {
	Container Container
	State     ContainerState
}
