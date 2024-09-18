package util

import (
	"fmt"
	"time"
)

func TimeSince(t time.Time) string {
	duration := time.Since(t)

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
