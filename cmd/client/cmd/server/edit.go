package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/naiba/cloudssh/cmd/client/dao"
	"github.com/naiba/cloudssh/internal/apiio"
	"github.com/naiba/cloudssh/internal/model"
	"github.com/naiba/cloudssh/pkg/xcrypto"
	"github.com/spf13/cobra"
)

// EditCmd ..
var EditCmd *cobra.Command

func init() {
	EditCmd = &cobra.Command{
		Use:   "edit",
		Short: "Edit server",
	}
	EditCmd.Flags().StringP("id", "i", "", "server id")
	EditCmd.Run = edit
}

func edit(cmd *cobra.Command, args []string) {
	id, _ := cmd.Flags().GetString("id")
	if id == "" {
		cmd.Usage()
		return
	}
	body, err := dao.API.Do("/server/"+id, "GET", nil)
	if err != nil {
		log.Println("API Request", err)
		return
	}
	var resp apiio.GetServerResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		log.Println("API Request", string(body), err)
		return
	}
	if !resp.Success {
		log.Println("API Request", resp.Message)
		return
	}
	if err := xcrypto.DecryptStruct(&resp.Data, dao.Conf.MasterKey); err != nil {
		log.Println("DecryptStruct", resp.Message)
		return
	}

	var req apiio.ServerRequest
	fmt.Print("!! let section empty will not change !!")
	fmt.Printf("Server IP(%s): ", resp.Data.IP)
	fmt.Scanf("%s", &req.IP)
	if req.IP == "" {
		req.IP = resp.Data.IP
	}
	fmt.Printf("SSH port(%s): ", resp.Data.Port)
	fmt.Scanf("%s", &req.Port)
	if req.Port == "" {
		req.Port = resp.Data.Port
	}
	fmt.Printf("Login user(%s): ", resp.Data.LoginUser)
	fmt.Scanf("%s", &req.LoginUser)
	if req.LoginUser == "" {
		req.LoginUser = resp.Data.LoginUser
	}
	fmt.Printf("Login method(1:authorizedKey,2:password)(%s): ", resp.Data.LoginWith)
	fmt.Scanf("%s", &req.LoginWith)
	if req.LoginWith == "" {
		req.LoginWith = resp.Data.LoginWith
	} else {
		if req.LoginWith != model.ServerLoginWithAuthorizedKey && req.LoginWith != model.ServerLoginWithPassword {
			log.Println("No such login method:", req.LoginWith)
			return
		}
	}
	fmt.Printf("Login Key (%s): ", resp.Data.Key)
	in := bufio.NewReader(os.Stdin)
	line, err := in.ReadString('\n')
	if err != nil {
		log.Println("Read authorizedKey file path or password", err)
		return
	}
	req.Key = line[:len(line)-1]
	if req.Key == "" {
		req.Key = resp.Data.Key
	} else if req.LoginWith == model.ServerLoginWithAuthorizedKey {
		info, err := os.Stat(req.Key)
		if err != nil {
			log.Println("Stat authorizedKey file", err)
			return
		}
		m := info.Mode()
		if m&(1<<2) == 0 {
			log.Println("Can't read file, please check file permission", req.Key)
			return
		}
		kb, err := ioutil.ReadFile(req.Key)
		if err != nil {
			log.Println("Read PrivateKey", err)
			return
		}
		req.Key = string(kb)
	}
	fmt.Printf("ServerName(%s): ", resp.Data.Name)
	fmt.Scanf("%s", &req.Name)
	if req.Name == "" {
		req.Name = resp.Data.Name
	}

	if err := xcrypto.EncryptStruct(&req, dao.Conf.MasterKey); err != nil {
		log.Println("EncryptStruct", err)
		return
	}

	body, err = dao.API.Do("/server/"+id, "PATCH", req)
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
