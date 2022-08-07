package raft

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestRand(t *testing.T) {
	for i := 0; i < 3; i++ {
		//rand.Seed(int64(i))
		fmt.Println(rand.Intn(100))
		fmt.Println(rand.Intn(100))
		fmt.Println(rand.Intn(100))
	}
}
