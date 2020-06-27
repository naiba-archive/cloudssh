package organization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/liamylian/x-rsa/golang/xrsa"
	"github.com/naiba/cloudssh/cmd/client/dao"
	"github.com/naiba/cloudssh/internal/apiio"
	"github.com/spf13/cobra"
)

// CreateCmd ..
var CreateCmd *cobra.Command

func init() {
	CreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create organization",
	}
	CreateCmd.Run = create
}

func create(cmd *cobra.Command, args []string) {
	var req apiio.NewOrgRequrest

	fmt.Print("Organization Name: ")
	fmt.Scanf("%s", &req.Name)

	publicKey, privateKey := bytes.NewBufferString(""), bytes.NewBufferString("")
	if err := xrsa.CreateKeys(publicKey, privateKey, 2048); err != nil {
		log.Println("CreateKeys", err)
		return
	}
	req.Pubkey = publicKey.String()
	xr, err := dao.Conf.GerUserXRsa()
	if err != nil {
		log.Println("GerUserXRsa", err)
		return
	}
	req.Prikey, err = xr.PublicEncrypt(privateKey.String())
	if err != nil {
		log.Println("EncryptWithPublicKey", err)
		return
	}

	orgXr, err := xrsa.NewXRsa(publicKey.Bytes(), privateKey.Bytes())
	if err != nil {
		log.Println("NewXRsa", err)
		return
	}

	req.Name, err = orgXr.PublicEncrypt(req.Name)
	if err != nil {
		log.Println("Encrypt prikey", err)
		return
	}

	body, err := dao.API.Do("/organization", "POST", req)
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
