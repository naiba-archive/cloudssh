package team

import (
	"github.com/spf13/cobra"

	"github.com/naiba/cloudssh/cmd/client/cmd/team/user"
)

// UserCmd ..
var UserCmd *cobra.Command

func init() {
	UserCmd = &cobra.Command{
		Use:   "user",
		Short: "manage team users",
	}
	UserCmd.AddCommand(user.AddCmd)
	UserCmd.AddCommand(user.DeleteCmd)
	UserCmd.AddCommand(user.ListCmd)
}
