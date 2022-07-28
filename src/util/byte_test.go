package util

import (
	"fmt"
	"testing"
)

func TestInt64toBytes(t *testing.T) {
	var x uint64=0x0102030405060708
	b:=Int64toBytes(x)
	for i:=0;i<8;i++{
		fmt.Printf("%02d",b[i]);
	}
	fmt.Println()
	y:= BytestoInt64(b)
	fmt.Println(y)
	Assert(x==y)
}
