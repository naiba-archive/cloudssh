package server

import (
	"log"

	"github.com/naiba/cloudssh/cmd/client/dao"
	"github.com/spf13/cobra"
)

// EditCmd ..
var EditCmd *cobra.Command

func init() {
	EditCmd = &cobra.Command{
		Use:   "edit",
		Short: "Edit server",
	}
	EditCmd.Flags().StringP("id", "i", "", "server id")
	EditCmd.MarkFlagRequired("id")
	EditCmd.Run = edit
}

func edit(cmd *cobra.Command, args []string) {
	id, _ := cmd.Flags().GetString("id")
	if id == "" {
		cmd.Usage()
		return
	}
	if err := dao.API.EditServer(id); err != nil {
		log.Println(err)
	}
}
