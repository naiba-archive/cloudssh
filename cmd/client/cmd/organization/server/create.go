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
	orgID, _ := cmd.Parent().Parent().PersistentFlags().GetUint64("oid")
	if orgID == 0 {
		log.Println("must set organization ID")
		return
	}
	dao.API.CreateServer(orgID)
}
