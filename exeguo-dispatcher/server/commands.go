package server

import (
	"context"
	"io"
	"log"
	"sync"
	"time"

	"github.com/jinzhu/gorm"

	pb "github.com/fabiosussetto/exeguo/exeguo-dispatcher/rpc"
	"google.golang.org/grpc"
)

func RunExecutionPlan(db *gorm.DB, execPlanRun *ExecutionPlanRun) {
	var wg sync.WaitGroup

	execPlan := execPlanRun.ExecutionPlan

	log.Printf("Starting execution plan: %d", execPlan.ID)

	wg.Add(len(execPlan.PlanHosts))

	for _, planHost := range execPlan.PlanHosts {
		go func(planHost ExecutionPlanHost) {
			defer wg.Done()

			runStatus := RunStatus{}

			db.FirstOrCreate(&runStatus, RunStatus{
				ExecutionPlanRunID:  execPlanRun.ID,
				ExecutionPlanHostID: planHost.TargetHostID,
			})

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
					log.Println("Received EOF")
					return
				}
				if err != nil {
					log.Printf("Failed to receive message : %v", err)
					return
				}

				log.Printf("Got cmd output: %#v", in)

				startedAt := time.Unix(0, in.StartTs)
				completedAt := time.Unix(0, in.StopTs)

				db.Model(&runStatus).Updates(RunStatus{
					Cmd:         in.Cmd,
					PID:         in.PID,
					Complete:    in.Complete,
					Stdout:      in.Stdout,
					Stderr:      in.Stderr,
					StartedAt:   &startedAt,
					CompletedAt: &completedAt,
					Runtime:     in.Runtime,
					ExitCode:    in.Exit,
				})
			}
		}(planHost)
	}
	wg.Wait()
}
