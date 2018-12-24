package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	hw "github.com/fabiosussetto/hello/hello-worker/lib"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var numWorkers int

var rootCmd = &cobra.Command{
	Use:   "hello",
	Short: "Hello executes bash commands",
	Long:  `Work in progress`,
	Run: func(cmd *cobra.Command, args []string) {
		start()
	},
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&numWorkers, "workers", "w", 4, "number of workers (defaults to 4)")
}

func start() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	pool := hw.NewWorkerPool(numWorkers)

	shutdownChan := make(chan os.Signal)

	signal.Notify(shutdownChan, syscall.SIGTERM)
	signal.Notify(shutdownChan, syscall.SIGINT)

	poolStatusChan, poolStatusForcedChan := pool.Start()

	select {
	case sig := <-shutdownChan:
		log.Infof("Caught signal %+v, gracefully stopping worker pool", sig)

		go func() {
			pool.Stop()
		}()

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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
