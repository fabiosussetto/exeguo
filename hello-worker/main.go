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

type Worker struct {
	ID     int
	pool   *WorkerPool
	logger *log.Entry
}

type Job struct {
	ID        int
	worker    *Worker
	cmd       *cmd.Cmd
	cmdStatus cmd.Status
}

func NewJob(ID int, command *cmd.Cmd) *Job {
	return &Job{
		ID:  ID,
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

type WorkerPool struct {
	NumWorkers int

	terminationFlag  uint64
	statusChan       chan struct{}
	gracefulStopChan chan os.Signal
	jobChan          chan *Job
	cancelChan       chan struct{}
	workersWg        sync.WaitGroup
	shutdownWg       sync.WaitGroup
}

func NewWorkerPool(numWorkers int) *WorkerPool {
	return &WorkerPool{
		NumWorkers: numWorkers,

		gracefulStopChan: make(chan os.Signal),
	}
}

func (pool *WorkerPool) Start() <-chan struct{} {
	pool.statusChan = make(chan struct{}, 1)
	pool.jobChan = make(chan *Job, 100)
	pool.cancelChan = make(chan struct{})

	pool.shutdownWg.Add(1)

	log.Infof("Starting %d workers", pool.NumWorkers)

	for i := 0; i < pool.NumWorkers; i++ {
		pool.workersWg.Add(1)
		w := NewWorker(pool, i)
		w.Start()
	}

	return pool.statusChan
}

func (pool *WorkerPool) Stop() {
	atomic.StoreUint64(&pool.terminationFlag, 1)
	close(pool.cancelChan)

	log.Info("Wait 5s to finish processing")

	time.Sleep(5 * time.Second)

	pool.shutdownWg.Done()

	// signal user we're shutting down
	close(pool.statusChan)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	pool := NewWorkerPool(4)

	signal.Notify(pool.gracefulStopChan, syscall.SIGTERM)
	signal.Notify(pool.gracefulStopChan, syscall.SIGINT)

	go func() {
		sig := <-pool.gracefulStopChan

		log.Infof("Caught signal %+v, stopping pool", sig)

		pool.Stop()
	}()

	poolChan := pool.Start()

	numProducers := 2
	var producerWg sync.WaitGroup

	for i := 0; i < numProducers; i++ {
		producerWg.Add(1)

		// simulate async producer
		go func(k int) {
			defer producerWg.Done()

			for i := 0; i < 1; i++ {
				cmd := cmd.NewCmd("../test_cmd", strconv.Itoa(rand.Intn(10)), strconv.Itoa(rand.Intn(5)))
				pool.jobChan <- NewJob((100*k)+i, cmd)

				time.Sleep(time.Duration(rand.Int31n(10000)) * time.Millisecond)
			}
		}(i)
	}

	producerWg.Wait()

	<-poolChan
}
