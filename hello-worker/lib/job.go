package lib

import (
	"github.com/go-cmd/cmd"
)

type Job struct {
	ID        uint64
	CmdStatus cmd.Status

	worker *Worker
	cmd    *cmd.Cmd
}

func NewJob(ID uint64, command *cmd.Cmd) *Job {
	return &Job{
		ID:  ID,
		cmd: command,
		CmdStatus: cmd.Status{
			Cmd:      "",
			PID:      0,
			Complete: false,
			Exit:     -1,
			Error:    nil,
			Runtime:  0,
		},
	}
}
