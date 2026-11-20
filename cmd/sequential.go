package cmd

// simple sequential MapReduce.
//
// go run mrsequential.go wc.so pg*.txt

import (
	"fmt"
	"os"
	"sort"

	"sistemterdistribusi/internal/mr"

	"github.com/spf13/cobra"
)

// createSequentialCmd runs MapReduce in one process.
// Use it to make the reference output (mr-out-0).
func createSequentialCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sequential <plugin.so> <inputfiles...>",
		Short: "Run MapReduce sequentially in one process",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			mapf, reducef, err := loadPlugin(args[0])
			if err != nil {
				return err
			}
			return runSequential(mapf, reducef, args[1:])
		},
	}
}

func runSequential(mapf mapFunc, reducef reduceFunc, files []string) error {
	// Read each input file and call Map.
	// Collect all Map output in one slice.
	var intermediate []mr.KeyValue
	for _, filename := range files {
		content, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("cannot read %v: %w", filename, err)
		}
		intermediate = append(intermediate, mapf(filename, string(content))...)
	}

	// Real MapReduce splits the data into NxM buckets.
	// This version keeps all data in one slice.
	sort.Slice(intermediate, func(i, j int) bool {
		return intermediate[i].Key < intermediate[j].Key
	})

	ofile, err := os.Create("mr-out-0")
	if err != nil {
		return err
	}
	defer ofile.Close()

	// Call Reduce one time for each different key.
	// Write each result to mr-out-0.
	for i := 0; i < len(intermediate); {
		j := i + 1
		for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
			j++
		}

		values := make([]string, 0, j-i)
		for k := i; k < j; k++ {
			values = append(values, intermediate[k].Value)
		}

		output := reducef(intermediate[i].Key, values)
		if _, err := fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, output); err != nil {
			return err
		}

		i = j
	}

	return nil
}
