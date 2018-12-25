package rpc

import (
	hw "github.com/fabiosussetto/hello/hello-worker/lib"
)

type ScheduleResult struct{}

type JobsRPC struct {
	WorkerPool *hw.WorkerPool
}

// var WorkerPool *hw.WorkerPool

func (t *JobsRPC) ScheduleCommand(jobCmd hw.JobCmd, reply *ScheduleResult) error {
	t.WorkerPool.RunCmd(jobCmd)

	*reply = ScheduleResult{}

	return nil
}
