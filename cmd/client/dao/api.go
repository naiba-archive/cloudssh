package dao

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/liamylian/x-rsa/golang/xrsa"
	"github.com/naiba/cloudssh/internal/apiio"
	"github.com/naiba/cloudssh/internal/model"
	"github.com/naiba/cloudssh/pkg/xcrypto"
	"github.com/shiena/ansicolor"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// APIClient ..
type APIClient struct {
}

// API ..
var API *APIClient

func init() {
	API = &APIClient{}
}

// GetUser ..
func (api *APIClient) GetUser(email string) (*apiio.UserInfoResponse, error) {
	body, err := api.Do("/user/?email="+email, "GET", nil)
	if err != nil {
		return nil, err
	}
	var resp apiio.UserInfoResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, err
	}
	return &resp, nil
}

// EditServer ..
func (api *APIClient) EditServer(id string) error {
	body, err := api.Do("/server/"+id, "GET", nil)
	if err != nil {
		return err
	}
	var resp apiio.GetServerResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return errors.New(resp.Message)
	}

	var req apiio.ServerRequest
	var orgXrsa *xrsa.XRsa
	if resp.Data.OwnerType == model.ServerOwnerTypeOrganization {
		req.OrganizationID = resp.Data.OwnerID
		orgXrsa, err = api.GetOrganizationXRsa(req.OrganizationID)
		if err != nil {
			return err
		}
		if err := xcrypto.DecryptStructWithXRsa(&resp.Data, orgXrsa); err != nil {
			return err
		}
	} else {
		if err := xcrypto.DecryptStruct(&resp.Data, Conf.MasterKey); err != nil {
			return err
		}
	}

	fmt.Println("!! let section empty will not change !!")
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
			return fmt.Errorf("No such login method: %s", req.LoginWith)
		}
	}
	fmt.Printf("Login Key (%s): ", resp.Data.Key)
	in := bufio.NewReader(os.Stdin)
	line, err := in.ReadString('\n')
	if err != nil {
		return err
	}
	req.Key = line[:len(line)-1]
	if req.Key == "" {
		req.Key = resp.Data.Key
	} else if req.LoginWith == model.ServerLoginWithAuthorizedKey {
		info, err := os.Stat(req.Key)
		if err != nil {
			return err
		}
		m := info.Mode()
		if m&(1<<2) == 0 {
			return fmt.Errorf("Can't read file, please check file permission: %s", req.Key)
		}
		kb, err := ioutil.ReadFile(req.Key)
		if err != nil {
			return err
		}
		req.Key = string(kb)
	}
	fmt.Printf("Server name(%s): ", resp.Data.Name)
	fmt.Scanf("%s", &req.Name)
	if req.Name == "" {
		req.Name = resp.Data.Name
	}

	if req.OrganizationID > 0 {
		if err := xcrypto.EncryptStructWithXRsa(&req, orgXrsa); err != nil {
			return err
		}
	} else {
		if err := xcrypto.EncryptStruct(&req, Conf.MasterKey); err != nil {
			return err
		}
	}

	body, err = api.Do("/server/"+id, "PATCH", req)
	if err != nil {
		return err
	}
	var resp1 apiio.Response
	if err = json.Unmarshal(body, &resp1); err != nil {
		return err
	}
	if !resp1.Success {
		return errors.New(resp1.Message)
	}
	log.Println(resp1.Message)
	return nil
}

