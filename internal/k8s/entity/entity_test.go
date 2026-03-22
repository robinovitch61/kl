package entity_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/robinovitch61/kl/internal/k8s/container"
	"github.com/robinovitch61/kl/internal/k8s/entity"
	"github.com/robinovitch61/kl/internal/k8s/k8s_log"
	"github.com/robinovitch61/kl/internal/k8s/k8s_model"
)

func newTestEntity(state entity.EntityState, containerState container.ContainerState) entity.Entity {
	return entity.Entity{
		Container: container.Container{
			Cluster:   "cluster1",
			Namespace: "namespace1",
			PodOwner:  "podOwner1",
			Pod:       "pod1",
			Name:      "container1",
			Status:    container.ContainerStatus{State: containerState},
		},
		State: state,
	}
}

func newTestTree() entity.Tree {
	return entity.NewEntityTree([]k8s_model.ClusterNamespaces{
		{Cluster: "cluster1", Namespaces: []string{"namespace1"}},
	})
}

func newTestScanner() k8s_log.LogScanner {
	_, cancel := context.WithCancel(context.Background())
	return k8s_log.NewLogScanner(
		container.Container{
			Cluster:   "cluster1",
			Namespace: "namespace1",
			PodOwner:  "podOwner1",
			Pod:       "pod1",
			Name:      "container1",
			Status:    container.ContainerStatus{State: container.ContainerRunning},
		},
		nil,
		cancel,
		nil,
	)
}

func newTestDelta(containerState container.ContainerState, toDelete, toActivate bool) container.ContainerDelta {
	return container.ContainerDelta{
		Time: time.Now(),
		Container: container.Container{
			Cluster:   "cluster1",
			Namespace: "namespace1",
			PodOwner:  "podOwner1",
			Pod:       "pod1",
			Name:      "container1",
			Status:    container.ContainerStatus{State: containerState},
		},
		ToDelete:   toDelete,
		ToActivate: toActivate,
	}
}

func assertState(t *testing.T, ent entity.Entity, expected entity.EntityState) {
	t.Helper()
	if ent.State != expected {
		t.Errorf("expected state %v, got %v", expected, ent.State)
	}
}

func assertActions(t *testing.T, got []entity.EntityAction, expected []entity.EntityAction) {
	t.Helper()
	if len(got) != len(expected) {
		t.Errorf("expected %d actions %v, got %d actions %v", len(expected), expected, len(got), got)
		return
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("action[%d]: expected %v, got %v", i, expected[i], got[i])
		}
	}
}

func assertPanics(t *testing.T, name string, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s should have panicked", name)
		}
	}()
	f()
}

// --- Activate ---

func TestActivate_FromInactive_WaitingContainer(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.Inactive, container.ContainerWaiting)
	tree.AddOrReplace(ent)

	result, _, actions := ent.Activate(tree)

	assertState(t, result, entity.WantScanning)
	assertActions(t, actions, []entity.EntityAction{})
}

func TestActivate_FromInactive_RunningContainer(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.Inactive, container.ContainerRunning)
	tree.AddOrReplace(ent)

	result, _, actions := ent.Activate(tree)

	assertState(t, result, entity.ScannerStarting)
	assertActions(t, actions, []entity.EntityAction{entity.StartScanner})
}

func TestActivate_FromInactive_TerminatedContainer(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.Inactive, container.ContainerTerminated)
	tree.AddOrReplace(ent)

	result, _, actions := ent.Activate(tree)

	assertState(t, result, entity.ScannerStarting)
	assertActions(t, actions, []entity.EntityAction{entity.StartScanner})
}

func TestActivate_FromInvalidState_Panics(t *testing.T) {
	for _, state := range []entity.EntityState{entity.WantScanning, entity.ScannerStarting, entity.Scanning, entity.ScannerStopping, entity.Deleted} {
		t.Run(state.String(), func(t *testing.T) {
			tree := newTestTree()
			ent := newTestEntity(state, container.ContainerRunning)
			tree.AddOrReplace(ent)
			assertPanics(t, fmt.Sprintf("Activate from %v", state), func() {
				ent.Activate(tree)
			})
		})
	}
}

// --- Deactivate ---

func TestDeactivate_FromWantScanning(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.WantScanning, container.ContainerWaiting)
	tree.AddOrReplace(ent)

	result, _, actions := ent.Deactivate(tree)

	assertState(t, result, entity.Inactive)
	assertActions(t, actions, []entity.EntityAction{entity.RemoveLogs})
}

func TestDeactivate_FromScanning(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.Scanning, container.ContainerRunning)
	tree.AddOrReplace(ent)

	result, _, actions := ent.Deactivate(tree)

	assertState(t, result, entity.ScannerStopping)
	assertActions(t, actions, []entity.EntityAction{entity.StopScanner})
}

