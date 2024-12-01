package model

import (
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/emirpasic/gods/trees/redblacktree"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/style"
)

type PageLogContainerName struct {
	Prefix        string
	ContainerName string
}

type PageLogContainerNames struct {
	Short PageLogContainerName
	Full  PageLogContainerName
}

type PageLogTimestamps struct {
	Short string
	Full  string
}

type PageLog struct {
	Log              Log
	ContainerColors  ContainerColors
	ContainerNames   PageLogContainerNames
	CurrentName      PageLogContainerName
	Timestamps       PageLogTimestamps
	CurrentTimestamp string
	Terminated       bool
	Styles           *style.Styles
}

func (l PageLog) Render() string {
	ts := ""
	if l.CurrentTimestamp != "" {
		ts = l.Styles.Green.Render(l.CurrentTimestamp)
	}
	label := ""
	if l.CurrentName.ContainerName != "" {
		if ts != "" {
			label += " "
		}
		label += l.RenderName(l.CurrentName, true)
	}
	prefix := ts + label
	if len(prefix) > 0 {
		if l.Log.Content != "" {
			prefix = prefix + " "
		}
	}
	return prefix + l.Log.Content
}

func (l PageLog) Equals(other interface{}) bool {
	otherLog, ok := other.(PageLog)
	if !ok {
		return false
	}
	return l.Log == otherLog.Log && l.Timestamps.Full == otherLog.Timestamps.Full
}

func (l PageLog) RenderName(name PageLogContainerName, includeStyle bool) string {
	var renderedPrefix, renderedName string
	if includeStyle {
		renderedPrefix = lipgloss.NewStyle().Background(l.ContainerColors.ID).Foreground(lipgloss.Color("#000000")).Render(name.Prefix)
		renderedName = lipgloss.NewStyle().Background(l.ContainerColors.Name).Foreground(lipgloss.Color("#000000")).Render(name.ContainerName)
	} else {
		renderedPrefix = name.Prefix
		renderedName = name.ContainerName
	}
	if lipgloss.Width(renderedPrefix) == 0 {
		return renderedName
	}
	return renderedPrefix + "/" + renderedName
}

type PageLogContainer struct {
	allLogs   *redblacktree.Tree
	ascending bool
}

func pageLogComparatorAsc(a, b interface{}) int {
	e1 := a.(PageLog)
	e2 := b.(PageLog)
	switch {
	case e1.Log.Timestamp.Before(e2.Log.Timestamp):
		return -1
	case e1.Log.Timestamp.After(e2.Log.Timestamp):
		return 1
	default:
		return 0
	}
}

func pageLogComparatorDesc(a, b interface{}) int {
	return -pageLogComparatorAsc(a, b)
}

func NewPageLogContainer(ascending bool) *PageLogContainer {
	comparator := pageLogComparatorAsc
	if !ascending {
		comparator = pageLogComparatorDesc
	}
	return &PageLogContainer{
		allLogs:   redblacktree.NewWith(comparator),
		ascending: ascending,
	}
}

func (lc *PageLogContainer) AppendLog(log PageLog, _ interface{}) {
	lc.allLogs.Put(log, nil)
}

func (lc *PageLogContainer) RemoveAllLogs() {
	lc.allLogs.Clear()
}

func (lc PageLogContainer) GetOrderedLogs() []PageLog {
	var allLogs []PageLog
	dev.Debug("iterating logs")
	defer dev.Debug("done iterating logs")
	for _, k := range lc.allLogs.Keys() {
		allLogs = append(allLogs, k.(PageLog))
	}
	return allLogs
}

func (lc PageLogContainer) Ascending() bool {
	return lc.ascending
}

func (lc *PageLogContainer) ToggleAscending() {
	if lc.ascending {
		lc.allLogs.Comparator = pageLogComparatorDesc
	} else {
		lc.allLogs.Comparator = pageLogComparatorAsc
	}
	lc.ascending = !lc.ascending
}
