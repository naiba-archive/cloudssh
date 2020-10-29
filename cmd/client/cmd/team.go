package cmd

import (
	"github.com/spf13/cobra"

	"github.com/naiba/cloudssh/cmd/client/cmd/team"
)

// TeamCmd ..
var TeamCmd *cobra.Command

func init() {
	TeamCmd = &cobra.Command{
		Use:   "team",
		Short: "team manage",
	}
	TeamCmd.PersistentFlags().Uint64P("oid", "o", 0, "team id")
	TeamCmd.AddCommand(team.CreateCmd)
	TeamCmd.AddCommand(team.ServerCmd)
	TeamCmd.AddCommand(team.EditCmd)
	TeamCmd.AddCommand(team.UserCmd)
	TeamCmd.AddCommand(team.DeleteCmd)
	TeamCmd.AddCommand(team.ListCmd)
}
