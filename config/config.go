package config

import (
	"encoding/json"
	"io/ioutil"
)

type Configuration struct {
	Databases DatabasesConf `json:"Databases"`
	Api       ApiConf       `json:"Api`
}

type DatabasesConf struct {
	MongoDb MongodbConf `json:"MongoDb"`
}
type ApiConf struct {
	Url        string `json:"Url"`
	TerminalId int    `json:"TerminalId"`
}

type MongodbConf struct {
	DatabaseName string `json:"DatabaseName"`
	Host         string `json:"Host"`
	Username     string `json:"Username"`
	Password     string `json:"Password"`
}

func (config *Configuration) ParseConfig() error {

	data, err := ioutil.ReadFile("config/config.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(data), config)
	if err != nil {
		return err
	}
	return nil
}
