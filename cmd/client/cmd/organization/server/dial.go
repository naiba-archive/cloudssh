package server

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/naiba/cloudssh/cmd/client/dao"
)

// DialCmd ..
var DialCmd *cobra.Command

func init() {
	DialCmd = &cobra.Command{
		Use:   "dial",
		Short: "Connect to server, you must set server's name or id",
	}
	DialCmd.Run = dial
	DialCmd.Flags().StringP("name", "n", "", "server name")
	DialCmd.Flags().StringP("id", "i", "", "server id")
}

func dial(cmd *cobra.Command, args []string) {
	name, _ := cmd.Flags().GetString("name")
	id, _ := cmd.Flags().GetString("id")
	orgID, _ := cmd.Parent().Parent().PersistentFlags().GetUint64("oid")
	if name == "" && id == "" {
		log.Println("You must set which server you want to connect")
		return
	}
	if orgID == 0 {
		log.Println("You must set organization id")
		return
	}
	dao.API.DialServer(orgID, name, id)
}
