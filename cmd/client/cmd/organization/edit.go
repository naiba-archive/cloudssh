package organization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/liamylian/x-rsa/golang/xrsa"
	"github.com/naiba/cloudssh/cmd/client/dao"
	"github.com/naiba/cloudssh/internal/apiio"
	"github.com/naiba/cloudssh/pkg/xcrypto"
	"github.com/spf13/cobra"
)

// EditCmd ..
var EditCmd *cobra.Command

func init() {
	EditCmd = &cobra.Command{
		Use:   "edit",
		Short: "Edit organization",
	}
	EditCmd.Run = edit
}

func edit(cmd *cobra.Command, args []string) {
	orgID, _ := cmd.Parent().PersistentFlags().GetUint64("oid")
	if orgID == 0 {
		log.Println("must set organization ID")
		return
	}
	body, err := dao.API.Do(fmt.Sprintf("/organization/%d", orgID), "GET", nil)
	if err != nil {
		log.Println("API GetOrganization", err)
		return
	}
	var resp apiio.GetOrganizationResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		log.Println("Unmarshal Response", err)
		return
	}
	if !resp.Success {
		log.Println(resp.Message)
		return
	}

	oldOrganizationXr, err := dao.API.GetOrganizationXRsa(orgID)
	if err != nil {
		log.Println("GetOrganizationXRsa", err)
		return
	}
	resp.Data.Organization.Name, err = oldOrganizationXr.PrivateDecrypt(resp.Data.Organization.Name)
	if err != nil {
		log.Println("PrivateDecrypt", err)
		return
	}
	var req apiio.OrgRequrest
	fmt.Println("!! let section empty will not change !!")
	fmt.Printf("Organization Name: (%s)", resp.Data.Organization.Name)
	fmt.Scanf("%s", &req.Name)
	if req.Name == "" {
		req.Name = resp.Data.Organization.Name
	}
	var reset string
	fmt.Print("Reset secret key (n/y):")
	fmt.Scanf("%s", &reset)
	if reset == "y" {
		publicKey, privateKey := bytes.NewBufferString(""), bytes.NewBufferString("")
		if err := xrsa.CreateKeys(publicKey, privateKey, 2048); err != nil {
			log.Println("CreateKeys", err)
			return
		}
		req.Pubkey = publicKey.String()
		newOrganizationXr, err := xrsa.NewXRsa(publicKey.Bytes(), privateKey.Bytes())
		if err != nil {
			log.Println("NewXRsa", err)
			return
		}
		req.Name, err = newOrganizationXr.PublicEncrypt(req.Name)
		if err != nil {
			log.Println("PublicEncrypt", err)
			return
		}
		// re encrypt organization_user data
		body, err = dao.API.Do(fmt.Sprintf("/organization/%d/user", orgID), "GET", req)
		if err != nil {
			log.Println("API Request", err)
			return
		}
		var respUser apiio.ListOrganizationUserResponse
		if err = json.Unmarshal(body, &respUser); err != nil {
			log.Println("API Request", string(body), err)
			return
		}
		if !respUser.Success {
			log.Println("API Request", respUser.Message)
			return
		}
		cs, err := xcrypto.NewCipherString(dao.Conf.User.Privatekey)
		if err != nil {
			log.Println("Load user privateKey", err)
			return
		}
		userPrivateKeyByte, err := cs.Decrypt(dao.Conf.MasterKey)
		if err != nil {
			log.Println("Decrypt user privateKey", err)
			return
		}
		for i := 0; i < len(respUser.Data.User); i++ {
			xr, err := xrsa.NewXRsa([]byte(respUser.Data.Key[respUser.Data.User[i].UserID]), userPrivateKeyByte)
			if err != nil {
				log.Println("NewXRsa", err)
				return
			}
			respUser.Data.User[i].PrivateKey, err = xr.PublicEncrypt(privateKey.String())
			if err != nil {
				log.Println("PublicEncrypt", err)
				return
			}
			req.Users = append(req.Users, respUser.Data.User[i])
		}
		// re encrypt organization_server data
		servers, err := dao.API.GetServers(orgID)
		if err != nil {
			log.Println("GetServers", err)
			return
		}
		for i := 0; i < len(servers); i++ {
			if err := xcrypto.EncryptStructWithXRsa(&servers[i], newOrganizationXr); err != nil {
				log.Println("EncryptStructWithXRsa", err)
				return
			}
			req.Servers = append(req.Servers, servers[i])
		}
	} else {
		req.Name, err = oldOrganizationXr.PublicEncrypt(req.Name)
		if err != nil {
			log.Println("PublicEncrypt", err)
			return
		}
	}

	body, err = dao.API.Do(fmt.Sprintf("/organization/%d", orgID), "PATCH", req)
	if err != nil {
		log.Println("API Request", err)
		return
	}
	var resp1 apiio.Response
	if err = json.Unmarshal(body, &resp1); err != nil {
		log.Println("API Request", string(body), err)
		return
	}
	if !resp1.Success {
		log.Println("API Request", resp1.Message)
		return
	}
	log.Println(resp1.Message)
}
