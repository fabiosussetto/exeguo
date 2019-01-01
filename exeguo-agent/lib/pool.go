package lib

import (
	"sync"
	"sync/atomic"

	"github.com/go-cmd/cmd"
	log "github.com/sirupsen/logrus"
)

type WorkerPool struct {
	NumWorkers int
	jobCounter uint64

	terminationFlag  uint64
	statusChan       chan struct{}
	statusForcedChan chan struct{}

	jobChan         chan *Job
	cancelChan      chan struct{}
	forceCancelChan chan struct{}
	workersWg       sync.WaitGroup
}

type JobCmd struct {
	Name string
	Args []string
	Env  []string
	Dir  string
}

func NewWorkerPool(numWorkers int) *WorkerPool {
	return &WorkerPool{
		NumWorkers: numWorkers,
	}
}

func (pool *WorkerPool) RunCmd(jobCmd JobCmd) *Job {
	atomic.AddUint64(&pool.jobCounter, 1)

	cmd := cmd.NewCmd(jobCmd.Name, jobCmd.Args...)

	job := NewJob(atomic.LoadUint64(&pool.jobCounter), cmd)

	go func() {
		pool.jobChan <- job
	}()

	return job
}

func (pool *WorkerPool) Start() (<-chan struct{}, <-chan struct{}) {
	pool.statusChan = make(chan struct{}, 1)
	pool.statusForcedChan = make(chan struct{}, 1)

	pool.jobChan = make(chan *Job, 100)
	pool.cancelChan = make(chan struct{})
	pool.forceCancelChan = make(chan struct{})

	log.Infof("Starting %d workers", pool.NumWorkers)

	for i := 0; i < pool.NumWorkers; i++ {
		pool.workersWg.Add(1)
		w := NewWorker(pool, i)
		w.Start()
	}

	return pool.statusChan, pool.statusForcedChan
}

func (pool *WorkerPool) Stop() {
	atomic.StoreUint64(&pool.terminationFlag, 1)
	close(pool.cancelChan)

	// wait for all workers to finish current work
	pool.workersWg.Wait()

	// signal user we're shutting down
	close(pool.statusChan)
}

func (pool *WorkerPool) ForceStop() {
	atomic.StoreUint64(&pool.terminationFlag, 1)
	close(pool.forceCancelChan)

	pool.workersWg.Wait()

	// signal user we're shutting down
	close(pool.statusForcedChan)
}