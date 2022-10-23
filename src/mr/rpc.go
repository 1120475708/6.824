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

// Add your RPC definitions here.
type GetTaskReq struct {
	IP   string
	Port string
}
type GetTaskResp struct {
	TaskType int
	FileName string
}

type GetWorkerStatusReq struct {
	FileName string
}
type GetWorkerStatusResp struct {
	Status int
}

// Cook up a unique-ish UNIX-domain socket Name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
