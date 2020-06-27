package cmd

import (
	"github.com/spf13/cobra"

	"github.com/naiba/cloudssh/cmd/client/cmd/organization"
)

// OrganizationCmd ..
var OrganizationCmd *cobra.Command

func init() {
	OrganizationCmd = &cobra.Command{
		Use:   "org",
		Short: "organization manage",
	}
	OrganizationCmd.PersistentFlags().Uint64P("oid", "o", 0, "organization id")
	OrganizationCmd.AddCommand(organization.CreateCmd)
	OrganizationCmd.AddCommand(organization.ServerCmd)
}
