package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/liamylian/x-rsa/golang/xrsa"
	"github.com/spf13/cobra"

	"github.com/naiba/cloudssh/cmd/client/dao"
	"github.com/naiba/cloudssh/internal/apiio"
	"github.com/naiba/cloudssh/pkg/xcrypto"
)

// PasswdCmd ..
var PasswdCmd *cobra.Command

func init() {
	PasswdCmd = &cobra.Command{
		Use:   "passwd",
		Short: "change password and reset your publickey",
	}
	PasswdCmd.Run = passwd
}

func passwd(cmd *cobra.Command, args []string) {
	var oldPass string
	fmt.Print("Old password to confirm: ")
	fmt.Scanf("%s", &oldPass)

	var newPass string
	fmt.Print("New password: ")
	fmt.Scanf("%s", &newPass)

	var req apiio.PasswdRequest
	req.OldPasswordHash = xcrypto.MakePassworkHash(oldPass, dao.Conf.MasterKey)
	newMasterKey := xcrypto.MakeKey(newPass, strings.ToLower(dao.Conf.User.Email))
	req.PasswordHash = xcrypto.MakePassworkHash(newPass, newMasterKey)
	encKey, err := xcrypto.MakeEncKey(newMasterKey.EncKey)
	if err != nil {
		log.Println("MakeEncKey", err)
		return
	}
	req.EncryptKey = encKey.ToString()

	publicKey, privateKey := bytes.NewBufferString(""), bytes.NewBufferString("")
	if err := xrsa.CreateKeys(publicKey, privateKey, 2048); err != nil {
		return
	}
	cs, err := xcrypto.Encrypt(privateKey.Bytes(), newMasterKey)
	if err != nil {
		log.Println("xcrypto.Encrypt", err)
		return
	}
	req.Privatekey = cs.ToString()
	req.Pubkey = publicKey.String()

	// server data
	servers, err := dao.API.GetServers(0)
	if err != nil {
		log.Println("GetServers", err)
		return
	}
	for i := 0; i < len(servers); i++ {
		if err := xcrypto.EncryptStruct(&servers[i], newMasterKey); err != nil {
			log.Println("EncryptStruct", servers[i], err)
			return
		}
	}
	req.Server = append(req.Server, servers...)

	// team data
	xr, err := xrsa.NewXRsa(publicKey.Bytes(), privateKey.Bytes())
	if err != nil {
		log.Println("NewXRsa", err)
		return
	}
	body, err := dao.API.Do("/user/team", "GET", nil)
	if err != nil {
		log.Println("API.Do", err)
		return
	}
	var teams apiio.ListTeamUserResponse
	if err = json.Unmarshal(body, &teams); err != nil {
		log.Println("Unmarshal", string(body), err)
		return
	}
	oldXr, err := dao.Conf.GerUserXRsa()
	if err != nil {
		log.Println("GerUserXRsa[old]", err)
		return
	}
	for i := 0; i < len(teams.Data.User); i++ {
		teams.Data.User[i].PrivateKey, err = oldXr.PrivateDecrypt(teams.Data.User[i].PrivateKey)
		if err != nil {
			log.Println("PublicEncrypt[old]", err)
			return
		}
		teams.Data.User[i].PrivateKey, err = xr.PublicEncrypt(teams.Data.User[i].PrivateKey)
		if err != nil {
			log.Println("PublicEncrypt[new]", err)
			return
		}
	}
	req.TeamUser = append(req.TeamUser, teams.Data.User...)

	// sync
	body, err = dao.API.Do("/user/passwd", "POST", req)
	if err != nil {
		log.Println("API Request", err)
		return
	}
	var resp apiio.UserResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		log.Println("API Request", string(body), err)
		return
	}
	if !resp.Success {
		log.Println("API Request", resp.Message)
		return
	}
	dao.Conf.User = resp.Data
	dao.Conf.MasterKey = newMasterKey
	err = dao.Conf.Save()
	log.Println(resp.Message, err)
}
