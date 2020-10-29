package user

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/naiba/cloudssh/cmd/client/dao"
	"github.com/naiba/cloudssh/internal/apiio"
	"github.com/naiba/cloudssh/internal/model"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// ListCmd ..
var ListCmd *cobra.Command

func init() {
	ListCmd = &cobra.Command{
		Use:   "list",
		Short: "list users from team",
	}
	ListCmd.Run = list
}

func list(cmd *cobra.Command, args []string) {
	teamID, _ := cmd.Parent().Parent().PersistentFlags().GetUint64("oid")
	if teamID == 0 {
		log.Println("must set team ID")
		return
	}

	body, err := dao.API.Do(fmt.Sprintf("/team/%d/user", teamID), "GET", nil)
	if err != nil {
		log.Println("API Request", err)
		return
	}
	var resp apiio.ListTeamUserResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		log.Println("API Request", string(body), err)
		return
	}
	if !resp.Success {
		log.Println("API Request", resp.Message)
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Email", "Permission"})
	for i := 0; i < len(resp.Data.User); i++ {
		table.Append([]string{
			fmt.Sprintf("%d", resp.Data.User[i].UserID),
			resp.Data.Email[resp.Data.User[i].UserID],
			model.GetPermissionComment(resp.Data.User[i].Permission),
		})
	}
	table.Render()
}
