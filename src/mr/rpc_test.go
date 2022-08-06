package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"testing"
	"time"
)

type Args struct {
	A int
}
type SimpleRPC struct {
}

func (s *SimpleRPC) Echo(a int, b *int) error {
	*b = a
	return nil
}

func RPCClient() {
	sockname := "/tmp/test.sock"
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	var a = 12345
	//Args{
	//	A: 1234,
	//}
	var b int
	before := time.Now()
	err = c.Call("SimpleRPC.Echo", a, &b)
	if err != nil {
		panic(err)
	}

	fmt.Printf("rpc cost %v", time.Now().Sub(before))
	fmt.Println(b)
}

func TestRPC(t *testing.T) {
	fmt.Printf("%v\n", time.Millisecond)
	fmt.Printf("%v\n", time.Microsecond)
	fmt.Printf("%v\n", time.Nanosecond)
	s := &SimpleRPC{}
	rpc.Register(s)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := "/tmp/test.sock"
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)

	RPCClient()
}
