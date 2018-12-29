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
	TargetHostID    uint       `sql:"type:uint REFERENCES target_hosts(id) ON DELETE CASCADE" json:"targetHostId" binding:"required"`
	TargetHost      TargetHost `json:"targetHost" binding:"required,dive"`
}

// type ExecutionPlanLog struct {
// 	ID        uint      `json:"id"`
// 	ExecutionPlanID uint       `sql:"type:uint REFERENCES execution_plans(id) ON DELETE CASCADE" json:"executionPlanId"`

// 	CreatedAt time.Time `json:"createdAt"`
// }

type RunStatus struct {
	ID                  uint `json:"id"`
	ExecutionPlanHostID uint `sql:"type:uint REFERENCES execution_plan_hosts(id) ON DELETE CASCADE" json:"executionPlanHostId" binding:"required"`

	Stdout      string
	Stderr      string
	StartedAt   *time.Time
	CompletedAt *time.Time
	Runtime     float64
	ExitCode    int
}
