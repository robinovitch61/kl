package internal

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/robinovitch61/kl/internal/command"
	"github.com/robinovitch61/kl/internal/k8s/container"
	"github.com/robinovitch61/kl/internal/k8s/entity"
	"github.com/robinovitch61/kl/internal/k8s/k8s_log"
	"github.com/robinovitch61/kl/internal/k8s/k8s_model"
	"github.com/robinovitch61/kl/internal/keymap"
	"github.com/robinovitch61/kl/internal/message"
	"github.com/robinovitch61/kl/internal/model"
	"github.com/robinovitch61/kl/internal/page"
	"github.com/robinovitch61/kl/internal/style"
	"github.com/robinovitch61/viewport/viewport/item"
)

func newTestModel() Model {
	km := keymap.DefaultKeyMap()
	theme := style.PickTheme("none")
	clusterNamespaces := []k8s_model.ClusterNamespaces{
		{Cluster: "test-cluster", Namespaces: []string{"default"}},
	}
	entityTree := entity.NewEntityTree(clusterNamespaces)

	width, height := 120, 40

	m := InitialModel(Config{
		ContainerLimit: -1,
		SinceTime:      model.NewSinceTime(time.Time{}, -1),
		Version:        "test",
	})
	m.state.width = width
	m.state.height = height
	m.state.initialized = true
	m.state.gotFirstContainers = true
	m.state.seenFirstContainer = true
	m.state.focusedPageType = page.EntitiesPageType
	m.state.rightPageType = page.LogsPageType
	m.data.theme = theme
	m.entityTree = entityTree

	contentHeight := height - 1
	m.pages = make(map[page.Type]page.GenericPage)
	m.pages[page.EntitiesPageType] = page.NewEntitiesPage(km, width, contentHeight, entityTree, theme)
	m.pages[page.LogsPageType] = page.NewLogsPage(km, width, contentHeight, false, theme)
	m.pages[page.SingleLogPageType] = page.NewSingleLogPage(km, width, contentHeight, theme)
	m.data.topBarHeight = 1

	m.pages[m.state.focusedPageType] = m.pages[m.state.focusedPageType].WithFocus()

	return m
}

func newAppTestContainer() container.Container {
	return container.Container{
		Cluster:   "test-cluster",
		Namespace: "default",
		PodOwner:  "my-app",
		Pod:       "my-app-abc123",
		Name:      "web",
		Status:    container.ContainerStatus{State: container.ContainerRunning},
	}
}

func newAppTestDelta(ct container.Container, toActivate bool) container.ContainerDelta {
	return container.ContainerDelta{
		Time:       time.Now(),
		Container:  ct,
		ToActivate: toActivate,
	}
}

func updateModel(t *testing.T, m Model, msg tea.Msg) Model {
	t.Helper()
	updated, _ := m.Update(msg)
	return updated.(Model)
}

func TestErrMsg_ShowsErrorScreen(t *testing.T) {
	m := newTestModel()

	m = updateModel(t, m, message.ErrMsg{Err: fmt.Errorf("something broke")})

	if m.state.err == nil {
		t.Fatal("expected err to be set")
	}

	view := m.View().Content
	if !strings.Contains(view, "something broke") {
		t.Errorf("expected view to contain error text, got:\n%s", view)
	}
	if !strings.Contains(view, "Error at") {
		t.Errorf("expected view to contain 'Error at' timestamp, got:\n%s", view)
	}
	if !strings.Contains(view, "ctrl+c to quit") {
		t.Errorf("expected view to contain quit instruction, got:\n%s", view)
	}
}

