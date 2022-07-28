package util

import (
	"time"
)

var start time.Time = time.Now()

type StopWatch struct {
	t    time.Time
	text string
}

func NewStopWatch(text string) (s StopWatch) {
	s.t = time.Now()
	s.text = text
	Printf("%s start\n", text)
	return
}

func (s *StopWatch) Stop() {
	now := time.Now()
	elapsed := now.Sub(s.t)
	elapsed_total := now.Sub(start)
	Printf("%s   end %.3f %.3f\n", s.text, float64(elapsed)/float64(time.Second), float64(elapsed_total)/float64(time.Second))
}

func PrintTime() {
	now := time.Now()
	elapsed := now.Sub(start)
	Println(elapsed)
}
