package util

import (
	"runtime"
	"time"
)

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	//fmt.Printf("mem: Alloc = %v MiB", bToMb(m.Alloc))
	//fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	//fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	//fmt.Printf("\tNumGC = %v\n", m.NumGC)

	Printf("mem alloc %dM heap %dM total %dM free %dM sys %dM gc %d\n",
		bToMb(m.Alloc), bToMb(m.HeapAlloc), bToMb(m.TotalAlloc), bToMb(m.Frees), bToMb(m.Sys), m.NumGC)

}
func mem_tick() {
	for range time.Tick(time.Second * 30) {
		PrintMemUsage()
	}
}

func init() {
	//go mem_tick()
}
