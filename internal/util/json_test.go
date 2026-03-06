package util

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func testColors() JSONColorStyles {
	return JSONColorStyles{
		Key:    lipgloss.NewStyle().Foreground(lipgloss.Red),
		String: lipgloss.NewStyle().Foreground(lipgloss.Green),
		Number: lipgloss.NewStyle().Foreground(lipgloss.Yellow),
		Bool:   lipgloss.NewStyle().Foreground(lipgloss.Magenta),
		Null:   lipgloss.NewStyle().Foreground(lipgloss.Cyan),
	}
}

func TestColorizeJSON_NonJSON(t *testing.T) {
	colors := testColors()
	tests := []struct {
		name  string
		input string
	}{
		{"plain text", "just plain text"},
		{"empty string", ""},
		{"number", "42"},
		{"starts with letter", "abc{def}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ColorizeJSON(tt.input, colors)
			if got != tt.input {
				t.Errorf("expected input unchanged, got %q", got)
			}
		})
	}
}

func TestColorizeJSON_EmptyStructures(t *testing.T) {
	// with no-color styles, output should match input
	noColor := JSONColorStyles{
		Key:    lipgloss.NewStyle(),
		String: lipgloss.NewStyle(),
		Number: lipgloss.NewStyle(),
		Bool:   lipgloss.NewStyle(),
		Null:   lipgloss.NewStyle(),
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty object", `{}`, `{}`},
		{"empty array", `[]`, `[]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ColorizeJSON(tt.input, noColor)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestColorizeJSON_PreservesStructure(t *testing.T) {
	// with no-color styles, output should exactly match input
	noColor := JSONColorStyles{
		Key:    lipgloss.NewStyle(),
		String: lipgloss.NewStyle(),
		Number: lipgloss.NewStyle(),
		Bool:   lipgloss.NewStyle(),
		Null:   lipgloss.NewStyle(),
	}

	tests := []struct {
		name  string
		input string
	}{
		{"simple object", `{"key":"value"}`},
		{"nested object", `{"a":{"b":"c"}}`},
		{"object with number", `{"count":42}`},
		{"object with bool", `{"ok":true}`},
		{"object with null", `{"val":null}`},
		{"object with false", `{"ok":false}`},
		{"object with negative number", `{"n":-3.14}`},
		{"object with array", `{"items":[1,2,3]}`},
		{"array of objects", `[{"a":1},{"b":2}]`},
		{"object with escaped string", `{"msg":"hello \"world\""}`},
		{"pretty printed", "{\n    \"key\": \"value\"\n}"},
		{"complex", `{"level":"info","msg":"request","status":200,"ok":true,"data":null}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ColorizeJSON(tt.input, noColor)
			if got != tt.input {
				t.Errorf("expected input preserved:\n  expected: %q\n  got:      %q", tt.input, got)
			}
		})
	}
}

func TestColorizeJSON_AppliesColors(t *testing.T) {
	colors := testColors()

	input := `{"key":"value"}`
	got := ColorizeJSON(input, colors)

	// verify it changed (colors were applied)
	if got == input {
		t.Error("expected colorized output to differ from input")
	}

	// verify structural characters are preserved (not colored)
	// The output should contain the key and value with ANSI codes around them
	if len(got) <= len(input) {
		t.Error("expected colorized output to be longer than input due to ANSI codes")
	}
}

func TestColorizeJSON_DistinguishesKeysFromValues(t *testing.T) {
	// use distinct styles for key vs string
	colors := JSONColorStyles{
		Key:    lipgloss.NewStyle().Bold(true),
		String: lipgloss.NewStyle().Italic(true),
		Number: lipgloss.NewStyle(),
		Bool:   lipgloss.NewStyle(),
		Null:   lipgloss.NewStyle(),
	}

	input := `{"name":"alice"}`
	got := ColorizeJSON(input, colors)

	keyStyled := colors.Key.Render(`"name"`)
	valStyled := colors.String.Render(`"alice"`)

	if got != "{"+keyStyled+":"+valStyled+"}" {
		t.Errorf("unexpected output: %q", got)
	}
}

