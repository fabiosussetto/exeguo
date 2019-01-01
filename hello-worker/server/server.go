package server

import (
	"log"
	"regexp"
	"strings"

	pb "github.com/fabiosussetto/hello/hello-sender/rpc"
	hw "github.com/fabiosussetto/hello/hello-worker/lib"
)

type JobServiceServer struct {
	WorkerPool *hw.WorkerPool
}

func (s *JobServiceServer) ScheduleCommand(command *pb.JobCommand, stream pb.JobService_ScheduleCommandServer) error {
	job := s.WorkerPool.RunCmd(hw.JobCmd{Name: command.Name, Args: regexp.MustCompile("\\s+").Split(command.Args, -1)})

	for jobStatus := range job.StdoutChan {
		statusUpdate := &pb.JobStatusUpdate{
			Cmd:      jobStatus.Cmd,
			PID:      int64(jobStatus.PID),
			Complete: jobStatus.Complete,
			Exit:     int64(jobStatus.Exit),
			StartTs:  jobStatus.StartTs,
			StopTs:   jobStatus.StopTs,
			Runtime:  float32(jobStatus.Runtime),
			Stdout:   strings.Join(jobStatus.Stdout, "\n"),
			Stderr:   strings.Join(jobStatus.Stderr, "\n"),
		}

		if err := stream.Send(statusUpdate); err != nil {
			// return err
			log.Fatalf("Failed to send a status update: %v", err)
		}

	}

	return nil
}
