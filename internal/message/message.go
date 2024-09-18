package message

type ErrMsg struct{ Err error }

func (e ErrMsg) Error() string { return e.Err.Error() }

type CleanupCompleteMsg struct{}

type BatchUpdateLogsMsg struct{}

type StartMaintainEntitySelectionMsg struct{}

type AttemptUpdateSinceTimeMsg struct{}

type UpdateSinceTimeTextMsg struct {
	UUID string
}
