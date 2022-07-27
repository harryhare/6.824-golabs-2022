package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import "os"
import "strconv"

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

const (
	TypeGet = iota
	TypeFinish
	TypeError
)

const (
	TypeMap = iota
	TypeReduce
	TypeQuit
)

type TaskRequest struct {
	Type   int // TypeTask,TypeFinish,TypeError
	SelfId int
	TaskId string // only valid when Type==TypeFinish
}

type TaskResponse struct {
	Type   int    //map,reduce,quit
	TaskId string // 每个task 可以被多次分配, task id 唯一
	Files  []string
}

// Add your RPC definitions here.

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
