package model

import "time"

type ContainerStatus struct {
	State        ContainerState
	RunningSince time.Time
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
