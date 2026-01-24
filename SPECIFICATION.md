# KL Application Specification

## Overview

**KL** is an interactive, cross-cluster, multi-container Kubernetes log viewer for the terminal. It enables users to view logs from multiple containers across multiple Kubernetes clusters, namespaces, and pods simultaneously in a unified, time-sorted view.

### Target Users
- Kubernetes developers debugging applications
- DevOps engineers managing multi-cluster infrastructure
- SREs investigating incidents
- Anyone who needs efficient Kubernetes log browsing

### Core Value Proposition
- Multi-cluster log viewing in a single interface
- Interactive container selection with pattern matching
- Real-time log streaming with unified timeline
- Advanced filtering (text and regex)
- Single-log inspection with JSON formatting

---

## Application Structure

### Views

The application has three main views:

1. **Entities View (Selection View)** - Left panel showing hierarchical container list
2. **Logs View** - Right panel showing interleaved log stream
3. **Single Log View** - Right panel showing expanded single log entry

### Layout Modes

**Split View (Default)**
```
┌─────────────────────────────────────────────────────┐
│ Top Bar                                             │
├────────────────────┬────────────────────────────────┤
│ Entities View      │ Logs View / Single Log View    │
│ (~40% width)       │ (~60% width)                   │
└────────────────────┴────────────────────────────────┘
│ Toast notification (when visible)                   │
```

**Full-Screen Mode**
```
┌─────────────────────────────────────────────────────┐
│ Top Bar                                             │
├─────────────────────────────────────────────────────┤
│ Currently Focused View (100% width)                 │
│                                                     │
└─────────────────────────────────────────────────────┘
```

---

## Example Views

### Default Split View (Entities + Logs)

```
┌──────────────────────────────────────────────────────────────────────────────────────┐
│ kl v0.5.0  Logs for the Last 1m  0/3/12                  ctrl+c to quit / ? for help │
├─────────────────────────────┬────────────────────────────────────────────────────────┤
│ (S)election                 │ (L)ogs Ascending                                       │
│─────────────────────────────│────────────────────────────────────────────────────────│
│                             │                                                        │
│ prod-cluster                │ 14:23:01.123 api-server     Starting HTTP server...    │
│   default                   │ 14:23:01.456 api-server     Listening on :8080         │
│   ├─api-deployment          │ 14:23:02.789 worker-1       Processing batch job       │
│   │ ├─api-pod-abc123        │ 14:23:03.012 api-server     GET /health 200 2ms        │
│   │ │ └─[x] api-server      │ 14:23:03.234 worker-1       Batch complete: 150 items  │
│   │ └─api-pod-def456        │ 14:23:04.567 api-server     GET /api/users 200 45ms    │
│   │   └─[ ] api-server      │ 14:23:05.890 worker-1       Starting next batch        │
│   └─worker-statefulset      │ 14:23:06.123 api-server     POST /api/data 201 89ms    │
│     └─worker-0              │ 14:23:07.456 worker-1       Connected to database      │
│       ├─[x] worker-1        │ 14:23:08.789 api-server     GET /metrics 200 5ms       │
│       └─[ ] sidecar         │ 14:23:09.012 worker-1       Query executed: 23ms       │
│   kube-system               │ 14:23:10.345 api-server     WebSocket connected        │
│   └─coredns                 │                                                        │
│     └─coredns-xyz789        │                                                        │
│       └─[ ] coredns         │                                                        │
│                             │                                                        │
│                             │────────────────────────────────────────────────────────│
│ '/' or 'r' to filter        │ '/' or 'r' to filter                                   │
└─────────────────────────────┴────────────────────────────────────────────────────────┘
```

### Entities View with Selection Highlight

The currently selected row is shown with inverse colors (highlighted):

```
│ (S)election                 │
│─────────────────────────────│
│                             │
│ prod-cluster                │
│   default                   │
│   └─api-deployment          │
│     ├─api-pod-abc123        │
│ ████│ └─[x] api-server██████│  ← Highlighted row (inverse colors)
│     └─api-pod-def456        │
│       └─[ ] api-server      │
```

### Container State Indicators in Entity List

```
│   └─various-deployment          │
│     ├─pod-running-abc           │
│     │ ├─[ ] container-inactive  │  ← Not selected
│     │ ├─[.] container-waiting   │  ← Selected, container pending/waiting
│     │ ├─[^] container-starting  │  ← Scanner initializing
│     │ ├─[x] container-scanning  │  ← Actively streaming logs
│     │ └─[v] container-stopping  │  ← Scanner shutting down
│     └─pod-deleted-xyz           │
│       └─[d] container-deleted   │  ← Container gone but logs retained
```

### Logs View - Ascending Order (Default)

Oldest logs at top, newest at bottom. Selection sticks to bottom as new logs arrive:

```
│ (L)ogs Ascending                                                              │
│───────────────────────────────────────────────────────────────────────────────│
│ 14:23:01.123 api-server     Starting HTTP server on port 8080                 │
│ 14:23:01.456 api-server     Loading configuration from /etc/config            │
│ 14:23:02.789 worker-1       Initializing worker process                       │
│ 14:23:03.012 api-server     Configuration loaded successfully                 │
│ 14:23:03.234 worker-1       Connected to message queue                        │
│ 14:23:04.567 api-server     Health check endpoint ready                       │
│ 14:23:05.890 worker-1       Processing message: order-12345                   │
│ 14:23:06.123 api-server     Incoming request: GET /api/orders                 │
│ 14:23:07.456 worker-1       Message processed successfully                    │
│ 14:23:08.789 api-server     Response sent: 200 OK (45ms)                      │  ← Selection here
│───────────────────────────────────────────────────────────────────────────────│
│ '/' or 'r' to filter                                                          │
```

### Logs View - Descending Order

Newest logs at top, oldest at bottom. Selection sticks to top as new logs arrive:

```
│ (L)ogs Descending                                                             │
│───────────────────────────────────────────────────────────────────────────────│
│ 14:23:08.789 api-server     Response sent: 200 OK (45ms)                      │  ← Selection here
│ 14:23:07.456 worker-1       Message processed successfully                    │
│ 14:23:06.123 api-server     Incoming request: GET /api/orders                 │
│ 14:23:05.890 worker-1       Processing message: order-12345                   │
│ 14:23:04.567 api-server     Health check endpoint ready                       │
│ 14:23:03.234 worker-1       Connected to message queue                        │
│ 14:23:03.012 api-server     Configuration loaded successfully                 │
│ 14:23:02.789 worker-1       Initializing worker process                       │
│ 14:23:01.456 api-server     Loading configuration from /etc/config            │
│ 14:23:01.123 api-server     Starting HTTP server on port 8080                 │
│───────────────────────────────────────────────────────────────────────────────│
│ '/' or 'r' to filter                                                          │
```

### Logs View - Different Timestamp Formats

**No Timestamp (t cycles to this):**
```
│ api-server     Starting HTTP server on port 8080                              │
│ api-server     Loading configuration from /etc/config                         │
│ worker-1       Initializing worker process                                    │
```

**Short Timestamp (t cycles to this):**
```
│ 14:23:01.123 api-server     Starting HTTP server on port 8080                 │
│ 14:23:01.456 api-server     Loading configuration from /etc/config            │
│ 14:23:02.789 worker-1       Initializing worker process                       │
```

**Full Timestamp (t cycles to this):**
```
│ 2024-01-15T14:23:01.123Z api-server     Starting HTTP server on port 8080     │
│ 2024-01-15T14:23:01.456Z api-server     Loading configuration from /etc/config│
│ 2024-01-15T14:23:02.789Z worker-1       Initializing worker process           │
```

### Logs View - Different Container Name Formats

**Short Name (c cycles to this):**
```
│ 14:23:01.123 api          Starting HTTP server on port 8080                   │
│ 14:23:02.789 wrk          Initializing worker process                         │
```

**No Name (c cycles to this):**
```
│ 14:23:01.123 Starting HTTP server on port 8080                                │
│ 14:23:02.789 Initializing worker process                                      │
```

**Full Name (c cycles to this):**
```
│ 14:23:01.123 prod/default/api-pod-abc123/api-server     Starting HTTP server  │
│ 14:23:02.789 prod/default/worker-0/worker-1             Initializing worker   │
```

### Logs View - Terminated Container

The `[TERMINATED]` suffix appears on the container name, not the log content:

```
│ 14:23:01.123 api-server [TERMINATED]  Starting HTTP server on port 8080       │
│ 14:23:02.789 api-server [TERMINATED]  Received SIGTERM                        │
│ 14:23:03.012 api-server [TERMINATED]  Graceful shutdown initiated             │
│ 14:23:04.567 api-server-new           Starting HTTP server on port 8080       │
```

### Logs View - With Active Filter (Matches Only)

Filter hides non-matching lines:

