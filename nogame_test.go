package main

import (
	"fmt"
	"testing"
	"time"
)

func assertEqual(t *testing.T, a, b interface{}) {
	x := fmt.Sprintf("%v", a)
	y := fmt.Sprintf("%v", b)
	if x != y {
		t.Errorf("%s does not equal %s", a, b)
	}
}

func clock(s string) time.Time {
	lt := "2006/01/02 15:04:05"
	c, _ := time.Parse(lt, s)
	return c
}

func TestBlockClock(t *testing.T) {
	bl := Create()

	now := clock("2014/01/01 12:00:00")
	lockAt := clock("2014/01/01 12:30:00")
	unlockAt := clock("2014/01/02 6:00:00")

	c := bl.Set(now)

	assertEqual(t, c.LockAt, lockAt)
	assertEqual(t, c.UnlockAt, unlockAt)
}

func TestAllow(t *testing.T) {
	bl := Create()
	now := time.Now()
	c := bl.Set(now)

	assertEqual(t, true, c.Allow())

	later, _ := time.ParseDuration("-1h")
	c = bl.Set(now.Add(later))
	assertEqual(t, false, c.Allow())
}
