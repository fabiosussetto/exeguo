package server

type HostIDExecutionPlanHost struct {
	TargetHostID uint `json:"targetHostId"`
}

type HostIDExecutionPlan struct {
	CmdName string `json:"cmdName" binding:"required"`
	Args    string `json:"args" binding:"required"`

	PlanHosts []HostIDExecutionPlanHost `json:"planHosts" binding:"required,dive"`
}
