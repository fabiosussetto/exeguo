package server

import (
	// "github.com/fabiosussetto/hello/hello-sender/server/endpoints"

	"io"
	"log"
	"net"

	pb "github.com/fabiosussetto/hello/hello-sender/rpc"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"google.golang.org/grpc"
)

type dispatcherServer struct {
}

func (s *dispatcherServer) StreamJobStatus(stream pb.DispatcherService_StreamJobStatusServer) error {
	for {
		jobStatusUpdate, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.JobStatusUpdatesResult{})
		}
		if err != nil {
			return err
		}
		log.Printf("Status stream: %s\n", jobStatusUpdate.StdinLine)
	}
}

func setupDB() *gorm.DB {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic("failed to connect database")
	}
	db.Exec("PRAGMA foreign_keys = ON;")
	// db.LogMode(true)

	// db.AutoMigrate(&Command{}, &CommandRun{})

	return db
}

func setupRPC() (*grpc.ClientConn, pb.JobServiceClient) {
	conn, err := grpc.Dial("localhost:1234", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	return conn, pb.NewJobServiceClient(conn)
}

func setupRPCServer() {
	log.Println("Starting gRPC server")

	lis, err := net.Listen("tcp", "localhost:1235")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterDispatcherServiceServer(grpcServer, &dispatcherServer{})

	grpcServer.Serve(lis)
}

func StartServer() {
	db := setupDB()
	defer db.Close()

	rpcConn, rpcClient := setupRPC()
	defer rpcConn.Close()

	go func() {
		setupRPCServer()
	}()

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
