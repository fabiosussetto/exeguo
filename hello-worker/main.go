package main

import (
	"math/rand"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/go-cmd/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	pool := NewWorkerPool(4)

	signal.Notify(pool.gracefulStopChan, syscall.SIGTERM)
	signal.Notify(pool.gracefulStopChan, syscall.SIGINT)

	numProducers := 2
	var producerWg sync.WaitGroup

	go func() {
		sig := <-pool.gracefulStopChan

		log.Infof("Caught signal %+v, stopping pool", sig)

		pool.Stop()
	}()

	poolChan := pool.Start()

	for i := 0; i < numProducers; i++ {
		producerWg.Add(1)

		// simulate async producer
		go func(k int) {
			defer producerWg.Done()

			for i := 0; i < 1; i++ {
				time.Sleep(time.Duration(rand.Int31n(10000)) * time.Millisecond)

				cmd := cmd.NewCmd("../test_cmd", strconv.Itoa(rand.Intn(10)), strconv.Itoa(rand.Intn(5)))
				pool.jobChan <- NewJob((100*k)+i, cmd)
			}
		}(i)
	}

	producerWg.Wait()

	<-poolChan
}
