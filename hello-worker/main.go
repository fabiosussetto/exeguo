package main

import (
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-cmd/cmd"
	log "github.com/sirupsen/logrus"
)

func testProducers(pool *WorkerPool) chan struct{} {
	numProducers := 2
	exitCh := make(chan struct{})

	for i := 0; i < numProducers; i++ {
		// simulate async producer
		go func(k int) {
			for i := 0; ; i++ {
				select {
				case <-exitCh:
					return
				default:
					time.Sleep(time.Duration(rand.Int31n(10000)) * time.Millisecond)

					cmd := cmd.NewCmd("../test_cmd", strconv.Itoa(rand.Intn(10)), strconv.Itoa(rand.Intn(5)))
					pool.jobChan <- NewJob((100*k)+i, cmd)
				}
			}
		}(i)
	}

	return exitCh
}

func main() {
	rand.Seed(time.Now().UnixNano())
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	pool := NewWorkerPool(4)

	shutdownChan := make(chan os.Signal, 1)

	signal.Notify(shutdownChan, syscall.SIGTERM)
	signal.Notify(shutdownChan, syscall.SIGINT)

	poolChan := pool.Start()
	producerExitChan := testProducers(pool)

	go func() {
		sig := <-shutdownChan

		log.Infof("Caught signal %+v, stopping pool and producers", sig)
		close(producerExitChan)

		pool.Stop()

	}()

	<-producerExitChan
	<-poolChan
}
