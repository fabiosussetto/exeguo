package server

import (
	"context"
	"log"

	pb "github.com/fabiosussetto/hello/hello-sender/rpc"
	"google.golang.org/grpc"
)

func runExecutionPlan(execPlan *ExecutionPlan) {

	for _, planHost := range execPlan.PlanHosts {
		conn, err := grpc.Dial(planHost.TargetHost.Address, grpc.WithInsecure())

		if err != nil {
			log.Fatalf("fail to dial: %v", err)
		}

		rpcClient := pb.NewJobServiceClient(conn)

		jobCmd := &pb.JobCommand{Name: execPlan.CmdName, Args: execPlan.Args}
		_, rpcErr := rpcClient.ScheduleCommand(context.Background(), jobCmd)

		if rpcErr != nil {
			log.Printf("%v.ScheduleCommand(_) = _, %v: ", rpcClient, err)
		}
	}

	// conn, err := grpc.Dial("localhost:1234", grpc.WithInsecure())
	// if err != nil {
	// 	log.Fatalf("fail to dial: %v", err)
	// }

	// // ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// // defer cancel()

	// jobCmd := &pb.JobCommand{Name: command.CmdName, Args: command.Args}
	// _, err := rpcClient.ScheduleCommand(context.Background(), jobCmd)

	// if err != nil {
	// 	log.Printf("%v.ScheduleCommand(_) = _, %v: ", rpcClient, err)
	// }
}