func TestColorizeJSON_ArrayValues(t *testing.T) {
	// strings in arrays should be string-colored, not key-colored
	colors := JSONColorStyles{
		Key:    lipgloss.NewStyle().Bold(true),
		String: lipgloss.NewStyle().Italic(true),
		Number: lipgloss.NewStyle().Underline(true),
		Bool:   lipgloss.NewStyle(),
		Null:   lipgloss.NewStyle(),
	}

	input := `["hello",42]`
	got := ColorizeJSON(input, colors)

	strStyled := colors.String.Render(`"hello"`)
	numStyled := colors.Number.Render(`42`)

	expected := "[" + strStyled + "," + numStyled + "]"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestColorizeJSON_AllValueTypes(t *testing.T) {
	colors := JSONColorStyles{
		Key:    lipgloss.NewStyle().Bold(true),
		String: lipgloss.NewStyle().Foreground(lipgloss.Green),
		Number: lipgloss.NewStyle().Foreground(lipgloss.Yellow),
		Bool:   lipgloss.NewStyle().Foreground(lipgloss.Magenta),
		Null:   lipgloss.NewStyle().Foreground(lipgloss.Cyan),
	}

	input := `{"s":"val","n":1,"b":true,"f":false,"x":null}`
	got := ColorizeJSON(input, colors)

	// verify all value types got colored (output is longer than input)
	if len(got) <= len(input) {
		t.Error("expected ANSI codes to make output longer")
	}

	// with no-color, should preserve
	noColor := JSONColorStyles{
		Key: lipgloss.NewStyle(), String: lipgloss.NewStyle(),
		Number: lipgloss.NewStyle(), Bool: lipgloss.NewStyle(), Null: lipgloss.NewStyle(),
	}
	preserved := ColorizeJSON(input, noColor)
	if preserved != input {
		t.Errorf("no-color should preserve input:\n  expected: %q\n  got:      %q", input, preserved)
	}
}

func TestColorizeJSON_EscapedStrings(t *testing.T) {
	noColor := JSONColorStyles{
		Key: lipgloss.NewStyle(), String: lipgloss.NewStyle(),
		Number: lipgloss.NewStyle(), Bool: lipgloss.NewStyle(), Null: lipgloss.NewStyle(),
	}

	tests := []struct {
		name  string
		input string
	}{
		{"escaped quote", `{"k":"val\"ue"}`},
		{"escaped backslash", `{"k":"val\\ue"}`},
		{"escaped newline", `{"k":"line1\nline2"}`},
		{"escaped tab", `{"k":"col1\tcol2"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ColorizeJSON(tt.input, noColor)
			if got != tt.input {
				t.Errorf("expected preserved:\n  expected: %q\n  got:      %q", tt.input, got)
			}
		})
	}
}

func TestColorizeJSON_LeadingWhitespace(t *testing.T) {
	noColor := JSONColorStyles{
		Key: lipgloss.NewStyle(), String: lipgloss.NewStyle(),
		Number: lipgloss.NewStyle(), Bool: lipgloss.NewStyle(), Null: lipgloss.NewStyle(),
	}

	input := "  {\"k\":1}"
	got := ColorizeJSON(input, noColor)
	if got != input {
		t.Errorf("expected %q, got %q", input, got)
	}
}

func TestColorizeJSON_NestedObjectKeyResumption(t *testing.T) {
	// after a nested object closes, the next string in the parent object should be a key
	colors := JSONColorStyles{
		Key:    lipgloss.NewStyle().Bold(true),
		String: lipgloss.NewStyle().Italic(true),
		Number: lipgloss.NewStyle(),
		Bool:   lipgloss.NewStyle(),
		Null:   lipgloss.NewStyle(),
	}

	input := `{"a":{"b":"c"},"d":"e"}`
	got := ColorizeJSON(input, colors)

	keyA := colors.Key.Render(`"a"`)
	keyB := colors.Key.Render(`"b"`)
	valC := colors.String.Render(`"c"`)
	keyD := colors.Key.Render(`"d"`)
	valE := colors.String.Render(`"e"`)

	expected := "{" + keyA + ":{" + keyB + ":" + valC + "}," + keyD + ":" + valE + "}"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestColorizeJSON_ArrayInObject(t *testing.T) {
	// strings inside an array value should be string-colored, not key-colored;
	// the key after the array should be key-colored
	colors := JSONColorStyles{
		Key:    lipgloss.NewStyle().Bold(true),
		String: lipgloss.NewStyle().Italic(true),
		Number: lipgloss.NewStyle(),
		Bool:   lipgloss.NewStyle(),
		Null:   lipgloss.NewStyle(),
	}

	input := `{"items":["a","b"],"next":"c"}`
	got := ColorizeJSON(input, colors)

	keyItems := colors.Key.Render(`"items"`)
	valA := colors.String.Render(`"a"`)
	valB := colors.String.Render(`"b"`)
	keyNext := colors.Key.Render(`"next"`)
	valC := colors.String.Render(`"c"`)

	expected := "{" + keyItems + ":[" + valA + "," + valB + "]," + keyNext + ":" + valC + "}"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestColorizeJSON_ScientificNotation(t *testing.T) {
	noColor := JSONColorStyles{
		Key: lipgloss.NewStyle(), String: lipgloss.NewStyle(),
		Number: lipgloss.NewStyle(), Bool: lipgloss.NewStyle(), Null: lipgloss.NewStyle(),
	}

	input := `{"n":1.5e10}`
	got := ColorizeJSON(input, noColor)
	if got != input {
		t.Errorf("expected %q, got %q", input, got)
	}
}

func TestColorizeJSON_MultilinePreservesStructure(t *testing.T) {
	noColor := JSONColorStyles{
		Key: lipgloss.NewStyle(), String: lipgloss.NewStyle(),
		Number: lipgloss.NewStyle(), Bool: lipgloss.NewStyle(), Null: lipgloss.NewStyle(),
	}

	input := "{\n    \"a\": 1,\n    \"b\": true,\n    \"c\": null\n}"
	got := ColorizeJSON(input, noColor)
	if got != input {
		t.Errorf("expected input preserved:\n  expected: %q\n  got:      %q", input, got)
	}
}

func TestPrettyPrintJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "not JSON",
			input:    "just plain text",
			expected: []string{"just plain text"},
		},
		{
			name:     "invalid JSON",
			input:    `{"key": broken}`,
			expected: []string{`{"key": broken}`},
		},
		{
			name:     "JSON array not pretty-printed",
			input:    `[1, 2, 3]`,
			expected: []string{`[1, 2, 3]`},
		},
		{
			name:  "simple object",
			input: `{"key":"value"}`,
			expected: []string{
				`{`,
				`    "key": "value"`,
				`}`,
			},
		},
		{
			name:  "nested object",
			input: `{"a":{"b":"c"}}`,
			expected: []string{
				`{`,
				`    "a": {`,
				`        "b": "c"`,
				`    }`,
				`}`,
			},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{""},
		},
		{
			name:     "empty object",
			input:    `{}`,
			expected: []string{`{}`},
		},
		{
			name:  "escaped newlines split into lines",
			input: `{"msg":"line1\nline2\nline3"}`,
			expected: []string{
				`{`,
				`    "msg": "line1`,
				`line2`,
				`line3"`,
				`}`,
			},
		},
		{
			name:  "escaped tabs replaced with spaces",
			input: `{"msg":"before\tafter"}`,
			expected: []string{
				`{`,
				`    "msg": "before    after"`,
				`}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PrettyPrintJSON(tt.input, nil)
			if len(got) != len(tt.expected) {
				t.Fatalf("expected %d lines, got %d\nexpected: %q\ngot:      %q", len(tt.expected), len(got), tt.expected, got)
			}
			for i := range tt.expected {
				if got[i] != tt.expected[i] {
					t.Errorf("line %d: expected %q, got %q", i, tt.expected[i], got[i])
				}
			}
		})
	}
}

func TestPrettyPrintJSON_WithColorize(t *testing.T) {
	colorize := func(s string) string {
		return "COLORED:" + s + ":END"
	}

	t.Run("colorize is applied to valid JSON", func(t *testing.T) {
		got := PrettyPrintJSON(`{"key":"value"}`, colorize)
		joined := strings.Join(got, "\n")
		if !strings.Contains(joined, "COLORED:") {
			t.Error("expected colorize to be applied")
		}
	})

	t.Run("colorize is not applied to non-JSON", func(t *testing.T) {
		got := PrettyPrintJSON("just plain text", colorize)
		if len(got) != 1 || got[0] != "just plain text" {
			t.Errorf("expected unchanged plain text, got %q", got)
		}
	})

	t.Run("nil colorize is safe", func(t *testing.T) {
		got := PrettyPrintJSON(`{"a":1}`, nil)
		if len(got) != 3 {
			t.Fatalf("expected 3 lines, got %d", len(got))
		}
	})
}

func TestPrettyPrintJSON_ColorizeWithEscapedNewlines(t *testing.T) {
	// When a JSON string value contains \n, PrettyPrintJSON splits it into
	// multiple display lines. The propagation logic should re-open the active
	// ANSI style on continuation lines and close it on non-final lines.
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Cyan)
	strStyle := lipgloss.NewStyle().Foreground(lipgloss.Green)
	colors := JSONColorStyles{
		Key: keyStyle, String: strStyle,
		Number: lipgloss.NewStyle(), Bool: lipgloss.NewStyle(), Null: lipgloss.NewStyle(),
	}
	colorize := func(s string) string {
		return ColorizeJSON(s, colors)
	}

	keyRendered := keyStyle.Render(`"msg"`)
	strOpen := "\x1b[32m" // green foreground SGR from strStyle
	reset := "\x1b[m"

	expected := []string{
		`{`,
		"    " + keyRendered + ": " + strOpen + `"line1` + reset,
		strOpen + "line2" + reset,
		strOpen + `line3"` + reset,
		`}`,
	}

	lines := PrettyPrintJSON(`{"msg":"line1\nline2\nline3"}`, colorize)

	if len(lines) != len(expected) {
		t.Fatalf("expected %d lines, got %d: %q", len(expected), len(lines), lines)
	}
	for i := range expected {
		if lines[i] != expected[i] {
			t.Errorf("line %d:\n  expected: %q\n  got:      %q", i, expected[i], lines[i])
		}
	}
}
