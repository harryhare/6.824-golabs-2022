package util

import (
	"testing"
	"time"
)

func TestPrintln(t *testing.T) {
	Println("test")
}

func TestPrintf(t *testing.T) {
	Printf("*%d*\n", 1234)
}

type D struct {
	A int
	B int
}

func TestPrint(t *testing.T) {
	d := &D{
		1,
		2,
	}
	Print(d)
}

func TestAppendPrintForce(t *testing.T) {
	Println("1")
	Println("2")
	FlushLog()
	Println("3")
	Println("4")
	FlushLog()
}
func TestAppendPrint(t *testing.T) {
	Println("1")
	time.Sleep(10 * time.Second)
	Println("2")
	time.Sleep(10 * time.Second)
	Println("3")
	time.Sleep(10 * time.Second)
	Println("4")
	time.Sleep(10 * time.Second)
}
