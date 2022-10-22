package mr

import (
	"fmt"
	"log"
)
import "net"
import "net/rpc"
import "net/http"

type Coordinator struct {
	// Your definitions here.
	files     []string
	reduceNum int
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

type task struct {
	taskName  string
	reduceNum int
}

func (c *Coordinator) GetTask(req *GetTaskReq, resp *GetTaskResp) error {

	resp.Name = c.files[0]
	resp.Num = c.reduceNum
	fmt.Println("Get Task Success!")
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
	ret := false

	// Your code here.

	return ret
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{
		files:     files,
		reduceNum: nReduce,
	}

	// Your code here.

	c.server()
	return &c
}
