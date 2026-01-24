# KL Application Specification

## Overview

**KL** is an interactive, cross-cluster, multi-container Kubernetes log viewer for the terminal. It enables users to view logs from multiple containers across multiple Kubernetes clusters, namespaces, and pods simultaneously in a unified, time-sorted view.

### Target Users
- Kubernetes developers debugging applications
- DevOps engineers managing multi-cluster infrastructure
- SREs investigating incidents

### Core Value Proposition
- Multi-cluster log viewing in a single interface
- Interactive container selection with pattern matching
- Real-time log streaming with unified timeline
- Advanced filtering (text and regex)
- Single-log inspection with JSON formatting

---

## Layout

### Views

The application has three main views:

1. **Entities View** - Left panel showing hierarchical container list
2. **Logs View** - Right panel showing interleaved log stream
3. **Single Log View** - Right panel showing expanded single log entry

### Split View (Default)
```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│ kl v0.5.0  Logs for the Last 1m  0/3/12                  ctrl+c to quit / ? for help│
├────────────────────────────┬────────────────────────────────────────────────────────┤
│ (S)election                │ (L)ogs Ascending                                       │
│────────────────────────────│────────────────────────────────────────────────────────│
│                            │                                                        │
│ prod-cluster               │ 14:23:01.123 api-server     Starting HTTP server...    │
│   default                  │ 14:23:01.456 api-server     Listening on :8080         │
│   ├─api-deployment         │ 14:23:02.789 worker-1       Processing batch job       │
│   │ ├─api-pod-abc123       │ 14:23:03.012 api-server     GET /health 200 2ms        │
│   │ │ └─[x] api-server     │ 14:23:03.234 worker-1       Batch complete: 150 items  │
│   │ └─api-pod-def456       │                                                        │
│   │   └─[ ] api-server     │                                                        │
│   └─worker-statefulset     │                                                        │
│     └─worker-0             │                                                        │
│       ├─[x] worker-1       │                                                        │
│       └─[ ] sidecar        │                                                        │
│                            │────────────────────────────────────────────────────────│
│ '/' or 'r' to filter       │ '/' or 'r' to filter                                   │
└────────────────────────────┴────────────────────────────────────────────────────────┘
```

### Full-Screen Mode
Press `S`, `L`, or `F` to toggle full-screen for the focused view.

### Top Bar Content
```
kl v0.5.0   Logs for the Last 1m34s   0/4/11 Pending/Selected/Total   [PAUSED]   ctrl+c to quit / ? for help
```
- **Pending**: containers in WantScanning or ScannerStarting states
- **Selected**: containers that may have logs (Scanning or Deleted states)
- **Total**: all discovered containers
- `[PAUSED]` shown when streaming is paused
- Right-side help text hidden if terminal too narrow

---

## Keyboard Reference

### Global
| Key | Action |
|-----|--------|
| Ctrl+C | Quit |
| ? | Toggle help overlay |
| p | Pause/resume log streaming |
| Ctrl+S | Save to file |
| Ctrl+Y | Copy to clipboard |
| 0-9 | Change time range |

### Navigation
| Key | Action |
|-----|--------|
| ↑/k | Move up |
| ↓/j | Move down |
| ←/→ | Pan left/right (when not wrapped) |
| g/Ctrl+G | Jump to top |
| G | Jump to bottom |
| u/Ctrl+U | Half page up |
| d/Ctrl+D | Half page down |
| b/Ctrl+B | Full page up |
| f/Ctrl+F | Full page down |

### View Switching
| Key | Action |
|-----|--------|
| s/l | Focus entities/logs view |
| S/L | Fullscreen entities/logs view |
| F | Toggle fullscreen |
| Escape | Exit Single Log View / Clear filter |

### Entities View
| Key | Action |
|-----|--------|
| Enter | Toggle container selection |
| Shift+R | Deselect all containers |

### Logs View
| Key | Action |
|-----|--------|
| Enter | Open single log |
| o | Toggle order (asc/desc) |
| t | Cycle timestamp format (none/short/full) |
| c | Cycle container name format (short/none/full) |
| w | Toggle line wrap |

### Filtering
| Key | Action |
|-----|--------|
| / | Text filter |
| r | Regex filter |
| n/N | Next/previous match |
| x | Toggle context (matches only vs all with highlighting) |
| Escape | Clear filter |
| Enter | Apply filter |

### Time Range
| Key | Lookback |
|-----|----------|
| 0 | Now onwards |
| 1 | 1 minute |
| 2 | 5 minutes |
| 3 | 15 minutes |
| 4 | 30 minutes |
| 5 | 1 hour |
| 6 | 3 hours |
| 7 | 12 hours |
| 8 | 24 hours |
| 9 | All time |

