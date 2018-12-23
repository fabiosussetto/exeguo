package main

import (
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-cmd/cmd"
	log "github.com/sirupsen/logrus"
)

var terminationFlag uint64

type Worker struct {
	ID         int
	wg         *sync.WaitGroup
	jobChan    <-chan *Job
	cancelChan <-chan struct{}
	logger     *log.Entry
}

type Job struct {
	ID         int
	secsToWait int
	worker     *Worker
	cmd        *cmd.Cmd
	cmdStatus  cmd.Status
}

func NewJob(ID int, command *cmd.Cmd) *Job {
	return &Job{
		ID: ID,
		// secsToWait: secsToWait,
		cmd: command,
		cmdStatus: cmd.Status{
			Cmd:      "",
			PID:      0,
			Complete: false,
			Exit:     -1,
			Error:    nil,
			Runtime:  0,
		},
	}
}

func NewWorker(ID int, wg *sync.WaitGroup, jobChan <-chan *Job, cancelChan <-chan struct{}) *Worker {
	return &Worker{
		ID,
		wg,
		jobChan,
		cancelChan,
		log.WithFields(log.Fields{"worker_id": ID}),
	}
}

func (w *Worker) Start() {
	go func() {
		defer w.wg.Done()

		for {
			select {
			case <-w.cancelChan:
				return

			case job := <-w.jobChan:
				// if cancelChan and jobChan have messages ready at the same time, go scheduler
				// randomly selected one of the select cases. So it can happen that the job is still
				// scheduled (and if very unlucky, it can happen more than once in a row too)
				if atomic.LoadUint64(&terminationFlag) == 1 {
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

	// jobCmd := cmd.NewCmd("../test_cmd", strconv.Itoa(job.secsToWait), "5")

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
		<-w.cancelChan
		w.logger.Infof("Stopping job #%d ...", job.ID)
		job.cmd.Stop()
		ticker.Stop()
	}()

	// Block waiting for command to exit, be stopped, or be killed
	finalStatus := <-statusChan

	if !finalStatus.Complete {
		w.logger.Warnf("Forced termination of job #%d", job.ID)
	} else {
		w.logger.Infof("Finished job #%d", job.ID)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	jobChan := make(chan *Job, 100)
	cancelChan := make(chan struct{})

	var wg sync.WaitGroup
	var shutdownWg sync.WaitGroup

	shutdownWg.Add(1)

	go func() {
		sig := <-gracefulStop

		log.Infof("Caught signal %+v", sig)

		atomic.StoreUint64(&terminationFlag, 1)
		close(cancelChan)

		log.Info("Wait for 3 second to finish processing")

		time.Sleep(5 * time.Second)

		log.Info("Exiting")
		shutdownWg.Done()
	}()

	numWorkers := 4

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		w := NewWorker(i, &wg, jobChan, cancelChan)
		w.Start()
	}

	for i := 0; i < 50; i++ {
		cmd := cmd.NewCmd("../test_cmd", strconv.Itoa(rand.Intn(10)), strconv.Itoa(rand.Intn(5)))
		jobChan <- NewJob(i, cmd)
	}

	wg.Wait()
	shutdownWg.Wait()
}
