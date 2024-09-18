package model

import (
	"github.com/google/uuid"
	"github.com/robinovitch61/kl/internal/util"
	"time"
)

type SinceTime struct {
	UUID         string
	Time         time.Time
	LookbackMins int
}

func NewSinceTime(t time.Time, lookbackMins int) SinceTime {
	return SinceTime{
		UUID:         uuid.New().String(),
		Time:         t,
		LookbackMins: lookbackMins,
	}
}

func (st SinceTime) TimeToNextUpdate() time.Duration {
	now := time.Now()
	diff := now.Sub(st.Time)
	var between time.Duration
	if diff < 10*time.Minute {
		between = time.Second
	} else {
		between = time.Minute
	}
	return util.DurationTilNext(st.Time, now, between)
}
