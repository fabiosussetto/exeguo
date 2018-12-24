package main

import (
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
)

type Worker struct {
	ID     int
	pool   *WorkerPool
	logger *log.Entry
}

func NewWorker(pool *WorkerPool, ID int) *Worker {
	return &Worker{
		pool:   pool,
		ID:     ID,
		logger: log.WithFields(log.Fields{"worker_id": ID}),
	}
}

func (w *Worker) Start() {
	go func() {
		defer w.pool.workersWg.Done()

		w.logger.Infof("Ready to process jobs")

		for {
			select {
			case <-w.pool.cancelChan:
				return

			case job := <-w.pool.jobChan:
				// if cancelChan and jobChan have messages ready at the same time, go scheduler
				// randomly selected one of the select cases. So it can happen that the job is still
				// scheduled (and if very unlucky, it can happen more than once in a row too)
				if atomic.LoadUint64(&w.pool.terminationFlag) == 1 {
					return
				}

				w.process(job)
			}
		}
	}()
}

func (w *Worker) process(job *Job) {
	job.worker = w

	w.logger.Infof("Processing job #%d", job.ID)

	statusChan := job.cmd.Start() // non-blocking

	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for range ticker.C {
			status := job.cmd.Status()
			job.cmdStatus = status
			// n := len(status.Stdout)
			// fmt.Println(status.Stdout[n-1])
			w.logger.Infof("Job #%d output: %s", job.ID, status.Stdout)
		}
	}()

	go func() {
		<-w.pool.cancelChan
		// TODO: no need to do this if we get this notification while worker is idle
		w.logger.Infof("Stopping job #%d ...", job.ID)
		job.cmd.Stop()
		ticker.Stop()
	}()

	// Block waiting for command to exit, be stopped, or be killed
	finalStatus := <-statusChan

	ticker.Stop()

	if !finalStatus.Complete {
		w.logger.Warnf("Forced termination of job #%d", job.ID)
	} else {
		w.logger.Infof("Finished job #%d", job.ID)
	}
}
