package cmd

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/naiba/cloudssh/cmd/client/dao"
	"github.com/naiba/cloudssh/internal/apiio"
	"github.com/naiba/cloudssh/pkg/xcrypto"
	"github.com/spf13/cobra"
)

// LoginCmd ..
var LoginCmd *cobra.Command

func init() {
	LoginCmd = &cobra.Command{
		Use:   "login",
		Short: "login into cloudssh instance",
	}
	LoginCmd.Flags().StringP("email", "u", "hi@example.com", "your email account")
	LoginCmd.Flags().StringP("password", "p", "********", "your password")
	LoginCmd.Flags().StringP("server", "s", "https://cssh.example.com", "a cloudssh server instance")
	LoginCmd.Run = login
}

func login(cmd *cobra.Command, args []string) {
	var (
		email    string
		password string
		server   string
	)
	email, _ = cmd.Flags().GetString("email")
	password, _ = cmd.Flags().GetString("password")
	server, _ = cmd.Flags().GetString("server")

	if dao.Conf.User.Token != "" && dao.Conf.User.TokenExpires.After(time.Now()) {
		log.Println("You already logged in", dao.Conf.Server, ", please logout first.")
		return
	}

	dao.Conf.Server = server

	var req apiio.LoginRequest
	req.Email = email
	dao.Conf.MasterKey = xcrypto.MakeKey(password, strings.ToLower(email))
	req.PasswordHash = xcrypto.MakePassworkHash(password, dao.Conf.MasterKey)
	body, err := dao.API.Do("/login", "POST", req)
	if err != nil {
		log.Println("API Request", err)
		return
	}
	var resp apiio.RegisterResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		log.Println("API Request", req, string(body), err)
		return
	}
	if !resp.Success {
		log.Println("API Request", resp.Message)
		return
	}
	dao.Conf.User = resp.Data
	err = dao.Conf.Save()
	log.Println("Login Success", "ID", resp.Data.ID, resp.Data.Email, err)
}