```
│ (L)ogs Ascending                                                              │
│───────────────────────────────────────────────────────────────────────────────│
│ 14:23:01.123 api-server     GET /api/users 200 45ms                           │
│ 14:23:03.456 api-server     GET /api/orders 200 32ms                          │
│ 14:23:05.789 api-server     GET /api/products 200 28ms                        │
│ 14:23:07.012 api-server     GET /api/users/123 200 15ms                       │
│ 14:23:09.345 api-server     GET /api/health 200 2ms                           │
│                                                                               │
│                                                                               │
│                                                                               │
│                                                                               │
│                                                                               │
│───────────────────────────────────────────────────────────────────────────────│
│ filter: GET (matches only)                                                    │
```

### Logs View - With Active Filter (Context Mode - x toggled)

All lines shown, matches highlighted, navigable with n/N:

```
│ (L)ogs Ascending                                                              │
│───────────────────────────────────────────────────────────────────────────────│
│ 14:23:01.123 api-server     [GET] /api/users 200 45ms                         │  ← "GET" highlighted
│ 14:23:02.456 worker-1       Processing batch job                              │
│ 14:23:03.456 api-server     [GET] /api/orders 200 32ms                        │  ← Current match (2/5)
│ 14:23:04.789 worker-1       Batch complete                                    │
│ 14:23:05.789 api-server     [GET] /api/products 200 28ms                      │  ← "GET" highlighted
│ 14:23:06.012 api-server     POST /api/orders 201 89ms                         │
│ 14:23:07.012 api-server     [GET] /api/users/123 200 15ms                     │  ← "GET" highlighted
│ 14:23:08.345 worker-1       Starting next job                                 │
│ 14:23:09.345 api-server     [GET] /api/health 200 2ms                         │  ← "GET" highlighted
│                                                                               │
│───────────────────────────────────────────────────────────────────────────────│
│ filter: GET (2/5, n/N to cycle)                                               │
```

### Logs View - Regex Filter

```
│───────────────────────────────────────────────────────────────────────────────│
│ regex filter: error|warn|fail  (12/156 matches)                               │
```

### Logs View - Invalid Regex

```
│───────────────────────────────────────────────────────────────────────────────│
│ invalid regex: [unclosed                                                      │
```

### Single Log View - Plain Text

```
┌──────────────────────────────────────────────────────────────────────────────────────┐
│ kl v0.5.0  Logs for the Last 1m  0/3/12                  ctrl+c to quit / ? for help │
├──────────────────────────────────────────────────────────────────────────────────────┤
│ Single Log                                                                           │
│ 2024-01-15T14:23:01.123Z  prod-cluster/default/api-pod-abc123/api-server             │
│──────────────────────────────────────────────────────────────────────────────────────│
│                                                                                      │
│ Starting HTTP server on port 8080. Configuration loaded from /etc/config/app.yaml.   │
│ Server ready to accept connections. TLS enabled with certificate from /etc/certs.    │
│ Registered 15 API endpoints. Health check available at /health.                      │
│                                                                                      │
│                                                                                      │
│                                                                                      │
│                                                                                      │
│                                                                                      │
│──────────────────────────────────────────────────────────────────────────────────────│
│ '/' or 'r' to filter                                                                 │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

### Single Log View - JSON Formatted

Raw log containing JSON is automatically formatted:

**Original log content:**
```
{"level":"info","timestamp":"2024-01-15T14:23:01Z","message":"Request processed","method":"POST","path":"/api/orders","status":201,"duration_ms":89,"user_id":"user-123"}
```

**Displayed in Single Log View:**
```
│ Single Log                                                                           │
│ 2024-01-15T14:23:01.123Z  prod-cluster/default/api-pod-abc123/api-server             │
│──────────────────────────────────────────────────────────────────────────────────────│
│                                                                                      │
│ {                                                                                    │
│     "level": "info",                                                                 │
│     "timestamp": "2024-01-15T14:23:01Z",                                             │
│     "message": "Request processed",                                                  │
│     "method": "POST",                                                                │
│     "path": "/api/orders",                                                           │
│     "status": 201,                                                                   │
│     "duration_ms": 89,                                                               │
│     "user_id": "user-123"                                                            │
│ }                                                                                    │
│                                                                                      │
│──────────────────────────────────────────────────────────────────────────────────────│
│ '/' or 'r' to filter                                                                 │
```

### Single Log View - Escaped Sequences Expanded

Log containing `\n` and `\t` escape sequences are expanded:

**Original log content:**
```
Error occurred:\n\tFile: /app/handler.go\n\tLine: 142\n\tMessage: Connection refused
```

**Displayed in Single Log View:**
```
│ Single Log                                                                           │
│──────────────────────────────────────────────────────────────────────────────────────│
│                                                                                      │
│ Error occurred:                                                                      │
│     File: /app/handler.go                                                            │
│     Line: 142                                                                        │
│     Message: Connection refused                                                      │
│                                                                                      │
```

### Full-Screen Entities View (Shift+S)

```
┌──────────────────────────────────────────────────────────────────────────────────────┐
│ kl v0.5.0  Logs for the Last 1m  0/3/12                  ctrl+c to quit / ? for help │
├──────────────────────────────────────────────────────────────────────────────────────┤
│ (S)election                                                                          │
│──────────────────────────────────────────────────────────────────────────────────────│
│                                                                                      │
│ prod-cluster                                                                         │
│   default                                                                            │
│   ├─api-deployment                                                                   │
│   │ ├─api-pod-abc123                                                                 │
│   │ │ └─[x] api-server                                                               │
│   │ └─api-pod-def456                                                                 │
│   │   └─[ ] api-server                                                               │
│   └─worker-statefulset                                                               │
│     └─worker-0                                                                       │
│       ├─[x] worker-1                                                                 │
│       └─[ ] sidecar                                                                  │
│   kube-system                                                                        │
│   └─coredns                                                                          │
│     └─coredns-xyz789                                                                 │
│       └─[ ] coredns                                                                  │
│                                                                                      │
│──────────────────────────────────────────────────────────────────────────────────────│
│ '/' or 'r' to filter                                                                 │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

### Full-Screen Logs View (Shift+L)

```
┌──────────────────────────────────────────────────────────────────────────────────────┐
│ kl v0.5.0  Logs for the Last 1m  0/3/12                  ctrl+c to quit / ? for help │
├──────────────────────────────────────────────────────────────────────────────────────┤
│ (L)ogs Ascending                                                                     │
│──────────────────────────────────────────────────────────────────────────────────────│
│ 14:23:01.123 api-server     Starting HTTP server on port 8080                        │
│ 14:23:01.456 api-server     Loading configuration from /etc/config                   │
│ 14:23:02.789 worker-1       Initializing worker process                              │
│ 14:23:03.012 api-server     Configuration loaded successfully                        │
│ 14:23:03.234 worker-1       Connected to message queue                               │
│ 14:23:04.567 api-server     Health check endpoint ready                              │
│ 14:23:05.890 worker-1       Processing message: order-12345                          │
│ 14:23:06.123 api-server     Incoming request: GET /api/orders                        │
│ 14:23:07.456 worker-1       Message processed successfully                           │
│ 14:23:08.789 api-server     Response sent: 200 OK (45ms)                             │
│ 14:23:09.012 worker-1       Waiting for next message                                 │
│ 14:23:10.345 api-server     Incoming request: POST /api/users                        │
│ 14:23:11.678 api-server     User created: user-456                                   │
│ 14:23:12.901 worker-1       Received message: user-created                           │
│ 14:23:13.234 worker-1       Processing user-created event                            │
│──────────────────────────────────────────────────────────────────────────────────────│
│ '/' or 'r' to filter                                                                 │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

### Top Bar Variations

**Normal State:**
```
│ kl v0.5.0  Logs for the Last 1m  0/3/12                        ctrl+c to quit / ? for help │
```

**With Pending Containers:**
```
│ kl v0.5.0  Logs for the Last 15m  2/5/20                       ctrl+c to quit / ? for help │
                                   ↑
                                   2 containers pending (waiting to start)
```

**Paused State:**
```
│ kl v0.5.0  Logs for the Last 1m  0/3/12  [PAUSED]              ctrl+c to quit / ? for help │
                                           ↑
                                           Highlighted/inverse text
