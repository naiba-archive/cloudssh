package organization

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
		Short: "List Organizations",
	}
	ListCmd.Run = list
}

func list(cmd *cobra.Command, args []string) {
	body, err := dao.API.Do("/organization", "GET", nil)
	if err != nil {
		log.Println("API", err)
		return
	}
	var resp apiio.ListOrganizationResponse
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
	for i := 0; i < len(resp.Data.Orgnazation); i++ {
		orgXrsa, err := dao.API.GetOrganizationXRsa(resp.Data.Orgnazation[i].ID)
		if err != nil {
			log.Println("GetOrganizationXRsa", err)
			return
		}
		resp.Data.Orgnazation[i].Name, err = orgXrsa.PrivateDecrypt(resp.Data.Orgnazation[i].Name)
		if err != nil {
			log.Println("PrivateDecrypt", err)
			return
		}
		table.Append([]string{
			fmt.Sprintf("%d", resp.Data.Orgnazation[i].ID),
			resp.Data.Orgnazation[i].Name,
			model.GetPermissionComment(resp.Data.Permission[resp.Data.Orgnazation[i].ID]),
			resp.Data.Orgnazation[i].CreatedAt.String(),
		})
	}
	table.Render()
}
