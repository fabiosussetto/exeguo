package server

type HostIDExecutionPlanHost struct {
	TargetHostID uint `json:"targetHostId"`
}

type HostIDExecutionPlan struct {
	Name    string `json:"name" binding:"required"`
	CmdName string `json:"cmdName" binding:"required"`
	Args    string `json:"args"`

	PlanHosts []HostIDExecutionPlanHost `json:"planHosts" binding:"required,dive"`
}
