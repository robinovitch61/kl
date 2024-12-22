package fixtures

import (
	"github.com/google/go-cmp/cmp"
	"runtime"
	"testing"
)

func Cmp(t *testing.T, expected, actual string) {
	_, file, line, _ := runtime.Caller(1)
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("\nTest failed at %s:%d\nMismatch (-expected, +actual):\n%s", file, line, diff)
	}
}
