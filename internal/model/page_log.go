package model

import (
	"time"

	"charm.land/lipgloss/v2"
	"github.com/emirpasic/gods/trees/redblacktree"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/k8s/k8s_log"
	"github.com/robinovitch61/kl/internal/k8s/k8s_model"
	"github.com/robinovitch61/kl/internal/style"
	"github.com/robinovitch61/viewport/viewport/item"
)

type PageLogContainerNames struct {
	Short k8s_model.ContainerNameAndPrefix
	Full  k8s_model.ContainerNameAndPrefix
}

// PageLog is a Log with metadata. It has mostly pointer fields for efficient copying
type PageLog struct {
	Log              *k8s_log.Log
	ContainerNames   *PageLogContainerNames
	CurrentName      *k8s_model.ContainerNameAndPrefix
	CurrentTimestamp string
	Terminated       bool
	Theme            *style.Theme
	PrettyPrinted    bool
	prettyItem       item.Item // cached MultiLineItem
}

func (l PageLog) GetItem() item.Item {
	if l.PrettyPrinted && l.prettyItem != nil {
		return l.prettyItem
	}
	prefix := l.renderPrefix(true)
	if prefix == "" {
		return l.Log.ContentItem
	}
	return item.NewConcat(item.NewItem(prefix), l.Log.ContentItem)
}

// BuildPrettyItemWithPrefix assembles the pretty-printed MultiLineItem for this log.
func (l *PageLog) BuildPrettyItemWithPrefix() {
	cached := l.Log.PrettyItems
	if cached == nil {
		l.prettyItem = nil
		return
	}

	// only segment 0 depends on the prefix — the rest are reused from the Log cache
	prefix := l.renderPrefix(true)
	segments := make([]item.SingleItem, len(cached))
	if prefix != "" {
		segments[0] = item.NewItem(prefix + cached[0].Content())
	} else {
		segments[0] = cached[0]
	}
	copy(segments[1:], cached[1:])
	l.prettyItem = item.NewMultiLineItem(segments...)
}

// renderPrefix returns the styled prefix (timestamp + container name + trailing space if needed)
func (l PageLog) renderPrefix(includeStyle bool) string {
	ts := ""
	if l.CurrentTimestamp != "" {
		if includeStyle && l.Theme != nil {
			ts = l.Theme.TimestampPrefix.Render(l.CurrentTimestamp)
		} else {
			ts = l.CurrentTimestamp
		}
	}
	label := ""
	if l.CurrentName != nil && l.CurrentName.ContainerName != "" {
		if ts != "" {
			label += " "
		}
		label += l.RenderName(*l.CurrentName, includeStyle)
	}

	prefix := ts + label
	if len(prefix) > 0 {
		if l.Log.ContentItem.Content() != "" {
			prefix = prefix + " "
		}
	}
	return prefix
}

func (l PageLog) Equals(other interface{}) bool {
	otherLog, ok := other.(PageLog)
	if !ok {
		return false
	}
	if l.Log == nil || otherLog.Log == nil {
		return false
	}
	// TODO LEO: make this method on Log
	return l.Log.ContentItem.Content() == otherLog.Log.ContentItem.Content() && l.Log.Timestamps.Full == otherLog.Log.Timestamps.Full
}

// RenderName renders a container name for display in log lines.
// When includeStyle is true, applies two deterministic foreground colors:
// one for the prefix (hashed from the full prefix for consistency across
// short and full display modes) and one for the container name (hashed from
// just the name, so e.g. "mycontainer" is always the same color).
func (l PageLog) RenderName(name k8s_model.ContainerNameAndPrefix, includeStyle bool) string {
	if includeStyle && l.Theme != nil {
		renderedName := l.Theme.ContainerColorStyle(name.ContainerName).Render(name.ContainerName)
		if lipgloss.Width(name.Prefix) == 0 {
			return renderedName
		}
		// always hash from the full prefix so short and full display modes get the same color
		colorKey := name.Prefix
		if l.ContainerNames != nil && l.ContainerNames.Full.Prefix != "" {
			colorKey = l.ContainerNames.Full.Prefix
		}
		renderedPrefix := l.Theme.ContainerColorStyle(colorKey).Render(name.Prefix)
		return renderedPrefix + "/" + renderedName
	}
	if lipgloss.Width(name.Prefix) == 0 {
		return name.ContainerName
	}
	return name.Prefix + "/" + name.ContainerName
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

func (lc PageLogContainer) Len() int {
	return lc.allLogs.Size()
}

// LastTimestamp returns the timestamp of the last log in the current ordering.
// In ascending mode this is the max timestamp; in descending mode the min.
func (lc PageLogContainer) LastTimestamp() (time.Time, bool) {
	node := lc.allLogs.Right()
	if node == nil {
		return time.Time{}, false
	}
	return node.Key.(PageLog).Log.Timestamp, true
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
	allKeys := lc.allLogs.Keys()
	for i := range allKeys {
		allLogs = append(allLogs, allKeys[i].(PageLog))
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
