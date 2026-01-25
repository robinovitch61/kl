package view

import (
	"testing"
	"time"

	"github.com/robinovitch61/kl/internal/domain"
)

func TestNewLogsView(t *testing.T) {
	lv := NewLogsView(80, 24, true)

	if lv.width != 80 {
		t.Errorf("expected width 80, got %d", lv.width)
	}
	if lv.height != 24 {
		t.Errorf("expected height 24, got %d", lv.height)
	}
	if !lv.ascending {
		t.Error("expected ascending=true")
	}
}

func TestLogsView_AppendLogs(t *testing.T) {
	lv := NewLogsView(80, 24, true)

	logs := []domain.Log{
		{
			Timestamp:   time.Now(),
			ContainerID: domain.ContainerID{Container: "c1"},
			Content:     "log line 1",
		},
		{
			Timestamp:   time.Now().Add(time.Second),
			ContainerID: domain.ContainerID{Container: "c1"},
			Content:     "log line 2",
		},
	}

	lv = lv.AppendLogs(logs)

	if len(lv.logs) != 2 {
		t.Errorf("expected 2 logs, got %d", len(lv.logs))
	}
}

func TestLogsView_ClearLogsForContainer(t *testing.T) {
	lv := NewLogsView(80, 24, true)

	c1 := domain.ContainerID{Container: "c1"}
	c2 := domain.ContainerID{Container: "c2"}

	logs := []domain.Log{
		{ContainerID: c1, Content: "c1 log"},
		{ContainerID: c2, Content: "c2 log"},
		{ContainerID: c1, Content: "c1 another"},
	}

	lv = lv.AppendLogs(logs)
	lv = lv.ClearLogsForContainer(c1)

	if len(lv.logs) != 1 {
		t.Errorf("expected 1 log remaining, got %d", len(lv.logs))
	}
	if lv.logs[0].ContainerID != c2 {
		t.Error("expected remaining log to be from c2")
	}
}

func TestLogsView_SetAscending(t *testing.T) {
	lv := NewLogsView(80, 24, true)

	lv = lv.SetAscending(false)
	if lv.ascending {
		t.Error("expected ascending=false")
	}

	lv = lv.SetAscending(true)
	if !lv.ascending {
		t.Error("expected ascending=true")
	}
}

func TestLogsView_ToggleTimestampFormat(t *testing.T) {
	lv := NewLogsView(80, 24, true)

	// Initial format is TimestampShort
	if lv.timestampFormat != TimestampShort {
		t.Errorf("expected TimestampShort, got %v", lv.timestampFormat)
	}

	lv = lv.ToggleTimestampFormat()
	if lv.timestampFormat != TimestampFull {
		t.Errorf("expected TimestampFull, got %v", lv.timestampFormat)
	}

	lv = lv.ToggleTimestampFormat()
	if lv.timestampFormat != TimestampNone {
		t.Errorf("expected TimestampNone, got %v", lv.timestampFormat)
	}

	lv = lv.ToggleTimestampFormat()
	if lv.timestampFormat != TimestampShort {
		t.Errorf("expected TimestampShort (wrap around), got %v", lv.timestampFormat)
	}
}

func TestLogsView_ToggleNameFormat(t *testing.T) {
	lv := NewLogsView(80, 24, true)

	// Initial format is NameShort
	if lv.nameFormat != NameShort {
		t.Errorf("expected NameShort, got %v", lv.nameFormat)
	}

	lv = lv.ToggleNameFormat()
	if lv.nameFormat != NameNone {
		t.Errorf("expected NameNone, got %v", lv.nameFormat)
	}

	lv = lv.ToggleNameFormat()
	if lv.nameFormat != NameFull {
		t.Errorf("expected NameFull, got %v", lv.nameFormat)
	}

	lv = lv.ToggleNameFormat()
	if lv.nameFormat != NameShort {
		t.Errorf("expected NameShort (wrap around), got %v", lv.nameFormat)
	}
}

func TestLogsView_SetSize(t *testing.T) {
	lv := NewLogsView(80, 24, true)

	lv = lv.SetSize(120, 40)

	if lv.width != 120 {
		t.Errorf("expected width 120, got %d", lv.width)
	}
	if lv.height != 40 {
		t.Errorf("expected height 40, got %d", lv.height)
	}
}
