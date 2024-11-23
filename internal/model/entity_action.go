package model

type EntityAction int

const (
	StartScanner EntityAction = iota
	StopScanner
	StopScannerKeepLogs
	RemoveEntity
	RemoveLogs
	MarkLogsTerminated
)

func (a EntityAction) String() string {
	switch a {
	case StartScanner:
		return "StartScanner"
	case StopScanner:
		return "StopScanner"
	case StopScannerKeepLogs:
		return "StopScannerKeepLogs"
	case RemoveEntity:
		return "RemoveEntity"
	case RemoveLogs:
		return "RemoveLogs"
	case MarkLogsTerminated:
		return "MarkLogsTerminated"
	default:
		return "Unknown"
	}
}