```

**All Time Range:**
```
│ kl v0.5.0  Logs for All Time  0/3/12                           ctrl+c to quit / ? for help │
```

**Narrow Terminal (help text hidden):**
```
│ kl v0.5.0  Logs for the Last 1m  0/3/12                                                    │
```

### Help Overlay

```
┌──────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                      │
│                                                                                      │
│         ╔════════════════════════════════════════════════════════════════╗           │
│         ║                Help (press any key to hide)                    ║           │
│         ╠════════════════════════════════════════════════════════════════╣           │
│         ║                                                                ║           │
│         ║  Navigation                    Actions                         ║           │
│         ║  ───────────                   ───────                         ║           │
│         ║  ↑/k      Move up              enter   Select/zoom             ║           │
│         ║  ↓/j      Move down            o       Toggle order            ║           │
│         ║  g        Jump to top          t       Cycle timestamp         ║           │
│         ║  G        Jump to bottom       c       Cycle container name    ║           │
│         ║  u/ctrl+u Half page up         w       Toggle wrap             ║           │
│         ║  d/ctrl+d Half page down       p       Pause/resume            ║           │
│         ║  b/ctrl+b Full page up         ctrl+s  Save to file            ║           │
│         ║  f/ctrl+f Full page down       ctrl+y  Copy to clipboard       ║           │
│         ║  ←        Pan left                                             ║           │
│         ║  →        Pan right            Filtering                       ║           │
│         ║                                ─────────                       ║           │
│         ║  Views                         /       Text filter             ║           │
│         ║  ─────                         r       Regex filter            ║           │
│         ║  s        Focus selection      n       Next match              ║           │
│         ║  l        Focus logs           N       Previous match          ║           │
│         ║  S        Fullscreen selection x       Toggle context          ║           │
│         ║  L        Fullscreen logs      esc     Clear filter            ║           │
│         ║  F        Toggle fullscreen                                    ║           │
│         ║                                                                ║           │
│         ║  Time Range: 0=now 1=1m 2=5m 3=15m 4=30m 5=1h 6=3h 7=12h 8=1d 9=all         ║           │
│         ║                                                                ║           │
│         ╚════════════════════════════════════════════════════════════════╝           │
│                                                                                      │
│                                                                                      │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

### Confirmation Prompt - Bulk Selection

```
┌──────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                      │
│                                                                                      │
│                                                                                      │
│                                                                                      │
│              ╔══════════════════════════════════════════════════════╗                │
│              ║                                                      ║                │
│              ║   Select 8 & deselect 3 visible containers?          ║                │
│              ║                                                      ║                │
│              ║       [ NO, CANCEL ]      [█YES, PROCEED█]           ║                │
│              ║                                  ↑                   ║                │
│              ║                           Currently selected         ║                │
│              ╚══════════════════════════════════════════════════════╝                │
│                                                                                      │
│                                                                                      │
│                                                                                      │
│                                                                                      │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

### Confirmation Prompt - Deselect All

```
              ╔══════════════════════════════════════════════════════╗
              ║                                                      ║
              ║   Deselect all 5 containers?                         ║
              ║                                                      ║
              ║       [█NO, CANCEL█]      [ YES, PROCEED ]           ║
              ║             ↑                                        ║
              ║      Currently selected                              ║
              ╚══════════════════════════════════════════════════════╝
```

### Toast Notifications

**File Save Success:**
```
│ 14:23:09.345 api-server     GET /api/health 200 2ms                                  │
│──────────────────────────────────────────────────────────────────────────────────────│
│ '/' or 'r' to filter                                                                 │
├──────────────────────────────────────────────────────────────────────────────────────┤
│ Saved to /Users/dev/kl_logs_20240115_142315.log                                      │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

**Clipboard Copy:**
```
├──────────────────────────────────────────────────────────────────────────────────────┤
│ Copied to clipboard                                                                  │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

**Time Range Change:**
```
├──────────────────────────────────────────────────────────────────────────────────────┤
│ Changing time range to start from 1 hour ago...                                      │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

**Selection Limit Reached:**
```
├──────────────────────────────────────────────────────────────────────────────────────┤
│ Selection limit reached (10 containers). Use --limit to increase.                    │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

### Initial Loading State (Before Containers Discovered)

```
┌──────────────────────────────────────────────────────────────────────────────────────┐
│ kl v0.5.0  Logs for the Last 1m  0/0/0                         ctrl+c to quit / ? for help │
├─────────────────────────────┬────────────────────────────────────────────────────────┤
│ (S)election                 │ (L)ogs Ascending                                       │
│─────────────────────────────│────────────────────────────────────────────────────────│
│                             │                                                        │
│ Subscribing to updates for: │                                                        │
│   Context: prod-cluster     │                                                        │
│   Namespace: default        │                                                        │
│                             │                                                        │
│                             │                                                        │
│                             │                                                        │
│                             │                                                        │
│                             │                                                        │
│ '/' or 'r' to filter        │ '/' or 'r' to filter                                   │
└─────────────────────────────┴────────────────────────────────────────────────────────┘
```

### Empty Logs View (No Containers Selected)

```
│ (L)ogs Ascending                                                                     │
│──────────────────────────────────────────────────────────────────────────────────────│
│                                                                                      │
│                                                                                      │
│                                                                                      │
│                         No containers selected.                                      │
│                    Select containers in the left panel.                              │
│                                                                                      │
│                                                                                      │
│                                                                                      │
│──────────────────────────────────────────────────────────────────────────────────────│
│ '/' or 'r' to filter                                                                 │
```

### Line Wrapping vs Horizontal Scroll

**With Wrapping Enabled (w toggles):**
```
│ 14:23:01.123 api-server     This is a very long log message that would normally    │
│                             extend beyond the visible area but with wrapping it    │
│                             continues on the next line automatically               │
│ 14:23:02.456 worker-1       Short message                                          │
```

**Without Wrapping (horizontal scroll with ←/→):**
```
│ 14:23:01.123 api-server     This is a very long log message that would normally...│
│ 14:23:02.456 worker-1       Short message                                         │
                                                                              ↑
                                                        "..." indicates truncation
```

**After Scrolling Right:**
```
│ ...message that would normally extend beyond the visible area of the terminal win...│
│ ...                                                                                 │
       ↑                                                                          ↑
"..." on left indicates content scrolled past             "..." on right indicates more
```

### Color Coding Example

Each container gets unique colors for visual distinction:

```
│ 14:23:01.123 [BLUE]api[/BLUE]-[CYAN]server[/CYAN]     Starting HTTP server        │
│ 14:23:02.456 [GREEN]wrk[/GREEN]-[YELLOW]worker-1[/YELLOW]   Processing batch      │
│ 14:23:03.789 [MAGENTA]svc[/MAGENTA]-[RED]sidecar[/RED]      Health check passed   │
```

The prefix portion (e.g., "api", "wrk", "svc") uses one color, and the full name portion uses another. This helps visually distinguish logs from different containers at a glance.

---

## Features

### 1. Container Discovery and Selection

#### Behavior
- Discovers containers across configured Kubernetes contexts
- Displays containers in hierarchical tree: Cluster → Namespace → Pod Owner → Pod → Container
- Containers can be selected/deselected to control which logs are displayed
- Selection state persists across container restarts

#### Entity Tree Display

Each level of the hierarchy displays using box-drawing characters:

```
k3d-test                                            ← Cluster
  default                                           ← Namespace
  └─flask-deployment <Deployment>                   ← Owner with type label
    └─flask-deployment-5477db84c5-w2dn7             ← Pod name
      ├─[x] flask-1 (running for 5m23s)             ← Container with state & duration
      └─[ ] flask-2 (running for 3m - NEW!)         ← New container annotation
