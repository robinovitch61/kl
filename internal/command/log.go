package command

import (
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/kl/internal/errtype"
	"github.com/robinovitch61/kl/internal/k8s"
	"github.com/robinovitch61/kl/internal/model"
	"time"
)

type StartedLogScannerMsg struct {
	LogScanner model.LogScanner
	Err        error
}

func StartLogScannerCmd(
	client k8s.Client,
	container model.Container,
	sinceTime time.Time,
) tea.Cmd {
	return func() tea.Msg {
		// update the container status just before getting a log stream in case status is not up to date
		status, err := client.GetContainerStatus(container)
		if err != nil {
			return StartedLogScannerMsg{
				LogScanner: model.LogScanner{Container: container},
				Err:        fmt.Errorf("error getting container status: %v", err),
			}
		}
		container.Status = status

		// exit early if the container is not running
		if status.State != model.ContainerRunning {
			return StartedLogScannerMsg{
				LogScanner: model.LogScanner{Container: container},
				Err:        fmt.Errorf("container %s is not running", container.Name),
			}
		}

		// attempt to create and start a log scanner from a k8s log stream
		scanner, cancel, err := client.GetLogStream(container, sinceTime)
		if err != nil {
			return StartedLogScannerMsg{
				LogScanner: model.LogScanner{Container: container},
				Err:        fmt.Errorf("error getting log stream: %v", err),
			}
		}
		ls := model.NewLogScanner(container, scanner, cancel)
		ls.StartReadingLogs()
		return StartedLogScannerMsg{LogScanner: ls}
	}
}

type GetNewLogsMsg struct {
	LogScanner   model.LogScanner
	NewLogs      []model.Log
	DoneScanning bool
	Err          error
}

func GetNextLogsCmd(ls model.LogScanner, duration time.Duration) tea.Cmd {
	return func() tea.Msg {
		for {
			collectedLogs, doneScanning, err := collectLogsForDuration(ls, duration)
			if err != nil {
				return GetNewLogsMsg{LogScanner: ls, Err: err}
			}
			if len(collectedLogs) > 0 || doneScanning {
				return GetNewLogsMsg{LogScanner: ls, NewLogs: collectedLogs, DoneScanning: doneScanning}
			}
		}
	}
}

func collectLogsForDuration(ls model.LogScanner, duration time.Duration) ([]model.Log, bool, error) {
	var collectedLogs []model.Log
	timeout := time.After(duration)

	for {
		select {
		case log := <-ls.LogChan:
			collectedLogs = append(collectedLogs, log)
		case err := <-ls.ErrChan:
			if errors.Is(err, errtype.LogScannerStoppedErr{}) {
				return collectedLogs, true, nil
			}
			return nil, false, err
		case <-timeout:
			return collectedLogs, false, nil
		}
	}
}

type LogScannersStoppedMsg struct {
	Containers []model.Container
	Restart    bool
	KeepLogs   bool
}

// StopLogScannerCmd stops an Entity's LogScanner
func StopLogScannerCmd(entity model.Entity, keepLogs bool) tea.Cmd {
	return func() tea.Msg {
		if entity.IsSelected() {
			entity.LogScanner.Cancel()
			return LogScannersStoppedMsg{Containers: []model.Container{entity.Container}, Restart: false, KeepLogs: keepLogs}
		}
		return LogScannersStoppedMsg{Containers: []model.Container{}, Restart: false, KeepLogs: keepLogs}
	}
}

func StopLogScannersInPrepForNewLookbackCmd(logScanners []model.LogScanner) tea.Cmd {
	return func() tea.Msg {
		var specs []model.Container
		for _, ls := range logScanners {
			ls.Cancel()
			specs = append(specs, ls.Container)
		}
		return LogScannersStoppedMsg{Containers: specs, Restart: true}
	}
}
