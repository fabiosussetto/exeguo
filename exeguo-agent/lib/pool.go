package lib

import (
	"fmt"
	"strings"
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

	jobs map[uint64]*Job
}

type JobCmd struct {
	JobId uint64
	Name  string
	Args  []string
	Env   []string
	Dir   string
}

func NewWorkerPool(numWorkers int) *WorkerPool {
	return &WorkerPool{
		NumWorkers: numWorkers,
		jobs:       make(map[uint64]*Job),
	}
}

func (pool *WorkerPool) RunCmd(jobCmd JobCmd) *Job {
	cmd := cmd.NewCmd("bash", "-c", fmt.Sprintf("%s %s", jobCmd.Name, strings.Join(jobCmd.Args, " ")))

	job := NewJob(jobCmd.JobId, cmd)

	pool.jobs[job.ID] = job

	go func() {
		pool.jobChan <- job
	}()

	return job
}

func (pool *WorkerPool) GetJob(JobID uint64) (*Job, bool) {
	job, ok := pool.jobs[JobID]
	return job, ok
}

func (pool *WorkerPool) StopJob(JobID uint64) bool {
	job, found := pool.GetJob(JobID)
	if !found {
		log.Warnf("Cannot find Job to stop with id %d", JobID)
		return false
	}
	job.Stop()
	return true
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
