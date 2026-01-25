package view

import (
	"testing"
	"time"

	"github.com/robinovitch61/kl/internal/domain"
)

func TestNewLogRow(t *testing.T) {
	log := domain.Log{
		Timestamp:   time.Now(),
		ContainerID: domain.ContainerID{Container: "test"},
		Content:     "test log content",
	}

	row := NewLogRow(log, TimestampShort, NameShort, "#ff0000")

	if row.Log().Content != "test log content" {
		t.Errorf("expected content 'test log content', got '%s'", row.Log().Content)
	}
}

func TestLogRow_GetItem(t *testing.T) {
	log := domain.Log{
		Timestamp:   time.Now(),
		ContainerID: domain.ContainerID{Container: "test"},
		Content:     "hello world",
	}

	row := NewLogRow(log, TimestampNone, NameNone, "")

	item := row.GetItem()
	if item.ContentNoAnsi() != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", item.ContentNoAnsi())
	}
}

func TestLogRow_Log(t *testing.T) {
	log := domain.Log{
		Timestamp:    time.Now(),
		ContainerID:  domain.ContainerID{Container: "c1", Pod: "p1"},
		Content:      "test",
		IsTerminated: true,
	}

	row := NewLogRow(log, TimestampFull, NameFull, "#00ff00")

	retrieved := row.Log()
	if retrieved.Content != "test" {
		t.Errorf("expected 'test', got '%s'", retrieved.Content)
	}
	if !retrieved.IsTerminated {
		t.Error("expected IsTerminated=true")
	}
	if retrieved.ContainerID.Container != "c1" {
		t.Errorf("expected container 'c1', got '%s'", retrieved.ContainerID.Container)
	}
}
