package server

import (
	"log"
	"net/http"

	"github.com/jinzhu/gorm"

	pb "github.com/fabiosussetto/hello/hello-sender/rpc"
	"github.com/gin-gonic/gin"
)

type Env struct {
	db        *gorm.DB
	rpcClient pb.JobServiceClient
}

func (e *Env) HostListEndpoint(c *gin.Context) {
	var hosts []TargetHost
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

func (e *Env) ExecutionPlanCreateEndpoint(c *gin.Context) {
	var execPlan ExecutionPlan

	if err := c.ShouldBindJSON(&execPlan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := e.db.Save(&execPlan).Error; err != nil {
		log.Printf("Error creating exec plan: %s", err)
	}

	go RunExecutionPlan(&execPlan)

	c.JSON(http.StatusCreated, execPlan)
}
