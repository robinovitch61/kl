package model

type ContainerListener struct {
	Cluster            string
	Namespace          string
	ContainerDeltaChan chan ContainerDelta
	Stop               func()
}
