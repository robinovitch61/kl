package util

import (
	"fmt"
	"runtime"
	"sort"
	"testing"
	"time"
)

func TimeSince(t time.Time) string {
	return FormatDuration(time.Since(t))
}

func FormatDuration(duration time.Duration) string {
	seconds := int(duration.Seconds())
	minutes := seconds / 60
	hours := minutes / 60
	days := hours / 24
	months := days / 30
	years := months / 12

	if years > 0 {
		if years < 10 {
			return fmt.Sprintf("%dy%dmo", years, months%12)
		}
		return fmt.Sprintf("%dy", years)
	} else if months > 0 {
		if months < 10 {
			return fmt.Sprintf("%dmo%dd", months, days%30)
		}
		return fmt.Sprintf("%dmo", months)
	} else if days > 0 {
		if days < 10 {
			return fmt.Sprintf("%dd%dh", days, hours%24)
		}
		return fmt.Sprintf("%dd", days)
	} else if hours > 0 {
		if hours < 10 {
			return fmt.Sprintf("%dh%dm", hours, minutes%60)
		}
		return fmt.Sprintf("%dh", hours)
	} else if minutes > 0 {
		if minutes < 10 {
			return fmt.Sprintf("%dm%ds", minutes, seconds%60)
		}
		return fmt.Sprintf("%dm", minutes)
	} else {
		return fmt.Sprintf("%ds", seconds)
	}
}

// DurationTilNext calculates the duration until the between occurrence of a periodic event.
//
// Parameters:
// - start: A start time for the periodic event
// - now: The current time
// - between: The duration between occurrences of the event
//
// Returns: The duration until the between occurrence of the event
//
// Examples:
//
//  1. For a minutely event:
//     start = 2023-01-01 00:00:00
//     now       = 2023-01-01 00:01:30
//     between      = 1 minute
//     Result: 30 seconds (until 2023-01-01 00:02:00)
//
//  2. For an hourly event:
//     start = 2023-01-01 00:00:00
//     now       = 2023-01-01 02:45:00
//     between      = 1 hour
//     Result: 15 minutes (until 2023-01-01 03:00:00)
//
//  3. For a custom duration event (every 45 minutes):
//     start = 2023-01-01 00:00:00
//     now       = 2023-01-01 01:20:00
//     between      = 45 minutes
//     Result: 25 minutes (until 2023-01-01 01:45:00)
func DurationTilNext(start time.Time, now time.Time, between time.Duration) time.Duration {
	elapsed := now.Sub(start)
	intervals := elapsed / between
	nextUpdate := start.Add(between * (intervals + 1))
	return nextUpdate.Sub(now)
}

func RunWithTimeout(t *testing.T, runTest func(t *testing.T), timeout time.Duration) {
	t.Helper()

	// warmup runs
	for i := 0; i < 3; i++ {
		runTest(t)
	}

	// actual measured runs
	var durations []time.Duration
	for i := 0; i < 3; i++ {
		done := make(chan struct{})
		var testErr error
		start := time.Now()

		go func() {
			defer func() {
				if r := recover(); r != nil {
					testErr = fmt.Errorf("test panicked: %v", r)
				}
				close(done)
			}()

			subT := &testing.T{}
			runTest(subT)
			if subT.Failed() {
				testErr = fmt.Errorf("test failed in goroutine")
			}
		}()

		select {
		case <-done:
			if testErr != nil {
				t.Fatal(testErr)
			}
			durations = append(durations, time.Since(start))
		case <-time.After(timeout):
			t.Fatalf("Test took too long: %v", timeout)
		}

		runtime.GC()
		time.Sleep(time.Millisecond * 10)
	}

	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})
	median := durations[len(durations)/2]
	t.Logf("Test timing: median=%v min=%v max=%v",
		median, durations[0], durations[len(durations)-1])
}
