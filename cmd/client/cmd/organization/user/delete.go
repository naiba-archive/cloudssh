package user

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/naiba/cloudssh/cmd/client/dao"
	"github.com/naiba/cloudssh/internal/apiio"
	"github.com/spf13/cobra"
)

// DeleteCmd ..
var DeleteCmd *cobra.Command

func init() {
	DeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "delete organization user(s)",
	}
	DeleteCmd.Flags().UintSlice("id", []uint{}, "sever id list --id=\"1,3,4\"")
	DeleteCmd.Run = delete
}

func delete(cmd *cobra.Command, args []string) {
	orgID, _ := cmd.Parent().Parent().PersistentFlags().GetUint64("oid")
	if orgID == 0 {
		log.Println("must set organization ID")
		return
	}
	id, _ := cmd.Flags().GetUintSlice("id")
	var req apiio.DeleteOrganizationRequest
	req.ID = id
	if len(req.ID) == 0 {
		log.Println("Please input server id list")
		return
	}
	fmt.Printf("Please type 'y' to confirm delete %+v:", req.ID)
	var confirm string
	fmt.Scanf("%s", &confirm)
	if confirm != "y" {
		return
	}

	body, err := dao.API.Do(fmt.Sprintf("/organization/%d/user/batch-delete", orgID), "POST", req)
	if err != nil {
		log.Println("API Request", err)
		return
	}
	var resp apiio.Response
	if err = json.Unmarshal(body, &resp); err != nil {
		log.Println("API Request", string(body), err)
		return
	}
	if !resp.Success {
		log.Println("API Request", resp.Message)
		return
	}
	log.Println(resp.Message)
}
