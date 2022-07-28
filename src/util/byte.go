package util

import "encoding/binary"

func Int64toBytes(v uint64) []byte {
	var a = make([]byte, 8)
	for i := uint(0); i < 8; i++ {
		a[7-i] = (byte)((v) & 0x0ff)
		v = v >> 8
	}
	return a
}

func BytestoInt64(a [] byte) uint64 {
	var v uint64
	for i := uint(0); i < 8; i++ {
		v = (v << 8) | (uint64)(a[i]);
	}
	return v
}

/* bold 的文档中给的代码：
 * https://github.com/boltdb/bolt
 * 其他参考：
 * https://studygolang.com/articles/16154
 */

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}