func TestScannerError_RestartsWithoutErrorScreen(t *testing.T) {
	m := newTestModel()

	// add a container via delta
	ct := newAppTestContainer()
	var deltaSet container.ContainerDeltaSet
	deltaSet.Add(newAppTestDelta(ct, true))
	m = updateModel(t, m, command.GetContainerDeltasMsg{DeltaSet: deltaSet})

	// entity should now be in ScannerStarting state
	ent := m.entityTree.GetEntity(ct)
	if ent == nil {
		t.Fatal("expected entity to exist in tree")
	}
	if ent.State != entity.ScannerStarting {
		t.Fatalf("expected ScannerStarting, got %v", ent.State)
	}

	// simulate scanner started successfully
	_, cancel := context.WithCancel(context.Background())
	scanner := k8s_log.NewLogScanner(ct, nil, cancel)
	m = updateModel(t, m, command.StartedLogScannerMsg{LogScanner: scanner})

	// entity should now be Scanning
	ent = m.entityTree.GetEntity(ct)
	if ent == nil {
		t.Fatal("expected entity to exist in tree")
	}
	if ent.State != entity.Scanning {
		t.Fatalf("expected Scanning, got %v", ent.State)
	}

	// send a scanner error (simulating connection reset)
	m = updateModel(t, m, command.GetNewLogsMsg{
		LogScanner: scanner,
		Err:        fmt.Errorf("read tcp: read: connection reset by peer"),
	})

	// should NOT go to error screen
	if m.state.err != nil {
		t.Fatalf("expected no error screen, got: %v", m.state.err)
	}

	// entity should be restarting (ScannerStarting)
	ent = m.entityTree.GetEntity(ct)
	if ent == nil {
		t.Fatal("expected entity to still exist in tree")
	}
	if ent.State != entity.ScannerStarting {
		t.Fatalf("expected ScannerStarting (restarted), got %v", ent.State)
	}

	// view should not show error screen
	view := m.View().Content
	if strings.Contains(view, "Error at") {
		t.Errorf("expected no error screen in view, got:\n%s", view)
	}
}

func TestContainerSelection_LogsAppearInView(t *testing.T) {
	m := newTestModel()

	// add a container with auto-activation
	ct := newAppTestContainer()
	var deltaSet container.ContainerDeltaSet
	deltaSet.Add(newAppTestDelta(ct, true))
	m = updateModel(t, m, command.GetContainerDeltasMsg{DeltaSet: deltaSet})

	// simulate scanner started successfully
	_, cancel := context.WithCancel(context.Background())
	scanner := k8s_log.NewLogScanner(ct, nil, cancel)
	m = updateModel(t, m, command.StartedLogScannerMsg{LogScanner: scanner})

	// send logs with recognizable content
	now := time.Now()
	m = updateModel(t, m, command.GetNewLogsMsg{
		LogScanner: scanner,
		NewLogs: []k8s_log.Log{
			{
				Timestamp:   now,
				Container:   ct,
				ContentItem: item.NewItem("hello from the container"),
			},
			{
				Timestamp:   now.Add(time.Second),
				Container:   ct,
				ContentItem: item.NewItem("second log line here"),
			},
		},
	})

	// flush the log buffer to the logs page
	m = updateModel(t, m, message.BatchUpdateLogsMsg{})

	view := m.View().Content
	if !strings.Contains(view, "hello from the container") {
		t.Errorf("expected view to contain first log line, got:\n%s", view)
	}
	if !strings.Contains(view, "second log line here") {
		t.Errorf("expected view to contain second log line, got:\n%s", view)
	}
}

func TestContainerArrival_ShowsInEntityView(t *testing.T) {
	m := newTestModel()

	ct := newAppTestContainer()
	var deltaSet container.ContainerDeltaSet
	deltaSet.Add(newAppTestDelta(ct, false))
	m = updateModel(t, m, command.GetContainerDeltasMsg{DeltaSet: deltaSet})

	view := m.View().Content
	if !strings.Contains(view, "web") {
		t.Errorf("expected view to contain container name 'web', got:\n%s", view)
	}
	if !strings.Contains(view, "my-app") {
		t.Errorf("expected view to contain pod owner 'my-app', got:\n%s", view)
	}
	if !strings.Contains(view, "test-cluster") {
		t.Errorf("expected view to contain cluster name 'test-cluster', got:\n%s", view)
	}
}
