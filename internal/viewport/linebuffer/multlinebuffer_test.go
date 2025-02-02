package linebuffer

import (
	"fmt"
	"github.com/charmbracelet/lipgloss/v2"
	"strings"
	"testing"
)

// helper to split a string into individual runes for testing
func splitIntoRunes(s string) []string {
	var result []string
	for _, r := range s {
		result = append(result, string(r))
	}
	return result
}

func TestMultiLineBuffer_Equivalence(t *testing.T) {
	// Test cases that should behave identically between LineBuffer and MultiLineBuffer
	tests := []struct {
		name         string
		input        string
		width        int
		continuation string
		toHighlight  string
		numPopLefts  int
	}{
		{
			name:         "simple string",
			input:        "hello world",
			width:        5,
			continuation: "",
			numPopLefts:  3,
		},
		//{
		//	name:         "unicode string",
		//	input:        "世界🌟世界",
		//	width:        4,
		//	continuation: "",
		//	numPopLefts:  3,
		//},
		//{
		//	name:         "ansi string",
		//	input:        "\x1b[38;2;255;0;0mhello\x1b[m world",
		//	width:        6,
		//	continuation: "",
		//	numPopLefts:  3,
		//},
		//{
		//	name:         "with continuation",
		//	input:        "hello world",
		//	width:        5,
		//	continuation: "...",
		//	numPopLefts:  3,
		//},
		//{
		//	name:        "with highlight",
		//	input:       "hello world test string",
		//	width:       10,
		//	toHighlight: "test",
		//	numPopLefts: 3,
		//},
		//{
		//	name:         "zero width",
		//	input:        "hello",
		//	width:        0,
		//	continuation: "...",
		//	numPopLefts:  2,
		//},
		//{
		//	name:         "empty string",
		//	input:        "",
		//	width:        5,
		//	continuation: "...",
		//	numPopLefts:  1,
		//},
	}

	highlightStyle := lipgloss.NewStyle().Background(lipgloss.Color("#FF0000"))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create MultiLineBuffer with same content split different ways
			splits := [][]LineBuffer{
				// Single buffer (should be identical to LineBuffer)
				{New(tt.input)},
				// Split into words
				func() []LineBuffer {
					words := strings.Fields(tt.input)
					var buffers []LineBuffer
					for i, word := range words {
						if i > 0 {
							buffers = append(buffers, New(" "))
						}
						buffers = append(buffers, New(word))
					}
					return buffers
				}(),
				//// Split into individual runes
				//func() []LineBuffer {
				//	var buffers []LineBuffer
				//	for _, s := range splitIntoRunes(tt.input) {
				//		buffers = append(buffers, New(s))
				//	}
				//	return buffers
				//}(),
			}

			for splitIdx, buffers := range splits {
				t.Run(fmt.Sprintf("a%d", splitIdx), func(t *testing.T) {
					// Create single LineBuffer
					singleBuf := New(tt.input)
					multiBuf := NewMulti(buffers...)

					// Test Width
					if singleBuf.Width() != multiBuf.Width() {
						t.Errorf("Width mismatch: single=%d, multi=%d", singleBuf.Width(), multiBuf.Width())
					}

					// Test Content
					if singleBuf.Content() != multiBuf.Content() {
						t.Errorf("Content mismatch:\nsingle=%q\nmulti=%q", singleBuf.Content(), multiBuf.Content())
					}

					// Test PopLeft sequence
					for i := 0; i < tt.numPopLefts; i++ {
						singleResult := singleBuf.PopLeft(tt.width, tt.continuation, tt.toHighlight, highlightStyle)
						multiResult := multiBuf.PopLeft(tt.width, tt.continuation, tt.toHighlight, highlightStyle)

						if singleResult != multiResult {
							t.Errorf("PopLeft %d mismatch:\nsingle=%q\nmulti=%q", i, singleResult, multiResult)
						}
					}

					// Test SeekToWidth and PopLeft combinations
					seekWidths := []int{0, 1, tt.width / 2, tt.width, tt.width * 2}
					for _, seekWidth := range seekWidths {
						singleBuf.SeekToWidth(seekWidth)
						multiBuf.SeekToWidth(seekWidth)

						singleResult := singleBuf.PopLeft(tt.width, tt.continuation, tt.toHighlight, highlightStyle)
						multiResult := multiBuf.PopLeft(tt.width, tt.continuation, tt.toHighlight, highlightStyle)

						if singleResult != multiResult {
							t.Errorf("After SeekToWidth(%d) PopLeft mismatch:\nsingle=%q\nmulti=%q",
								seekWidth, singleResult, multiResult)
						}
					}

					// Test WrappedLines
					maxLinesTests := []int{-1, 0, 1, 2, 5}
					for _, maxLines := range maxLinesTests {
						singleWrapped := singleBuf.WrappedLines(tt.width, maxLines, tt.toHighlight, highlightStyle)
						multiWrapped := multiBuf.WrappedLines(tt.width, maxLines, tt.toHighlight, highlightStyle)

						if len(singleWrapped) != len(multiWrapped) {
							t.Errorf("WrappedLines length mismatch for maxLines=%d: single=%d, multi=%d",
								maxLines, len(singleWrapped), len(multiWrapped))
						}

						for i := range singleWrapped {
							if singleWrapped[i] != multiWrapped[i] {
								t.Errorf("WrappedLines[%d] mismatch for maxLines=%d:\nsingle=%q\nmulti=%q",
									i, maxLines, singleWrapped[i], multiWrapped[i])
							}
						}
					}
				})
			}
		})
	}
}

