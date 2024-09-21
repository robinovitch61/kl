package viewport

import (
	"testing"
)

func TestViewport_StringWidth(t *testing.T) {
	testCases := []struct {
		input         string
		expectedWidth int
	}{
		{
			"\x1b[7mhi\x1b[38;2;214;125;17mthere\x1b[0mblah\x1b[0mextra",
			len("hithereblahextra"),
		},
		{
			"\x1b[31mHello, World!\x1b[0m",
			13, // Expected width of "Hello, World!" without ANSI codes
		},
		{
			"\x1b[1mBold Text\x1b[0m",
			9, // Expected width of "Bold Text"
		},
		{
			"No ANSI here, just plain text",
			29, // Expected width of the plain string
		},
		{
			"\x1b[1m\x1b[0m",
			0, // Only bold and reset codes, no text
		},
	}

	for _, tc := range testCases {
		result := stringWidth(tc.input)
		if result != tc.expectedWidth {
			t.Errorf("For input '%s', expected width %d, but got %d", tc.input, tc.expectedWidth, result)
		}
	}
}
