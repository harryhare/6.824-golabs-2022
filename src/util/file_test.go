package util

import (
	"fmt"
	"testing"
)

func TestFileExist(t *testing.T) {
	fmt.Println(FileExist("/tmp/tmp.txt"))
	fmt.Println(FileExist("/tmp/tmp1.txt"))
}
