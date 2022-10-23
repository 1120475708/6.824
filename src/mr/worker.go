package mr

import (
	"fmt"
	"time"
)
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

//
// main/mrworker.go calls this function.
//

func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {
	c := WorkerProcess{
		IP:          IPAddress,
		Port:        GetPort(),
		StatusTable: map[string]int{},
	}
	c.server()

	for {
		resp := GetTaskResp{}
		req := GetTaskReq{}
		call("Coordinator.GetTask", &req, &resp, IPAddress, CoordinatorPort)
		switch resp.TaskType {
		case MapTask:

		case ReduceTask:

		case WaitingTask:
			time.Sleep(1 * time.Second)
		case ExitTask:
			return
		}

	}

	//resp := GetTaskResp{}
	//req := GetTaskReq{}
	//call("Coordinator.GetTask", &req, &resp)
	//
	//fmt.Println("Debug ", resp)
	//
	//time.Sleep(100 * time.Second)

	// Your worker implementation here.

	// uncomment to send the Example RPC to the coordinator.
	// CallExample()

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
	ok := call("Coordinator.Example", &args, &reply, IPAddress, CoordinatorPort)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}

}
