# Key Interfaces

Interfaces here are "primary building blocks" - packages, key types, and their relationships.

---

## External Library: bubbleo

Location: `github.com/robinovitch61/bubbleo` (local: ~/tesla/bubbleo)

### viewport.Model[T]

A generic scrollable viewport displaying objects of type T. Key capabilities:
- Selection: track and navigate to selected item
- Wrapping: toggle text wrap mode
- Sticky scroll: stay at top/bottom when content updates
- Highlighting: apply styles to specific text ranges
- File saving: save content to file via hotkey

### filterableviewport.Model[T]

Wraps `viewport.Model[T]` adding text/regex filtering. Key capabilities:
- Text or regex pattern matching
- "Matches only" vs "highlight all" modes
- Next/previous match navigation
- Incremental updates via `AppendObjects`

### viewport.Object Interface

Any type displayed in a viewport must implement:
```go
type Object interface {
    GetItem() item.Item
}
```

Create items via `item.NewItem(styledString)` which handles ANSI codes, Unicode widths, wrapping, and match extraction.

---

## Async Model: BubbleTea Commands

All asynchronous operations use BubbleTea's `tea.Cmd` and message passing - not direct goroutine management. This keeps concurrency concerns inside the framework:

```
Action triggers tea.Cmd → Cmd executes async work → Returns tea.Msg → Model.Update handles result
```

Packages that do async work (k8s, logscanner) expose functions returning `tea.Cmd`. The Model dispatches these commands; results arrive as messages in Update.

---

## Package Architecture

```
internal/
├── domain/      # Pure value types - no external deps
├── tree/        # Hierarchical tree for selection view
├── k8s/         # Kubernetes client abstraction (returns tea.Cmd)
├── logscanner/  # Log streaming coordination (returns tea.Cmd)
├── view/        # View components (tree, logs, single-log)
└── app.go       # Main Model orchestrating everything
```

### Dependency Direction
```
app.go
  ├── view/        (view components)
  ├── tree/        (tree management)
  ├── logscanner/  (log streaming)
  └── k8s/         (kubernetes client)
        │
        └── All packages depend on domain/
```

---

## internal/domain/

Pure value types with no external dependencies (except `time`). These form the vocabulary for all other packages.

### ContainerID

Uniquely identifies a container across all clusters:
```go
type ContainerID struct {
    Cluster, Namespace, Pod, Container string
}
```

### Container

Kubernetes container with discovery metadata:
```go
type Container struct {
    ID           ContainerID
    OwnerName    string     // "api-deployment"
    OwnerType    string     // "Deployment", "StatefulSet", "Job", etc.
    StartedAt    time.Time
    IsRunning    bool
    IsTerminated bool
}
```

### ContainerState

The 6-state scanning state machine:
```
Inactive        - Not selected, no log scanning
WantScanning    - Selected, waiting for container to become ready
ScannerStarting - Initializing log stream connection
Scanning        - Actively streaming logs
ScannerStopping - Gracefully shutting down scanner
Deleted         - Container removed from cluster, logs retained
```

### Log

A single log line with its source:
```go
type Log struct {
    Timestamp    time.Time
    ContainerID  ContainerID
    Content      string
    IsTerminated bool  // was container terminated when log emitted
}
```

### TimeRange

Lookback selection (keys 0-9):
```go
type TimeRange struct {
    Key      int           // 0-9 key press
    Duration time.Duration // 0=now onwards, -1=all time
}
```

---

## internal/tree/

Manages the hierarchical tree for the selection view.

### Concept

The tree transforms a flat list of `SelectableContainer` into a hierarchical display:
```
cluster-name
  namespace
  └─owner-name <Deployment>
    └─pod-name
      └─[x] container-name (running for 5m) - NEW
```

### Node

A single row in the tree. Different kinds: Cluster, Namespace, Owner, Pod, Container.
- Implements `viewport.Object` via `GetItem()`
- Only Container nodes have selection state
- Tracks depth and tree-drawing prefix characters

### Tree

Immutable tree structure with key operations:
- `Update(containers []SelectableContainer) Tree` - rebuild from containers
- `Nodes() []Node` - flattened list for viewport display
- `ToggleSelection(idx int) (Tree, []StateChange)` - toggle container at index
- `DeselectAll() (Tree, []StateChange)` - deselect all containers

### StateChange

Returned by selection operations to indicate required scanner actions:
```go
type StateChange struct {
    ContainerID ContainerID
    FromState   ContainerState
    ToState     ContainerState
}
```

---

## internal/k8s/

Abstracts Kubernetes API operations. Returns `tea.Cmd` for async work.

### Client

Wraps Kubernetes client for a single cluster. Used internally by Manager.

