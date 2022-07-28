package util

import (
	"reflect"
	"unsafe"
)

func Interface2Key(v interface{}) int64 {
	switch v.(type) {
	case int64:
		return v.(int64)
	case float32:
		return int64(v.(float32))
	case float64:
		return int64(v.(float64))
	case string:
		// todo 判断是time吗
		return int64(String2Int(v.(string)))
	}
	return 0
}

func Bytes2Str(slice []byte) string {
	return *(*string)(unsafe.Pointer(&slice))
}

func Str2Bytes(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

func String2Int(s string) int32 {
	b := Str2Bytes(s)
	//return (int64(b[0]) << 56) | (int64(b[1]) << 48) | (int64(b[2]) << 40) | (int64(b[3]) << 32) | (int64(b[4]) << 24) | (int64(b[5]) << 16) | (int64(b[6]) << 8) | (int64(b[7]) << 0)
	return (int32(b[0]) << 24) | (int32(b[1]) << 16) | (int32(b[2]) << 8) | (int32(b[3]) << 0)
}

//2021-11-16 00:02:12
func Time2Int(t string) int32 {
	b := Str2Bytes(t)
	Assert(len(b) == 19)
	Assert(b[4] == '-')
	Assert(b[7] == '-')
	Assert(b[10] == ' ')
	Assert(b[13] == ':')
	Assert(b[16] == ':')
	y := int(b[0]-'0')*1000 + int(b[1]-'0')*100 + int(b[2]-'0')*10 + int(b[3]-'0')
	m := int(b[5]-'0')*10 + int(b[6]-'0')
	d := int(b[8]-'0')*10 + int(b[9]-'0')
	h := int(b[11]-'0')*10 + int(b[12]-'0')
	n := int(b[14]-'0')*10 + int(b[15]-'0')
	s := int(b[17]-'0')*10 + int(b[18]-'0')
	return int32((((((y-2000)*12+m-1)*31+d-1)*24+h)*60+n)*60 + s)
}
