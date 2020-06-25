package cmd

import (
	"encoding/json"
	"log"

	"github.com/naiba/cloudssh/cmd/client/dao"
	"github.com/naiba/cloudssh/internal/apiio"
	"github.com/naiba/cloudssh/internal/model"
	"github.com/spf13/cobra"
)

// LogoutCmd ..
var LogoutCmd *cobra.Command

func init() {
	LogoutCmd = &cobra.Command{
		Use:   "login",
		Short: "login into cloudssh instance",
	}
	email = LogoutCmd.Flags().StringP("email", "u", "hi@example.com", "your email account")
	password = LogoutCmd.Flags().StringP("password", "p", "********", "your password")
	server = LogoutCmd.Flags().StringP("server", "s", "https://cssh.example.com", "a cloudssh server instance")
	LogoutCmd.Run = logout
}

func logout(cmd *cobra.Command, args []string) {
	body, err := dao.API.Do("/logout", "GET", nil)
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
	dao.Conf.User = model.User{}
	err = dao.Conf.Save()
	log.Println("Logout Success", err)
}
