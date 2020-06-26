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

// CreateCmd ..
var CreateCmd *cobra.Command

func init() {
	CreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create server",
	}
	CreateCmd.Run = create
}

func create(cmd *cobra.Command, args []string) {
	var req apiio.NewServerRequest

	fmt.Print("Please type your server IP: ")
	fmt.Scanf("%s", &req.IP)
	fmt.Print("Please type SSH port: ")
	fmt.Scanf("%s", &req.Port)
	fmt.Print("Please type login user: ")
	fmt.Scanf("%s", &req.User)
	fmt.Print("Please type login method(1:authorizedKey,2:password): ")
	fmt.Scanf("%s", &req.LoginWith)
	if req.LoginWith != model.ServerLoginWithAuthorizedKey && req.LoginWith != model.ServerLoginWithPassword {
		log.Println("No such login method:", req.LoginWith)
		return
	}
	fmt.Print("Please type authorizedKey file path or password: ")
	in := bufio.NewReader(os.Stdin)
	line, err := in.ReadString('\n')
	if err != nil {
		log.Println("Read authorizedKey file path or password", err)
		return
	}
	req.Key = line[:len(line)-1]
	if req.LoginWith == model.ServerLoginWithAuthorizedKey {
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
	fmt.Print("Please type a server note: ")
	fmt.Scanf("%s", &req.Name)

	if err := xcrypto.EncryptStruct(&req, dao.Conf.MasterKey); err != nil {
		log.Println("EncryptStruct", err)
		return
	}

	body, err := dao.API.Do("/user/server", "POST", req)
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
