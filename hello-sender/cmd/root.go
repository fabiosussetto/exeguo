package cmd

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
	"regexp"

	"github.com/spf13/cobra"

	hw "github.com/fabiosussetto/hello/hello-worker/lib"
	workerRpc "github.com/fabiosussetto/hello/hello-worker/rpc"
)

func connectToWorker(host string) *rpc.Client {
	client, err := rpc.DialHTTP("tcp", host)
	if err != nil {
		log.Fatal("Connection error: ", err)
	}

	return client
}

func sendCommandToExec(client *rpc.Client, jobCmd *hw.JobCmd) {
	var reply workerRpc.ScheduleResult
	client.Call("JobsRPC.ScheduleCommand", jobCmd, &reply)
}

func Execute() {
	var host string
	var rawArgs string

	jobCmd := hw.JobCmd{}

	var rootCmd = &cobra.Command{
		Use:   "hello",
		Short: "Hello executes bash commands",
		Long:  `Work in progress`,
		Run: func(cmd *cobra.Command, args []string) {
			client := connectToWorker(host)

			jobCmd.Args = regexp.MustCompile("\\s+").Split(rawArgs, -1)
			sendCommandToExec(client, &jobCmd)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&host, "worker-host", "H", "localhost:1234", "Worker host")
	rootCmd.PersistentFlags().StringVarP(&jobCmd.Name, "command", "c", "", "Command to execute")
	rootCmd.PersistentFlags().StringVarP(&rawArgs, "args", "a", "", "Command arguments")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