```

**Tree Node Elements**:
- `├─` / `└─` / `│` - Box-drawing characters showing tree hierarchy (tree is always fully expanded; use filtering to reduce visible items)
- `<Deployment>`, `<StatefulSet>`, etc. - Owner type label
- `(running for Xm)` - Container running duration
- `- NEW` - Containers started within last 3 minutes
- State indicators: `[ ]`, `[.]`, `[^]`, `[x]`, `[v]`, `[d]`

#### Auto-Selection (via CLI)
- Match/ignore patterns for: clusters, namespaces, pods, pod owners, containers
- Kubernetes label selectors
- Limit on number of auto-selected containers

#### Visual Indicators
Container state indicators in the entity list:
- `[ ]` - Inactive (not selected)
- `[.]` - Waiting to scan (container not yet running)
- `[^]` - Scanner starting
- `[x]` - Scanning (actively collecting logs)
- `[v]` - Scanner stopping
- `[d]` - Deleted (container gone but logs retained)

#### Interactions
| Action | Behavior |
|--------|----------|
| Navigate (↑/↓/j/k) | Move selection highlight |
| Enter on single item | Toggle selection of that container |
| Enter on filtered list | Bulk toggle visible containers (with confirmation prompt) |
| Shift+R | Deselect all containers (with confirmation) |

---

### 2. Log Streaming and Display

#### Behavior
- Streams logs from all selected containers in real-time
- Logs are interleaved by timestamp into unified timeline
- Supports ascending (oldest first) or descending (newest first) order
- Batches log updates (200ms intervals) for performance
- Can be paused to freeze the current view

#### Log Line Format
Each log line displays:
```
[timestamp] [container-name] [log-content]
```

#### Timestamp Formats (cyclable)
1. **None** - No timestamp shown
2. **Short** - Time only (HH:MM:SS.mmm)
3. **Full** - Full ISO timestamp

#### Container Name Formats (cyclable with 'c' key)

1. **Short** - Abbreviated pod name + full container name
   - Format: `[pod-abbrev]/[container-name]`
   - Example: `fl..7/flask-1` (pod "flask-deployment-5477db84c5-w2dn7" abbreviated)
   - Keeps first 2 and last 2 characters of pod suffix with ".." in middle

2. **None** - No name shown (just log content)

3. **Full** - Complete container path
   - Format: `[cluster]/[namespace]/[pod]/[container]`
   - Example: `k3d-test/default/flask-deployment-5477db84c5-w2dn7/flask-1`

#### Terminated Container Handling
- When a container terminates, its logs remain visible
- The container name in log lines displays `[TERMINATED]` suffix (e.g., `api-server [TERMINATED]`)
- If container restarts, new logs stream in

#### Visual Features
- Each container gets unique color coding (ID and name portions)
- Line wrapping toggle (wrapped vs horizontal scroll)
- Sticky scroll: stays at bottom (descending) or top (ascending) as new logs arrive

#### Scroll Position Indicator
When content overflows the viewport, a scroll position indicator appears at the bottom:
- Format: `X% (N/M)` where X is percentage through content, N is current position, M is total items
- Example: `25% (5/20)` - showing item 5 of 20, 25% through the list
- Only shown when scrolling has begun or content exceeds viewport height

#### Interactions
| Action | Behavior |
|--------|----------|
| Navigate (↑/↓/j/k) | Move through log lines |
| Page navigation (f/b, d/u) | Full/half page scrolling |
| Jump (g/G) | Jump to top/bottom |
| Horizontal pan (←/→) | Scroll horizontally when not wrapped |
| Enter | Open selected log in Single Log View |
| o | Toggle ascending/descending order |
| t | Cycle timestamp format |
| c | Cycle container name format |
| w | Toggle line wrapping |
| p | Pause/resume log streaming |

---

### 3. Single Log View

#### Behavior
- Displays a single log entry in full detail
- Header shows full timestamp with timezone and complete container path
- JSON content is automatically formatted with 4-space indentation
- Escaped sequences (\n, \t) are expanded for readability (\t → 4 spaces)
- Supports both horizontal and vertical scrolling

#### Header Format
```
Single Log  '/' or 'r' to filter
2024-12-30T16:23:43.820-08:00 | k3d-test/default/flask-deployment/flask-deployment-5477db84c5-w2dn7/flask-1
```

Container path components are color-coded:
- Cluster/namespace/owner path in one color
- Container name in bright highlight color

#### Interactions
| Action | Behavior |
|--------|----------|
| Escape | Return to Logs View (or clear filter if active) |
| Navigate (↑/↓) | Vertical scroll |
| Horizontal pan (←/→) | Horizontal scroll |
| Ctrl+Y | Copy log content to clipboard (without ANSI codes) |
| / | Enter text filter mode |
| r | Enter regex filter mode |
| n/N | Navigate to next/previous match |

---

### 4. Filtering

#### Filter Modes
1. **Text Filter** - Exact substring matching
2. **Regex Filter** - Regular expression pattern matching

#### Context Display Modes
1. **Matches Only** - Show only lines matching the filter
2. **Contextual** - Show all lines with matches highlighted, navigate between matches

#### Behavior
- Filter applies to the currently focused view
- Real-time filtering as user types
- Invalid regex patterns show error indicator
- Match count displayed in filter bar

#### Filter Bar Display States
- Empty: `'/' or 'r' to filter`
- Text filter active: `filter: [query]`
- Regex filter active: `regex filter: [pattern]`
- Invalid regex: `invalid regex: [pattern]`
- Context mode OFF (matches only): `filter: [query] (matches only)`
- Context mode ON with matches: `filter: [query] (N/M, n/N to cycle)`
- Context mode ON while editing: `filter: [query] (N/M, enter to apply)`
- Context mode ON with no matches: `filter: [query] (no matches)`

Example: `filter: error (3/217, n/N to cycle)` - showing match 3 of 217, use n/N keys to navigate.

#### Interactions
| Action | Behavior |
|--------|----------|
| / | Enter text filter mode |
| r | Enter regex filter mode |
| Type | Update filter in real-time |
| Enter | Apply filter and exit edit mode |
| Escape | Clear filter |
| n | Jump to next match |
| N | Jump to previous match |
| x | Toggle context display (matches only vs all with highlighting) |

---

### 5. Time Range Control

#### Behavior
- Controls how far back to fetch logs from containers
- When changed, all active scanners restart with new time range
- Applies to all selected containers

#### Time Range Options
| Key | Lookback | Description |
|-----|----------|-------------|
| 0 | Now onwards | Only show new logs from this point forward |
| 1 | 1 minute | Last 1 minute of logs |
| 2 | 5 minutes | Last 5 minutes of logs |
| 3 | 15 minutes | Last 15 minutes of logs |
| 4 | 30 minutes | Last 30 minutes of logs |
| 5 | 1 hour | Last 1 hour of logs |
| 6 | 3 hours | Last 3 hours of logs |
| 7 | 12 hours | Last 12 hours of logs |
| 8 | 24 hours | Last 24 hours (1 day) of logs |
| 9 | All time | All available logs (no time limit) |

#### Toast Notification
When time range changes, display: "Changing time range to start from [X ago]..."

---

### 6. View Navigation

#### Focus States
- One view is always "focused" and receives keyboard input
- Unfocused view remains visible in split mode

#### Interactions
| Action | Behavior |
|--------|----------|
| s | Focus Entities View |
| l | Focus Logs View |
| S (Shift+s) | Full-screen Entities View |
| L (Shift+l) | Full-screen Logs View |
| F (Shift+f) | Toggle full-screen mode |

---

### 7. File Export

#### Save to File
- Saves all visible logs to a file
- Removes ANSI color codes for clean text output
- Default filename: `kl_logs_[timestamp].log`
- Shows toast notification with file path on success

#### Copy to Clipboard
- Copies content to system clipboard
- In Logs View: copies all visible logs
- In Single Log View: copies the single log
- Removes ANSI codes
- Shows toast confirmation

#### Interactions
| Action | Behavior |
|--------|----------|
| Ctrl+S | Save to file |
| Ctrl+Y | Copy to clipboard |

---

### 8. Top Bar

#### Content (left to right)
1. Application name/version (e.g., "kl demo" or "kl v0.5.0")
2. Current time range with live counter (e.g., "Logs for the Last 1m34s")
   - Counter updates in real-time showing elapsed time
3. Container counts: `X/Y/Z Pending/Selected/Total`
   - Pending = containers in ScannerStarting or WantScanning states
   - Selected = containers that may have logs (Scanning or Deleted states)
   - Total = all discovered containers
4. Pause indicator (when paused): `[PAUSED]` (highlighted)
5. Right side: "ctrl+c to quit / ? for help"

#### Example Top Bar
```
kl demo   Logs for the Last 1m34s   0/4/11 Pending/Selected/Total       ctrl+c to quit / ? for help
```

#### Responsive Behavior
- Right-side help text hidden if terminal too narrow
- Background color: lilac

---

### 9. Help Overlay

#### Trigger
- Press `?` to show help overlay
- Press any key to dismiss

#### Behavior
- Displayed as centered modal overlay when user presses `?`
- Overlays the current view (does not replace it)
- Background remains visible but inactive
- Does not interrupt log streaming or other background operations
- Dismissed by pressing any key (including `?` again)

#### Visual Design
- Bordered title: "Help (press any key to hide)"
- Key bindings displayed in multi-column layout (11 items per column)
- Time range bindings displayed below in separate columns (6 items per column)
- Dynamically generated from keymap definitions

#### Help Content

The help overlay displays two sections of key bindings:

**General Key Bindings** (displayed in columns of 11 rows):
```
 enter  select/deselect containers    ↑/k   scroll up
 R      deselect all containers       ↓/j   scroll down
 l      focus logs                    b     pgup
 L      logs fullscreen               f     pgdn
 s      focus selection               g     top
 S      selection fullscreen          G     bottom
 F      toggle fullscreen             ctrl+s save focus to file
 w      toggle line wrap              p     pause/resume logs
 t      show short/full/no timestamps enter zoom on log
 c      show short/full/no names      esc   back to all logs
 o      reverse timestamp order       ctrl+y copy zoomed log
 /      edit filter                   ctrl+c quit
 r      regex filter                  ?     show/hide help
 esc    discard filter
 enter  apply filter
 n      next filter match
 N      prev filter match
 x      filter matches only
```

**Time Range Bindings** (displayed below):
```
 0-9    change log start time
 0      now onwards
 1      1m
 2      5m
 3      15m
 4      30m
 5      1h
 6      3h
 7      12h
 8      1d
 9      all time
