package domain

import "time"

// TimeRange represents lookback time selection (keys 0-9)
type TimeRange struct {
	Key      int           // 0-9 key press
	Duration time.Duration // 0 = now onwards, -1 = all time
}

// SinceTime returns the time to use for log lookback
func (tr TimeRange) SinceTime() time.Time {
	if tr.Duration < 0 {
		// All time - return zero time
		return time.Time{}
	}
	if tr.Duration == 0 {
		// Now onwards
		return time.Now()
	}
	return time.Now().Add(-tr.Duration)
}

// NewTimeRange creates a TimeRange from a key press
func NewTimeRange(key int) TimeRange {
	durations := map[int]time.Duration{
		0: 0,                      // now onwards
		1: 1 * time.Minute,        // 1m
		2: 5 * time.Minute,        // 5m
		3: 15 * time.Minute,       // 15m
		4: 30 * time.Minute,       // 30m
		5: 1 * time.Hour,          // 1h
		6: 3 * time.Hour,          // 3h
		7: 12 * time.Hour,         // 12h
		8: 24 * time.Hour,         // 24h
		9: -1 * time.Nanosecond,   // all time (negative)
	}
	d, ok := durations[key]
	if !ok {
		d = 1 * time.Minute // default
	}
	return TimeRange{Key: key, Duration: d}
}
