package k8s_model

type ClusterNamespaces struct {
	Cluster    string
	Namespaces []string
}

type ContainerNameAndPrefix struct {
	Prefix        string
	ContainerName string
}

type PodOwnerMetadata struct {
	OwnerType string
}
