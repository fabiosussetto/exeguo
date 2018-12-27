package server

import (
	"context"
	"regexp"
	"strings"

	pb "github.com/fabiosussetto/hello/hello-sender/rpc"
	hw "github.com/fabiosussetto/hello/hello-worker/lib"
)

type JobServiceServer struct {
	WorkerPool *hw.WorkerPool
}

func (s *JobServiceServer) ScheduleCommand(ctx context.Context, jobCmd *pb.JobCommand) (*pb.JobCommandResult, error) {
	myJobCmd := hw.JobCmd{Name: jobCmd.Name, Args: regexp.MustCompile("\\s+").Split(jobCmd.Args, -1)}

	job := s.WorkerPool.RunCmd(myJobCmd)

	return &pb.JobCommandResult{JobId: job.ID}, nil
}

func (s *JobServiceServer) QueryJobStatus(ctx context.Context, req *pb.JobStatusRequest) (*pb.JobStatus, error) {
	job := s.WorkerPool.GetJobByID(req.JobId)

	jobStatus := &pb.JobStatus{
		CommandName: job.CmdStatus.Cmd,
		StdOut:      strings.Join(job.CmdStatus.Stdout, "\n"),
	}

	return jobStatus, nil
}
