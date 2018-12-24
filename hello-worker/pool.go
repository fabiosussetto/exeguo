package main

import (
	"os"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
)

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
