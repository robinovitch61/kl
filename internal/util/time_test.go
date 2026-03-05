package util_test

import (
	"testing"
	"time"

	"github.com/robinovitch61/kl/internal/util"
)

func TestDurationTilNext(t *testing.T) {
	parseTime := func(s string) time.Time {
		t, err := time.Parse("2006-01-02 15:04:05", s)
		if err != nil {
			panic(err)
		}
		return t
	}

	tests := []struct {
		name     string
		start    time.Time
		now      time.Time
		between  time.Duration
		expected time.Duration
	}{
		{
			name:     "Minutely event",
			start:    parseTime("2023-01-01 00:00:00"),
			now:      parseTime("2023-01-01 00:01:30"),
			between:  time.Minute,
			expected: 30 * time.Second,
		},
		{
			name:     "Hourly event",
			start:    parseTime("2023-01-01 00:00:00"),
			now:      parseTime("2023-01-01 02:45:00"),
			between:  time.Hour,
			expected: 15 * time.Minute,
		},
		{
			name:     "Custom duration event (45 minutes)",
			start:    parseTime("2023-01-01 00:00:00"),
			now:      parseTime("2023-01-01 01:20:00"),
			between:  45 * time.Minute,
			expected: 10 * time.Minute,
		},
		{
			name:     "Event just passed",
			start:    parseTime("2023-01-01 00:00:00"),
			now:      parseTime("2023-01-01 00:00:01"),
			between:  time.Minute,
			expected: 59 * time.Second,
		},
		{
			name:     "Event about to occur",
			start:    parseTime("2023-01-01 00:00:00"),
			now:      parseTime("2023-01-01 00:59:59"),
			between:  time.Hour,
			expected: 1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.DurationTilNext(tt.start, tt.now, tt.between)
			if result != tt.expected {
				t.Errorf("TimeToNextUpdate(%v, %v, %v) = %v, want %v",
					tt.start, tt.now, tt.between, result, tt.expected)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"zero", 0, "0s"},
		{"one second", time.Second, "1s"},
		{"seconds", 45 * time.Second, "45s"},
		{"one minute with seconds", 90 * time.Second, "1m30s"},
		{"minutes under 10", 5*time.Minute + 15*time.Second, "5m15s"},
		{"minutes over 10", 15 * time.Minute, "15m"},
		{"one hour with minutes", time.Hour + 30*time.Minute, "1h30m"},
		{"hours under 10", 5*time.Hour + 45*time.Minute, "5h45m"},
		{"hours over 10", 15 * time.Hour, "15h"},
		{"one day with hours", 30 * time.Hour, "1d6h"},
		{"days under 10", 5*24*time.Hour + 12*time.Hour, "5d12h"},
		{"days over 10", 15 * 24 * time.Hour, "15d"},
		{"one month with days", 45 * 24 * time.Hour, "1mo15d"},
		{"months under 10", 9 * 30 * 24 * time.Hour, "9mo0d"},
		{"months over 10", 11 * 30 * 24 * time.Hour, "11mo"},
		{"one year with months", 400 * 24 * time.Hour, "1y1mo"},
		{"years over 10", 11 * 365 * 24 * time.Hour, "11y"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.FormatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("util.FormatDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}
