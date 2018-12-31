package server

import (
	"context"
	"io"
	"log"
	"sync"

	pb "github.com/fabiosussetto/hello/hello-sender/rpc"
	"google.golang.org/grpc"
)

func RunExecutionPlan(execPlan *ExecutionPlan) {
	var wg sync.WaitGroup
	wg.Add(len(execPlan.PlanHosts))

	for _, planHost := range execPlan.PlanHosts {
		go func(planHost ExecutionPlanHost) {
			defer wg.Done()

			log.Printf("Connecting to client: %v", planHost)
			conn, err := grpc.Dial(planHost.TargetHost.Address, grpc.WithInsecure())

			if err != nil {
				log.Printf("fail to dial: %v", err)
				return
			}

			rpcClient := pb.NewJobServiceClient(conn)
			jobCmd := &pb.JobCommand{Name: execPlan.CmdName, Args: execPlan.Args}
			stream, err := rpcClient.ScheduleCommand(context.Background(), jobCmd)

			if err != nil {
				log.Printf("fail to create rpc stream: %v", err)
				return
			}

			for {
				in, err := stream.Recv()
				if err == io.EOF {
					// read done.
					// close(waitc)
					log.Println("Received EOF")
					return
				}
				if err != nil {
					log.Printf("Failed to receive message : %v", err)
					return
				}

				log.Printf("Got cmd output: %s", in.StdinLine)
			}
		}(planHost)
	}
	wg.Wait()
}
