package server

import (
	"log"
	"net"

	pb "github.com/fabiosussetto/hello/hello-sender/rpc"
	rpcserver "github.com/fabiosussetto/hello/hello-sender/rpcserver"
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

	db.AutoMigrate(&TargetHost{}, &ExecutionPlan{}, &ExecutionPlanHost{}, &RunStatus{})

	return db
}

// func setupRPC() (*grpc.ClientConn, pb.JobServiceClient) {
// 	conn, err := grpc.Dial("localhost:1234", grpc.WithInsecure())
// 	if err != nil {
// 		log.Fatalf("fail to dial: %v", err)
// 	}
// 	return conn, pb.NewJobServiceClient(conn)
// }

func setupRPCServer(db *gorm.DB) {
	log.Println("Starting gRPC server")

	lis, err := net.Listen("tcp", "localhost:1235")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterDispatcherServiceServer(grpcServer, &rpcserver.DispatcherServer{DB: db})

	grpcServer.Serve(lis)

	// TODO: use grpcServer.GracefulStop
}

func StartServer() {
	db := setupDB()
	defer db.Close()

	// rpcConn, rpcClient := setupRPC()
	// defer rpcConn.Close()

	go func() {
		setupRPCServer(db)
	}()

	router := gin.Default()

	// env := &Env{db: db, rpcClient: rpcClient}
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
		}

	}

	router.Run() // listen and serve on 0.0.0.0:8080
}
