package linebuffer

import (
	"github.com/charmbracelet/lipgloss/v2"
	"testing"
)

var equivalentLineBuffers = map[string][]LineBufferer{
	// TODO LEO: add ansi, unicode
	"hello world": {
		New("hello world"),
		NewMulti(
			New("hello"),
			New(" world"),
		),
		NewMulti(
			New("hel"),
			New("lo "),
			New("wo"),
			New("rld"),
		),
		NewMulti(
			New("h"),
			New("e"),
			New("l"),
			New("l"),
			New("o"),
			New(" "),
			New("w"),
			New("o"),
			New("r"),
			New("l"),
			New("d"),
		),
	},
}

func TestMultiLineBuffer_Width(t *testing.T) {
	for _, eq := range equivalentLineBuffers {
		for i, lb := range eq {
			if lb.Width() != eq[0].Width() {
				t.Errorf("expected %d, got %d for line buffer %d", eq[0].Width(), lb.Width(), i)
			}
		}
	}
}

func TestMultiLineBuffer_Content(t *testing.T) {
	for _, eq := range equivalentLineBuffers {
		for i, lb := range eq {
			if lb.Content() != eq[0].Content() {
				t.Errorf("expected %q, got %q for line buffer %d", eq[0].Content(), lb.Content(), i)
			}
		}
	}
}

func TestMultiLineBuffer_SeekToWidth(t *testing.T) {
	tests := []struct {
		name            string
		key             string
		seekToWidth     int
		takeWidth       int
		continuation    string
		expectedPopLeft string
	}{
		{
			name:            "hello world 0",
			key:             "hello world",
			seekToWidth:     0,
			takeWidth:       7,
			continuation:    "",
			expectedPopLeft: "hello w",
		},
		{
			name:            "hello world 1",
			key:             "hello world",
			seekToWidth:     1,
			takeWidth:       7,
			continuation:    "",
			expectedPopLeft: "ello wo",
		},
		{
			name:            "hello world end",
			key:             "hello world",
			seekToWidth:     10,
			takeWidth:       3,
			continuation:    "",
			expectedPopLeft: "d",
		},
		{
			name:            "hello world past end",
			key:             "hello world",
			seekToWidth:     11,
			takeWidth:       3,
			continuation:    "",
			expectedPopLeft: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, eq := range equivalentLineBuffers[tt.key] {
				eq.SeekToWidth(tt.seekToWidth)
				if actual := eq.PopLeft(tt.takeWidth, tt.continuation, "", lipgloss.NewStyle()); actual != tt.expectedPopLeft {
					t.Errorf("for %s, expected %q, got %q", eq.Repr(), tt.expectedPopLeft, actual)
				}
			}
		})
	}
}

func TestMultiLineBuffer_PopLeft(t *testing.T) {
	// TODO LEO
}
