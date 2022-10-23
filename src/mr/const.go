package mr

const (
	IPAddress       = "127.0.0.1"
	CoordinatorPort = ":1234"
)

const (
	MapTask = iota
	ReduceTask
	WaitingTask
	ExitTask
)

const (
	InProcess = iota
	Finished
)
