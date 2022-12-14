package mr

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)
import "net"
import "os"
import "net/rpc"
import "net/http"

const (
	TaskToDo = iota
	TaskFinished
	TaskRuning
)
const (
	PhaseMap = iota
	PhaseReduce
)
const (
	timeout  = 5 * time.Second
	interval = 5 * time.Second
)

type MapTask struct {
	Id     string
	File   []string
	WorkId int    // 1-Nreduce
	Worker string //which worker is on this task
	Status int
	start  time.Time
}

type ReduceTask struct {
	Id     string
	WorkId int    // 1-Nreduce
	Worker string // which worker is on this task
	Status int
	start  time.Time
}

type Coordinator struct {
	Nmap               int
	NReduce            int
	TaskTodo           chan interface{}
	MapTaskFinished    []*MapTask
	ReduceTaskFinished []*ReduceTask

	MapTasks    map[string]*MapTask
	ReduceTasks map[string]*ReduceTask
	MapLock     sync.Mutex
	ReduceLock  sync.Mutex

	MapDone    chan int
	ReduceDone chan int
}

func (c *Coordinator) handleGet(args *TaskRequest, reply *TaskResponse) {

	task := <-c.TaskTodo
	if task == nil {
		reply = &TaskResponse{
			Type: TypeQuit,
		}
		return
	}
	switch task.(type) {
	case *MapTask:
		mapTask := task.(*MapTask)
		*reply = TaskResponse{
			Type:    TypeMap,
			TaskId:  mapTask.Id,
			Files:   mapTask.File,
			WorkId:  mapTask.WorkId,
			NMap:    c.Nmap,
			NReduce: c.NReduce,
		}
		mapTask.Status = TaskRuning
		mapTask.start = time.Now()
		mapTask.Worker = args.SelfId

		c.MapLock.Lock()
		c.MapTasks[mapTask.Id] = mapTask
		c.MapLock.Unlock()
		fmt.Printf("assign task %s to %s\n", mapTask.Id, mapTask.Worker)
	case *ReduceTask:
		reduceTask := task.(*ReduceTask)
		*reply = TaskResponse{
			Type:    TypeReduce,
			TaskId:  reduceTask.Id,
			WorkId:  reduceTask.WorkId,
			NMap:    c.Nmap,
			NReduce: c.NReduce,
		}
		reduceTask.Status = TaskRuning
		reduceTask.start = time.Now()
		reduceTask.Worker = args.SelfId

		c.ReduceLock.Lock()
		c.ReduceTasks[reduceTask.Id] = reduceTask
		c.ReduceLock.Unlock()
		fmt.Printf("assign task %s to %s\n", reduceTask.Id, reduceTask.Worker)

	default:
		panic(errors.New("unexpect task type"))
	}
}

func (c *Coordinator) checkTimeout() {
	fmt.Println("check timeout")
	now := time.Now()
	c.MapLock.Lock()
	for k, v := range c.MapTasks {
		if now.Sub(v.start) > timeout {
			delete(c.MapTasks, k)
			v.Status = TaskToDo
			c.TaskTodo <- v
			fmt.Printf("task %s timeout\n", v.Id)
		}
	}
	c.MapLock.Unlock()

	c.ReduceLock.Lock()
	for k, v := range c.ReduceTasks {
		if now.Sub(v.start) > timeout {
			delete(c.ReduceTasks, k)
			v.Status = TaskToDo
			c.TaskTodo <- v
			fmt.Printf("task %s timeout", v.Id)
		}
	}
	c.ReduceLock.Unlock()
}

