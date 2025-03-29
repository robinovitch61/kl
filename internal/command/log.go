package command

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/k8s/client"
	"github.com/robinovitch61/kl/internal/k8s/container"
	"github.com/robinovitch61/kl/internal/k8s/entity"
	"github.com/robinovitch61/kl/internal/k8s/k8s_log"
	"time"
)

type StartedLogScannerMsg struct {
	LogScanner k8s_log.LogScanner
	Err        error
}

func StartLogScannerCmd(
	client client.K8sClient,
	container container.Container,
	sinceTime time.Time,
) tea.Cmd {
	return func() tea.Msg {
		dev.Debug(fmt.Sprintf("cmd running to start log scanner for container %v", container.HumanReadable()))
		// update the container status just before getting a log stream in case status is not up to date
		status, err := client.GetContainerStatus(container)
		if err != nil {
			return StartedLogScannerMsg{
				LogScanner: k8s_log.LogScanner{Container: container},
				Err:        fmt.Errorf("error getting container status: %v", err),
			}
		}
		container.Status = status

		// attempt to create and start a log scanner from a k8s log stream
		scanner, cancel, err := client.GetLogStream(container, sinceTime)
		if err != nil {
			return StartedLogScannerMsg{
				LogScanner: k8s_log.LogScanner{Container: container},
				Err:        fmt.Errorf("error getting log stream: %v", err),
			}
		}
		ls := k8s_log.NewLogScanner(container, scanner, cancel)
		ls.StartReadingLogs()
		return StartedLogScannerMsg{LogScanner: ls}
	}
}

type GetNewLogsMsg struct {
	LogScanner   k8s_log.LogScanner
	NewLogs      []k8s_log.Log
	DoneScanning bool
	Err          error
}

func GetNextLogsCmd(ls k8s_log.LogScanner, duration time.Duration) tea.Cmd {
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
	collectedLogs []k8s_log.Log
	doneScanning  bool
	err           error
}

func collectLogsForDuration(ls k8s_log.LogScanner, duration time.Duration) collectedLogsResult {
	var collectedLogs []k8s_log.Log
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
			return collectedLogsResult{collectedLogs: collectedLogs, doneScanning: true, err: err}
		case <-timeout:
			return collectedLogsResult{collectedLogs: collectedLogs, doneScanning: false, err: nil}
		}
	}
}

type StoppedLogScannersMsg struct {
	Containers []container.Container
	Restart    bool
	KeepLogs   bool
}

// StopLogScannerCmd stops an Entity's LogScanner
func StopLogScannerCmd(entity entity.Entity, keepLogs bool) tea.Cmd {
	return func() tea.Msg {
		if entity.LogScanner != nil {
			entity.LogScanner.Cancel()
			return StoppedLogScannersMsg{Containers: []container.Container{entity.Container}, Restart: false, KeepLogs: keepLogs}
		}
		return StoppedLogScannersMsg{Containers: []container.Container{}, Restart: false, KeepLogs: keepLogs}
	}
}

func StopLogScannersInPrepForNewSinceTimeCmd(logScanners []k8s_log.LogScanner) tea.Cmd {
	return func() tea.Msg {
		var specs []container.Container
		for _, ls := range logScanners {
			ls.Cancel()
			specs = append(specs, ls.Container)
		}
		return StoppedLogScannersMsg{Containers: specs, Restart: true, KeepLogs: false}
	}
}
