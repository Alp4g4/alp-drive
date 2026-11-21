package system

import (
	"time"

	"alp-drive/internal/application/ports/clock"
)

var _ clock.Clock = (*Clock)(nil)

type Clock struct{}

func NewClock() *Clock {
	return &Clock{}
}

func (c *Clock) Now() time.Time {
	return time.Now().UTC()
}
