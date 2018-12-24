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
		go func() {
			for i := 0; ; i++ {
				select {
				case <-exitCh:
					return
				default:
					time.Sleep(time.Duration(rand.Int31n(10000)) * time.Millisecond)

					cmd := cmd.NewCmd("../test_cmd", strconv.Itoa(rand.Intn(10)), strconv.Itoa(rand.Intn(5)))
					pool.jobChan <- NewJob(cmd)
				}
			}
		}()
	}

	return exitCh
}

func main() {
	rand.Seed(time.Now().UnixNano())
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	pool := NewWorkerPool(4)

	shutdownChan := make(chan os.Signal)

	signal.Notify(shutdownChan, syscall.SIGTERM)
	signal.Notify(shutdownChan, syscall.SIGINT)

	poolStatusChan, poolStatusForcedChan := pool.Start()
	producerExitChan := testProducers(pool)

	select {
	case sig := <-shutdownChan:
		log.Infof("Caught signal %+v, gracefully stopping worker pool", sig)

		go func() {
			pool.Stop()
		}()

		close(producerExitChan)

		select {
		case forceSig := <-shutdownChan:
			log.Warnf("Caught signal %+v, forcing pool shutdown", forceSig)
			pool.ForceStop()

		case <-poolStatusChan:
			log.Infof("Pool has been gracefully terminated")
			os.Exit(0)

		case <-poolStatusForcedChan:
			log.Warnf("Pool has been forcefully terminated")
			os.Exit(0)
		}
	}
}