---

## Entity Tree

### Hierarchy
```
cluster-name
  namespace
  └─owner-name <OwnerType>
    └─pod-name
      └─[x] container-name (running for 5m23s) - NEW
```

**Tree elements:**
- `├─` / `└─` / `│` - Box-drawing characters for hierarchy
- `<Deployment>`, `<StatefulSet>`, etc. - Owner type label
- `(running for Xm)` - Container running duration
- `- NEW` - Containers started within last 3 minutes

### State Indicators
| Indicator | State | Description |
|-----------|-------|-------------|
| `[ ]` | Inactive | Not selected |
| `[.]` | WantScanning | Selected, waiting for container to be ready |
| `[^]` | ScannerStarting | Initializing log scanner |
| `[x]` | Scanning | Actively streaming logs |
| `[v]` | ScannerStopping | Shutting down scanner |
| `[d]` | Deleted | Container gone, logs retained |

### Selection Behavior
- Enter on single container toggles its selection
- Enter on filtered list bulk-toggles visible containers (with confirmation if ≥5 changes)
- Shift+R deselects all (with confirmation)

---

## Logs View

### Log Line Format
```
[timestamp] [container-name] [log-content]
```

### Timestamp Formats (cycle with `t`)
1. **None** - No timestamp
2. **Short** - `14:23:01.123`
3. **Full** - `2024-01-15T14:23:01.123Z`

### Container Name Formats (cycle with `c`)
1. **Short** - `fl..7/flask-1` (abbreviated pod + container)
2. **None** - Just log content
3. **Full** - `cluster/namespace/pod/container`

### Order Behavior
- **Ascending** (default): oldest first, selection sticks to bottom
- **Descending**: newest first, selection sticks to top

### Terminated Containers
Container name displays `[TERMINATED]` suffix:
```
14:23:02.789 api-server [TERMINATED]  Graceful shutdown initiated
```

### Scroll Position Indicator
When content overflows: `25% (5/20)` - showing item 5 of 20

---

## Single Log View

Displays a single log entry in detail:
- Full timestamp with timezone
- Complete container path (color-coded)
- JSON auto-formatted with 4-space indentation
- Escape sequences (`\n`, `\t`) expanded

Navigate with ↑/↓ and ←/→. Press Escape to return.

---

## Filtering

### Modes
- **Text filter** (`/`) - Exact substring matching
- **Regex filter** (`r`) - Regular expression matching

### Context Display (toggle with `x`)
- **Matches only** - Shows only matching lines
- **Contextual** - Shows all lines with matches highlighted

### Filter Bar States
```
'/' or 'r' to filter                    # No filter
filter: error (matches only)            # Text filter, matches only mode
filter: error (3/217, n/N to cycle)     # Text filter, contextual mode
regex filter: error|warn (12/156)       # Regex filter
invalid regex: [unclosed                # Invalid regex
```

---

## Help Overlay

Press `?` to show, any key to dismiss. Displays key bindings in multi-column layout, dynamically generated from keymap definitions.

---

## Confirmation Prompts

Shown for:
- Bulk selection changes (≥5 containers)
- Deselect all

Double-bordered modal with "NO, CANCEL" / "YES, PROCEED" buttons. Navigate with ←/→, h/l, or Tab.

---

## Toast Notifications

Non-blocking notifications at bottom of screen, auto-dismiss after 5 seconds:
- File save confirmations
- Clipboard copy
- Time range changes
- Selection limit reached

---

## State Machine

### Entity States

```
States:
  Inactive        - Not selected, no scanning
  WantScanning    - Selected but container not ready
  ScannerStarting - Initializing log scanner
  Scanning        - Actively collecting logs
  ScannerStopping - Shutting down scanner
  Deleted         - Container removed but logs retained
```

### State Transitions

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

### Connection Behavior
- Timeout: 30 seconds before retry
- Retry with exponential backoff: 1s, 2s, 4s, 8s, max 30s
- Max 10 retries before giving up

### Pause Behavior
- Logs continue buffering but not displayed
- `[PAUSED]` shown in top bar
- Press `p` to resume and flush buffer

### Multiple Container Coordination
- Each container has independent state machine
- Logs interleave by timestamp
- Time range and pause apply to all

---

## CLI Configuration

All flags can be set via environment variable `KL_<FLAG_NAME>` (uppercase, hyphens to underscores).

