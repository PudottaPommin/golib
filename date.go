package golib

import (
	"time"
)

func TimeIsSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func TimeIsWithinRange(t, t1, t2 time.Time) bool {
	if t1.Equal(t2) {
		return t.Equal(t1)
	}
	return t.Equal(t1) || t.Equal(t2) || (t.After(t1) && t.Before(t2))
}
