package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	pb "github.com/fabiosussetto/hello/hello-sender/rpc"
)

func sendCommand(client pb.JobServiceClient, jobCmd *pb.JobCommand) *pb.JobCommandResult {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Calling ScheduleCommand")

	jobResult, err := client.ScheduleCommand(ctx, jobCmd)

	if err != nil {
		log.Fatalf("%v.ScheduleCommand(_) = _, %v: ", client, err)
	}
	log.Println(jobResult)

	return jobResult
}

func queryJobStatus(client pb.JobServiceClient, jobID uint64) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Calling QueryJobStatus")

	jobStatus, err := client.QueryJobStatus(ctx, &pb.JobStatusRequest{JobId: jobID})

	if err != nil {
		log.Fatalf("%v.QueryJobStatus(_) = _, %v: ", client, err)
	}
	log.Println(jobStatus)
}

func Execute() {
	var host string

	jobCmd := pb.JobCommand{}

	var rootCmd = &cobra.Command{
		Use:   "hello",
		Short: "Hello executes bash commands",
		Long:  `Work in progress`,
		Run: func(cmd *cobra.Command, args []string) {

			conn, err := grpc.Dial(host, grpc.WithInsecure())
			if err != nil {
				log.Fatalf("fail to dial: %v", err)
			}
			defer conn.Close()
			client := pb.NewJobServiceClient(conn)

			for i := 0; i < 10; i++ {
				jobResult := sendCommand(client, &jobCmd)

				time.Sleep(2 * time.Second)

				queryJobStatus(client, jobResult.JobId)
			}

		},
	}

	rootCmd.PersistentFlags().StringVarP(&host, "worker-host", "H", "localhost:1234", "Worker host")
	rootCmd.PersistentFlags().StringVarP(&jobCmd.Name, "command", "c", "", "Command to execute")
	rootCmd.PersistentFlags().StringVarP(&jobCmd.Args, "args", "a", "", "Command arguments")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
