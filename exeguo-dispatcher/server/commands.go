package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"log"
	"sync"
	"time"

	"github.com/jinzhu/gorm"

	pb "github.com/fabiosussetto/exeguo/exeguo-dispatcher/rpc"
	"github.com/fabiosussetto/exeguo/security"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func CreateHostGrpcClient(db *gorm.DB, targetHost *TargetHost) (pb.JobServiceClient, *grpc.ClientConn, error) {
	var (
		caCertConfig         Config
		caKeyConfig          Config
		dispatcherCertConfig Config
		dispatcherKeyConfig  Config
	)

	if db.Where(&Config{Key: "tls.ca_cert"}).First(&caCertConfig).RecordNotFound() {
		log.Fatalln("Cannot load CA cert from DB config")
	}

	if db.Where(&Config{Key: "tls.ca_key"}).First(&caKeyConfig).RecordNotFound() {
		log.Fatalln("Cannot load CA key from DB config")
	}

	if db.Where(&Config{Key: "tls.dispatcher_cert"}).First(&dispatcherCertConfig).RecordNotFound() {
		log.Fatalln("Cannot load CA cert from DB config")
	}

	if db.Where(&Config{Key: "tls.dispatcher_key"}).First(&dispatcherKeyConfig).RecordNotFound() {
		log.Fatalln("Cannot load CA key from DB config")
	}

	caCert, err := security.ParseCertFromPEM([]byte(caCertConfig.Value))

	dispatcherCert, err := tls.X509KeyPair([]byte(dispatcherCertConfig.Value), []byte(dispatcherKeyConfig.Value))

	if err != nil {
		log.Printf("could not read ca certificate: %s", err)
		return nil, nil, err
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(caCert)

	creds := credentials.NewTLS(&tls.Config{
		ServerName:   targetHost.Address, // TODO: decide if ip or hostname
		Certificates: []tls.Certificate{dispatcherCert},
		RootCAs:      certPool,
	})

	// conn, err := grpc.Dial(planHost.TargetHost.Address, grpc.WithInsecure())
	conn, err := grpc.Dial(targetHost.Address, grpc.WithTransportCredentials(creds))

	if err != nil {
		log.Printf("fail to dial: %v", err)
		return nil, nil, err
	}

	rpcClient := pb.NewJobServiceClient(conn)
	return rpcClient, conn, nil
}

func RunExecutionPlan(db *gorm.DB, execPlanRun *ExecutionPlanRun) {
	var wg sync.WaitGroup

	execPlan := execPlanRun.ExecutionPlan

	log.Printf("Starting execution plan: %d - Plan run %d", execPlan.ID, execPlanRun.ID)

	wg.Add(len(execPlan.PlanHosts))

	for _, planHost := range execPlan.PlanHosts {
		go func(planHost ExecutionPlanHost) {
			defer wg.Done()

			runStatus := RunStatus{}

			err := db.FirstOrCreate(&runStatus, RunStatus{
				ExecutionPlanRunID:  execPlanRun.ID,
				ExecutionPlanHostID: planHost.ID,
			}).Error

			if err != nil {
				log.Printf("fail to find/create model: %v - %s | %s", err, execPlanRun.ID, planHost.TargetHostID)
			}

			log.Printf("Connecting to client: %s (%s)", planHost.TargetHost.Name, planHost.TargetHost.Address)

			rpcClient, conn, err := CreateHostGrpcClient(db, &planHost.TargetHost)

			if err != nil {
				log.Printf("fail to create rpc client: %v", err)
				return
			}

			defer conn.Close()

			jobCmd := &pb.JobCommand{JobId: uint64(execPlanRun.ID), Name: execPlan.CmdName, Args: execPlan.Args}
			stream, err := rpcClient.ScheduleCommand(context.Background(), jobCmd)

			if err != nil {
				log.Printf("fail to create rpc stream: %v", err)
				return
			}

			for {
				in, err := stream.Recv()
				if err == io.EOF {
					log.Println("Received EOF")
					return
				}
				if err != nil {
					log.Printf("Failed to receive message : %v", err)
					return
				}

				log.Printf("Got cmd output: %#v", in)

				runStatus = RunStatus{
					Cmd:      in.Cmd,
					PID:      in.PID,
					Complete: in.Complete,
					Stdout:   in.Stdout,
					Stderr:   in.Stderr,
					Runtime:  in.Runtime,
					ExitCode: in.Exit,
				}

				if in.StartTs != 0 {
					startedAt := time.Unix(0, in.StartTs)
					runStatus.StartedAt = &startedAt
				}

				if in.StopTs != 0 {
					stoppedAt := time.Unix(0, in.StartTs)
					runStatus.CompletedAt = &stoppedAt
				}

				db.Model(&runStatus).Updates(runStatus)
			}
		}(planHost)
	}
	wg.Wait()
}

func StopExecutionPlan(db *gorm.DB, execPlanRun *ExecutionPlanRun) {
	var wg sync.WaitGroup

	execPlan := execPlanRun.ExecutionPlan

	log.Printf("Stopping execution plan: %d", execPlan.ID)

	wg.Add(len(execPlanRun.RunStatuses))

	for _, runStatus := range execPlanRun.RunStatuses {
		go func(runStatus RunStatus) {
			defer wg.Done()

			planHost := runStatus.ExecutionPlanHost
			targetHost := planHost.TargetHost

			log.Printf("Connecting to client: %s (%s)", targetHost.Name, targetHost.Address)

			rpcClient, conn, err := CreateHostGrpcClient(db, &targetHost)

			if err != nil {
				log.Printf("fail to create rpc client: %v", err)
				return
			}

			defer conn.Close()

			hostDeadline := time.Now().Add(time.Duration(15) * time.Second)
			ctx, _ := context.WithDeadline(context.Background(), hostDeadline)

			_, err = rpcClient.StopCommand(ctx, &pb.StopCommandRequest{JobId: uint64(execPlanRun.ID)})

			if err != nil {
				log.Printf("fail to call StopCommand: %v", err)
				return
			}

		}(runStatus)
	}
	wg.Wait()
}
