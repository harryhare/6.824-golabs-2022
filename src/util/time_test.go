package util

import (
	"testing"
	"time"
)

func TestStopWatch(t *testing.T) {
	s := NewStopWatch("test")
	time.Sleep(2 * time.Millisecond)
	s.Stop()
}
