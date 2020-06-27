package server

import (
	"fmt"
	"log"
	"os"

	"github.com/naiba/cloudssh/cmd/client/dao"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// ListCmd ..
var ListCmd *cobra.Command

func init() {
	ListCmd = &cobra.Command{
		Use:   "list",
		Short: "List server",
	}
	ListCmd.Run = list
}

func list(cmd *cobra.Command, args []string) {
	orgID, _ := cmd.Parent().Parent().PersistentFlags().GetUint64("oid")
	if orgID == 0 {
		log.Println("must set organization ID")
		return
	}
	servers, err := dao.API.GetServers(orgID)
	if err != nil {
		log.Println("API.GetServers", err)
		return
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "IP", "User", "CreatedAt"})
	for i := 0; i < len(servers); i++ {
		table.Append([]string{
			fmt.Sprintf("%d", servers[i].ID),
			servers[i].Name,
			servers[i].IP,
			servers[i].LoginUser,
			servers[i].CreatedAt.String(),
		})
	}
	table.Render()
}
