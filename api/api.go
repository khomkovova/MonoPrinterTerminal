package api

import (
	"MonoPrinter/rsaparser"
	"MonoPrinterTerminal/config"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type FileInfo struct {
	UniqueId     string
	Filename     string
	PrintingDate string

	UploadDate string
	NumberPage int
	Size       string

	IdPrinter int
	Status    string
}

type API struct {
	Url        string `json:"url"`
	Token      string `json:"token`
	TerminalId int    `json:"terminalid"`
}

func (api *API) InitConfig() error {
	var conf config.Configuration
	err := conf.ParseConfig()
	if err != nil {
		return err
	}
	api.Url = conf.Api.Url
	api.TerminalId = conf.Api.TerminalId
	return nil
}

func (api *API) GetNewToken() error {
	pubPEM, err := ioutil.ReadFile("config/terminalPublicKey.key")
	if err != nil {
		return err
	}
	publicKey, err := rsaparser.ParseRsaPublicKeyFromPemStr(string(pubPEM))
	if err != nil {
		return errors.New("Bad public key")
	}
	layout := "2006-01-02T15:04:05"
	message := []byte("{\"terminalId\":" + strconv.Itoa(api.TerminalId) + ", \"createDate\":\"" + time.Now().Add(time.Minute*20).Format(layout) + "\"}")
	fmt.Println("Plain token = " + string(message))
	label := []byte("")
	hash := sha256.New()
	ciphertext, _ := rsa.EncryptOAEP(hash, rand.Reader, publicKey, message, label)

	sEnc := base64.StdEncoding.EncodeToString(ciphertext)
	_ = ioutil.WriteFile("config/terminalToken.key", []byte(sEnc), 0644)
	api.Token = sEnc
	fmt.Println("Token = ", sEnc)

	return nil
}

func (api *API) GetFileList() (err error, files []FileInfo) {
	err = api.GetNewToken()
	if err != nil {
		return errors.New("Token don't get "), files
	}
	link := "/api/terminal/files"
	req, err := http.NewRequest("GET", api.Url+link, nil)
	if err != nil {
		return err, files
	}
	cookie := http.Cookie{Name: "token", Value: api.Token}
	req.AddCookie(&cookie)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return err, files
	}
	data, _ := ioutil.ReadAll(response.Body)
	err = json.Unmarshal(data, &files)
	if err != nil {
		return err, files
	}
	return nil, files
}
