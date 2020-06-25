package dao

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	home "github.com/mitchellh/go-homedir"
	"github.com/naiba/cloudssh/internal/model"
	"github.com/naiba/cloudssh/pkg/xcrypto"
)

// Config ..
type Config struct {
	User      model.User
	MasterKey xcrypto.CryptoKey
	Server    string
}

// Conf ..
var Conf *Config
var confFile string

// InitConfig ..
func InitConfig() {
	Conf = &Config{}
	homeDir, err := home.Dir()
	if err != nil {
		panic(err)
	}
	confFile = homeDir + "/.cloudssh.json"
	content, err := ioutil.ReadFile(confFile)
	if err != nil {
		log.Println("InitConfig ReadFile error:", err)
		return
	}
	if err = json.Unmarshal(content, Conf); err != nil {
		log.Println("InitConfig Unmarshal error:", err)
	}
}

// Save ..
func (c *Config) Save() error {
	content, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(confFile, content, os.FileMode(0655))
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
