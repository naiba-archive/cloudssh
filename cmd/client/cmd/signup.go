package cmd

import (
	"bytes"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/liamylian/x-rsa/golang/xrsa"
	"github.com/spf13/cobra"

	"github.com/naiba/cloudssh/cmd/client/dao"
	"github.com/naiba/cloudssh/internal/apiio"
	"github.com/naiba/cloudssh/pkg/xcrypto"
)

// SignUpCmd ..
var SignUpCmd *cobra.Command

func init() {
	SignUpCmd = &cobra.Command{
		Use:   "signup",
		Short: "signup into cloudssh instance",
	}
	SignUpCmd.Flags().StringP("email", "u", "hi@example.com", "your email account")
	SignUpCmd.Flags().StringP("password", "p", "********", "your password")
	SignUpCmd.Flags().StringP("server", "s", "https://cssh.example.com", "a cloudssh server instance")
	SignUpCmd.Run = signup
}

func signup(cmd *cobra.Command, args []string) {
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

	var req apiio.RegisterRequest
	req.Email = email
	dao.Conf.MasterKey = xcrypto.MakeKey(password, strings.ToLower(email))
	req.PasswordHash = xcrypto.MakePassworkHash(password, dao.Conf.MasterKey)
	encKey, err := xcrypto.MakeEncKey(dao.Conf.MasterKey.EncKey)
	if err != nil {
		log.Println("MakeEncKey", err)
		return
	}
	req.EncryptKey = encKey.ToString()

	publicKey := bytes.NewBufferString("")
	privateKey := bytes.NewBufferString("")

	if err := xrsa.CreateKeys(publicKey, privateKey, 2048); err != nil {
		return
	}
	cs, err := xcrypto.Encrypt(privateKey.Bytes(), dao.Conf.MasterKey)
	if err != nil {
		log.Println("xcrypto.Encrypt", err)
		return
	}
	req.Privatekey = cs.ToString()
	req.Pubkey = string(publicKey.Bytes())
	body, err := dao.API.Do("/auth/signup", "POST", req)
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
	log.Println("Signup Success", "ID", resp.Data.ID, resp.Data.Email, err)
}