// DialServer ..
func (api *APIClient) DialServer(orgID uint64, name, id string) {
	servers, err := api.GetServers(orgID)
	if err != nil {
		log.Println("API.GetServers", err)
		return
	}
	var server *model.Server
	for i := 0; i < len(servers); i++ {
		if (name != "" && servers[i].Name == name) || (id != "" && fmt.Sprintf("%d", servers[i].ID) == id) {
			server = &servers[i]
		}
	}
	if server == nil {
		log.Printf("server %s(%s) not found\n", name, id)
		return
	}
	var conf ssh.ClientConfig
	conf.Timeout = time.Second * 8
	conf.User = server.LoginUser
	conf.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	switch server.LoginWith {
	case model.ServerLoginWithAuthorizedKey:
		privateKey, err := ssh.ParsePrivateKey([]byte(server.Key))
		if err != nil {
			log.Println("ssh.ParsePrivateKey", err)
			return
		}
		conf.Auth = append(conf.Auth, ssh.PublicKeys(privateKey))
	case model.ServerLoginWithPassword:
		conf.Auth = append(conf.Auth, ssh.Password(server.Key))
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", server.IP, server.Port), &conf)
	if err != nil {
		log.Println("ssh.Dial", err)
		return
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		log.Println("client.NewSession", err)
		return
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	fd := int(os.Stdin.Fd())
	state, err := terminal.MakeRaw(fd)
	if err != nil {
		log.Printf("terminal make raw: %s", err)
		return
	}
	defer terminal.Restore(fd, state)

	w, h, err := terminal.GetSize(fd)
	if err != nil {
		log.Printf("terminal get size: %s", err)
		return
	}

	if err := session.RequestPty("xterm-256color", h, w, modes); err != nil {
		log.Printf("request for pseudo terminal failed: %s", err)
		return
	}

	session.Stdout = ansicolor.NewAnsiColorWriter(os.Stdout)
	session.Stderr = ansicolor.NewAnsiColorWriter(os.Stderr)
	session.Stdin = os.Stdin

	if err := session.Shell(); err != nil {
		log.Printf("failed to start shell: %s", err)
		return
	}

	if err := session.Wait(); err != nil {
		if e, ok := err.(*ssh.ExitError); ok {
			switch e.ExitStatus() {
			case 130:
				return
			}
		}
		log.Printf("ssh: %s", err)
	}
}

// BatchDeleteServer ..
func (api *APIClient) BatchDeleteServer(id []uint, organizationID uint64) {
	var req apiio.DeleteServerRequest
	req.ID = id
	req.OrganizationID = organizationID
	if len(req.ID) == 0 {
		log.Println("Please input server id list")
		return
	}
	fmt.Printf("Please type 'y' to confirm delete %+v:", req.ID)
	var confirm string
	fmt.Scanf("%s", &confirm)
	if confirm != "y" {
		return
	}

	body, err := api.Do("/server/batch-delete", "POST", req)
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

// GetServers ..
func (api *APIClient) GetServers(organizationID uint64) ([]model.Server, error) {
	var apiEndpoint string
	if organizationID > 0 {
		apiEndpoint = fmt.Sprintf("/organization/%d/server", organizationID)
	} else {
		apiEndpoint = "/server"
	}
	body, err := api.Do(apiEndpoint, "GET", nil)
	if err != nil {
		return nil, err
	}
	var resp apiio.ListServerResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, errors.New(resp.Message)
	}
	var orgXrsa *xrsa.XRsa
	if organizationID > 0 {
		orgXrsa, err = api.GetOrganizationXRsa(organizationID)
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < len(resp.Data); i++ {
		if organizationID > 0 {
			err := xcrypto.DecryptStructWithXRsa(&resp.Data[i], orgXrsa)
			if err != nil {
				return nil, err
			}
		} else {
			err := xcrypto.DecryptStruct(&resp.Data[i], Conf.MasterKey)
			if err != nil {
				return nil, err
			}
		}
	}
	return resp.Data, nil
}

// CreateServer ..
func (api *APIClient) CreateServer(organizationID uint64) {
	var req apiio.ServerRequest
	fmt.Print("Server IP: ")
	fmt.Scanf("%s", &req.IP)
	fmt.Print("SSH port: ")
	fmt.Scanf("%s", &req.Port)
	fmt.Print("Login user: ")
	fmt.Scanf("%s", &req.LoginUser)
	fmt.Print("Login method(1:apithorizedKey,2:password): ")
	fmt.Scanf("%s", &req.LoginWith)
	if req.LoginWith != model.ServerLoginWithAuthorizedKey && req.LoginWith != model.ServerLoginWithPassword {
		log.Println("No such login method:", req.LoginWith)
		return
	}
	fmt.Print("AuthorizedKey file path or password: ")
	in := bufio.NewReader(os.Stdin)
	line, err := in.ReadString('\n')
	if err != nil {
		log.Println("Read apithorizedKey file path or password", err)
		return
	}
	req.Key = line[:len(line)-1]
	if req.LoginWith == model.ServerLoginWithAuthorizedKey {
		info, err := os.Stat(req.Key)
		if err != nil {
			log.Println("Stat apithorizedKey file", err)
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
	fmt.Print("Server name: ")
	fmt.Scanf("%s", &req.Name)

	if organizationID > 0 {
		xr, err := api.GetOrganizationXRsa(organizationID)
		if err != nil {
			log.Println("GetOrganizationXRsa", err)
			return
		}
		if err := xcrypto.EncryptStructWithXRsa(&req, xr); err != nil {
			log.Println("EncryptStructWithXRsa", err)
			return
		}
		req.OrganizationID = organizationID
	} else {
		if err := xcrypto.EncryptStruct(&req, Conf.MasterKey); err != nil {
			log.Println("EncryptStruct", err)
			return
		}
	}

	body, err := api.Do("/server", "POST", req)
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

// GetOrganizationByID ..
func (api *APIClient) GetOrganizationByID(organizationID uint64) (*apiio.MyOrganizationInfo, error) {
	body, err := api.Do(fmt.Sprintf("/organization/%d", organizationID), "GET", nil)
	if err != nil {
		return nil, err
	}
	var resp apiio.GetOrganizationResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, errors.New(resp.Message)
	}
	return &resp.Data, nil
}

// Do ..
func (api *APIClient) Do(url, method string, data interface{}) ([]byte, error) {
	var req *http.Request
	var err error
	if data != nil {
		x, err := json.Marshal(&data)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest(method, Conf.Server+url, bytes.NewReader(x))
		if err != nil {
			return nil, err
		}
		req.Header.Set("content-type", "application/json")
	} else {
		req, err = http.NewRequest(method, Conf.Server+url, nil)
		if err != nil {
			return nil, err
		}
	}
	if Conf.User.TokenExpires.After(time.Now()) {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", Conf.User.Token))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// GetOrganizationXRsa ..
func (api *APIClient) GetOrganizationXRsa(id uint64) (*xrsa.XRsa, error) {
	info, err := api.GetOrganizationByID(id)
	if err != nil {
		return nil, err
	}
	xr, err := Conf.GerUserXRsa()
	if err != nil {
		return nil, err
	}
	orgPkStr, err := xr.PrivateDecrypt(info.OrganizationUser.PrivateKey)
	if err != nil {
		return nil, err
	}
	return xrsa.NewXRsa([]byte(info.Organization.Pubkey), []byte(orgPkStr))
}
