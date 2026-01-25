package logscanner

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/kl/internal/domain"
	"github.com/robinovitch61/kl/internal/k8s"
	"github.com/robinovitch61/kl/internal/tree"
)

// LogBatchMsg contains logs collected from a scanner
type LogBatchMsg struct {
	Logs        []domain.Log
	ContainerID domain.ContainerID
}

// StopReason indicates why a scanner stopped
type StopReason int

const (
	ReasonUserDeselected StopReason = iota
	ReasonContainerTerminated
	ReasonContainerDeleted
	ReasonTimeRangeChanged
	ReasonError
)

// ScannerStoppedMsg indicates a scanner has stopped
type ScannerStoppedMsg struct {
	ContainerID domain.ContainerID
	Reason      StopReason
}

// Coordinator manages all log scanners
type Coordinator struct {
	manager   *k8s.Manager
	timeRange domain.TimeRange
	scanners  map[domain.ContainerID]bool // tracks active scanners
}

// NewCoordinator creates a new scanner coordinator
func NewCoordinator(manager *k8s.Manager, timeRange domain.TimeRange) *Coordinator {
	return &Coordinator{
		manager:   manager,
		timeRange: timeRange,
		scanners:  make(map[domain.ContainerID]bool),
	}
}

// HandleStateChange processes a container state transition
// Returns tea.Cmd to start/stop scanners as needed
func (c *Coordinator) HandleStateChange(change tree.StateChange) tea.Cmd {
	// TODO: implement
	return nil
}

// SetTimeRange changes the time range, returns command to restart scanners
func (c *Coordinator) SetTimeRange(tr domain.TimeRange) tea.Cmd {
	// TODO: implement
	c.timeRange = tr
	return nil
}

// Shutdown stops all scanners
func (c *Coordinator) Shutdown() tea.Cmd {
	// TODO: implement
	return nil
}
