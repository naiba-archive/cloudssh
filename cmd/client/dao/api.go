package dao

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/naiba/cloudssh/internal/apiio"
	"github.com/naiba/cloudssh/internal/model"
	"github.com/naiba/cloudssh/pkg/xcrypto"
)

// APIClient ..
type APIClient struct {
}

// API ..
var API *APIClient

func init() {
	API = &APIClient{}
}

// GetServers ..
func (au *APIClient) GetServers() ([]model.Server, error) {
	body, err := au.Do("/server", "GET", nil)
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
	for i := 0; i < len(resp.Data); i++ {
		err := xcrypto.DecryptStruct(&resp.Data[i], Conf.MasterKey)
		if err != nil {
			return nil, err
		}
	}
	return resp.Data, nil
}

// Do ..
func (au *APIClient) Do(url, method string, data interface{}) ([]byte, error) {
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
