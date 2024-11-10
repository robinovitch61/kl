package command

import (
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/kl/internal/dev"
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
		dev.Debug(fmt.Sprintf("cmd running to start log scanner for container %v", container.HumanReadable()))
		// update the container status just before getting a log stream in case status is not up to date
		status, err := client.GetContainerStatus(container)
		if err != nil {
			return StartedLogScannerMsg{
				LogScanner: model.LogScanner{Container: container},
				Err:        fmt.Errorf("error getting container status: %v", err),
			}
		}
		container.Status = status

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
			logs := collectLogsForDuration(ls, duration)
			if logs.err != nil {
				return GetNewLogsMsg{LogScanner: ls, Err: logs.err}
			}
			if len(logs.collectedLogs) > 0 || logs.doneScanning {
				return GetNewLogsMsg{LogScanner: ls, NewLogs: logs.collectedLogs, DoneScanning: logs.doneScanning}
			}
		}
	}
}

type collectedLogsResult struct {
	collectedLogs []model.Log
	doneScanning  bool
	err           error
}

func collectLogsForDuration(ls model.LogScanner, duration time.Duration) collectedLogsResult {
	var collectedLogs []model.Log
	timeout := time.After(duration)

	for {
		select {
		case log, ok := <-ls.LogChan:
			if !ok {
				// channel is closed
				return collectedLogsResult{collectedLogs: collectedLogs, doneScanning: true, err: nil}
			}
			collectedLogs = append(collectedLogs, log)
		case err := <-ls.ErrChan:
			if errors.Is(err, errtype.LogScannerStoppedErr{}) {
				return collectedLogsResult{collectedLogs: collectedLogs, doneScanning: true, err: nil}
			}
			return collectedLogsResult{collectedLogs: collectedLogs, doneScanning: true, err: err}
		case <-timeout:
			return collectedLogsResult{collectedLogs: collectedLogs, doneScanning: false, err: nil}
		}
	}
}

type StoppedLogScannersMsg struct {
	Containers []model.Container
	Restart    bool
	KeepLogs   bool
}

// StopLogScannerCmd stops an Entity's LogScanner
func StopLogScannerCmd(entity model.Entity, keepLogs bool) tea.Cmd {
	return func() tea.Msg {
		if entity.LogScanner != nil {
			entity.LogScanner.Cancel()
			return StoppedLogScannersMsg{Containers: []model.Container{entity.Container}, Restart: false, KeepLogs: keepLogs}
		}
		return StoppedLogScannersMsg{Containers: []model.Container{}, Restart: false, KeepLogs: keepLogs}
	}
}

func StopLogScannersInPrepForNewSinceTimeCmd(logScanners []model.LogScanner) tea.Cmd {
	return func() tea.Msg {
		var specs []model.Container
		for _, ls := range logScanners {
			ls.Cancel()
			specs = append(specs, ls.Container)
		}
		return StoppedLogScannersMsg{Containers: specs, Restart: true, KeepLogs: false}
	}
}