func (c *Coordinator) handleFinish(args *TaskRequest, reply *TaskResponse) {
	taskId := args.TaskId
	if strings.HasPrefix(taskId, "map") {
		c.MapLock.Lock()
		defer c.MapLock.Unlock()

		task := c.MapTasks[taskId]
		if task == nil {
			return
		}

		task.Status = TaskFinished
		delete(c.MapTasks, taskId)
		c.MapTaskFinished = append(c.MapTaskFinished, task)

		fmt.Printf("task %s finished from %s\n", taskId, args.SelfId)

		if len(c.MapTaskFinished) == c.Nmap {
			close(c.MapDone)
			fmt.Println("map work done!")
			go c.produceReduceTasks()
		}
		return
	}
	if strings.HasPrefix(taskId, "reduce") {
		c.ReduceLock.Lock()
		defer c.ReduceLock.Unlock()
		task := c.ReduceTasks[taskId]
		if task == nil {
			return
		}

		task.Status = TaskFinished
		delete(c.ReduceTasks, taskId)
		c.ReduceTaskFinished = append(c.ReduceTaskFinished, task)
		fmt.Printf("task %s finished from %s\n", taskId, args.SelfId)

		if len(c.ReduceTaskFinished) == c.NReduce {
			close(c.ReduceDone)
			fmt.Println("reduce work done!")
		}
		return
	}
}

func (c *Coordinator) handleError(args *TaskRequest, reply *TaskResponse) {
	taskId := args.TaskId
	if strings.HasPrefix(taskId, "map") {
		c.MapLock.Lock()
		defer c.MapLock.Unlock()

		task := c.MapTasks[taskId]
		if task == nil {
			return
		}

		task.Status = TaskToDo
		delete(c.MapTasks, taskId)
		c.TaskTodo <- task
		return
	}
	if strings.HasPrefix(taskId, "reduce") {
		c.ReduceLock.Lock()
		defer c.ReduceLock.Unlock()
		task := c.ReduceTasks[taskId]
		if task == nil {
			return
		}

		task.Status = TaskToDo
		delete(c.ReduceTasks, taskId)
		c.TaskTodo <- task
		return
	}
}

func (c *Coordinator) RequestTask(args *TaskRequest, reply *TaskResponse) error {
	switch args.Type {
	case TypeGet:
		c.handleGet(args, reply)
	case TypeFinish:
		c.handleFinish(args, reply)
	case TypeError:
		c.handleError(args, reply)
	}
	return nil
}

func (c *Coordinator) server() {
	err := rpc.Register(c)
	if err != nil {
		panic(err)
	}
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

func (c *Coordinator) Done() bool {
	<-c.ReduceDone
	return true
}

func (c *Coordinator) produceMapTasks(files []string) {
	for i, file := range files {
		c.TaskTodo <- &MapTask{
			Id:     fmt.Sprintf("map_%d", i+1),
			WorkId: i,
			File:   []string{file},
			Status: TaskToDo,
		}
	}
	fmt.Printf("add %d map task to queue\n", len(files))
}
func (c *Coordinator) produceReduceTasks() {
	for i := 0; i < c.NReduce; i++ {
		c.TaskTodo <- &ReduceTask{
			Id:     fmt.Sprintf("reduce_%d", i+1),
			WorkId: i,
			Status: TaskToDo,
		}
	}

	fmt.Printf("add %d reduce task to queue\n", c.NReduce)
}
func checkTimeout(c *Coordinator) {
	for range time.Tick(interval) {
		c.checkTimeout()
	}
}

func MakeCoordinator(files []string, nReduce int) *Coordinator {
	fmt.Printf("files %v\n", files)
	fmt.Printf("nReduce %d\n", nReduce)
	c := Coordinator{}
	c.Nmap = len(files)
	c.NReduce = nReduce
	chanSize := c.NReduce
	if c.Nmap > chanSize {
		chanSize = c.Nmap
	}
	c.TaskTodo = make(chan interface{}, chanSize)
	c.MapTasks = map[string]*MapTask{}
	c.ReduceTasks = map[string]*ReduceTask{}
	c.MapDone = make(chan int, 0)
	c.ReduceDone = make(chan int, 0)
	go c.produceMapTasks(files)

	c.server()

	go checkTimeout(&c)

	return &c
}
