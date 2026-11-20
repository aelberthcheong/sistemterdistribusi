package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

// A worker has this much time to finish a task.
// After that, the coordinator gives the task to another worker.
const taskTimeout = 10 * time.Second

type taskState int

const (
	idle taskState = iota
	inProgress
	completed
)

type task struct {
	state     taskState
	startedAt time.Time
	filename  string // map tasks only
}

type Coordinator struct {
	mu          sync.Mutex
	nMap        int
	nReduce     int
	mapTasks    []task
	reduceTasks []task
	mapsDone    int
	reducesDone int
}

// GetTask gives a task to a worker.
// Reduce tasks start only after all map tasks are complete.
func (c *Coordinator) GetTask(args *GetTaskArgs, reply *GetTaskReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	reply.NMap = c.nMap
	reply.NReduce = c.nReduce

	switch {
	case c.mapsDone < c.nMap:
		if id := pickTask(c.mapTasks); id >= 0 {
			reply.Type = MapTask
			reply.TaskID = id
			reply.Filename = c.mapTasks[id].filename
		} else {
			reply.Type = WaitTask
		}
	case c.reducesDone < c.nReduce:
		if id := pickTask(c.reduceTasks); id >= 0 {
			reply.Type = ReduceTask
			reply.TaskID = id
		} else {
			reply.Type = WaitTask
		}
	default:
		reply.Type = ExitTask
	}
	return nil
}

// ReportTask marks a task as complete.
// A second report for the same task has no effect.
func (c *Coordinator) ReportTask(args *ReportTaskArgs, reply *ReportTaskReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch args.Type {
	case MapTask:
		if markDone(c.mapTasks, args.TaskID) {
			c.mapsDone++
		}
	case ReduceTask:
		if markDone(c.reduceTasks, args.TaskID) {
			c.reducesDone++
		}
	}
	return nil
}

// pickTask finds a task that is idle or has timed out.
// It marks the task as in progress and returns its index.
// It returns -1 if no task is free.
func pickTask(tasks []task) int {
	now := time.Now()
	for i := range tasks {
		t := &tasks[i]
		if t.state == idle || (t.state == inProgress && now.Sub(t.startedAt) > taskTimeout) {
			t.state = inProgress
			t.startedAt = now
			return i
		}
	}
	return -1
}

// markDone returns true if the task changes to complete now.
func markDone(tasks []task, id int) bool {
	if id < 0 || id >= len(tasks) || tasks[id].state == completed {
		return false
	}
	tasks[id].state = completed
	return true
}

// server starts a thread that listens for RPCs from worker.go.
func (c *Coordinator) server(sockname string) {
	rpc.Register(c)
	rpc.HandleHTTP()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatalf("listen error %s: %v", sockname, e)
	}
	go http.Serve(l, nil)
}

// Done returns true when all reduce tasks are complete.
// The coordinator command calls it at intervals.
func (c *Coordinator) Done() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.reducesDone == c.nReduce
}

// MakeCoordinator creates a coordinator with one map task for each input file.
// nReduce is the number of reduce tasks.
func MakeCoordinator(sockname string, files []string, nReduce int) *Coordinator {
	c := &Coordinator{
		nMap:        len(files),
		nReduce:     nReduce,
		mapTasks:    make([]task, len(files)),
		reduceTasks: make([]task, nReduce),
	}
	for i, f := range files {
		c.mapTasks[i].filename = f
	}

	c.server(sockname)
	return c
}
