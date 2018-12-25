package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"regexp"
	"strings"

	pb "github.com/fabiosussetto/hello/hello-sender/rpc"
	hw "github.com/fabiosussetto/hello/hello-worker/lib"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var numWorkers int

var workerPool *hw.WorkerPool

var rootCmd = &cobra.Command{
	Use:   "hello",
	Short: "Hello executes bash commands",
	Long:  `Work in progress`,
	Run: func(cmd *cobra.Command, args []string) {
		start()
	},
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&numWorkers, "workers", "w", 4, "number of workers (defaults to 4)")
}

type jobServiceServer struct {
	WorkerPool *hw.WorkerPool
}

func (s *jobServiceServer) ScheduleCommand(ctx context.Context, jobCmd *pb.JobCommand) (*pb.JobCommandResult, error) {
	myJobCmd := hw.JobCmd{Name: jobCmd.Name, Args: regexp.MustCompile("\\s+").Split(jobCmd.Args, -1)}

	job := s.WorkerPool.RunCmd(myJobCmd)

	return &pb.JobCommandResult{JobId: job.ID}, nil
}

func (s *jobServiceServer) QueryJobStatus(ctx context.Context, req *pb.JobStatusRequest) (*pb.JobStatus, error) {
	job := s.WorkerPool.GetJobByID(req.JobId)

	jobStatus := &pb.JobStatus{
		CommandName: job.CmdStatus.Cmd,
		StdOut:      strings.Join(job.CmdStatus.Stdout, "\n"),
	}

	return jobStatus, nil
}

func startServer(workerPool *hw.WorkerPool) {
	// lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	log.Infoln("Starting gRPC server")

	lis, err := net.Listen("tcp", "localhost:1234")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterJobServiceServer(grpcServer, &jobServiceServer{WorkerPool: workerPool})

	grpcServer.Serve(lis)
}

func start() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	workerPool = hw.NewWorkerPool(numWorkers)

	go func() {
		startServer(workerPool)
	}()

	shutdownChan := make(chan os.Signal)

	// signal.Notify(shutdownChan, syscall.SIGTERM, syscall.SIGINT)
	signal.Notify(shutdownChan, os.Interrupt)

	poolStatusChan, poolStatusForcedChan := workerPool.Start()

	select {
	case sig := <-shutdownChan:
		log.Infof("Caught signal %+v, gracefully stopping worker pool", sig)

		go func() {
			workerPool.Stop()
		}()

		select {
		case forceSig := <-shutdownChan:
			log.Warnf("Caught signal %+v, forcing pool shutdown", forceSig)
			workerPool.ForceStop()

		case <-poolStatusChan:
			log.Infof("Pool has been gracefully terminated")
			return

		case <-poolStatusForcedChan:
			log.Warnf("Pool has been forcefully terminated")
			return
		}
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
