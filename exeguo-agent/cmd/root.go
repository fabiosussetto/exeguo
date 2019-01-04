package cmd

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"

	hw "github.com/fabiosussetto/exeguo/exeguo-agent/lib"
	workerServer "github.com/fabiosussetto/exeguo/exeguo-agent/server"
	pb "github.com/fabiosussetto/exeguo/exeguo-dispatcher/rpc"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var numWorkers int

var workerPool *hw.WorkerPool
var bindAddress string

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
	rootCmd.PersistentFlags().StringVarP(&bindAddress, "host", "H", "localhost:1234", "host:port to listen on")

	rootCmd.AddCommand(GenCredentialsCmd)
}

func startServer(workerPool *hw.WorkerPool) {
	log.Infoln("Starting gRPC server")

	lis, err := net.Listen("tcp", bindAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// tlsConfig := &tls.Config{
	// 	Certificates: []tls.Certificate{insecure.Cert},
	// 	ClientCAs:    insecure.CertPool,
	// 	ClientAuth:   tls.VerifyClientCertIfGiven,
	// }

	// grpcServer := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))

	// Load the certificates from disk
	certificate, err := tls.LoadX509KeyPair("../certs/server_cert.pem", "../certs/server_key.pem")
	if err != nil {
		log.Fatalf("could not load server key pair: %s", err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("../certs/ca_cert.pem")
	if err != nil {
		log.Fatalf("could not read ca certificate: %s", err)
	}

	// Append the client certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("failed to append client certs: %s", err)
	}

	// creds, err := credentials.NewServerTLSFromFile("../certs/server_cert.pem", "../certs/server_key.pem")
	// grpcServer := grpc.NewServer(grpc.Creds(creds))

	// Create the TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	})

	grpcServer := grpc.NewServer(grpc.Creds(creds))

	pb.RegisterJobServiceServer(grpcServer, &workerServer.JobServiceServer{WorkerPool: workerPool})

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