```

#### Key Binding Reference (by category)

**Navigation**
| Key | Action |
|-----|--------|
| ↑/k | Move up one line |
| ↓/j | Move down one line |
| g/Ctrl+G | Jump to top |
| G | Jump to bottom |
| u/ctrl+u | Half page up |
| d/ctrl+d | Half page down |
| b/ctrl+b | Full page up |
| f/ctrl+f | Full page down |
| ← | Pan left (unwrapped mode) |
| → | Pan right (unwrapped mode) |

**Views**
| Key | Action |
|-----|--------|
| s | Focus selection/entities panel |
| l | Focus logs panel |
| S | Fullscreen selection view |
| L | Fullscreen logs view |
| F | Toggle fullscreen mode |

**Actions**
| Key | Action |
|-----|--------|
| enter | Select container or zoom into log |
| o | Toggle ascending/descending order |
| t | Cycle timestamp format |
| c | Cycle container name format |
| w | Toggle line wrap |
| p | Pause/resume streaming |
| ctrl+s | Save to file |
| ctrl+y | Copy to clipboard |

**Filtering**
| Key | Action |
|-----|--------|
| / | Text filter |
| r | Regex filter |
| n | Next match |
| N | Previous match |
| x | Toggle context display |
| esc | Clear filter |

**Time Range**
Reference: `0=now 1=1m 2=5m 3=15m 4=30m 5=1h 6=3h 7=12h 8=1d 9=all`

- `0` = From now onwards (only new logs)
- `9` = All available logs (no time limit)

---

### 10. Confirmation Prompts

#### Use Cases
1. Bulk selection changes (e.g., "Select 10 & deselect 5 visible containers?")
2. Deselect all containers

#### Visual Design
- Double-bordered modal dialog
- Centered on screen
- Two buttons: "NO, CANCEL" and "YES, PROCEED"
- Selected button highlighted/inverted

#### Interactions
| Action | Behavior |
|--------|----------|
| ←/→ or h/l or Tab | Toggle between Cancel/Proceed |
| Enter | Execute selected action |
| Escape | Cancel and close |

---

### 11. Toast Notifications

#### Behavior
- Non-blocking notifications at bottom of screen
- Auto-dismiss after 5 seconds
- Used for: file save confirmations, clipboard copy, time range changes, selection limits

---

## State Machines

### Container Log Streaming State Machine (Detailed)

This describes the complete lifecycle of a container's log streaming, including all user interactions at each stage.

#### States

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          CONTAINER LOG STATES                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  UNSELECTED ──────► SELECTED_WAITING ──────► CONNECTING ──────► STREAMING   │
│       ▲                    │                      │                  │      │
│       │                    │                      │                  │      │
│       └────────────────────┴──────────────────────┴──────────────────┘      │
│                           (user deselects)                                  │
│                                                                             │
│  STREAMING ──────► DISCONNECTING ──────► UNSELECTED                         │
│                    (user deselects or container terminates)                 │
│                                                                             │
│  Any State (selected) ──────► DELETED (container removed from cluster)     │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

#### State Definitions

| State | Description | Visual Indicator | Logs Visible |
|-------|-------------|------------------|--------------|
| **Unselected** | Container exists but user has not selected it | `[ ]` | No |
| **Selected_Waiting** | User selected, but container not ready (pending/init) | `[.]` | No |
| **Connecting** | Establishing connection to Kubernetes log stream | `[^]` | No |
| **Streaming** | Actively receiving logs | `[x]` | Yes |
| **Disconnecting** | Gracefully closing log stream | `[v]` | Yes (frozen) |
| **Deleted** | Container no longer exists in cluster | `[d]` | Yes (terminated) |

#### Detailed State Behaviors

##### UNSELECTED
```
Display: [ ] container-name
Behavior:
  - Container appears in entity list but no logs are fetched
  - No network activity for this container
  - Pressing Enter transitions to SELECTED_WAITING or CONNECTING

User Actions Available:
  ✓ Enter → Select container (start streaming)
  ✓ Navigate → Move to other containers
  ✓ Filter → Filter entity list
  ✗ Logs interaction → N/A (no logs exist)
```

##### SELECTED_WAITING
```
Display: [.] container-name
Behavior:
  - User has selected container but it's not ready to stream
  - Occurs when container is: Pending, ContainerCreating, Init, CrashLoopBackOff
  - Automatically transitions to CONNECTING when container becomes Running
  - System polls container status periodically

User Actions Available:
  ✓ Enter → Deselect container (return to UNSELECTED)
  ✓ Navigate → Move to other containers
  ✓ Time range change → Queued for when streaming starts
  ✗ Logs interaction → N/A (no logs yet)

Auto-transitions:
  - Container becomes Running → CONNECTING
  - Container deleted → DELETED (with no logs)
  - User deselects → UNSELECTED
```

##### CONNECTING
```
Display: [^] container-name
Behavior:
  - Establishing WebSocket/HTTP connection to Kubernetes API
  - Requesting logs since configured time range
  - Timeout: 30 seconds before retry
  - On success: transition to STREAMING with initial log batch
  - On failure: retry with exponential backoff (1s, 2s, 4s, 8s, max 30s)

User Actions Available:
  ✓ Enter → Cancel connection attempt, return to UNSELECTED
  ✓ Navigate → Move to other containers
  ✓ Time range change → Restart connection with new time range
  ✗ Logs interaction → N/A (no logs yet)

Auto-transitions:
  - Connection successful → STREAMING
  - Connection failed → Retry CONNECTING (with backoff)
  - Max retries exceeded (10) → DISCONNECTED
  - Container deleted → DELETED
  - User deselects → UNSELECTED (cancel connection)
```

##### STREAMING
```
Display: [x] container-name
Behavior:
  - Actively receiving log lines from Kubernetes
  - New logs appear in real-time (batched every 200ms for performance)
  - Logs are interleaved by timestamp with other streaming containers
  - Maintains connection health via periodic heartbeat

User Actions Available:
  ✓ Enter → Deselect container (transition to DISCONNECTING)
  ✓ Navigate → Move through logs or containers
  ✓ Filter → Filter logs by text or regex
  ✓ Time range change → DISCONNECTING then CONNECTING with new range
  ✓ Pause → Stop displaying new logs (still buffered)
  ✓ Copy/Save → Export current logs
  ✓ Single log view → Zoom into specific log entry

Auto-transitions:
  - Connection lost → DISCONNECTED
  - Container terminates → STREAMING with [TERMINATED] suffix on last log
  - Container deleted → DELETED (logs retained)
  - User deselects → DISCONNECTING
```

##### DISCONNECTING
```
Display: [v] container-name
Behavior:
  - Gracefully closing the log stream connection
  - Waiting for any buffered logs to flush
  - Typically very brief (< 1 second)
  - Logs remain visible during this transition

User Actions Available:
  ✓ Navigate → Move through existing logs
  ✓ Filter → Filter existing logs
  ✓ Copy/Save → Export existing logs
  ✗ Enter → Queued (will re-select after disconnect completes)

Auto-transitions:
  - Disconnect complete + user initiated → UNSELECTED (logs removed)
  - Disconnect complete + time range change → CONNECTING (logs cleared, restart)
  - Disconnect complete + container terminated → SELECTED_WAITING (keep logs)
```

##### DELETED
```
Display: [d] container-name  (with [TERMINATED] on logs)
Behavior:
  - Container no longer exists in Kubernetes cluster
  - All existing logs are retained and marked as terminated
  - Container remains in entity list until user deselects
  - Useful for debugging crashed containers

User Actions Available:
  ✓ Enter → Remove container from list and clear its logs
  ✓ Navigate → Browse terminated logs
  ✓ Filter → Filter terminated logs
  ✓ Copy/Save → Export terminated logs
  ✓ Single log view → Examine specific log entries

Auto-transitions:
  - None (terminal state until user action)
  - User deselects → Container removed from list entirely
```

#### State Transition Diagram with User Actions

```
                                    User presses Enter
                                          │
    ┌─────────────────────────────────────┼─────────────────────────────────────┐
    │                                     ▼                                     │
    │   ┌──────────────┐            ┌───────────────────┐                       │
    │   │  UNSELECTED  │───────────►│  SELECTED_WAITING │◄─────────┐            │
    │   │    [ ]       │  Enter     │       [.]         │          │            │
    │   └──────────────┘            └─────────┬─────────┘          │            │
    │          ▲                              │                    │            │
    │          │                    Container │ Ready              │            │
    │          │                              ▼                    │            │
    │          │                    ┌───────────────────┐          │            │
    │   Enter  │                    │    CONNECTING     │──────────┤            │
    │  (desel) │                    │       [^]         │  Failure │            │
    │          │                    └─────────┬─────────┘  (retry) │            │
    │          │                              │                    │            │
    │          │                    Success   │                    │            │
    │          │                              ▼                    │            │
    │          │                    ┌───────────────────┐          │            │
    │          ├────────────────────│    STREAMING      │◄─────────┘            │
    │          │      Enter         │       [x]         │                       │
    │          │     (desel)        └────────┬──────────┘                       │
    │          │                             │                                  │
    │          │                             │ Container                        │
    │          │                             │ Terminated                       │
    │          │                             ▼                                  │
    │          │                    ┌───────────────────┐                       │
    │          └────────────────────│  SCANNER_STOPPING │                       │
    │                  Enter        │       [v]         │                       │
    │                               └─────────┬─────────┘                       │
    │                                         │                                 │
    │                                         ▼                                 │
    │                                   (Back to UNSELECTED                     │
    │                                    or SELECTED_WAITING)                   │
    │                                                                           │
    │   Container deleted from cluster at any point:                            │
    │                                                                           │
    │   Any State ─────────────────► ┌───────────────────┐                      │
    │   (while selected)             │     DELETED       │                      │
    │            Container           │       [d]         │                      │
    │             Deleted            └─────────┬─────────┘                      │
    │                                          │                                │
    │                                   Enter  │  (remove from list)            │
    │                                          ▼                                │
    │                                   (Entity removed)                        │
    └───────────────────────────────────────────────────────────────────────────┘
