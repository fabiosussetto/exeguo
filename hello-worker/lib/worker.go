package lib

import (
	"context"
	"io"
	"strings"
	"sync/atomic"
	"time"

	pb "github.com/fabiosussetto/hello/hello-sender/rpc"
	"github.com/go-cmd/cmd"
	log "github.com/sirupsen/logrus"
)

type Worker struct {
	ID     int
	pool   *WorkerPool
	logger *log.Entry

	gRPCStream pb.DispatcherService_StreamJobStatusClient
}

func NewWorker(pool *WorkerPool, ID int) *Worker {
	gRPCStream, err := pool.dispatcherClient.StreamJobStatus(context.Background())

	if err != nil {
		log.Fatalf("%v.StreamJobStatus(_) = _, %v", pool.dispatcherClient, err)
	}

	return &Worker{
		pool:       pool,
		ID:         ID,
		logger:     log.WithFields(log.Fields{"worker_id": ID}),
		gRPCStream: gRPCStream,
	}
}

func (w *Worker) Start() {
	go func() {
		defer w.pool.workersWg.Done()

		defer w.closeStatusUpdateStreaming()

		w.logger.Infof("Ready to process jobs")

		for {
			select {
			case <-w.pool.cancelChan:
				w.logger.Infof("Received shutdown signal, won't process any new jobs")
				return

			case job := <-w.pool.jobChan:
				// if cancelChan and jobChan have messages ready at the same time, go scheduler
				// randomly selected one of the select cases. So it can happen that the job is still
				// scheduled (and if very unlucky, it can happen more than once in a row too)
				if atomic.LoadUint64(&w.pool.terminationFlag) == 1 {
					return
				}

				w.logger.Infof("Processing job #%d", job.ID)

				job.worker = w
				w.process(job)
			}
		}
	}()
}

func (w *Worker) streamStatusUpdate(status *cmd.Status) {
	statusUpdate := &pb.JobStatusUpdate{StdinLine: strings.Join(status.Stdout, "\n")}

	err := w.gRPCStream.Send(statusUpdate)

	if err == io.EOF {
		w.logger.Warn("Lost gRPC connection to Dispatcher")
		return
	}

	if err != nil {
		w.logger.Fatalf("%v.Send(%v) = %v", w.gRPCStream, statusUpdate, err)
	}
}

func (w *Worker) closeStatusUpdateStreaming() {
	_, err := w.gRPCStream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", w.gRPCStream, err, nil)
	}
	w.logger.Info("Closed gRPC streaming to Dispatcher")
}

func (w *Worker) process(job *Job) {
	statusChan := job.cmd.Start() // non-blocking

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			status := job.cmd.Status()
			job.CmdStatus = status
			// n := len(status.Stdout)
			// fmt.Println(status.Stdout[n-1])
			w.logger.Infof("Job #%d output: %s", job.ID, status.Stdout)

			w.streamStatusUpdate(&status)
		}
	}()

	select {
	case <-w.pool.forceCancelChan:
		w.logger.Warnf("Forcefully stopping job #%d ...", job.ID)
		job.cmd.Stop()
		ticker.Stop()
	case finalStatus := <-statusChan:
		ticker.Stop()
		job.CmdStatus = finalStatus

		w.streamStatusUpdate(&finalStatus)

		if !finalStatus.Complete {
			w.logger.Warnf("Forced termination of job #%d", job.ID)
			return
		}

		w.logger.Infof("Job #%d completed. Output: %s", job.ID, finalStatus.Stdout)
	}
}
