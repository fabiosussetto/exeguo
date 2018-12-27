package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"

	pb "github.com/fabiosussetto/hello/hello-sender/rpc"
	"github.com/gin-gonic/gin"
)

type Env struct {
	db        *gorm.DB
	rpcClient pb.JobServiceClient
}

func (e *Env) ListCommandsEndpoint(c *gin.Context) {
	var commands []Command
	e.db.Preload("Runs").Find(&commands)

	c.JSON(http.StatusOK, commands)
}

func (e *Env) CreateCommandEndpoint(c *gin.Context) {
	var command Command
	if err := c.ShouldBindJSON(&command); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	e.db.Create(&command)

	c.JSON(http.StatusCreated, command)
}

func (e *Env) CommandDetailEndpoint(c *gin.Context) {
	var command Command

	if e.db.First(&command, c.Param("id")).RecordNotFound() {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}

	c.JSON(http.StatusOK, command)
}

func (e *Env) DeleteCommandEndpoint(c *gin.Context) {
	var command Command

	if e.db.First(&command, c.Param("id")).RecordNotFound() {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}

	e.db.Delete(&command)
	c.JSON(http.StatusOK, nil)
}

func (e *Env) CreateCommandRunEndpoint(c *gin.Context) {
	var commandRun CommandRun
	if err := c.ShouldBindJSON(&commandRun); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var command Command

	if e.db.First(&command, commandRun.CommandID).RecordNotFound() {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}

	if err := e.db.Create(&commandRun).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		jobCmd := &pb.JobCommand{Name: command.CmdName, Args: command.Args}
		_, err := e.rpcClient.ScheduleCommand(ctx, jobCmd)

		if err != nil {
			log.Printf("%v.ScheduleCommand(_) = _, %v: ", e.rpcClient, err)
		}
	}()

	c.JSON(http.StatusCreated, commandRun)
}
