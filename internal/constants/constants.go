package constants

import (
	"regexp"
	"time"
)

// *********************************************************************************************************************
// THESE ARE KEY TO GOOD PERFORMANCE & RESPONSIVENESS IN HIGH LOG VOLUME SETTINGS (EXACT VALUES DETERMINED BY FEEL)

// SingleContainerLogCollectionDuration controls the amount of time an individual container's log scanner will collect
// logs for before returning them to the main Model via a tea.Msg
var SingleContainerLogCollectionDuration = 150 * time.Millisecond

// GetNextContainerDeltasDuration controls the amount of time a container listener will collect containers deltas
// before returning them to the main Model via a tea.Msg
var GetNextContainerDeltasDuration = 300 * time.Millisecond

// BatchUpdateLogsInterval controls the cadence at which the main Model actually updates the logs page with all
// the newly acquired logs from all the containers. In between updates, it accumulates logs from received messages
var BatchUpdateLogsInterval = 200 * time.Millisecond

// AttemptUpdateSinceTimeInterval controls the cadence at which a "since time" update is attempted
var AttemptUpdateSinceTimeInterval = 500 * time.Millisecond

// *********************************************************************************************************************

var AnsiRegex = regexp.MustCompile("\x1b\\[[0-9;]*m")

var EmptySequenceRegex = regexp.MustCompile("\x1b\\[[0-9;]+m\x1b\\[m")

// LeftPageWidthFraction controls the width of the left page as a fraction of the terminal width
const LeftPageWidthFraction = 2. / 5.

// MinCharsEachSideShortNames controls the minimum number of characters to show on each side of container shortnames
const MinCharsEachSideShortNames = 2

// ConfirmSelectionActionsThreshold controls the number of actions that require confirmation before executing
const ConfirmSelectionActionsThreshold = 5

// KeyPressToLookbackMins maps the keyboard number keys to minutes of log lookback
var KeyPressToLookbackMins = map[int]int{
	0: 0, // from now on
	1: 1,
	2: 5,
	3: 15,
	4: 30,
	5: 60,
	6: 180,  // 3hrs
	7: 720,  // 12hrs
	8: 1440, // 24hrs
	9: -1,   // max
}

// NewContainerThreshold controls when a container is annotated as new in the tree
const NewContainerThreshold = 3 * time.Minute

// InitialLookbackMins controls the initial number of minutes to look back in logs
var InitialLookbackMins = 1

// AttemptMaintainEntitySelectionAfterFirstContainer controls how long to delay after the first container before
// attempting to maintain the currently selected Entity in the tree. The thinking goes that the tree may not be fully
// populated yet and the user won't even have time to orient themselves and then the selection is somewhere in the
// middle of the tree. But after a short amount of time, they will have actively selected something and we can try to
// maintain that selection
var AttemptMaintainEntitySelectionAfterFirstContainer = 1 * time.Second
