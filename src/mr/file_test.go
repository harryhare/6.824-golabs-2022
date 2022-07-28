package mr

import (
	"6.824/util"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestReadFile(t *testing.T) {
	file, err := os.OpenFile("/tmp/kv_file.json", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		panic(file)
	}
	enc := json.NewEncoder(file)
	for i := 0; i < 10; i++ {
		kv := &KeyValue{
			Key:   fmt.Sprintf("key-%d", i),
			Value: fmt.Sprintf("value-%d", i),
		}
		err = enc.Encode(kv)
		if err != nil {
			panic(err)
		}
	}
	file.Close()
}
func TestWriteFile(t *testing.T) {

	file, err := os.OpenFile("/tmp/kv_file.json", os.O_RDONLY, 0666)
	if err != nil {
		panic(file)
	}
	dec := json.NewDecoder(file)
	kva := []KeyValue{}
	for {
		var kv KeyValue
		if err := dec.Decode(&kv); err != nil {
			break
		}
		kva = append(kva, kv)
	}
	util.Print(kva)
}