func TestDeactivate_FromDeleted(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.Deleted, container.ContainerTerminated)
	tree.AddOrReplace(ent)

	_, _, actions := ent.Deactivate(tree)

	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d: %v", len(actions), actions)
	}
	hasRemoveLogs := false
	hasRemoveEntity := false
	for _, a := range actions {
		if a == entity.RemoveLogs {
			hasRemoveLogs = true
		}
		if a == entity.RemoveEntity {
			hasRemoveEntity = true
		}
	}
	if !hasRemoveLogs || !hasRemoveEntity {
		t.Errorf("expected RemoveLogs and RemoveEntity, got %v", actions)
	}
}

func TestDeactivate_FromInvalidState_Panics(t *testing.T) {
	for _, state := range []entity.EntityState{entity.Inactive, entity.ScannerStarting, entity.ScannerStopping} {
		t.Run(state.String(), func(t *testing.T) {
			tree := newTestTree()
			ent := newTestEntity(state, container.ContainerRunning)
			tree.AddOrReplace(ent)
			assertPanics(t, fmt.Sprintf("Deactivate from %v", state), func() {
				ent.Deactivate(tree)
			})
		})
	}
}

// --- Restart ---

func TestRestart_FromScanning(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.Scanning, container.ContainerRunning)
	tree.AddOrReplace(ent)

	result, _, actions := ent.Restart(tree)

	assertState(t, result, entity.ScannerStopping)
	assertActions(t, actions, []entity.EntityAction{})
}

func TestRestart_FromInvalidState_Panics(t *testing.T) {
	for _, state := range []entity.EntityState{entity.Inactive, entity.WantScanning, entity.ScannerStarting, entity.ScannerStopping, entity.Deleted} {
		t.Run(state.String(), func(t *testing.T) {
			tree := newTestTree()
			ent := newTestEntity(state, container.ContainerRunning)
			tree.AddOrReplace(ent)
			assertPanics(t, fmt.Sprintf("Restart from %v", state), func() {
				ent.Restart(tree)
			})
		})
	}
}

// --- Delete ---

func TestDelete_FromInactive(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.Inactive, container.ContainerRunning)
	tree.AddOrReplace(ent)

	_, _, actions := ent.Delete(tree, newTestDelta(container.ContainerTerminated, true, false))

	assertActions(t, actions, []entity.EntityAction{})
	// entity should be removed from tree
	if tree.GetEntity(ent.Container) != nil {
		t.Error("entity should have been removed from tree")
	}
}

func TestDelete_FromWantScanning(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.WantScanning, container.ContainerWaiting)
	tree.AddOrReplace(ent)

	result, _, actions := ent.Delete(tree, newTestDelta(container.ContainerTerminated, true, false))

	assertState(t, result, entity.Deleted)
	assertActions(t, actions, []entity.EntityAction{})
	// entity should still be in tree (visible as deleted)
	if tree.GetEntity(ent.Container) == nil {
		t.Error("entity should still be in tree")
	}
}

func TestDelete_FromScannerStarting(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.ScannerStarting, container.ContainerRunning)
	tree.AddOrReplace(ent)

	_, _, actions := ent.Delete(tree, newTestDelta(container.ContainerTerminated, true, false))

	assertActions(t, actions, []entity.EntityAction{entity.StopScanner})
	if tree.GetEntity(ent.Container) != nil {
		t.Error("entity should have been removed from tree")
	}
}

func TestDelete_FromScanning(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.Scanning, container.ContainerRunning)
	tree.AddOrReplace(ent)

	result, _, actions := ent.Delete(tree, newTestDelta(container.ContainerTerminated, true, false))

	assertState(t, result, entity.Deleted)
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d: %v", len(actions), actions)
	}
	hasStopKeep := false
	hasMark := false
	for _, a := range actions {
		if a == entity.StopScannerKeepLogs {
			hasStopKeep = true
		}
		if a == entity.MarkLogsTerminated {
			hasMark = true
		}
	}
	if !hasStopKeep || !hasMark {
		t.Errorf("expected StopScannerKeepLogs and MarkLogsTerminated, got %v", actions)
	}
}

func TestDelete_FromScannerStopping(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.ScannerStopping, container.ContainerRunning)
	tree.AddOrReplace(ent)

	_, _, actions := ent.Delete(tree, newTestDelta(container.ContainerTerminated, true, false))

	assertActions(t, actions, []entity.EntityAction{entity.StopScanner})
	if tree.GetEntity(ent.Container) != nil {
		t.Error("entity should have been removed from tree")
	}
}

