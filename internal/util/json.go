package util

import (
	"bytes"
	"encoding/json"
	"strings"

	"charm.land/lipgloss/v2"
)

// PrettyPrintJSON attempts to pretty-print JSON input. Returns the input as-is if not valid JSON.
// The optional colorize function is applied after indentation but before splitting escaped
// newlines/tabs within string values, so ANSI codes don't get split across lines.
func PrettyPrintJSON(input string, colorize func(string) string) []string {
	var raw map[string]interface{}

	err := json.Unmarshal([]byte(input), &raw)
	if err != nil {
		return []string{input}
	}

	var prettyJSON bytes.Buffer
	encoder := json.NewEncoder(&prettyJSON)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(raw)
	if err != nil {
		return []string{input}
	}

	pretty := strings.TrimRight(prettyJSON.String(), "\n")

	// colorize after indentation but before splitting escaped sequences,
	// so ANSI codes wrap complete JSON tokens and don't span split lines
	if colorize != nil {
		pretty = colorize(pretty)
	}

	lines := strings.Split(pretty, "\n")

	var result []string
	for i := range lines {
		if strings.Contains(lines[i], "\\n") || strings.Contains(lines[i], "\\t") {
			lines[i] = strings.ReplaceAll(lines[i], "\\t", "    ")
			parts := strings.Split(lines[i], "\\n")
			// propagate active ANSI style across \\n splits so each
			// fragment line has balanced open/reset sequences
			var activeStyle string
			for j := range parts {
				if j > 0 && activeStyle != "" {
					parts[j] = activeStyle + parts[j]
				}
				activeStyle = lastActiveANSI(parts[j])
				if j < len(parts)-1 && activeStyle != "" {
					parts[j] += "\x1b[m"
				}
			}
			result = append(result, parts...)
		} else {
			result = append(result, lines[i])
		}
	}
	return result
}

// JSONColorStyles holds the lipgloss styles for each JSON element type.
type JSONColorStyles struct {
	Key, String, Number, Bool, Null lipgloss.Style
}

// ColorizeJSON applies ANSI color codes to JSON tokens in the input string.
// Returns the input unchanged if it's not valid JSON.
func ColorizeJSON(input string, colors JSONColorStyles) string {
	if len(input) == 0 {
		return input
	}

	// skip leading whitespace to find first non-space char
	firstNonSpace := 0
	for firstNonSpace < len(input) && (input[firstNonSpace] == ' ' || input[firstNonSpace] == '\t' || input[firstNonSpace] == '\n' || input[firstNonSpace] == '\r') {
		firstNonSpace++
	}
	if firstNonSpace >= len(input) {
		return input
	}
	if input[firstNonSpace] != '{' && input[firstNonSpace] != '[' {
		return input
	}

	// validate that the input is actually valid JSON before colorizing,
	// otherwise text like "[INF] some log message" would be partially colorized
	if !json.Valid([]byte(input)) {
		return input
	}

	var buf strings.Builder
	buf.Grow(len(input) * 2) // rough estimate with ANSI codes

	// write leading whitespace
	if firstNonSpace > 0 {
		buf.WriteString(input[:firstNonSpace])
	}

	i := firstNonSpace
	// stack tracks context: true = inside object (strings can be keys), false = inside array
	var objectStack []bool
	expectKey := false

	for i < len(input) {
		ch := input[i]
		switch {
		case ch == '{':
			objectStack = append(objectStack, true)
			expectKey = true
			buf.WriteByte(ch)
			i++

		case ch == '}':
			if len(objectStack) > 0 {
				objectStack = objectStack[:len(objectStack)-1]
			}
			// after closing }, restore expectKey based on enclosing context
			if len(objectStack) > 0 && objectStack[len(objectStack)-1] {
				expectKey = false // will be set true on next comma
			}
			buf.WriteByte(ch)
			i++

		case ch == '[':
			objectStack = append(objectStack, false)
			buf.WriteByte(ch)
			i++

		case ch == ']':
			if len(objectStack) > 0 {
				objectStack = objectStack[:len(objectStack)-1]
			}
			if len(objectStack) > 0 && objectStack[len(objectStack)-1] {
				expectKey = false
			}
			buf.WriteByte(ch)
			i++

		case ch == '"':
			end := findStringEnd(input, i)
			token := input[i:end]
			inObject := len(objectStack) > 0 && objectStack[len(objectStack)-1]
			if inObject && expectKey {
				buf.WriteString(colors.Key.Render(token))
				expectKey = false
			} else {
				buf.WriteString(colors.String.Render(token))
			}
			i = end

		case ch == ':':
			buf.WriteByte(ch)
			i++

		case ch == ',':
			buf.WriteByte(ch)
			inObject := len(objectStack) > 0 && objectStack[len(objectStack)-1]
			if inObject {
				expectKey = true
			}
			i++

		case ch == 't', ch == 'f':
			end := findKeywordEnd(input, i)
			buf.WriteString(colors.Bool.Render(input[i:end]))
			i = end

		case ch == 'n':
			end := findKeywordEnd(input, i)
			buf.WriteString(colors.Null.Render(input[i:end]))
			i = end

		case ch == '-' || (ch >= '0' && ch <= '9'):
			end := findNumberEnd(input, i)
			buf.WriteString(colors.Number.Render(input[i:end]))
			i = end

		default:
			// whitespace and other characters passed through
			buf.WriteByte(ch)
			i++
		}
	}

	return buf.String()
}

// findStringEnd returns the index just past the closing quote of a JSON string
// starting at position start (which must point to the opening '"').
func findStringEnd(s string, start int) int {
	i := start + 1 // skip opening quote
	for i < len(s) {
		if s[i] == '\\' {
			i += 2 // skip escaped character
			continue
		}
		if s[i] == '"' {
			return i + 1 // past closing quote
		}
		i++
	}
	return len(s) // unterminated string
}

// findKeywordEnd returns the index past the end of an alphabetic keyword
// (true, false, null) starting at position start.
func findKeywordEnd(s string, start int) int {
	i := start
	for i < len(s) && s[i] >= 'a' && s[i] <= 'z' {
		i++
	}
	return i
}

// findNumberEnd returns the index past the end of a JSON number starting at
// position start.
func findNumberEnd(s string, start int) int {
	i := start
	for i < len(s) {
		ch := s[i]
		if (ch >= '0' && ch <= '9') || ch == '.' || ch == '-' || ch == '+' || ch == 'e' || ch == 'E' {
			i++
		} else {
			break
		}
	}
	return i
}
