package view

import (
	"github.com/robinovitch61/bubbleo/viewport/item"
	"github.com/robinovitch61/kl/internal/domain"
)

// TimestampFormat controls timestamp display
type TimestampFormat int

const (
	TimestampNone TimestampFormat = iota
	TimestampShort
	TimestampFull
)

// NameFormat controls container name display
type NameFormat int

const (
	NameShort NameFormat = iota
	NameNone
	NameFull
)

// LogRow wraps a Log for viewport display
type LogRow struct {
	log             domain.Log
	timestampFormat TimestampFormat
	nameFormat      NameFormat
	containerColor  string
}

// NewLogRow creates a LogRow with display settings
func NewLogRow(log domain.Log, tf TimestampFormat, nf NameFormat, color string) LogRow {
	return LogRow{
		log:             log,
		timestampFormat: tf,
		nameFormat:      nf,
		containerColor:  color,
	}
}

// GetItem implements viewport.Object
func (r LogRow) GetItem() item.Item {
	// TODO: implement proper formatting based on timestampFormat and nameFormat
	return item.NewItem(r.log.Content)
}

// Log returns the underlying log
func (r LogRow) Log() domain.Log {
	return r.log
}
