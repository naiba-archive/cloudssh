package server

import (
	"log"

	"github.com/naiba/cloudssh/cmd/client/dao"
	"github.com/spf13/cobra"
)

// DeleteCmd ..
var DeleteCmd *cobra.Command

func init() {
	DeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "delete server(s)",
	}
	DeleteCmd.Flags().UintSlice("id", []uint{}, "sever id list --id=\"1,3,4\"")
	DeleteCmd.Run = delete
}

func delete(cmd *cobra.Command, args []string) {
	teamID, _ := cmd.Parent().Parent().PersistentFlags().GetUint64("oid")
	if teamID == 0 {
		log.Println("must set team ID")
		return
	}
	id, _ := cmd.Flags().GetUintSlice("id")
	dao.API.BatchDeleteServer(id, teamID)
}
