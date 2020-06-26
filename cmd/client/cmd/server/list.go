package server

import (
	"fmt"
	"log"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/naiba/cloudssh/cmd/client/dao"
)

// ListCmd ..
var ListCmd *cobra.Command

func init() {
	ListCmd = &cobra.Command{
		Use:   "list",
		Short: "List servers",
	}
	ListCmd.Run = list
}

func list(cmd *cobra.Command, args []string) {
	servers, err := dao.API.GetServers()
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
			servers[i].User,
			servers[i].CreatedAt.String(),
		})
	}
	table.Render()
}
