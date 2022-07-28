package mr

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
)
import "log"
import "net/rpc"
import "hash/fnv"

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

var SelfId = fmt.Sprintf("worker-%d", rand.Int31()%1000)

//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.

	// uncomment to send the Example RPC to the coordinator.
	// CallExample()
	for true {
		task := GetTask()
		if task == nil {
			fmt.Println("rpc call failed, quit")
			break
		}
		if task.Type == TypeQuit {
			fmt.Println("quit command received, quit")
			break
		}
		if task.Type == TypeMap {
			reduce := task.NReduce
			files := []*os.File{}
			var err error
			for i := 0; i < reduce; i++ {
				filename := fmt.Sprintf("mr-%d-%d", task.WorkId, i)
				f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
				if err != nil {
					fmt.Errorf("open file %s error, %s", filename, err)
					break
				}
				files = append(files, f)
			}
			if err != nil {
				ErrorTask(task)
				continue
			}
			for _, filename := range task.Files {
				file, err := os.Open(filename)
				if err != nil {
					log.Fatalf("cannot open %v", filename)
				}
				content, err := ioutil.ReadAll(file)
				if err != nil {
					log.Fatalf("cannot read %v", filename)
				}
				file.Close()
				kva := mapf(filename, string(content))
				// todo write to file
				for _, kv := range kva {
					i := ihash(kv.Key) % reduce
					files[i].Write()
				}
			}
			continue
		}
		if task.Type == TypeReduce {
			intermediate := []KeyValue{}
			for _, filename := range task.Files {
				file, err := os.Open(filename)
				if err != nil {
					log.Fatalf("cannot open %v", filename)
				}
				content, err := ioutil.ReadAll(file)
				if err != nil {
					log.Fatalf("cannot read %v", filename)
				}
				file.Close()
				kva := mapf(filename, string(content))
				intermediate = append(intermediate, kva...)
			}
			continue
		}
	}

}

func ErrorTask(t *TaskResponse) *TaskResponse {
	args := &TaskRequest{
		Type:   TypeError,
		SelfId: SelfId,
		TaskId: t.TaskId,
	}
	reply := &TaskResponse{}
	ok := call("Coordinator.RequestTask", &args, &reply)
	if ok {
		fmt.Printf("reply %v\n", reply)
	} else {
		fmt.Printf("call failed!\n")
		return nil
	}
	return reply
}

func FinishTask(t *TaskResponse) *TaskResponse {
	args := &TaskRequest{
		Type:   TypeFinish,
		SelfId: SelfId,
		TaskId: t.TaskId,
	}
	reply := &TaskResponse{}
	ok := call("Coordinator.RequestTask", &args, &reply)
	if ok {
		fmt.Printf("reply %v\n", reply)
	} else {
		fmt.Printf("call failed!\n")
		return nil
	}
	return reply
}

func GetTask() *TaskResponse {
	args := &TaskRequest{
		Type:   TypeGet,
		SelfId: SelfId,
		TaskId: "null",
	}
	reply := &TaskResponse{}
	ok := call("Coordinator.RequestTask", &args, &reply)
	if ok {
		fmt.Printf("reply %v\n", reply)
	} else {
		fmt.Printf("call failed!\n")
		return nil
	}
	return reply
}

//
// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
//
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

//
// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