func TestDelete_UpdatesContainerStatus(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.Scanning, container.ContainerRunning)
	tree.AddOrReplace(ent)

	delta := newTestDelta(container.ContainerTerminated, true, false)
	result, _, _ := ent.Delete(tree, delta)

	if result.Container.Status.State != container.ContainerTerminated {
		t.Errorf("expected container status to be updated to Terminated, got %v", result.Container.Status.State)
	}
}

// --- Create ---

func TestCreate_ToActivate_RunningContainer(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.Inactive, container.ContainerRunning)

	result, _, actions := ent.Create(tree, newTestDelta(container.ContainerRunning, false, true))

	assertState(t, result, entity.ScannerStarting)
	assertActions(t, actions, []entity.EntityAction{entity.StartScanner})
}

func TestCreate_ToActivate_TerminatedContainer(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.Inactive, container.ContainerTerminated)

	result, _, actions := ent.Create(tree, newTestDelta(container.ContainerTerminated, false, true))

	assertState(t, result, entity.ScannerStarting)
	assertActions(t, actions, []entity.EntityAction{entity.StartScanner})
}

func TestCreate_ToActivate_WaitingContainer(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.Inactive, container.ContainerWaiting)

	result, _, actions := ent.Create(tree, newTestDelta(container.ContainerWaiting, false, true))

	assertState(t, result, entity.WantScanning)
	assertActions(t, actions, []entity.EntityAction{})
}

func TestCreate_ToActivate_UnknownContainer(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.Inactive, container.ContainerUnknown)

	result, _, actions := ent.Create(tree, newTestDelta(container.ContainerUnknown, false, true))

	assertState(t, result, entity.WantScanning)
	assertActions(t, actions, []entity.EntityAction{})
}

func TestCreate_NoActivate(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.Inactive, container.ContainerRunning)

	result, _, actions := ent.Create(tree, newTestDelta(container.ContainerRunning, false, false))

	assertState(t, result, entity.Inactive)
	assertActions(t, actions, []entity.EntityAction{})
}

// --- Update ---

func TestUpdate_WantScanning_ContainerStartsRunning(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.WantScanning, container.ContainerWaiting)
	tree.AddOrReplace(ent)

	result, _, actions := ent.Update(tree, newTestDelta(container.ContainerRunning, false, false))

	assertState(t, result, entity.ScannerStarting)
	assertActions(t, actions, []entity.EntityAction{entity.StartScanner})
}

func TestUpdate_WantScanning_ContainerTerminated(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.WantScanning, container.ContainerWaiting)
	tree.AddOrReplace(ent)

	result, _, actions := ent.Update(tree, newTestDelta(container.ContainerTerminated, false, false))

	assertState(t, result, entity.ScannerStarting)
	assertActions(t, actions, []entity.EntityAction{entity.StartScanner})
}

func TestUpdate_WantScanning_ContainerStillWaiting(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.WantScanning, container.ContainerWaiting)
	tree.AddOrReplace(ent)

	result, _, actions := ent.Update(tree, newTestDelta(container.ContainerWaiting, false, false))

	assertState(t, result, entity.WantScanning)
	assertActions(t, actions, []entity.EntityAction{})
}

func TestUpdate_Scanning_ContainerTerminates(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.Scanning, container.ContainerRunning)
	tree.AddOrReplace(ent)

	result, _, actions := ent.Update(tree, newTestDelta(container.ContainerTerminated, false, false))

	assertState(t, result, entity.WantScanning)
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d: %v", len(actions), actions)
	}
	hasStopKeep := false
	hasMark := false
	for _, a := range actions {
		if a == entity.StopScannerKeepLogs {
			hasStopKeep = true
		}
		if a == entity.MarkLogsTerminated {
			hasMark = true
		}
	}
	if !hasStopKeep || !hasMark {
		t.Errorf("expected StopScannerKeepLogs and MarkLogsTerminated, got %v", actions)
	}
}

func TestUpdate_Scanning_ContainerStillRunning(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.Scanning, container.ContainerRunning)
	tree.AddOrReplace(ent)

	result, _, actions := ent.Update(tree, newTestDelta(container.ContainerRunning, false, false))

	assertState(t, result, entity.Scanning)
	assertActions(t, actions, []entity.EntityAction{})
}

func TestUpdate_Inactive_NoTransition(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.Inactive, container.ContainerRunning)
	tree.AddOrReplace(ent)

	result, _, actions := ent.Update(tree, newTestDelta(container.ContainerRunning, false, false))

	assertState(t, result, entity.Inactive)
	assertActions(t, actions, []entity.EntityAction{})
}

