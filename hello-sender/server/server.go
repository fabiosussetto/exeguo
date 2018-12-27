package server

import (
	// "github.com/fabiosussetto/hello/hello-sender/server/endpoints"
	"log"

	pb "github.com/fabiosussetto/hello/hello-sender/rpc"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"google.golang.org/grpc"
)

func setupDB() *gorm.DB {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic("failed to connect database")
	}
	db.Exec("PRAGMA foreign_keys = ON;")
	db.LogMode(true)
	db.AutoMigrate(&Command{}, &CommandRun{})

	return db
}

func setupRPC() (*grpc.ClientConn, pb.JobServiceClient) {
	conn, err := grpc.Dial("localhost:1234", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	return conn, pb.NewJobServiceClient(conn)
}

func StartServer() {
	db := setupDB()
	defer db.Close()

	rpcConn, rpcClient := setupRPC()
	defer rpcConn.Close()

	// Migrate the schema

	router := gin.Default()

	env := &Env{db: db, rpcClient: rpcClient}

	v1 := router.Group("/v1")
	{
		commandsR := v1.Group("/commands")
		{
			commandsR.GET("/", env.ListCommandsEndpoint)
			commandsR.POST("/", env.CreateCommandEndpoint)
			commandsR.GET("/:id", env.CommandDetailEndpoint)
			commandsR.DELETE("/:id", env.DeleteCommandEndpoint)
		}

		commandRunR := v1.Group("/command-runs")
		{
			commandRunR.POST("/", env.CreateCommandRunEndpoint)
		}

	}

	router.Run() // listen and serve on 0.0.0.0:8080
}
