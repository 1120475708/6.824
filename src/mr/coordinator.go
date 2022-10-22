package mr

import (
	"log"
	"sync"
	"time"
)
import "net"
import "net/rpc"
import "net/http"

type Coordinator struct {
	// Your definitions here.
	Files             []string
	ReduceNum         int
	IsClosed          bool
	MapTask           chan string
	ReduceTask        chan MapWorker
	MapTaskStatus     map[string]MapWorker
	ReduceTaskStatus  map[string]bool
	LeftMapTaskNum    int
	LeftReduceTaskNum int
	TaskType          int
	lock              sync.Mutex
}

type MapWorker struct {
	IsFinished bool
	IP         string
	Port       string
	FileName   string
}

// Your code here -- RPC handlers for the worker to call.

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

func (c *Coordinator) GetTask(req *GetTaskReq, resp *GetTaskResp) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	resp.TaskType = c.TaskType
	switch resp.TaskType {
	case MapTask:
		resp.FileName = <-c.MapTask
	case ReduceTask:
	case WaitingTask:
	case ExitTask:

	}

	return nil

}

//
// start a thread that listens for RPCs from worker.go
//
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1234")
	//sockname := coordinatorSock()
	//os.Remove(sockname)
	//l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
//
func (c *Coordinator) Done() bool {
	// Your code here.
	return c.IsClosed
}
func Daemon(c *Coordinator) {
	for {
		c.lock.Lock()

		c.lock.Unlock()
		time.Sleep(time.Second)

	}
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	mapTask := make(chan string, 64)
	for _, file := range files {
		mapTask <- file
	}
	c := Coordinator{
		Files:             files,
		ReduceNum:         nReduce,
		IsClosed:          false,
		MapTask:           mapTask,
		ReduceTask:        nil,
		MapTaskStatus:     make(map[string]MapWorker),
		ReduceTaskStatus:  nil,
		LeftMapTaskNum:    len(files),
		LeftReduceTaskNum: 0,
		TaskType:          MapTask,
	}

	// Your code here.
	go Daemon(&c)
	c.server()
	return &c
}
