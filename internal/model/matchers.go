package model

import (
	"fmt"
	"regexp"
)

type ValidRegex struct {
	*regexp.Regexp
}

func NewValidRegex(pattern string) (*ValidRegex, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &ValidRegex{re}, nil
}

type Selectors struct {
	isEmpty    bool
	cluster    *ValidRegex
	namespace  *ValidRegex
	deployment *ValidRegex
	pod        *ValidRegex
	container  *ValidRegex
}

type NewSelectorArgs struct {
	Cluster    string
	Namespace  string
	Deployment string
	Pod        string
	Container  string
}

func NewSelectors(args NewSelectorArgs) (*Selectors, error) {
	isEmpty := args.Cluster == "" && args.Namespace == "" && args.Deployment == "" && args.Pod == "" && args.Container == ""
	clusterRe, err := NewValidRegex(args.Cluster)
	if err != nil {
		return nil, fmt.Errorf("cluster: %v", err)
	}
	namespaceRe, err := NewValidRegex(args.Namespace)
	if err != nil {
		return nil, fmt.Errorf("namespace: %v", err)
	}
	deploymentRe, err := NewValidRegex(args.Deployment)
	if err != nil {
		return nil, fmt.Errorf("deployment: %v", err)
	}
	podRe, err := NewValidRegex(args.Pod)
	if err != nil {
		return nil, fmt.Errorf("pod: %v", err)
	}
	containerRe, err := NewValidRegex(args.Container)
	if err != nil {
		return nil, fmt.Errorf("container: %v", err)
	}
	return &Selectors{
		isEmpty:    isEmpty,
		cluster:    clusterRe,
		namespace:  namespaceRe,
		deployment: deploymentRe,
		pod:        podRe,
		container:  containerRe,
	}, nil
}

func (m Selectors) SelectContainer(c Container) bool {
	return !m.isEmpty &&
		m.cluster.MatchString(c.Cluster) &&
		m.namespace.MatchString(c.Namespace) &&
		m.deployment.MatchString(c.Deployment) &&
		m.pod.MatchString(c.Pod) &&
		m.container.MatchString(c.Name)
}
