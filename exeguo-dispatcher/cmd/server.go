package cmd

import (
	s "github.com/fabiosussetto/exeguo/exeguo-dispatcher/server"
	"github.com/spf13/cobra"
)

var config = &s.ServerConfig{}

func init() {
	ServerCmd.PersistentFlags().StringVarP(&config.ServerAddress, "host", "H", "localhost:8080", "address:port to listen on (defaults to localhost:8080)")
	ServerCmd.PersistentFlags().StringVarP(&config.PathToDB, "db-file", "", "./exeguo.sqlite", "Path to the db file. Will be created if non-existant.")
}

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Start server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		s.StartServer(*config)
	},
}