### Manager

Coordinates container watching across multiple clusters:
- `NewManager(kubeconfig, contexts) (Manager, error)` - create manager for contexts
- `WatchContainersCmd(namespaces, selector) tea.Cmd` - returns command that watches for container changes

The command produces `ContainerDeltasMsg` periodically (batched over 300ms per constants.go).

```go
type ContainerDeltasMsg struct {
    Deltas []ContainerDelta
}

type ContainerDelta struct {
    Container Container
    IsRemoved bool
}
```

---

## internal/logscanner/

Manages log streaming via `tea.Cmd`. Translates container state changes into commands that stream logs.

### Coordinator

One per cluster, returns commands for log streaming:
- `NewCoordinator(client, timeRange) Coordinator`
- `HandleStateChange(change StateChange) tea.Cmd` - returns command to start/stop log streaming
- `SetTimeRange(tr TimeRange) tea.Cmd` - returns command to restart streams with new time range
- `Shutdown() tea.Cmd` - returns command to stop all streams

### Key Messages

Commands produce these messages:
- `LogBatchMsg{Logs, ContainerID}` - batch of logs (every ~150ms per constants.go)
- `ScannerStoppedMsg{ContainerID, Reason}` - streaming ended (user deselect, container deleted, error, etc.)

---

## internal/view/

View components using bubbleo viewports.

### TreeView

Displays the tree in `filterableviewport.Model[tree.Node]`:
- `SetTree(Tree)` - update displayed tree
- `SelectedNode() *tree.Node` - get currently selected node
- Handles filtering via filterableviewport

### LogsView

Displays interleaved logs in `filterableviewport.Model[LogRow]`:
- `AppendLogs([]Log)` - add new logs, re-sort
- `ClearLogsForContainer(id)` - remove logs for deselected container
- `ToggleTimestampFormat()` / `ToggleNameFormat()` - cycle display formats
- `SetAscending(bool)` - change sort order
- Filtering handled by filterableviewport

### LogRow

Wraps `Log` for viewport display. Implements `viewport.Object`.
Formats: timestamp (none/short/full), container name (short/none/full).

### SingleLogView

Displays single expanded log in `viewport.Model[SingleLogLine]`:
- JSON auto-formatting with indentation
- Escape sequence expansion (\n, \t)
- `PlainText()` for clipboard copy

---

## internal/app.go

Main BubbleTea Model orchestrating all components.

### Key State

- `k8sManager` + `coordinators` - Kubernetes and log scanner management
- `containers map[ContainerID]SelectableContainer` - all known containers
- `tree` - current tree state
- `timeRange`, `paused` - global settings
- `treeView`, `logsView`, `singleLogView` - view components
- `pendingLogs` - buffer for batch UI updates

### View Modes

- Split view (tree left, logs right)
- Fullscreen tree or logs
- Single log overlay

### UI Overlays

- Help overlay (? key)
- Toast notifications (auto-dismiss)
- Confirmation prompts (bulk selection)

### Message Routing

1. `ContainerDeltasMsg` → update containers map → rebuild tree
2. `LogBatchMsg` → accumulate in pendingLogs
3. `BatchUpdateLogsMsg` (every 200ms) → flush pendingLogs to LogsView
4. `StateChange` from tree → route to logscanner Coordinator
5. Keyboard → route to focused view or handle globally

---

## State Machine Reference

From SPECIFICATION.md - Container state transitions:

```
User selects container:
  Inactive → ScannerStarting (if running) or WantScanning (if pending)

User deselects container:
  WantScanning → Inactive
  Scanning → ScannerStopping → Inactive
  Deleted → (removed from list)

Container becomes running:
  WantScanning → ScannerStarting → Scanning

Container terminates:
  Scanning → WantScanning (logs retained, marked terminated)

Container deleted from cluster:
  Scanning → Deleted (logs retained)
  Other states → removed

Time range changed:
  Scanning → ScannerStopping → ScannerStarting → Scanning
```

---

## Key Implementation Notes

1. **BubbleTea Commands**: All async work via `tea.Cmd` - no direct goroutine management. K8s watching and log streaming are encapsulated in commands that produce messages.

2. **Immutability**: Prefer value types. Tree and view methods return new instances.

3. **viewport.Object**: tree.Node, LogRow, SingleLogLine all implement `GetItem()`.

4. **Batching**: Three batch intervals coordinate performance:
   - 150ms: scanner log collection
   - 200ms: UI log updates
   - 300ms: container delta batching

5. **Color assignment**: Container colors via deterministic MD5 hash (see color/ package).

6. **Terminal detection**: Foreground/background detection for adaptive styling (see style/ package).
