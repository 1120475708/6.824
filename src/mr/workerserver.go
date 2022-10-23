package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type WorkerProcess struct {
	IP          string
	Port        string
	StatusTable map[string]int
}

func (c *WorkerProcess) GetWorkerStatus(req *GetWorkerStatusReq, resp *GetWorkerStatusResp) error {
	fileName := req.FileName
	resp.Status = c.StatusTable[fileName]

	return nil
}

func (c *WorkerProcess) server() {
	err := rpc.Register(c)
	if err != nil {
		log.Fatal("Register fail: ", err)
		return
	}
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", c.Port)
	//sockname := coordinatorSock()
	//os.Remove(sockname)
	//l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)

}