```

**Note:** Connection failures during CONNECTING are retried automatically with exponential backoff. This is handled internally and not exposed as a separate visible state.

#### Time Range Change Behavior

When user presses a time range key (1-9, 0):

```
Current State     │ Behavior
──────────────────┼─────────────────────────────────────────────────────────
UNSELECTED        │ Time range stored for when container is selected
SELECTED_WAITING  │ Time range stored for when connection starts
CONNECTING        │ Cancel current connection, restart with new time range
STREAMING         │ → SCANNER_STOPPING → Clear logs → CONNECTING with new range
DELETED           │ No effect (logs are historical)
```

#### Pause Behavior

When user presses 'p' to pause:

```
Current State     │ Behavior
──────────────────┼─────────────────────────────────────────────────────────
STREAMING         │ Logs continue arriving but buffer without display
                  │ "[PAUSED]" shown in top bar
                  │ Press 'p' again: flush buffer, resume display
──────────────────┼─────────────────────────────────────────────────────────
Other States      │ Pause state is remembered, applies when STREAMING starts
```

#### Multiple Container Coordination

When multiple containers are selected:

```
Behavior:
  - Each container has independent state machine
  - Logs from all STREAMING containers interleave by timestamp
  - Time range changes apply to ALL selected containers
  - Pause applies to entire log view (all containers)
  - Container counts in top bar: "Pending/Selected/Total"
    - Pending = SELECTED_WAITING + CONNECTING
    - Selected = STREAMING + DELETED
    - Total = All discovered containers

Example top bar:
  "2/5/20" means:
    - 2 containers pending/connecting
    - 5 containers actively streaming or with retained logs (deleted)
    - 20 total containers discovered
```

#### Error Handling

```
Error Type              │ Behavior
────────────────────────┼──────────────────────────────────────────────────
Network timeout         │ Auto-retry with exponential backoff (internal)
Auth failure (401/403)  │ CONNECTING → UNSELECTED + toast error
Container not found     │ → DELETED state
Rate limited (429)      │ Extended backoff (60s minimum)
API server unavailable  │ Auto-retry with backoff (internal)
Invalid log format      │ Skip malformed line, continue streaming
```

### Entity State Machine (Summary)

Manages the lifecycle of each container's log scanner.

```
States:
  Inactive        - Not selected, no scanning
  WantScanning    - Selected but container not ready (waiting/pending)
  ScannerStarting - Initializing log scanner
  Scanning        - Actively collecting logs
  ScannerStopping - Shutting down scanner
  Deleted         - Container removed from cluster but logs retained

Transitions:
  User selects inactive entity:
    Inactive → ScannerStarting (if container running)
    Inactive → WantScanning (if container waiting)

  User deselects entity:
    WantScanning → Inactive
    Scanning → ScannerStopping → Inactive
    Deleted → (entity removed, logs cleared)

  Container becomes running:
    WantScanning → ScannerStarting → Scanning

  Container terminates:
    Scanning → WantScanning (logs retained, marked terminated)

  Container deleted from cluster:
    Scanning → Deleted (logs retained, marked terminated)
    Other states → Entity removed

  Time range changed:
    Scanning → ScannerStopping → ScannerStarting → Scanning (with new time range)
```

### Application Initialization State

```
1. App starts
2. Request terminal colors (for theming)
3. Receive window size → initialize pages
4. Start container discovery listener
5. Receive first container list → enable interaction
6. Begin log streaming for auto-selected containers
```

### Pause State

```
Unpaused (default):
  - Log buffer flushed to view every 200ms
  - New logs appear in real-time

Paused:
  - Logs continue buffering but not displayed
  - "[PAUSED]" shown in top bar
  - Toggle with 'p' key resumes and flushes buffer
```

---

## Keyboard Reference

### Global Keys
| Key | Action |
|-----|--------|
| Ctrl+C | Quit application |
| ? | Toggle help overlay |
| p | Pause/resume log streaming |
| Ctrl+S | Save logs to file |
| Ctrl+Y | Copy to clipboard |
| 0-9 | Change time range |

### Navigation
| Key | Action |
|-----|--------|
| ↑/k | Move up |
| ↓/j | Move down |
| ← | Pan left (when not wrapped) |
| → | Pan right (when not wrapped) |
| g/Ctrl+G | Jump to top |
| G | Jump to bottom |
| u/Ctrl+U | Half page up |
| d/Ctrl+D | Half page down |
| b/Ctrl+B | Full page up |
| f/Ctrl+F | Full page down |

### View Switching
| Key | Action |
|-----|--------|
| s | Focus Entities View |
| l | Focus Logs View |
| S | Fullscreen Entities View |
| L | Fullscreen Logs View |
| F | Toggle fullscreen |
| Escape | Exit Single Log View / Clear filter |

### Selection (Entities View)
| Key | Action |
|-----|--------|
| Enter | Toggle selection |
| Shift+R | Deselect all |

### Logs View
| Key | Action |
|-----|--------|
| Enter | Open single log |
| o | Toggle order (asc/desc) |
| t | Cycle timestamp format |
| c | Cycle container name format |
| w | Toggle line wrap |

### Filtering
| Key | Action |
|-----|--------|
| / | Text filter |
| r | Regex filter |
| n | Next match |
| N | Previous match |
| x | Toggle context display |
| Escape | Clear filter |
| Enter | Apply filter |

---

## Configuration Options

All configuration is provided via command-line flags. Each flag can also be set via environment variable using the pattern `KL_<FLAG_NAME>` (uppercase, hyphens replaced with underscores).

### Kubernetes Context Configuration

#### `--context`
```
Type:        String (comma-separated list)
Default:     Current context from kubeconfig
Env:         KL_CONTEXT
Example:     --context prod-us-east,prod-us-west,staging
```

**Description**: Specifies which Kubernetes contexts to connect to. When multiple contexts are provided, the application connects to all of them simultaneously, discovering containers across all clusters.

**Behavior**:
- If not specified, uses the current context from kubeconfig
- Multiple contexts enable cross-cluster log viewing
- Each context appears as a top-level node in the entity tree
- Invalid contexts show an error toast but don't prevent other contexts from loading

**Example Usage**:
```bash
# Single context
kl --context production

# Multiple contexts (view logs across clusters)
kl --context prod-east,prod-west,staging

# All contexts matching a pattern (use with --mclust)
kl --mclust "prod-.*"
```

---

#### `--kubeconfig`
```
Type:        String (file path)
Default:     $KUBECONFIG or ~/.kube/config
Env:         KL_KUBECONFIG (or standard KUBECONFIG)
Example:     --kubeconfig /path/to/custom/config
```

**Description**: Path to the kubeconfig file containing cluster credentials and context definitions.

**Behavior**:
- Supports standard kubeconfig format
- Can contain multiple contexts
- If file doesn't exist or is invalid, application exits with error

---

#### `--mclust` (Match Clusters)
```
Type:        String (regular expression)
Default:     None (no auto-matching)
Env:         KL_MCLUST
Example:     --mclust "prod-.*"
```

**Description**: Regular expression pattern to auto-select clusters/contexts. All contexts matching the pattern will be connected to.

**Behavior**:
- Applied against context names in kubeconfig
- Case-sensitive matching
- Combines with `--context` (additive)
- Use with `--iclust` to exclude specific matches

---

#### `--iclust` (Ignore Clusters)
```
Type:        String (regular expression)
Default:     None
Env:         KL_ICLUST
Example:     --iclust ".*-backup$"
```

**Description**: Regular expression pattern to exclude clusters/contexts from selection.

**Behavior**:
- Applied after `--context` and `--mclust`
- Matching contexts are excluded even if explicitly listed
- Useful for excluding backup or test clusters

---

### Namespace Configuration

#### `-n, --namespace`
```
Type:        String (comma-separated list)
Default:     Current namespace from kubeconfig context
Env:         KL_NAMESPACE
Example:     -n kube-system,monitoring,app
```

**Description**: Specifies which namespaces to watch for containers.

**Behavior**:
- If not specified, uses the default namespace from the current kubeconfig context
- Limits container discovery to specified namespaces
- Reduces API load compared to all-namespaces
- Multiple namespaces can be specified
- Non-existent namespaces are silently ignored

**Example Usage**:
```bash
# Single namespace
kl -n production

# Multiple namespaces
kl -n frontend,backend,database

