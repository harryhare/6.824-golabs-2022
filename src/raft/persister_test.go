package raft

import (
	"6.824/labgob"
	"bytes"
	"fmt"
	"testing"
)

func TestBasic(t *testing.T) {
	persister := MakePersister()

	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	e.Encode(1234)
	data := w.Bytes()
	persister.SaveRaftState(data)

	var term int
	rdata := persister.ReadRaftState()
	r := bytes.NewBuffer(rdata)
	d := labgob.NewDecoder(r)
	err := d.Decode(&term)
	if err != nil {
		panic(err)
	}
	fmt.Println(term)
	// Output:
	// 1234
}
