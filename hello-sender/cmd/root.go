package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Execute() {
	// var host string

	// jobCmd := pb.JobCommand{}

	var rootCmd = &cobra.Command{
		Use:   "hello",
		Short: "Hello executes bash commands",
		Long:  `Work in progress`,
	}

	rootCmd.AddCommand(ServerCmd)
	// cmdEcho.AddCommand(cmdTimes)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
