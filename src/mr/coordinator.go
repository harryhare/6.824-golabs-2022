package mr

import (
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

type MapTask struct {
	Id     string
	File   []string
	Worker int //which worker is on this task
	Status int
	start  time.Time
}

type ReduceTask struct {
	Id     string
	WorkId int // 1-Nreduce
	Worker int // which worker is on this task
	Status int
	start  time.Time
}

type Coordinator struct {
	Nmap               int
	NReduce            int
	MapTaskTodo        []*MapTask
	ReduceTaskTodo     []*ReduceTask
	MapTaskFinished    []*MapTask
	ReduceTaskFinished []*ReduceTask

	MapTasks    map[string]*MapTask
	ReduceTasks map[string]*ReduceTask
	MapLock     sync.Mutex
	ReduceLock  sync.Mutex

	MapDone    chan int
	ReduceDone chan int
}

func (c *Coordinator) HandleGet(args *TaskRequest, reply *TaskResponse) {

}

func (c *Coordinator) HandleFinish(args *TaskRequest, reply *TaskResponse) {
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

		if len(c.MapTaskFinished) == c.Nmap {
			close(c.MapDone)
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

		if len(c.ReduceTaskFinished) == c.NReduce {
			close(c.ReduceDone)
		}
		return
	}
}

func (c *Coordinator) HandleError(args *TaskRequest, reply *TaskResponse) {
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
		c.MapTaskTodo = append(c.MapTaskTodo, task)
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
		c.ReduceTaskTodo = append(c.ReduceTaskTodo, task)
		return
	}
}

func (c *Coordinator) RequestTask(args *TaskRequest, reply *TaskResponse) error {
	switch args.Type {
	case TypeGet:
		c.HandleGet(args, reply)
	case TypeFinish:
		c.HandleFinish(args, reply)
	case TypeError:
		c.HandleError(args, reply)
	}
	return nil
}

func (c *Coordinator) server() {
	rpc.Register(c)
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

func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}
	c.Nmap = len(files)
	c.NReduce = nReduce
	c.MapTaskTodo = []*MapTask{}
	c.ReduceTaskTodo = []*ReduceTask{}
	c.MapTasks = map[string]*MapTask{}
	c.ReduceTasks = map[string]*ReduceTask{}
	c.MapDone = make(chan int, 0)
	c.ReduceDone = make(chan int, 0)
	for i, file := range files {
		c.MapTaskTodo = append(c.MapTaskTodo, &MapTask{
			Id:     fmt.Sprintf("map_%d", i+1),
			File:   []string{file},
			Status: TaskToDo,
		})
	}
	for i := 0; i < nReduce; i++ {
		c.ReduceTaskTodo = append(c.ReduceTaskTodo, &ReduceTask{
			Id:     fmt.Sprintf("reduce_%d", i+1),
			WorkId: i + 1,
			Status: TaskToDo,
		})
	}
	c.server()
	return &c
}
