package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type runner struct {
}

func main() {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	shutdownChan := make(chan int)

	go func() {
		sig := <-gracefulStop
		fmt.Printf("caught sig: %+v \n", sig)

		// used to stop processing goroutines (infinite for loop below) to do work while we clean up
		close(shutdownChan)

		fmt.Println("Wait for 3 second to finish processing")

		time.Sleep(3 * time.Second)
		os.Exit(0)
	}()

	var wg sync.WaitGroup
	parallel := 4
	wg.Add(parallel)

	for i := 0; i < parallel; i++ {
		go func(workerIndex int) {
			defer wg.Done()
			for {
				select {
				default:
					secsToWait := rand.Intn(5)
					fmt.Printf("[worker %d] Doing some work for %d seconds...\n", workerIndex, secsToWait)
					time.Sleep(time.Duration(secsToWait) * time.Second)
					fmt.Printf("[worker %d] ... done! \n", workerIndex)
				case <-shutdownChan:
					break
				}
			}
		}(i)
	}

	wg.Wait()
}
