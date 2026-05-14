package clock

import "time"

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func NewRealClock() RealClock {
	return RealClock{}
}

func (RealClock) Now() time.Time {
	return time.Now().UTC()
}