# Combined with context
kl --context prod -n critical-apps
```

---

#### `-A, --all-namespaces`
```
Type:        Boolean flag
Default:     false
Env:         KL_ALL_NAMESPACES=true
Example:     -A
```

**Description**: Watch all namespaces in the cluster(s) for containers.

**Behavior**:
- Overrides `-n, --namespace` when set
- Discovers containers across all namespaces
- May result in large entity lists on busy clusters
- Useful with `--mns`/`--ins` to filter

**Caution**: On large clusters, this can result in thousands of containers. Consider using with `--limit` or namespace filters.

---

#### `--mns` (Match Namespaces)
```
Type:        String (regular expression)
Default:     None
Env:         KL_MNS
Example:     --mns "^app-.*"
```

**Description**: Regular expression to auto-select namespaces. Containers in matching namespaces will be auto-selected on startup.

**Behavior**:
- Applied to namespace names
- Containers in matching namespaces start in CONNECTING state
- Combine with `-A` to first discover all, then auto-select matching

---

#### `--ins` (Ignore Namespaces)
```
Type:        String (regular expression)
Default:     None
Env:         KL_INS
Example:     --ins "^kube-"
```

**Description**: Regular expression to exclude namespaces from display.

**Behavior**:
- Matching namespaces are completely hidden from entity list
- Reduces clutter from system namespaces
- Applied before `--mns`

**Common Usage**:
```bash
# Ignore Kubernetes system namespaces
kl -A --ins "^kube-|^istio-"
```

---

### Pod and Container Matching

#### `--mpod` (Match Pods)
```
Type:        String (regular expression)
Default:     None
Env:         KL_MPOD
Example:     --mpod "api-server-.*"
```

**Description**: Regular expression to auto-select pods. All containers in matching pods will be auto-selected.

**Behavior**:
- Matched against pod names
- All containers within matched pods are selected
- Applied after namespace filtering

---

#### `--ipod` (Ignore Pods)
```
Type:        String (regular expression)
Default:     None
Env:         KL_IPOD
Example:     --ipod ".*-debug$"
```

**Description**: Regular expression to exclude pods from display.

**Behavior**:
- Matching pods are hidden from entity list
- Their containers are never discovered
- Useful for excluding debug/utility pods

---

#### `--mown` (Match Owners)
```
Type:        String (regular expression)
Default:     None
Env:         KL_MOWN
Example:     --mown "api-deployment|worker-statefulset"
```

**Description**: Regular expression to auto-select pod owners (Deployments, StatefulSets, DaemonSets, Jobs, etc.).

**Behavior**:
- Matched against owner reference names
- Selects all pods (and their containers) owned by matching resources
- Useful for selecting by deployment name rather than pod name

**Example Usage**:
```bash
# Select all pods from specific deployments
kl --mown "frontend|backend|api"

# Select all DaemonSet pods
kl --mown ".*-daemonset"
```

---

#### `--iown` (Ignore Owners)
```
Type:        String (regular expression)
Default:     None
Env:         KL_IOWN
Example:     --iown "metrics-.*"
```

**Description**: Regular expression to exclude pod owners from display.

**Behavior**:
- All pods owned by matching resources are hidden
- Applied after `--mown`

---

#### `--mc` (Match Containers)
```
Type:        String (regular expression)
Default:     None
Env:         KL_MC
Example:     --mc "^app$|^worker$"
```

**Description**: Regular expression to auto-select containers by name.

**Behavior**:
- Matched against container names within pods
- Most granular selection option
- Applied after pod filtering

**Example Usage**:
```bash
# Select only main application containers, not sidecars
kl --mc "^(app|api|worker)$"

# Exclude istio sidecars
kl --ic "istio-proxy|envoy"
```

---

#### `--ic` (Ignore Containers)
```
Type:        String (regular expression)
Default:     None
Env:         KL_IC
Example:     --ic "istio-proxy|linkerd-proxy"
```

**Description**: Regular expression to exclude containers from display.

**Behavior**:
- Matching containers are hidden from entity list
- Other containers in the same pod remain visible
- Common use: hiding service mesh sidecars

---

#### `-l, --selector`
```
Type:        String (Kubernetes label selector)
Default:     None
Env:         KL_SELECTOR
Example:     -l "app=nginx,tier in (frontend, backend)"
```

**Description**: Kubernetes label selector to filter pods by labels.

**Behavior**:
- Uses standard Kubernetes label selector syntax
- Supports equality (`=`, `==`, `!=`) and set-based (`in`, `notin`, `exists`) operators
- Applied at the Kubernetes API level (efficient)
- Combines with other filters (AND logic)

**Selector Syntax**:
```bash
# Equality-based
-l "app=nginx"
-l "environment!=production"

# Set-based
-l "tier in (frontend, backend)"
-l "version notin (v1, v2)"

# Existence
-l "app"           # has label 'app'
-l "!debug"        # does not have label 'debug'

# Combined
-l "app=nginx,tier in (frontend, backend),!canary"
```

---

#### `--ignore-owner-types`
```
Type:        String (comma-separated list)
Default:     None
Env:         KL_IGNORE_OWNER_TYPES
Example:     --ignore-owner-types Job,CronJob
```

**Description**: Exclude pods owned by specific Kubernetes resource types.

**Valid Values**:
- `Deployment`
- `StatefulSet`
- `DaemonSet`
- `ReplicaSet`
- `Job`
- `CronJob`
- `ReplicationController`
- `Node` (for static pods)

**Example Usage**:
```bash
# Ignore batch jobs, only show long-running services
kl --ignore-owner-types Job,CronJob
```

---

### Selection Limits

#### `--limit`
```
Type:        Integer
Default:     -1 (unlimited)
Env:         KL_LIMIT
Example:     --limit 20
```

**Description**: Maximum number of containers to auto-select.

**Behavior**:
- Limits how many containers transition from UNSELECTED to CONNECTING
- Applied in discovery order (not deterministic across runs)
- Toast notification shown when limit is reached
- User can manually select additional containers beyond limit
- Set to -1 or omit for unlimited; any positive value sets the limit

**Use Case**: Prevents overwhelming the system when patterns match many containers.

```bash
# Auto-select up to 10 containers matching pattern
kl --mpod "api-.*" --limit 10
```

---

### Display and Startup Options

#### `-f, --log-filter`
```
Type:        String
Default:     None (no filter)
Env:         KL_LOG_FILTER
Example:     -f "error"
```

**Description**: Initial text filter applied to logs on startup.

**Behavior**:
- Filter is active immediately when logs appear
- Equivalent to pressing `/` and typing the filter
- Case-sensitive exact substring match
- Can be cleared with Escape key during session

---

#### `-r, --log-regex`
```
Type:        String (regular expression)
Default:     None (no filter)
Env:         KL_LOG_REGEX
Example:     -r "error|warn|fail"
```

**Description**: Initial regex filter applied to logs on startup.

**Behavior**:
- Regex filter active immediately
- Equivalent to pressing `r` and typing the pattern
- Invalid regex shows error and exits
- Error if both `-f` and `-r` specified (mutually exclusive)

---

#### `-d, --desc`
```
Type:        Boolean flag
Default:     false (ascending order)
Env:         KL_DESC=true
Example:     -d
```

**Description**: Start with logs in descending order (newest first).

**Behavior**:
- Newest logs appear at top of view
- Selection sticks to top as new logs arrive
- Can be toggled with `o` key during session

---

#### `--logs-view`
```
Type:        Boolean flag
Default:     false (entities view focused)
Env:         KL_LOGS_VIEW=true
Example:     --logs-view
```

**Description**: Start with the logs view focused instead of the entities view.

**Behavior**:
- Logs panel receives keyboard focus on startup
- Useful when using auto-selection patterns
- Entity panel still visible in split mode

---

#### `--since`
```
Type:        Duration string (Go duration format)
Default:     "1m" (1 minute)
Env:         KL_SINCE
Example:     --since 1h
```

**Description**: How far back to fetch logs when connecting to containers.

**Valid Formats** (standard Go duration syntax):
```
30s     - 30 seconds
5m      - 5 minutes
2h      - 2 hours
1h30m   - 1 hour 30 minutes
24h     - 24 hours (1 day)
72h     - 72 hours (3 days)
0 or 0s - From now onwards (no historical logs)
```

**Behavior**:
- Applied when container enters CONNECTING state
- Larger values mean more initial logs to fetch
- Can be changed during session with number keys (0-9)
- Very large values may cause slow initial load
- To get "all available logs", use a very large duration (e.g., `8760h` for 1 year) or press `9` during the session

**Example Usage**:
```bash
# Last hour of logs
kl --since 1h

# Last 3 days
kl --since 72h

