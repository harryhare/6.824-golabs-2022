package util

import (
	"fmt"
	"testing"
)

func TestTime2Int(t *testing.T) {
	//fmt.Println(Time2Int("2021-11-16 00:02:12"))
	fmt.Println(Time2Int("2000-01-02 03:04:05"))
	fmt.Println(Time2Int("2022-12-31 23:59:59"))
}
