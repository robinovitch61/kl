package container

import "time"

type ContainerStatus struct {
	State         ContainerState
	StartedAt     time.Time
	TerminatedAt  time.Time
	WaitingFor    string
	TerminatedFor string
}

type ContainerState int

const (
	ContainerUnknown ContainerState = iota
	ContainerRunning
	ContainerTerminated
	ContainerWaiting
)

func (s ContainerState) String() string {
	switch s {
	case ContainerRunning:
		return "running"
	case ContainerTerminated:
		return "terminated"
	case ContainerWaiting:
		return "waiting"
	default:
		return "unknown"
	}
}
