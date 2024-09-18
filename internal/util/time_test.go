package util

import (
	"testing"
	"time"
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
			result := DurationTilNext(tt.start, tt.now, tt.between)
			if result != tt.expected {
				t.Errorf("TimeToNextUpdate(%v, %v, %v) = %v, want %v",
					tt.start, tt.now, tt.between, result, tt.expected)
			}
		})
	}
}