func TestMultiLineBuffer_EmptyBuffers(t *testing.T) {
	tests := []struct {
		name     string
		buffers  []LineBuffer
		expected string
	}{
		{
			name:     "no buffers",
			buffers:  nil,
			expected: "",
		},
		{
			name: "some empty buffers",
			buffers: []LineBuffer{
				New(""),
				New("hello"),
				New(""),
				New("world"),
				New(""),
			},
			expected: "helloworld",
		},
		{
			name: "all empty buffers",
			buffers: []LineBuffer{
				New(""),
				New(""),
				New(""),
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			multi := NewMulti(tt.buffers...)
			if content := multi.Content(); content != tt.expected {
				t.Errorf("expected content %q, got %q", tt.expected, content)
			}
		})
	}
}

func TestMultiLineBuffer_NestedAnsi(t *testing.T) {
	// Test complex nested ANSI scenarios
	tests := []struct {
		name     string
		buffers  []LineBuffer
		width    int
		expected string
	}{
		{
			name: "split within ANSI sequence",
			buffers: []LineBuffer{
				New("\x1b[31mhe"),
				New("llo\x1b[m"),
			},
			width:    5,
			expected: "\x1b[31mhello\x1b[m",
		},
		{
			name: "multiple nested styles",
			buffers: []LineBuffer{
				New("\x1b[31m"),
				New("he"),
				New("\x1b[1m"),
				New("ll"),
				New("\x1b[4m"),
				New("o"),
				New("\x1b[m"),
			},
			width:    5,
			expected: "\x1b[31mhe\x1b[1mll\x1b[4mo\x1b[m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			multi := NewMulti(tt.buffers...)
			result := multi.PopLeft(tt.width, "", "", lipgloss.NewStyle())
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMultiLineBuffer_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		buffers      []LineBuffer
		width        int
		continuation string
		expected     []string
	}{
		{
			name: "split unicode character boundary",
			buffers: []LineBuffer{
				New("世"),
				New("界"),
			},
			width:    2,
			expected: []string{"世", "界"},
		},
		{
			name: "continuation across buffer boundary",
			buffers: []LineBuffer{
				New("abc"),
				New("def"),
			},
			width:        3,
			continuation: "...",
			expected:     []string{"...", "..."},
		},
		{
			name: "mixed width characters",
			buffers: []LineBuffer{
				New("a世"),
				New("b界"),
			},
			width:    3,
			expected: []string{"a世", "b界"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			multi := NewMulti(tt.buffers...)
			for i, expected := range tt.expected {
				result := multi.PopLeft(tt.width, tt.continuation, "", lipgloss.NewStyle())
				if result != expected {
					t.Errorf("PopLeft %d: expected %q, got %q", i, expected, result)
				}
			}
		})
	}
}
