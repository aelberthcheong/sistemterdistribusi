package cmd

// start the coordinator process, which is implemented
// in ../mr/coordinator.go
//
// go run mrcoordinator.go pg*.txt

import (
	"time"

	"github.com/spf13/cobra"

	"sistemterdistribusi/internal/mr"
)

func createCoordinatorCmd() *cobra.Command {
	var nReduce int

	cmd := &cobra.Command{
		Use:   "coordinator <sockname> <inputfiles...>",
		Short: "Start the MapReduce coordinator",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			m := mr.MakeCoordinator(args[0], args[1:], nReduce)

			// Wait until the job is done or the user stops the process.
			for !m.Done() {
				select {
				case <-cmd.Context().Done():
					return cmd.Context().Err()
				case <-time.After(time.Second):
				}
			}

			// Give the workers time to read the exit reply.
			time.Sleep(time.Second)
			return nil
		},
	}

	cmd.Flags().IntVar(&nReduce, "nreduce", 10, "number of reduce tasks")

	return cmd
}