### Context & Kubeconfig
| Flag | Default | Description |
|------|---------|-------------|
| `--context` | Current context | Comma-separated list of contexts |
| `--kubeconfig` | `~/.kube/config` | Path to kubeconfig file |
| `--mclust` | None | Regex to match cluster names |
| `--iclust` | None | Regex to exclude clusters |

### Namespace
| Flag | Default | Description |
|------|---------|-------------|
| `-n, --namespace` | Current namespace | Comma-separated namespaces |
| `-A, --all-namespaces` | false | Watch all namespaces |
| `--mns` | None | Regex to match namespaces |
| `--ins` | None | Regex to exclude namespaces |

### Pod & Container Matching
| Flag | Description |
|------|-------------|
| `--mpod` / `--ipod` | Match/ignore pods by name regex |
| `--mown` / `--iown` | Match/ignore pod owners by name regex |
| `--mc` / `--ic` | Match/ignore containers by name regex |
| `-l, --selector` | Kubernetes label selector |
| `--ignore-owner-types` | Exclude by owner type (Job, CronJob, etc.) |
| `--limit` | Max containers to auto-select (-1 = unlimited) |

### Display Options
| Flag | Default | Description |
|------|---------|-------------|
| `-f, --log-filter` | None | Initial text filter |
| `-r, --log-regex` | None | Initial regex filter |
| `-d, --desc` | false | Start in descending order |
| `--logs-view` | false | Start with logs view focused |
| `--since` | 1m | Initial time range (Go duration) |

### Filter Evaluation Order
```
1. --context / --mclust / --iclust   → Clusters
2. -n / -A / --mns / --ins           → Namespaces
3. --ignore-owner-types              → Owner types
4. --mown / --iown                   → Owner names
5. --mpod / --ipod                   → Pod names
6. -l (selector)                     → Pod labels
7. --mc / --ic                       → Container names
8. --limit                           → Selection cap
```

"Ignore" patterns always take precedence. "Match" patterns cause auto-selection.

### Usage Examples
```bash
# Basic
kl                                    # Default namespace, current context
kl -n production                      # Specific namespace
kl -A                                 # All namespaces

# Multi-cluster
kl --context prod-east,prod-west -n critical
kl --mclust "prod-.*" -A

# Filtering
kl -n myapp --mown "api-deployment"
kl -l "app=nginx,version=v2"
kl -A --ic "istio-proxy|envoy"

# Debugging
kl -n prod --since 1h -r "error|panic" -d
kl -A --mown "critical-.*" --limit 50
```

---

## Visual Styling

### Colors
- **Lilac** (Color 189) - Top bar, borders, headers
- **Green** (Color 46) - Timestamps
- **Container colors** - Deterministically assigned from 13-color palette using MD5 hash

### Terminal Adaptation
- Detects terminal background (dark/light)
- Adjusts foreground colors for readability
- Falls back to defaults if detection fails

---

## Technical Details

### Component Hierarchy
```
App
├── Top Bar
├── Pages Container
│   ├── Entities Page → FilterableViewport
│   ├── Logs Page → FilterableViewport
│   └── Single Log Page → FilterableViewport
├── Help Overlay (conditional)
├── Prompt Modal (conditional)
└── Toast Notification (conditional)
```

### Performance Constants
| Constant | Value | Description |
|----------|-------|-------------|
| SingleContainerLogCollectionDuration | 150ms | Log collection batch window |
| GetNextContainerDeltasDuration | 300ms | Container discovery batch window |
| BatchUpdateLogsInterval | 200ms | UI update interval |
| NewContainerThreshold | 3 minutes | "NEW" annotation threshold |
| ConfirmSelectionActionsThreshold | 5 | Changes requiring confirmation |
| LeftPageWidthFraction | 40% | Entities panel width |

### GKE Authentication
- Validates `gke-gcloud-auth-plugin` in PATH when required
- Shows helpful error with installation hint if missing

### Debug Features
| Variable | Description |
|----------|-------------|
| `KL_DEBUG` | Enable debug logging |
| `KL_DEBUG_PATH` | Debug log file path (default: `kl.log`) |

---

## Verification Checklist

1. Container discovery across contexts
2. Selection starts/stops log streaming
3. Logs interleave by timestamp
4. Real-time updates (when not paused)
5. Text and regex filtering with context toggle
6. Time range changes restart scanners
7. Split/fullscreen view navigation
8. Single log JSON formatting and escape expansion
9. File export without ANSI codes
10. All keyboard shortcuts function
11. Terminal resize adaptation
12. Container lifecycle handling (start/stop/restart)
