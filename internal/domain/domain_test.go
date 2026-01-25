package domain

import (
	"testing"
	"time"
)

func TestContainerID(t *testing.T) {
	id := ContainerID{
		Cluster:   "prod",
		Namespace: "default",
		Pod:       "api-pod-123",
		Container: "api",
	}

	if id.Cluster != "prod" {
		t.Errorf("expected Cluster 'prod', got %q", id.Cluster)
	}
	if id.Namespace != "default" {
		t.Errorf("expected Namespace 'default', got %q", id.Namespace)
	}
}

func TestContainerState(t *testing.T) {
	// Verify state ordering
	if StateInactive >= StateWantScanning {
		t.Error("StateInactive should be less than StateWantScanning")
	}
	if StateScanning >= StateScannerStopping {
		t.Error("StateScanning should be less than StateScannerStopping")
	}
}

func TestLogsSortAscending(t *testing.T) {
	now := time.Now()
	logs := Logs{
		{Timestamp: now.Add(2 * time.Second), Content: "third"},
		{Timestamp: now, Content: "first"},
		{Timestamp: now.Add(1 * time.Second), Content: "second"},
	}

	sorted := logs.SortAscending()

	if sorted[0].Content != "first" {
		t.Errorf("expected first log to be 'first', got %q", sorted[0].Content)
	}
	if sorted[1].Content != "second" {
		t.Errorf("expected second log to be 'second', got %q", sorted[1].Content)
	}
	if sorted[2].Content != "third" {
		t.Errorf("expected third log to be 'third', got %q", sorted[2].Content)
	}

	// Original should be unchanged
	if logs[0].Content != "third" {
		t.Error("original logs should be unchanged")
	}
}

func TestLogsSortDescending(t *testing.T) {
	now := time.Now()
	logs := Logs{
		{Timestamp: now, Content: "first"},
		{Timestamp: now.Add(2 * time.Second), Content: "third"},
		{Timestamp: now.Add(1 * time.Second), Content: "second"},
	}

	sorted := logs.SortDescending()

	if sorted[0].Content != "third" {
		t.Errorf("expected first log to be 'third', got %q", sorted[0].Content)
	}
	if sorted[2].Content != "first" {
		t.Errorf("expected last log to be 'first', got %q", sorted[2].Content)
	}
}

func TestNewTimeRange(t *testing.T) {
	tests := []struct {
		key      int
		wantDur  time.Duration
	}{
		{0, 0},
		{1, 1 * time.Minute},
		{5, 1 * time.Hour},
		{9, -1 * time.Nanosecond},
	}

	for _, tt := range tests {
		tr := NewTimeRange(tt.key)
		if tr.Key != tt.key {
			t.Errorf("NewTimeRange(%d).Key = %d, want %d", tt.key, tr.Key, tt.key)
		}
		if tr.Duration != tt.wantDur {
			t.Errorf("NewTimeRange(%d).Duration = %v, want %v", tt.key, tr.Duration, tt.wantDur)
		}
	}
}

func TestTimeRangeSinceTime(t *testing.T) {
	// All time should return zero time
	tr := NewTimeRange(9)
	since := tr.SinceTime()
	if !since.IsZero() {
		t.Errorf("all time SinceTime should be zero, got %v", since)
	}

	// Now onwards should be close to now
	tr = NewTimeRange(0)
	since = tr.SinceTime()
	if time.Since(since) > 1*time.Second {
		t.Error("now onwards SinceTime should be close to now")
	}

	// 1 minute ago should be roughly 1 minute in the past
	tr = NewTimeRange(1)
	since = tr.SinceTime()
	diff := time.Since(since)
	if diff < 59*time.Second || diff > 61*time.Second {
		t.Errorf("1m SinceTime should be ~1m ago, got %v ago", diff)
	}
}
