package cmd

import (
	"github.com/naiba/cloudssh/cmd/client/cmd/server"
	"github.com/spf13/cobra"
)

// ServerCmd ..
var ServerCmd *cobra.Command

func init() {
	ServerCmd = &cobra.Command{
		Use:   "server",
		Short: "server manager",
	}
	ServerCmd.AddCommand(server.ListCmd)
	ServerCmd.AddCommand(server.CreateCmd)
	ServerCmd.AddCommand(server.DialCmd)
	ServerCmd.AddCommand(server.EditCmd)
	ServerCmd.AddCommand(server.DeleteCmd)
}
