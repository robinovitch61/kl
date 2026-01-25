package domain

import (
	"sort"
	"time"
)

// Log represents a single log entry
type Log struct {
	Timestamp    time.Time
	ContainerID  ContainerID
	Content      string
	IsTerminated bool
}

// Logs is a sortable slice of Log
type Logs []Log

// SortAscending returns logs sorted oldest first
func (l Logs) SortAscending() Logs {
	result := make(Logs, len(l))
	copy(result, l)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})
	return result
}

// SortDescending returns logs sorted newest first
func (l Logs) SortDescending() Logs {
	result := make(Logs, len(l))
	copy(result, l)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.After(result[j].Timestamp)
	})
	return result
}
