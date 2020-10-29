package server

import (
	"log"

	"github.com/naiba/cloudssh/cmd/client/dao"
	"github.com/spf13/cobra"
)

// CreateCmd ..
var CreateCmd *cobra.Command

func init() {
	CreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create server",
	}
	CreateCmd.Run = create
}

func create(cmd *cobra.Command, args []string) {
	teamID, _ := cmd.Parent().Parent().PersistentFlags().GetUint64("oid")
	if teamID == 0 {
		log.Println("must set team ID")
		return
	}
	dao.API.CreateServer(teamID)
}
