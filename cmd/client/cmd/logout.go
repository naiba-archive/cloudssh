package cmd

import (
	"encoding/json"
	"log"

	"github.com/naiba/cloudssh/cmd/client/dao"
	"github.com/naiba/cloudssh/internal/apiio"
	"github.com/spf13/cobra"
)

// LogoutCmd ..
var LogoutCmd *cobra.Command

func init() {
	LogoutCmd = &cobra.Command{
		Use:   "logout",
		Short: "user logout",
	}
	LogoutCmd.Flags().BoolP("force", "f", false, "force logout")
	LogoutCmd.Run = logout
}

func logout(cmd *cobra.Command, args []string) {
	var flag bool
	defer func() {
		force, _ := cmd.Flags().GetBool("force")
		if force || flag {
			dao.Conf = &dao.Config{}
			dao.Conf.Save()
		}
	}()
	body, err := dao.API.Do("/auth/logout", "GET", nil)
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
	if resp.Success {
		log.Println(resp.Message)
	}
	flag = true
}
