package server

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/shiena/ansicolor"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/naiba/cloudssh/cmd/client/dao"
	"github.com/naiba/cloudssh/internal/model"
)

// DialCmd ..
var DialCmd *cobra.Command

func init() {
	DialCmd = &cobra.Command{
		Use:   "dial",
		Short: "Connect to server, you must set server's name or id",
	}
	DialCmd.Run = dial
	DialCmd.Flags().StringP("name", "n", "", "server name")
	DialCmd.Flags().StringP("id", "i", "", "server id")
}

func dial(cmd *cobra.Command, args []string) {
	name, _ := cmd.Flags().GetString("name")
	id, _ := cmd.Flags().GetString("id")
	if name == "" && id == "" {
		log.Println("You must set which server you want to connect")
		return
	}
	servers, err := dao.API.GetServers()
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
	conf.User = server.User
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
