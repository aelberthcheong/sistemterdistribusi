package mr

// TaskType tells the worker what to do next.
type TaskType int

const (
	MapTask    TaskType = iota
	ReduceTask          // all map tasks are done
	WaitTask            // no task is free now, so ask again later
	ExitTask            // all work is done, so stop
)

type GetTaskArgs struct{}

type GetTaskReply struct {
	Type     TaskType
	TaskID   int
	Filename string // input file, map tasks only
	NMap     int    // number of map tasks
	NReduce  int    // number of reduce tasks
}

type ReportTaskArgs struct {
	Type   TaskType
	TaskID int
}

type ReportTaskReply struct{}
