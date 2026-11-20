// cmd/root.go
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

// Set at compile time with ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func createRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "mr",
		Short:         "MapReduce coordinator, worker, and sequential runner",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.SetVersionTemplate(fmt.Sprintf(
		"version %s (commit: %s, built at: %s)\n",
		version, commit, date,
	))

	return cmd
}

// Execute runs the root command. It returns the process exit code.
func Execute() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rootCmd := createRootCommand()
	rootCmd.AddCommand(
		createCoordinatorCmd(),
		createWorkerCmd(),
		createSequentialCmd(),
	)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", rootCmd.Name(), err)
		return 1
	}

	return 0
}
