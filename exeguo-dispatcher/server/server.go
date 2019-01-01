package server

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func setupDB() *gorm.DB {
	db, err := gorm.Open("sqlite3", "test.sqlite")
	if err != nil {
		panic("failed to connect database")
	}

	db.Exec("PRAGMA foreign_keys = ON;")
	db.LogMode(true)

	db.AutoMigrate(&TargetHost{}, &ExecutionPlan{}, &ExecutionPlanHost{}, &ExecutionPlanRun{}, &RunStatus{})

	return db
}

func StartServer() {
	db := setupDB()
	defer db.Close()

	router := gin.Default()

	env := &Env{db: db}

	v1 := router.Group("/v1")
	{
		commandsR := v1.Group("/hosts")
		{
			commandsR.GET("/", env.HostListEndpoint)
			commandsR.POST("/", env.HostCreateEndpoint)
			// commandsR.GET("/:id", env.CommandDetailEndpoint)
			commandsR.PUT("/:id", env.HostUpdateEndpoint)
			commandsR.DELETE("/:id", env.HostDeleteEndpoint)
		}

		commandRunR := v1.Group("/exec-plans")
		{
			commandRunR.POST("/", env.ExecutionPlanCreateEndpoint)
			commandRunR.GET("/:id", env.ExecutionPlanDetailEndpoint)
		}

		execPlanRunRoute := v1.Group("/exec-plan-runs")
		{
			execPlanRunRoute.POST("/", env.ExecutionPlanRunCreateEndpoint)
			execPlanRunRoute.GET("/:id", env.ExecutionPlanRunDetailEndpoint)
		}

	}

	router.Run() // listen and serve on 0.0.0.0:8080
}