package cmd

import (
	"fmt"
	"log"
	"net/rpc"
	"os"

	"github.com/spf13/cobra"

	hw "github.com/fabiosussetto/hello/hello-worker/lib"
	workerRpc "github.com/fabiosussetto/hello/hello-worker/rpc"
)

var rootCmd = &cobra.Command{
	Use:   "hello",
	Short: "Hello executes bash commands",
	Long:  `Work in progress`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.DialHTTP("tcp", "localhost:1234")
		if err != nil {
			log.Fatal("Connection error: ", err)
		}

		var reply workerRpc.ScheduleResult

		jobCmd := hw.JobCmd{
			Name: "/Users/fabio/go/src/github.com/fabiosussetto/hello/test_cmd",
			Args: []string{"5", "1"},
		}

		client.Call("JobsRPC.ScheduleCommand", jobCmd, &reply)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
