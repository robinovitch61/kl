package fixtures

import (
	"github.com/google/go-cmp/cmp"
	"runtime"
	"testing"
)

func Cmp(t *testing.T, expected, actual string) {
	_, file, line, _ := runtime.Caller(1)
	testName := t.Name()
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("\nTest %q failed at %s:%d\nDiff (-expected +actual):\n%s", testName, file, line, diff)
	}
}
