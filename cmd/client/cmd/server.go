package cmd

import (
	"github.com/spf13/cobra"

	"github.com/naiba/cloudssh/cmd/client/cmd/server"
)

// ServerCmd ..
var ServerCmd *cobra.Command

func init() {
	ServerCmd = &cobra.Command{
		Use:   "server",
		Short: "server manage",
	}
	ServerCmd.AddCommand(server.ListCmd)
	ServerCmd.AddCommand(server.CreateCmd)
	ServerCmd.AddCommand(server.DialCmd)
	ServerCmd.AddCommand(server.EditCmd)
	ServerCmd.AddCommand(server.DeleteCmd)
}
