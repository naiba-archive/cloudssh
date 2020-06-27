package dao

import (
	"encoding/json"
	"io/ioutil"
)

// Config ..
type Config struct {
	Debug  bool
	DBDSN  string
	Domain string
}

// Conf ..
var Conf *Config

// InitConfig ..
func InitConfig(filepath string) error {
	Conf = &Config{}
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	return json.Unmarshal(content, Conf)
}
