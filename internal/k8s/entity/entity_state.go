package entity

type EntityState int

const (
	// Inactive is a dormant state
	Inactive EntityState = iota

	// WantScanning is a state where the entity will start scanning when the container state allows it
	WantScanning

	// ScannerStarting is a state where the entity's log scanner is being initiated
	ScannerStarting

	// Scanning is a state where the entity's log scanner is actively scanning
	Scanning

	// ScannerStopping is a state where the entity's log scanner is being stopped
	ScannerStopping

	// Deleted is a state where the entity's container is garbage collected in the cluster but the entity is still
	// visually selected in the entity tree
	Deleted

	// Removed is a state where the entity is removed from the entity tree
	// it is not currently used
	Removed
)

func (s EntityState) String() string {
	switch s {
	case Inactive:
		return "Inactive"
	case WantScanning:
		return "WantScanning"
	case ScannerStarting:
		return "ScannerStarting"
	case Scanning:
		return "Scanning"
	case ScannerStopping:
		return "ScannerStopping"
	case Deleted:
		return "Deleted"
	case Removed:
		return "Removed"
	default:
		return "Unknown"
	}
}

func (s EntityState) StatusIndicator() string {
	if s == WantScanning {
		return "[.]"
	} else if s == ScannerStarting {
		return "[^]"
	} else if s == Scanning || s == Deleted {
		return "[x]"
	} else if s == ScannerStopping {
		return "[v]"
	}
	return "[ ]"
}

func (s EntityState) ActivatesWhenSelected() bool {
	switch s {
	case Scanning, WantScanning, Deleted:
		return false
	default:
		return true
	}
}

func (s EntityState) MayHaveLogs() bool {
	switch s {
	case Scanning, ScannerStopping, Deleted:
		return true
	default:
		return false
	}
}
