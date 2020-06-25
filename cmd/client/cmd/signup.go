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

// SignUpCmd ..
var SignUpCmd *cobra.Command

var (
	email    *string
	password *string
	server   *string
)

func init() {
	SignUpCmd = &cobra.Command{
		Use:   "signup",
		Short: "signup into cloudssh instance",
	}
	email = SignUpCmd.Flags().StringP("email", "u", "hi@example.com", "your email account")
	password = SignUpCmd.Flags().StringP("password", "p", "********", "your password")
	server = SignUpCmd.Flags().StringP("server", "s", "https://cssh.example.com", "a cloudssh server instance")
	SignUpCmd.Run = signup
}

func signup(cmd *cobra.Command, args []string) {
	if dao.Conf.User.Token != "" && dao.Conf.User.TokenExpires.After(time.Now()) {
		log.Println("You already logged in", dao.Conf.Server, ", please logout first.")
		return
	}

	dao.Conf.Server = *server

	var req apiio.RegisterRequest
	req.Email = *email
	dao.Conf.MasterKey = xcrypto.MakeKey(*password, strings.ToLower(*email))
	req.PasswordHash = xcrypto.MakePassworkHash(*password, dao.Conf.MasterKey)
	encKey, err := xcrypto.MakeEncKey(dao.Conf.MasterKey.EncKey)
	if err != nil {
		log.Println("MakeEncKey", err)
		return
	}
	req.EncryptKey = encKey.ToString()
	prikey, pubkey, err := xcrypto.GenerateKeyPair(2048)
	if err != nil {
		log.Println("GenerateKeyPair", err)
		return
	}
	req.Privatekey = string(xcrypto.PrivateKeyToBytes(prikey))
	pubkeyByte, err := xcrypto.PublicKeyToBytes(pubkey)
	if err != nil {
		log.Println("PublicKeyToBytes", err)
		return
	}
	req.Pubkey = string(pubkeyByte)
	body, err := dao.API.Do("/signup", "POST", req)
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
	log.Println("Signup Success", resp.Data.ID, resp.Data.Email, err)
}
