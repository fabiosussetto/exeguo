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

func RunExecutionPlan(db *gorm.DB, execPlanRun *ExecutionPlanRun) {
	var wg sync.WaitGroup

	execPlan := execPlanRun.ExecutionPlan

	log.Printf("Starting execution plan: %d", execPlan.ID)

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

	wg.Add(len(execPlan.PlanHosts))

	for _, planHost := range execPlan.PlanHosts {
		go func(planHost ExecutionPlanHost) {
			defer wg.Done()

			runStatus := RunStatus{}

			db.FirstOrCreate(&runStatus, RunStatus{
				ExecutionPlanRunID:  execPlanRun.ID,
				ExecutionPlanHostID: planHost.TargetHostID,
			})

			log.Printf("Connecting to client: %s (%s)", planHost.TargetHost.Name, planHost.TargetHost.Address)

			// Create a certificate pool from the certificate authority
			certPool := x509.NewCertPool()

			caCert, err := security.ParseCertFromPEM([]byte(caCertConfig.Value))

			dispatcherCert, err := tls.X509KeyPair([]byte(dispatcherCertConfig.Value), []byte(dispatcherKeyConfig.Value))

			// caCert, err := tls.X509KeyPair([]byte(caCertConfig.Value), []byte(caKeyConfig.Value))

			if err != nil {
				log.Printf("could not read ca certificate: %s", err)
				return
			}

			// Append the certificates from the CA
			certPool.AddCert(caCert)

			creds := credentials.NewTLS(&tls.Config{
				ServerName:   planHost.TargetHost.Address, // TODO: decide if ip or hostname
				Certificates: []tls.Certificate{dispatcherCert},
				RootCAs:      certPool,
			})

			// conn, err := grpc.Dial(planHost.TargetHost.Address, grpc.WithInsecure())
			conn, err := grpc.Dial(planHost.TargetHost.Address, grpc.WithTransportCredentials(creds))

			if err != nil {
				log.Printf("fail to dial: %v", err)
				return
			}

			rpcClient := pb.NewJobServiceClient(conn)
			jobCmd := &pb.JobCommand{Name: execPlan.CmdName, Args: execPlan.Args}
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

				startedAt := time.Unix(0, in.StartTs)
				completedAt := time.Unix(0, in.StopTs)

				db.Model(&runStatus).Updates(RunStatus{
					Cmd:         in.Cmd,
					PID:         in.PID,
					Complete:    in.Complete,
					Stdout:      in.Stdout,
					Stderr:      in.Stderr,
					StartedAt:   &startedAt,
					CompletedAt: &completedAt,
					Runtime:     in.Runtime,
					ExitCode:    in.Exit,
				})
			}
		}(planHost)
	}
	wg.Wait()
}
