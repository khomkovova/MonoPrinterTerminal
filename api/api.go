package api

import (
	"MonoPrinter/rsaparser"
	"MonoPrinterTerminal/config"
	"MonoPrinterTerminal/uploadFile"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	// "fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

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
	// fmt.Println("Plain token = " + string(message))
	label := []byte("")
	hash := sha256.New()
	ciphertext, _ := rsa.EncryptOAEP(hash, rand.Reader, publicKey, message, label)

	sEnc := base64.StdEncoding.EncodeToString(ciphertext)
	_ = ioutil.WriteFile("config/terminalToken.key", []byte(sEnc), 0644)
	api.Token = sEnc
	// fmt.Println("Token = ", sEnc)

	return nil
}

func (api *API) ChangeFileStatus(fileInfo uploadFile.FileInfo) (err error) {
	err = api.GetNewToken()
	if err != nil {
		return errors.New("Token don't get ")
	}
	link := "/api/terminal/files?uniqueid=" + fileInfo.UniqueId
	type status struct {
		Status string `json:"Status"`
	}
	var st status
	st.Status = fileInfo.Status
	dataB, err := json.Marshal(st)

	if err != nil {
		return errors.New("Bad json")
	}
	req, err := http.NewRequest("PUT", api.Url+link, bytes.NewBuffer(dataB))
	if err != nil {
		return err
	}
	cookie := http.Cookie{Name: "token", Value: api.Token}
	req.AddCookie(&cookie)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	data, _ := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if len(data) < 2 {
		return errors.New("Response is veru small")
	}
	return nil
}

func (api *API) DownloadFile(fileInfo uploadFile.FileInfo) (err error, fileData []byte) {
	err = api.GetNewToken()
	if err != nil {
		return errors.New("Token don't get "), fileData
	}
	link := "/api/terminal/files?uniqueid=" + fileInfo.UniqueId
	req, err := http.NewRequest("GET", api.Url+link, nil)
	if err != nil {
		return err, fileData
	}
	cookie := http.Cookie{Name: "token", Value: api.Token}
	req.AddCookie(&cookie)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return err, fileData
	}
	fileData, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return err, fileData
	}

	return nil, fileData
}

func (api *API) GetFileList() (err error, files []uploadFile.FileInfo) {
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