# Only new logs from now onwards
kl --since 0
```

---

### Filter Evaluation Order

When multiple filters are specified, they are evaluated in this order:

```
1. --context / --mclust / --iclust     → Which clusters to connect to
2. -n / -A / --mns / --ins             → Which namespaces to watch
3. --ignore-owner-types                 → Exclude by owner type
4. --mown / --iown                      → Match/exclude by owner name
5. --mpod / --ipod                      → Match/exclude by pod name
6. -l (selector)                        → Filter by pod labels
7. --mc / --ic                          → Match/exclude by container name
8. --limit                              → Cap number of auto-selections
```

**Logic**:
- "Ignore" patterns (`--i*`) always take precedence (exclusions win)
- "Match" patterns (`--m*`) cause auto-selection of matching items
- Items not matched by any pattern remain visible but unselected
- User can manually select any visible item regardless of patterns

---

### Configuration Examples

#### Basic Usage
```bash
# View logs from default namespace in current context
kl

# View logs from specific namespace
kl -n production

# View logs from all namespaces
kl -A
```

#### Multi-Cluster Setup
```bash
# View logs across multiple production clusters
kl --context prod-us-east,prod-us-west,prod-eu -n critical-services

# All clusters matching pattern
kl --mclust "prod-.*" -A
```

#### Filtering by Application
```bash
# All pods from a specific deployment
kl -n myapp --mown "api-deployment"

# Pods matching label selector
kl -l "app=nginx,version=v2"

# Exclude sidecars
kl -A --ic "istio-proxy|envoy|linkerd"
```

#### Debugging Specific Issues
```bash
# Last hour of logs, filtered to errors, newest first
kl -n production --since 1h -r "error|exception|panic" -d

# Specific pod with all containers
kl -n myapp --mpod "api-server-abc123"
```

#### Large Cluster Management
```bash
# Limit auto-selection to prevent overload
kl -A --mown "critical-.*" --limit 50

# Exclude system namespaces
kl -A --ins "^kube-|^istio-|^monitoring$"
```

---

### Environment Variable Reference

| Flag | Environment Variable | Example |
|------|---------------------|---------|
| `--context` | `KL_CONTEXT` | `KL_CONTEXT=prod,staging` |
| `--kubeconfig` | `KL_KUBECONFIG` | `KL_KUBECONFIG=/custom/config` |
| `--mclust` | `KL_MCLUST` | `KL_MCLUST=prod-.*` |
| `--iclust` | `KL_ICLUST` | `KL_ICLUST=.*-backup` |
| `-n, --namespace` | `KL_NAMESPACE` | `KL_NAMESPACE=app,db` |
| `-A, --all-namespaces` | `KL_ALL_NAMESPACES` | `KL_ALL_NAMESPACES=true` |
| `--mns` | `KL_MNS` | `KL_MNS=^app-` |
| `--ins` | `KL_INS` | `KL_INS=^kube-` |
| `--mpod` | `KL_MPOD` | `KL_MPOD=api-.*` |
| `--ipod` | `KL_IPOD` | `KL_IPOD=.*-test` |
| `--mown` | `KL_MOWN` | `KL_MOWN=frontend` |
| `--iown` | `KL_IOWN` | `KL_IOWN=batch-.*` |
| `--mc` | `KL_MC` | `KL_MC=^app$` |
| `--ic` | `KL_IC` | `KL_IC=sidecar` |
| `-l, --selector` | `KL_SELECTOR` | `KL_SELECTOR=app=nginx` |
| `--ignore-owner-types` | `KL_IGNORE_OWNER_TYPES` | `KL_IGNORE_OWNER_TYPES=Job` |
| `--limit` | `KL_LIMIT` | `KL_LIMIT=25` |
| `-f, --log-filter` | `KL_LOG_FILTER` | `KL_LOG_FILTER=error` |
| `-r, --log-regex` | `KL_LOG_REGEX` | `KL_LOG_REGEX=error\|warn` |
| `-d, --desc` | `KL_DESC` | `KL_DESC=true` |
| `--logs-view` | `KL_LOGS_VIEW` | `KL_LOGS_VIEW=true` |
| `--since` | `KL_SINCE` | `KL_SINCE=1h` |

**Precedence**: Command-line flags override environment variables.

---

## Visual Styling

### Color Scheme
- **Lilac** (Color 189) - Top bar, borders, headers
- **Green** (Color 46) - Timestamps
- **Blue** (Color 6) - Accent elements
- **Container colors** - Each container assigned unique colors for name prefix and name

### Text Styles
- Normal text
- Bold
- Inverse (for selections, focused elements)
- Underline
- Alt inverse (secondary selections)

### Terminal Adaptation
- Detects terminal background color (dark/light)
- Adjusts foreground colors for readability
- Falls back to defaults if detection fails

---

## Component Hierarchy

```
App
├── Top Bar
├── Pages Container
│   ├── Entities Page
│   │   └── FilterableViewport
│   │       ├── Filter Bar
│   │       └── Viewport (selectable list)
│   ├── Logs Page
│   │   └── FilterableViewport
│   │       ├── Filter Bar
│   │       └── Viewport (log lines)
│   └── Single Log Page
│       └── FilterableViewport
│           ├── Filter Bar
│           └── Viewport (formatted log)
├── Help Overlay (conditional)
├── Prompt Modal (conditional)
└── Toast Notification (conditional)
```

---

## Additional Technical Details

### Container Color Assignment

Each container is assigned a consistent color based on its name using MD5 hashing:

**Color Palette (13 colors)**:
- Blue (#58A2EE)
- Bright Green (#3FE34B)
- Purple (#7c60d7)
- Red (#FD2C4C)
- Orange (#FE7A00)
- Yellow (#FAF81C)
- Teal (#56EBD3)
- Green (#42952E)
- Light Pink (#FFACE6)
- Bright Pink (#FE16F4)
- Gold (#D6A112)
- Beige (#FFDAB9)
- Tomato (#FF7E6A)

**Behavior**:
- Color is deterministically assigned using MD5 hash of container name
- Same container name always gets same color across sessions
- Ensures visual consistency for debugging

---

### GKE Authentication Plugin

When connecting to Google Kubernetes Engine (GKE) clusters:

**Automatic Validation**:
- Application checks if `gke-gcloud-auth-plugin` is in system PATH
- Only triggered when kubeconfig authInfo requires this plugin

**Error Handling**:
- If plugin not found, displays helpful error message
- Includes installation hint from kubeconfig
- Does not prevent other (non-GKE) contexts from loading

**Example Error**:
```
gke-gcloud-auth-plugin not found in system PATH for context gke-prod.
  - [installation hint from kubeconfig] and ensure 'google-cloud-sdk/bin' is in your system's PATH.
```

---

### Developer/Debug Features

Hidden environment variables for troubleshooting:

| Variable | Description | Default |
|----------|-------------|---------|
| `KL_DEBUG` | Enable debug logging to file (any non-empty value) | Disabled |
| `KL_DEBUG_PATH` | Path to debug log file | `kl.log` |
| `KL_PPROF_SERVER` | Enable pprof profiling server on port 6060 | Disabled |

**Debug Logging**:
- Logs all tea.Msg updates with timestamps
- Skips high-frequency messages (BatchUpdateLogsMsg, BlinkMsg)
- Useful for diagnosing state machine issues

---

### Performance Tuning Constants

Internal timing values that affect responsiveness:

| Constant | Value | Description |
|----------|-------|-------------|
| SingleContainerLogCollectionDuration | 150ms | How long each container collects logs before sending |
| GetNextContainerDeltasDuration | 300ms | How long to collect container changes before updating |
| BatchUpdateLogsInterval | 200ms | How often logs view receives batched updates |
| CheckStylesLoadedDuration | 200ms | Delay before checking terminal color detection |
| AttemptUpdateSinceTimeInterval | 500ms | Retry interval for time range changes |

---

### Layout Constants

| Constant | Value | Description |
|----------|-------|-------------|
| LeftPageWidthFraction | 2/5 (40%) | Width of entities panel in split view |
| MinCharsEachSideShortNames | 2 | Minimum characters shown on each side of abbreviated names |
| ConfirmSelectionActionsThreshold | 5 | Number of selection changes requiring confirmation prompt |
| NewContainerThreshold | 3 minutes | Containers newer than this are annotated as "new" |

---

### New Container Annotation

Containers discovered within the last 3 minutes display a visual indicator marking them as new. This helps users identify recently started containers.

---

## Verification Checklist

When implementing, verify:

1. **Container Discovery** - Containers appear as they're discovered across contexts
2. **Selection** - Selecting containers starts log streaming; deselecting stops it
3. **Log Interleaving** - Logs from multiple containers merge by timestamp
4. **Real-time Updates** - New logs appear automatically (when not paused)
5. **Filtering** - Text and regex filters work; context toggle functions
6. **Time Range** - Changing time range restarts scanners correctly
7. **View Navigation** - Split/fullscreen modes; focus switching
8. **Single Log** - JSON formatting; escape sequence expansion
9. **File Export** - Saves/copies without ANSI codes
10. **Keyboard Shortcuts** - All documented keys function correctly
11. **Terminal Resize** - Layout adapts to terminal size changes
12. **Container Lifecycle** - Handles container starts/stops/restarts gracefully
