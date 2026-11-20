package mr

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/rpc"
	"os"
	"sort"
	"time"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// Use ihash(key) % NReduce to choose the reduce task for a key.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

var coordSockName string // socket for the coordinator

// Worker asks the coordinator for tasks until the coordinator says stop.
// It also stops if the coordinator is not reachable.
func Worker(sockname string, mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	coordSockName = sockname

	for {
		reply := GetTaskReply{}
		if !call("Coordinator.GetTask", &GetTaskArgs{}, &reply) {
			return
		}

		switch reply.Type {
		case MapTask:
			if err := doMap(mapf, reply.Filename, reply.TaskID, reply.NReduce); err != nil {
				log.Printf("map task %d failed: %v", reply.TaskID, err)
				continue
			}
			call("Coordinator.ReportTask",
				&ReportTaskArgs{Type: MapTask, TaskID: reply.TaskID}, &ReportTaskReply{})
		case ReduceTask:
			if err := doReduce(reducef, reply.TaskID, reply.NMap); err != nil {
				log.Printf("reduce task %d failed: %v", reply.TaskID, err)
				continue
			}
			call("Coordinator.ReportTask",
				&ReportTaskArgs{Type: ReduceTask, TaskID: reply.TaskID}, &ReportTaskReply{})
		case WaitTask:
			time.Sleep(500 * time.Millisecond)
		case ExitTask:
			return
		}
	}
}

// doMap runs the map function on one file.
// It writes nReduce files named mr-<mapID>-<reduceID>.
func doMap(mapf func(string, string) []KeyValue, filename string, mapID, nReduce int) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	buckets := make([][]KeyValue, nReduce)
	for _, kv := range mapf(filename, string(content)) {
		r := ihash(kv.Key) % nReduce
		buckets[r] = append(buckets[r], kv)
	}

	for r, bucket := range buckets {
		name := fmt.Sprintf("mr-%d-%d", mapID, r)
		err := writeAtomic(name, func(w io.Writer) error {
			enc := json.NewEncoder(w)
			for _, kv := range bucket {
				if err := enc.Encode(&kv); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// doReduce reads the files mr-*-<reduceID> from all map tasks.
// It groups the pairs by key and writes the result to mr-out-<reduceID>.
func doReduce(reducef func(string, []string) string, reduceID, nMap int) error {
	var kva []KeyValue
	for m := 0; m < nMap; m++ {
		f, err := os.Open(fmt.Sprintf("mr-%d-%d", m, reduceID))
		if err != nil {
			return err
		}
		dec := json.NewDecoder(f)
		for {
			var kv KeyValue
			if err := dec.Decode(&kv); err != nil {
				f.Close()
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}
			kva = append(kva, kv)
		}
	}

	sort.Slice(kva, func(i, j int) bool { return kva[i].Key < kva[j].Key })

	return writeAtomic(fmt.Sprintf("mr-out-%d", reduceID), func(w io.Writer) error {
		for i := 0; i < len(kva); {
			j := i + 1
			for j < len(kva) && kva[j].Key == kva[i].Key {
				j++
			}
			values := make([]string, 0, j-i)
			for k := i; k < j; k++ {
				values = append(values, kva[k].Value)
			}
			if _, err := fmt.Fprintf(w, "%v %v\n", kva[i].Key, reducef(kva[i].Key, values)); err != nil {
				return err
			}
			i = j
		}
		return nil
	})
}

// call sends an RPC request to the coordinator and waits for the reply.
// It returns false if the coordinator is not reachable or the call fails.
func call(rpcname string, args interface{}, reply interface{}) bool {
	c, err := rpc.DialHTTP("unix", coordSockName)
	if err != nil {
		// The coordinator exits when the job is done. This is normal.
		return false
	}
	defer c.Close()

	if err := c.Call(rpcname, args, reply); err != nil {
		log.Printf("%d: call failed err %v", os.Getpid(), err)
		return false
	}
	return true
}
