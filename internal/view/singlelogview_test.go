package view

import (
	"testing"
	"time"

	"github.com/robinovitch61/kl/internal/domain"
)

func TestNewSingleLogView(t *testing.T) {
	sv := NewSingleLogView(80, 24)

	if sv.width != 80 {
		t.Errorf("expected width 80, got %d", sv.width)
	}
	if sv.height != 24 {
		t.Errorf("expected height 24, got %d", sv.height)
	}
}

func TestSingleLogView_SetLog(t *testing.T) {
	sv := NewSingleLogView(80, 24)

	log := domain.Log{
		Timestamp:   time.Now(),
		ContainerID: domain.ContainerID{Container: "test"},
		Content:     "test log content",
	}

	sv = sv.SetLog(log)

	if sv.Log() == nil {
		t.Error("expected log to be set")
	}
	if sv.Log().Content != "test log content" {
		t.Errorf("expected content 'test log content', got %q", sv.Log().Content)
	}
}

func TestSingleLogView_SetSize(t *testing.T) {
	sv := NewSingleLogView(80, 24)

	sv = sv.SetSize(120, 40)

	if sv.width != 120 {
		t.Errorf("expected width 120, got %d", sv.width)
	}
	if sv.height != 40 {
		t.Errorf("expected height 40, got %d", sv.height)
	}
}

func TestSingleLogView_PlainText(t *testing.T) {
	sv := NewSingleLogView(80, 24)

	// No log set
	if sv.PlainText() != "" {
		t.Error("expected empty string when no log set")
	}

	// With log
	log := domain.Log{
		Content: "plain text content",
	}
	sv = sv.SetLog(log)

	if sv.PlainText() != "plain text content" {
		t.Errorf("expected 'plain text content', got %q", sv.PlainText())
	}
}

func TestSingleLogView_Log_Nil(t *testing.T) {
	sv := NewSingleLogView(80, 24)

	if sv.Log() != nil {
		t.Error("expected nil when no log set")
	}
}

func TestSingleLogLine_GetItem(t *testing.T) {
	line := NewSingleLogLine("test content")

	item := line.GetItem()
	if item.ContentNoAnsi() != "test content" {
		t.Errorf("expected 'test content', got %q", item.ContentNoAnsi())
	}
}
