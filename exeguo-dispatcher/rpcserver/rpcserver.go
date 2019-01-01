package rpcserver

import (
	"io"
	"log"

	pb "github.com/fabiosussetto/exeguo/exeguo-dispatcher/rpc"
	"github.com/jinzhu/gorm"
)

type DispatcherServer struct {
	DB *gorm.DB
}

func (s *DispatcherServer) StreamJobStatus(stream pb.DispatcherService_StreamJobStatusServer) error {
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
