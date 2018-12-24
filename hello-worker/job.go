package main

import (
	"sync/atomic"

	"github.com/go-cmd/cmd"
)

var jobCounter uint64

type Job struct {
	ID        uint64
	worker    *Worker
	cmd       *cmd.Cmd
	cmdStatus cmd.Status
}

func NewJob(command *cmd.Cmd) *Job {
	atomic.AddUint64(&jobCounter, 1)

	return &Job{
		ID:  atomic.LoadUint64(&jobCounter),
		cmd: command,
		cmdStatus: cmd.Status{
			Cmd:      "",
			PID:      0,
			Complete: false,
			Exit:     -1,
			Error:    nil,
			Runtime:  0,
		},
	}
}
