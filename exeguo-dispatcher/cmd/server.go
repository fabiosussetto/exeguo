package cmd

import (
	s "github.com/fabiosussetto/exeguo/exeguo-dispatcher/server"
	"github.com/spf13/cobra"
)

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Start server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		s.StartServer()
	},
}
