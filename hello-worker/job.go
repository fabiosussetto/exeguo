package main

import (
	"github.com/go-cmd/cmd"
)

type Job struct {
	ID        int
	worker    *Worker
	cmd       *cmd.Cmd
	cmdStatus cmd.Status
}

func NewJob(ID int, command *cmd.Cmd) *Job {
	return &Job{
		ID:  ID,
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
