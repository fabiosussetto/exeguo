package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/jinzhu/gorm"

	pb "github.com/fabiosussetto/exeguo/exeguo-dispatcher/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func RunExecutionPlan(db *gorm.DB, execPlanRun *ExecutionPlanRun) {
	var wg sync.WaitGroup

	execPlan := execPlanRun.ExecutionPlan

	log.Printf("Starting execution plan: %d", execPlan.ID)

	wg.Add(len(execPlan.PlanHosts))

	for _, planHost := range execPlan.PlanHosts {
		go func(planHost ExecutionPlanHost) {
			defer wg.Done()

			runStatus := RunStatus{}

			db.FirstOrCreate(&runStatus, RunStatus{
				ExecutionPlanRunID:  execPlanRun.ID,
				ExecutionPlanHostID: planHost.TargetHostID,
			})

			log.Printf("Connecting to client: %v", planHost)

			// Load the client certificates from disk
			certificate, err := tls.LoadX509KeyPair("../certs/client_cert.pem", "../certs/client_key.pem")
			if err != nil {
				log.Printf("could not load client key pair: %s\n", err)
				return
			}

			// Create a certificate pool from the certificate authority
			certPool := x509.NewCertPool()
			ca, err := ioutil.ReadFile("../certs/ca_cert.pem")
			if err != nil {
				log.Printf("could not read ca certificate: %s", err)
				return
			}

			// Append the certificates from the CA
			if ok := certPool.AppendCertsFromPEM(ca); !ok {
				log.Printf("failed to append ca certs: %s", err)
				return
			}

			creds := credentials.NewTLS(&tls.Config{
				ServerName:   "127.0.0.1",
				Certificates: []tls.Certificate{certificate},
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
