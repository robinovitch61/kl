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

type Matchers struct {
	AutoSelectMatcher Matcher
	IgnoreMatcher     Matcher
}

type Matcher struct {
	isEmpty   bool
	cluster   *ValidRegex
	namespace *ValidRegex
	podOwner  *ValidRegex
	pod       *ValidRegex
	container *ValidRegex
}

type NewMatcherArgs struct {
	Cluster   string
	Container string
	PodOwner  string
	Namespace string
	Pod       string
}

func NewMatcher(args NewMatcherArgs) (*Matcher, error) {
	isEmpty := args.Cluster == "" && args.Namespace == "" && args.PodOwner == "" && args.Pod == "" && args.Container == ""
	clusterRe, err := NewValidRegex(args.Cluster)
	if err != nil {
		return nil, fmt.Errorf("cluster: %v", err)
	}
	namespaceRe, err := NewValidRegex(args.Namespace)
	if err != nil {
		return nil, fmt.Errorf("namespace: %v", err)
	}
	podOwnerRe, err := NewValidRegex(args.PodOwner)
	if err != nil {
		return nil, fmt.Errorf("podOwner: %v", err)
	}
	podRe, err := NewValidRegex(args.Pod)
	if err != nil {
		return nil, fmt.Errorf("pod: %v", err)
	}
	containerRe, err := NewValidRegex(args.Container)
	if err != nil {
		return nil, fmt.Errorf("container: %v", err)
	}
	return &Matcher{
		isEmpty:   isEmpty,
		cluster:   clusterRe,
		namespace: namespaceRe,
		podOwner:  podOwnerRe,
		pod:       podRe,
		container: containerRe,
	}, nil
}

func (m Matcher) MatchesContainer(c Container) bool {
	return !m.isEmpty &&
		m.cluster.MatchString(c.Cluster) &&
		m.namespace.MatchString(c.Namespace) &&
		m.podOwner.MatchString(c.PodOwner) &&
		m.pod.MatchString(c.Pod) &&
		m.container.MatchString(c.Name)
}
