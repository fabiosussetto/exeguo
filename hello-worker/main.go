package main

import (
	"github.com/fabiosussetto/hello/hello-worker/cmd"
	// "github.com/go-cmd/cmd"
)

// func testProducers(pool *WorkerPool) chan struct{} {
// 	numProducers := 2
// 	exitCh := make(chan struct{})

// 	for i := 0; i < numProducers; i++ {
// 		// simulate async producer
// 		go func() {
// 			for i := 0; ; i++ {
// 				select {
// 				case <-exitCh:
// 					return
// 				default:
// 					time.Sleep(time.Duration(rand.Int31n(10000)) * time.Millisecond)

// 					cmd := cmd.NewCmd("../test_cmd", strconv.Itoa(rand.Intn(10)), strconv.Itoa(rand.Intn(5)))
// 					pool.jobChan <- NewJob(cmd)
// 				}
// 			}
// 		}()
// 	}

// 	return exitCh
// }

func main() {
	cmd.Execute()
}
