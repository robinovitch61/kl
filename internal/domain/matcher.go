package domain

import (
	"regexp"
	"time"
)

// Matcher holds regex patterns for matching container attributes
type Matcher struct {
	Cluster   *regexp.Regexp
	Namespace *regexp.Regexp
	PodOwner  *regexp.Regexp
	Pod       *regexp.Regexp
	Container *regexp.Regexp
}

// NewMatcherArgs holds the string patterns for creating a Matcher
type NewMatcherArgs struct {
	Cluster   string
	Namespace string
	PodOwner  string
	Pod       string
	Container string
}

// NewMatcher creates a Matcher from string patterns
func NewMatcher(args NewMatcherArgs) (*Matcher, error) {
	m := &Matcher{}
	var err error

	if args.Cluster != "" {
		m.Cluster, err = regexp.Compile(args.Cluster)
		if err != nil {
			return nil, err
		}
	}
	if args.Namespace != "" {
		m.Namespace, err = regexp.Compile(args.Namespace)
		if err != nil {
			return nil, err
		}
	}
	if args.PodOwner != "" {
		m.PodOwner, err = regexp.Compile(args.PodOwner)
		if err != nil {
			return nil, err
		}
	}
	if args.Pod != "" {
		m.Pod, err = regexp.Compile(args.Pod)
		if err != nil {
			return nil, err
		}
	}
	if args.Container != "" {
		m.Container, err = regexp.Compile(args.Container)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

// Matchers holds both auto-select and ignore matchers
type Matchers struct {
	AutoSelectMatcher Matcher
	IgnoreMatcher     Matcher
}

// LogFilter represents a text or regex filter for logs
type LogFilter struct {
	Value   string
	IsRegex bool
}

// SinceTime wraps a start time with lookback context
type SinceTime struct {
	StartTime time.Time
	Minutes   int
}

// NewSinceTime creates a SinceTime from a start time and lookback minutes
func NewSinceTime(startTime time.Time, minutes int) SinceTime {
	return SinceTime{StartTime: startTime, Minutes: minutes}
}
