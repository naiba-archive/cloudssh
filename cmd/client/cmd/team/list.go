package team

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/naiba/cloudssh/cmd/client/dao"
	"github.com/naiba/cloudssh/internal/apiio"
	"github.com/naiba/cloudssh/internal/model"
)

// ListCmd ..
var ListCmd *cobra.Command

func init() {
	ListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Teams",
	}
	ListCmd.Run = list
}

func list(cmd *cobra.Command, args []string) {
	body, err := dao.API.Do("/team", "GET", nil)
	if err != nil {
		log.Println("API", err)
		return
	}
	var resp apiio.ListTeamResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		log.Println("Unmarshal", string(body), err)
		return
	}
	if !resp.Success {
		log.Println(resp.Message)
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Permission", "CreatedAt"})
	for i := 0; i < len(resp.Data.Teamnazation); i++ {
		teamXrsa, err := dao.API.GetTeamXRsa(resp.Data.Teamnazation[i].ID)
		if err != nil {
			log.Println("GetTeamXRsa", err)
			return
		}
		resp.Data.Teamnazation[i].Name, err = teamXrsa.PrivateDecrypt(resp.Data.Teamnazation[i].Name)
		if err != nil {
			log.Println("PrivateDecrypt", err)
			return
		}
		table.Append([]string{
			fmt.Sprintf("%d", resp.Data.Teamnazation[i].ID),
			resp.Data.Teamnazation[i].Name,
			model.GetPermissionComment(resp.Data.Permission[resp.Data.Teamnazation[i].ID]),
			resp.Data.Teamnazation[i].CreatedAt.String(),
		})
	}
	table.Render()
}
