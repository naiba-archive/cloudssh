package server

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/naiba/cloudssh/cmd/client/dao"
	"github.com/naiba/cloudssh/internal/apiio"
	"github.com/naiba/cloudssh/pkg/xcrypto"
)

// ListCmd ..
var ListCmd *cobra.Command

func init() {
	ListCmd = &cobra.Command{
		Use:   "ls",
		Short: "List servers",
	}
	ListCmd.Run = list
}

func list(cmd *cobra.Command, args []string) {
	body, err := dao.API.Do("/user/server", "GET", nil)
	if err != nil {
		log.Println("API Request", err)
		return
	}
	var resp apiio.ListServerResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		log.Println("API Request", string(body), err)
		return
	}
	if !resp.Success {
		log.Println("API Request", resp.Message)
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "IP", "User", "CreatedAt"})
	for i := 0; i < len(resp.Data); i++ {
		xcrypto.DecryptStruct(&resp.Data[i], dao.Conf.MasterKey)
		table.Append([]string{
			fmt.Sprintf("%d", resp.Data[i].ID),
			resp.Data[i].Name,
			resp.Data[i].IP,
			resp.Data[i].User,
			resp.Data[i].CreatedAt.String(),
		})
	}

	table.Render()
}
