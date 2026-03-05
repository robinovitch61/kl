package util

import (
	"regexp"
	"runtime"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/google/go-cmp/cmp"
)

var ansiRegex = regexp.MustCompile(`(\x1b\[[0-9;]*m.*?\x1b\[0?m)`)

// GetUniqueShortNames takes in the set of names and tries to create a unique set of substrings of them,
// starting at minChars length and increasing if necessary. If fromRight is true, it starts from the
// end of each name.
// Returns a mapping of name -> short name with the names in nameSet as keys.
func GetUniqueShortNames(nameSet map[string]bool, fromRight bool, minChars int) map[string]string {
	for {
		foundUniqueSet := true
		alreadySeenFromShortName := make(map[string]bool)
		shortNameFromName := make(map[string]string)
		for name := range nameSet {
			var shortName string
			if fromRight {
				shortName = name[len(name)-min(minChars, len(name)):]
			} else {
				shortName = name[:min(minChars, len(name))]
			}
			if alreadySeenFromShortName[shortName] {
				foundUniqueSet = false
				break
			}
			alreadySeenFromShortName[shortName] = true
			shortNameFromName[name] = shortName
		}
		if foundUniqueSet {
			for name, shortName := range shortNameFromName {
				if len(shortName) < len(name) {
					if fromRight {
						shortNameFromName[name] = ".." + shortName
					} else {
						shortNameFromName[name] = shortName + ".."
					}
				}
			}
			return shortNameFromName
		}
		minChars++
	}
}

// GetUniqueShortNamesFromEdges does the same thing as GetUniqueShortNames, but removes chars in the middle
// of the name instead of the beginning or end.
// if e.g. apple would return app..ple, just return apple
func GetUniqueShortNamesFromEdges(nameSet map[string]bool, minCharsEachSide int) map[string]string {
	for {
		foundUniqueSet := true
		alreadySeenFromShortName := make(map[string]bool)
		shortNameFromName := make(map[string]string)
		for name := range nameSet {
			if len(name) <= 2*minCharsEachSide {
				shortNameFromName[name] = name
				continue
			}
			shortName := name[:minCharsEachSide] + ".." + name[len(name)-minCharsEachSide:]
			if alreadySeenFromShortName[shortName] {
				foundUniqueSet = false
				break
			}
			alreadySeenFromShortName[shortName] = true
			shortNameFromName[name] = shortName
		}
		if foundUniqueSet {
			return shortNameFromName
		}
		minCharsEachSide++
	}
}

func JoinWithEqualSpacing(width int, items ...string) string {
	if len(items) == 0 {
		return ""
	}

	totalContentWidth := 0
	for _, item := range items {
		totalContentWidth += lipgloss.Width(item)
	}

	if width <= 0 {
		return ""
	}

	if totalContentWidth <= width {
		// if enough space, proceed with equal spacing
		if len(items) == 1 {
			return items[0]
		}

		totalSpacing := width - totalContentWidth
		baseSpacing := totalSpacing / (len(items) - 1)
		extraSpacing := totalSpacing % (len(items) - 1)

		var result strings.Builder

		for i, item := range items {
			result.WriteString(item)
			if i < len(items)-1 {
				spaces := baseSpacing
				if i < extraSpacing {
					spaces++
				}
				result.WriteString(strings.Repeat(" ", spaces))
			}
		}

		return result.String()
	}
	// if not enough space, truncate from the right
	var result strings.Builder
	remainingWidth := width

	for _, item := range items {
		itemWidth := lipgloss.Width(item)
		if remainingWidth <= 0 {
			break
		}
		if itemWidth > remainingWidth {
			result.WriteString(lipgloss.NewStyle().MaxWidth(remainingWidth).Render(item))
			break
		}
		result.WriteString(item)
		remainingWidth -= itemWidth
	}

	return result.String()
}

// StyleStyledString is for styling a string that contains ANSI escape codes.
func StyleStyledString(s string, st lipgloss.Style) string {
	split := ansiRegex.Split(s, -1)
	matches := ansiRegex.FindAllString(s, -1)

	finalResult := ""
	for i, section := range split {
		if section != "" {
			finalResult += st.Render(section)
		}
		if i < len(split)-1 && i < len(matches) {
			finalResult += matches[i]
		}
	}
	return finalResult
}

// SanitizeTerminalSequences removes terminal control sequences that could affect
// the terminal (cursor movement, screen clearing, etc.) while preserving ANSI
// styling sequences (SGR: colors, bold, underline, etc.).
func SanitizeTerminalSequences(s string) string {
	var buf strings.Builder
	buf.Grow(len(s))
	i := 0
	for i < len(s) {
		b := s[i]

		if b == '\x1b' {
			if i+1 >= len(s) {
				// lone ESC at end
				i++
				continue
			}
			next := s[i+1]

			// CSI sequence: ESC [
			if next == '[' {
				j := i + 2
				// skip parameter bytes (0x30-0x3F)
				for j < len(s) && s[j] >= 0x30 && s[j] <= 0x3F {
					j++
				}
				// skip intermediate bytes (0x20-0x2F)
				for j < len(s) && s[j] >= 0x20 && s[j] <= 0x2F {
					j++
				}
				// final byte (0x40-0x7E)
				if j < len(s) && s[j] >= 0x40 && s[j] <= 0x7E {
					if s[j] == 'm' {
						// SGR sequence - keep styling
						buf.WriteString(s[i : j+1])
					}
					i = j + 1
					continue
				}
				// incomplete CSI - skip ESC
				i++
				continue
			}

			// OSC sequence: ESC ] ... (BEL or ST)
			if next == ']' {
				j := i + 2
				for j < len(s) {
					if s[j] == '\x07' {
						j++
						break
					}
					if s[j] == '\x1b' && j+1 < len(s) && s[j+1] == '\\' {
						j += 2
						break
					}
					j++
				}
				i = j
				continue
			}

			// DCS, PM, APC sequences: ESC P, ESC ^, ESC _
			if next == 'P' || next == '^' || next == '_' {
				j := i + 2
				for j < len(s) {
					if s[j] == '\x1b' && j+1 < len(s) && s[j+1] == '\\' {
						j += 2
						break
					}
					j++
				}
				i = j
				continue
			}

			// other single-char escape sequences (e.g. ESC c = reset) - skip
			i += 2
			continue
		}

		// remove C0 control chars except tab (already converted to spaces above)
		if b < 0x20 && b != '\t' {
			i++
			continue
		}
		// remove DEL
		if b == 0x7F {
			i++
			continue
		}

		buf.WriteByte(b)
		i++
	}
	return buf.String()
}

// CmpStr compares two strings and fails the test if they are not equal
func CmpStr(t *testing.T, expected, actual string) {
	_, file, line, _ := runtime.Caller(1)
	testName := t.Name()
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("\nTest %q failed at %s:%d\nDiff (-expected +actual):\n%s", testName, file, line, diff)
	}
}
