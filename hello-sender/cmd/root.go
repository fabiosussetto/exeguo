package cmd

import (
	"context"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	pb "github.com/fabiosussetto/hello/hello-sender/rpc"
)

type jobServiceServer struct {
	// savedFeatures []*pb.Feature // read-only after initialized

	// mu         sync.Mutex // protects routeNotes
	// routeNotes map[string][]*pb.RouteNote

}

// func (s *jobServiceServer) ScheduleCommand(context.Context, *pb.JobCommand) (*pb.JobCommandResult, error) {

// }

func connectToWorker(host string) *rpc.Client {
	client, err := rpc.DialHTTP("tcp", host)
	if err != nil {
		log.Fatal("Connection error: ", err)
	}

	return client
}

func sendCommand(client pb.JobServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Calling ScheduleCommand")

	jobCmd := &pb.JobCommand{Name: "ls", Args: "-la /"}

	jobResult, err := client.ScheduleCommand(ctx, jobCmd)

	if err != nil {
		log.Fatalf("%v.ScheduleCommand(_) = _, %v: ", client, err)
	}
	log.Println(jobResult)
}

// func sendCommandToExec(client *rpc.Client, jobCmd *hw.JobCmd) {
// 	var reply workerRpc.ScheduleResult
// 	client.Call("JobsRPC.ScheduleCommand", jobCmd, &reply)
// }

func Execute() {
	var host string
	// var rawArgs string

	// jobCmd := hw.JobCmd{}

	var rootCmd = &cobra.Command{
		Use:   "hello",
		Short: "Hello executes bash commands",
		Long:  `Work in progress`,
		Run: func(cmd *cobra.Command, args []string) {
			// client := connectToWorker(host)

			// jobCmd.Args = regexp.MustCompile("\\s+").Split(rawArgs, -1)
			// sendCommandToExec(client, &jobCmd)

			conn, err := grpc.Dial(host, grpc.WithInsecure())
			if err != nil {
				log.Fatalf("fail to dial: %v", err)
			}
			defer conn.Close()
			client := pb.NewJobServiceClient(conn)

			sendCommand(client)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&host, "worker-host", "H", "localhost:1234", "Worker host")
	// rootCmd.PersistentFlags().StringVarP(&jobCmd.Name, "command", "c", "", "Command to execute")
	// rootCmd.PersistentFlags().StringVarP(&rawArgs, "args", "a", "", "Command arguments")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
