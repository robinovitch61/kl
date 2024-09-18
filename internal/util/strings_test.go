package util

import (
	"testing"
)

func TestGetUniqueShortNames(t *testing.T) {
	tests := []struct {
		nameSet   map[string]bool
		fromRight bool
		minChars  int
		expected  map[string]string
	}{
		{
			nameSet: map[string]bool{
				"apple":  true,
				"banana": true,
				"cherry": true,
			},
			fromRight: false,
			minChars:  2,
			expected: map[string]string{
				"apple":  "ap..",
				"banana": "ba..",
				"cherry": "ch..",
			},
		},
		{
			nameSet: map[string]bool{
				"apple":   true,
				"apricot": true,
				"banana":  true,
			},
			fromRight: false,
			minChars:  1,
			expected: map[string]string{
				"apple":   "app..",
				"apricot": "apr..",
				"banana":  "ban..",
			},
		},
		{
			nameSet: map[string]bool{
				"apple":  true,
				"papple": true,
				"grape":  true,
			},
			fromRight: true,
			minChars:  3,
			expected: map[string]string{
				"apple":  "apple",
				"papple": "papple",
				"grape":  "grape",
			},
		},
	}

	for _, test := range tests {
		result := GetUniqueShortNames(test.nameSet, test.fromRight, test.minChars)
		for k, v := range test.expected {
			if result[k] != v {
				t.Errorf("For name '%s', expected short name '%s' but got '%s'", k, v, result[k])
			}
		}
	}
}

// same test for GetUniqueShortNamesFromEdges
func TestGetUniqueShortNamesFromSides(t *testing.T) {
	tests := []struct {
		nameSet          map[string]bool
		numCharsEachSide int
		expected         map[string]string
	}{
		{
			nameSet: map[string]bool{
				"apple":  true,
				"banana": true,
				"cherry": true,
			},
			numCharsEachSide: 1,
			expected: map[string]string{
				"apple":  "a..e",
				"banana": "b..a",
				"cherry": "c..y",
			},
		},
		{
			nameSet: map[string]bool{
				"apple":  true,
				"banana": true,
				"cherry": true,
			},
			numCharsEachSide: 2,
			expected: map[string]string{
				"apple":  "ap..le",
				"banana": "ba..na",
				"cherry": "ch..ry",
			},
		},
		{
			nameSet: map[string]bool{
				"apple":  true,
				"banana": true,
				"cherry": true,
			},
			numCharsEachSide: 3,
			expected: map[string]string{
				"apple":  "apple",
				"banana": "banana",
				"cherry": "cherry",
			},
		},
		{
			nameSet: map[string]bool{
				"appsamele": true,
				"appdiffle": true,
			},
			numCharsEachSide: 1,
			expected: map[string]string{
				"appsamele": "app..ele",
				"appdiffle": "app..fle",
			},
		},
	}

	for _, test := range tests {
		result := GetUniqueShortNamesFromEdges(test.nameSet, test.numCharsEachSide)
		for k, v := range test.expected {
			if result[k] != v {
				t.Errorf("For name '%s', expected short name '%s' but got '%s'", k, v, result[k])
			}
		}
	}
}