func TestUpdate_UpdatesContainer(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.Inactive, container.ContainerWaiting)
	tree.AddOrReplace(ent)

	delta := newTestDelta(container.ContainerRunning, false, false)
	result, _, _ := ent.Update(tree, delta)

	if result.Container.Status.State != container.ContainerRunning {
		t.Errorf("expected container status to be updated to Running, got %v", result.Container.Status.State)
	}
}

// --- ScannerStarted ---

func TestScannerStarted_Success(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.ScannerStarting, container.ContainerRunning)
	tree.AddOrReplace(ent)

	scanner := newTestScanner()
	result, _, actions := ent.ScannerStarted(tree, nil, scanner)

	assertState(t, result, entity.Scanning)
	assertActions(t, actions, []entity.EntityAction{})
	if result.LogScanner == nil {
		t.Error("expected LogScanner to be set")
	}
}

func TestScannerStarted_Error(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.ScannerStarting, container.ContainerRunning)
	tree.AddOrReplace(ent)

	scanner := newTestScanner()
	result, _, actions := ent.ScannerStarted(tree, fmt.Errorf("connection refused"), scanner)

	assertState(t, result, entity.Inactive)
	assertActions(t, actions, []entity.EntityAction{})
	if result.LogScanner != nil {
		t.Error("expected LogScanner to be nil on error")
	}
}

func TestScannerStarted_DeletedEntity(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.Deleted, container.ContainerTerminated)
	tree.AddOrReplace(ent)

	scanner := newTestScanner()
	result, _, actions := ent.ScannerStarted(tree, nil, scanner)

	// should stay Deleted and cancel the scanner (race condition path)
	assertState(t, result, entity.Deleted)
	assertActions(t, actions, []entity.EntityAction{})
}

func TestScannerStarted_AlreadyScanning(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.Scanning, container.ContainerRunning)
	tree.AddOrReplace(ent)

	scanner := newTestScanner()
	result, _, actions := ent.ScannerStarted(tree, nil, scanner)

	// should stay Scanning and cancel the duplicate scanner
	assertState(t, result, entity.Scanning)
	assertActions(t, actions, []entity.EntityAction{})
}

func TestScannerStarted_FromInvalidState_Panics(t *testing.T) {
	for _, state := range []entity.EntityState{entity.Inactive, entity.WantScanning, entity.ScannerStopping} {
		t.Run(state.String(), func(t *testing.T) {
			tree := newTestTree()
			ent := newTestEntity(state, container.ContainerRunning)
			tree.AddOrReplace(ent)
			scanner := newTestScanner()
			assertPanics(t, fmt.Sprintf("ScannerStarted from %v", state), func() {
				ent.ScannerStarted(tree, nil, scanner)
			})
		})
	}
}

// --- ScannerStopped ---

func TestScannerStopped_FromScannerStopping(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.ScannerStopping, container.ContainerRunning)
	scanner := newTestScanner()
	ent.LogScanner = &scanner
	tree.AddOrReplace(ent)

	result, _, actions := ent.ScannerStopped(tree)

	assertState(t, result, entity.Inactive)
	assertActions(t, actions, []entity.EntityAction{})
	if result.LogScanner != nil {
		t.Error("expected LogScanner to be cleared")
	}
}

func TestScannerStopped_FromDeleted(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.Deleted, container.ContainerTerminated)
	scanner := newTestScanner()
	ent.LogScanner = &scanner
	tree.AddOrReplace(ent)

	result, _, actions := ent.ScannerStopped(tree)

	assertState(t, result, entity.Deleted)
	assertActions(t, actions, []entity.EntityAction{})
	if result.LogScanner != nil {
		t.Error("expected LogScanner to be cleared")
	}
}

func TestScannerStopped_FromWantScanning(t *testing.T) {
	tree := newTestTree()
	ent := newTestEntity(entity.WantScanning, container.ContainerWaiting)
	scanner := newTestScanner()
	ent.LogScanner = &scanner
	tree.AddOrReplace(ent)

	result, _, actions := ent.ScannerStopped(tree)

	assertState(t, result, entity.WantScanning)
	assertActions(t, actions, []entity.EntityAction{})
	if result.LogScanner != nil {
		t.Error("expected LogScanner to be cleared")
	}
}

func TestScannerStopped_FromInvalidState_Panics(t *testing.T) {
	for _, state := range []entity.EntityState{entity.Inactive, entity.ScannerStarting, entity.Scanning} {
		t.Run(state.String(), func(t *testing.T) {
			tree := newTestTree()
			ent := newTestEntity(state, container.ContainerRunning)
			tree.AddOrReplace(ent)
			assertPanics(t, fmt.Sprintf("ScannerStopped from %v", state), func() {
				ent.ScannerStopped(tree)
			})
		})
	}
}
