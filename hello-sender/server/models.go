package server

import (
	"time"
)

type TargetHost struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address" binding:"required"`
}

type ExecutionPlan struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	CmdName string `json:"cmdName" binding:"required"`
	Args    string `json:"args" binding:"required"`

	PlanHosts []ExecutionPlanHost `json:"planHosts" binding:"required,dive"`
}

type ExecutionPlanHost struct {
	ID              uint       `json:"id"`
	ExecutionPlanID uint       `sql:"type:uint REFERENCES execution_plans(id) ON DELETE CASCADE" json:"executionPlanId"`
	TargetHostID    uint       `sql:"type:uint REFERENCES target_hosts(id) ON DELETE CASCADE" json:"targetHostId"`
	TargetHost      TargetHost `json:"targetHost" binding:"required,dive"`
}

type ExecutionPlanRun struct {
	ID              uint          `json:"id"`
	ExecutionPlanID uint          `sql:"type:uint REFERENCES execution_plans(id) ON DELETE CASCADE" json:"executionPlanId" binding:"required"`
	ExecutionPlan   ExecutionPlan `json:"executionPlan" binding:"-"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	RunStatuses []RunStatus `json:"runStatuses"`
}

type RunStatus struct {
	ID                  uint `json:"id"`
	ExecutionPlanRunID  uint `sql:"type:uint REFERENCES execution_plan_runs(id) ON DELETE CASCADE" json:"executionPlanRunId"`
	ExecutionPlanHostID uint `sql:"type:uint REFERENCES execution_plan_hosts(id) ON DELETE CASCADE" json:"executionPlanHostId"`
	ExecutionPlanHost   ExecutionPlanHost

	Cmd         string
	PID         int64
	Complete    bool
	Stdout      string
	Stderr      string
	StartedAt   *time.Time
	CompletedAt *time.Time
	Runtime     float32
	ExitCode    int64
}
