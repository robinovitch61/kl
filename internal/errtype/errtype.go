package errtype

type LogScannerStoppedErr struct{}

func (e LogScannerStoppedErr) Error() string {
	return "log scanner stopped"
}
