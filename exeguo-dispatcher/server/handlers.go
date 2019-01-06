package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"

	pb "github.com/fabiosussetto/exeguo/exeguo-dispatcher/rpc"
	"github.com/gin-gonic/gin"
)

type Env struct {
	db        *gorm.DB
	rpcClient pb.JobServiceClient
}

func (e *Env) HostListEndpoint(c *gin.Context) {
	var hosts []TargetHost

	e.db.Find(&hosts)

	c.JSON(http.StatusOK, hosts)
}

func (e *Env) HostCreateEndpoint(c *gin.Context) {
	var host TargetHost
	if err := c.ShouldBindJSON(&host); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	e.db.Create(&host)
	c.JSON(http.StatusCreated, host)
}

func (e *Env) HostUpdateEndpoint(c *gin.Context) {
	var host TargetHost

	if e.db.First(&host, c.Param("id")).RecordNotFound() {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}

	c.BindJSON(&host)

	e.db.Save(&host)
	c.JSON(http.StatusOK, host)
}

func (e *Env) HostDeleteEndpoint(c *gin.Context) {
	var host TargetHost

	if e.db.First(&host, c.Param("id")).RecordNotFound() {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}

	e.db.Delete(&host)
	c.JSON(http.StatusOK, nil)
}

////

func (e *Env) ExecutionPlanCreateEndpoint(c *gin.Context) {
	var execPlanSchema HostIDExecutionPlan

	if err := c.ShouldBindJSON(&execPlanSchema); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var hosts []TargetHost
	var hostIds []uint
	var planHosts []ExecutionPlanHost

	for _, planHost := range execPlanSchema.PlanHosts {
		hostIds = append(hostIds, planHost.TargetHostID)
	}

	e.db.Where("id in (?)", hostIds).Find(&hosts)

	for _, host := range hosts {
		planHosts = append(planHosts, ExecutionPlanHost{TargetHostID: host.ID})
	}

	execPlan := &ExecutionPlan{
		CmdName:   execPlanSchema.CmdName,
		Args:      execPlanSchema.Args,
		PlanHosts: planHosts,
	}

	if err := e.db.Save(&execPlan).Error; err != nil {
		log.Printf("Error creating exec plan: %s", err)
	}

	var savedExecPlan ExecutionPlan

	e.db.Preload("PlanHosts.TargetHost").First(&savedExecPlan, execPlan.ID)

	c.JSON(http.StatusCreated, savedExecPlan)
}

////

func (e *Env) ExecutionPlanDetailEndpoint(c *gin.Context) {
	var execPlan ExecutionPlan

	q := e.db.Preload("PlanHosts.TargetHost")

	if q.First(&execPlan, c.Param("id")).RecordNotFound() {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}

	c.JSON(http.StatusOK, execPlan)
}

///

func (e *Env) ExecutionPlanRunCreateEndpoint(c *gin.Context) {
	var execPlanRun ExecutionPlanRun

	if err := c.ShouldBindJSON(&execPlanRun); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := e.db.Save(&execPlanRun).Error; err != nil {
		log.Printf("Error creating exec plan run: %s", err)
	}

	q := e.db.Preload("ExecutionPlan.PlanHosts.TargetHost")

	var savedExecPlanRun ExecutionPlanRun

	q.First(&savedExecPlanRun, execPlanRun.ID)

	go RunExecutionPlan(e.db, &savedExecPlanRun)

	c.JSON(http.StatusCreated, savedExecPlanRun)
}

func (e *Env) ExecutionPlanRunDetailEndpoint(c *gin.Context) {
	var execPlanRun ExecutionPlanRun

	q := e.db.Preload("ExecutionPlan").Preload("RunStatuses.ExecutionPlanHost.TargetHost")

	if q.First(&execPlanRun, c.Param("id")).RecordNotFound() {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}

	c.JSON(http.StatusOK, execPlanRun)
}

type HostStatus struct {
	HostId uint `json:"hostId" binding:"required"`
}

type HostStatusResponse struct {
	Connected bool
	Error     string
}

func (e *Env) HostStatusEndpoint(c *gin.Context) {
	var hostStatusReq HostStatus

	if err := c.ShouldBindJSON(&hostStatusReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var host TargetHost

	if e.db.First(&host, hostStatusReq.HostId).RecordNotFound() {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}

	log.Printf("Sending heartbeat request to host id: %s", host.ID)

	rpcClient, conn, err := CreateHostGrpcClient(e.db, &host)
	defer conn.Close()

	if err != nil {
		log.Printf("Error connecting to hearbeat handpoint: %s", err)
		c.JSON(http.StatusOK, &HostStatusResponse{Connected: false})
	}

	hostDeadline := time.Now().Add(time.Duration(15) * time.Second)
	ctx, _ := context.WithDeadline(context.Background(), hostDeadline)

	_, err = rpcClient.Heartbeat(ctx, &pb.HeartbeatPing{})

	if err != nil {
		c.JSON(http.StatusOK, &HostStatusResponse{Connected: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, &HostStatusResponse{Connected: true})
}

type StopPlanRequest struct {
	ExecutionPlanRunId uint `json:"executionPlanRunId" binding:"required"`
}

type StopPlanResponse struct {
	OK bool
}

func (e *Env) CreateStopPlanRequest(c *gin.Context) {
	var (
		stopReq     StopPlanRequest
		execPlanRun ExecutionPlanRun
	)

	if err := c.ShouldBindJSON(&stopReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if e.db.Preload("RunStatuses.ExecutionPlanHost.TargetHost").First(&execPlanRun, stopReq.ExecutionPlanRunId).RecordNotFound() {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}

	StopExecutionPlan(e.db, &execPlanRun)
	c.JSON(http.StatusOK, &StopPlanResponse{OK: true})
}
