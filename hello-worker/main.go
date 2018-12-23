package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-cmd/cmd"
)

var terminationFlag uint64

type Worker struct {
	ID         int
	wg         *sync.WaitGroup
	jobChan    <-chan *Job
	cancelChan <-chan struct{}
}

type Job struct {
	ID         int
	secsToWait int
	worker     *Worker
	cmdStatus  cmd.Status
}

func NewJob(ID int, secsToWait int) *Job {
	return &Job{
		ID:         ID,
		secsToWait: secsToWait,
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
	return &Worker{ID, wg, jobChan, cancelChan}
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

	msg := fmt.Sprintf("[Worker %d] Processing job #%d for %d seconds...", w.ID, job.ID, job.secsToWait)
	fmt.Println(msg)

	jobCmd := cmd.NewCmd("../test_cmd", strconv.Itoa(job.secsToWait), "5")

	statusChan := jobCmd.Start() // non-blocking

	ticker := time.NewTicker(1 * time.Second)

	// Print last line of stdout every 1s
	go func() {
		for range ticker.C {
			status := jobCmd.Status()
			job.cmdStatus = status
			// n := len(status.Stdout)
			// fmt.Println(status.Stdout[n-1])
			fmt.Printf("[Worker %d] Job #%d output: %s\n", w.ID, job.ID, status.Stdout)
		}
	}()

	go func() {
		<-w.cancelChan
		fmt.Printf("[Worker %d] Stopping job #%d ...\n", w.ID, job.ID)
		jobCmd.Stop()
		ticker.Stop()
	}()

	// Block waiting for command to exit, be stopped, or be killed
	finalStatus := <-statusChan

	if !finalStatus.Complete {
		fmt.Printf("[Worker %d] Forced termination of job #%d\n", w.ID, job.ID)
	} else {
		fmt.Printf("[Worker %d] Finished job #%d\n", w.ID, job.ID)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

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
		fmt.Printf("caught sig: %+v \n", sig)

		atomic.StoreUint64(&terminationFlag, 1)
		close(cancelChan)

		fmt.Println("Wait for 3 second to finish processing")

		time.Sleep(5 * time.Second)

		fmt.Println("Exiting")
		shutdownWg.Done()
		// os.Exit(0)
	}()

	numWorkers := 4

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		w := NewWorker(i, &wg, jobChan, cancelChan)
		w.Start()
	}

	for i := 0; i < 50; i++ {
		jobChan <- NewJob(i, rand.Intn(10))
	}

	wg.Wait()
	shutdownWg.Wait()
}
