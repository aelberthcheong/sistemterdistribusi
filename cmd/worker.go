package cmd

// start a worker process, which is implemented
// in ../mr/worker.go. typically there will be
// multiple worker processes, talking to one coordinator.
//
// go run mrworker.go wc.so

import (
	"github.com/spf13/cobra"

	"sistemterdistribusi/internal/mr"
)

func createWorkerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "worker <plugin.so> <sockname>",
		Short: "Start a MapReduce worker",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			mapf, reducef, err := loadPlugin(args[0])
			if err != nil {
				return err
			}

			mr.Worker(args[1], mapf, reducef)
			return nil
		},
	}
}
