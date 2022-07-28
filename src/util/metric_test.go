package util

import (
	"testing"
	"time"
)

func TestPrintMemUsage(t *testing.T) {
	a := make([]byte, 1024*1024)
	for i := 0; i < len(a); i++ {
		a[i] = byte(i)
	}
	PrintMemUsage()
}
func TestPrintMemUsage2(t *testing.T) {
	time.Sleep(10 * time.Second)
}
