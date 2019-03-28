package main

import (
	"MonoPrinterTerminal/api"
	"MonoPrinterTerminal/config"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/mgo.v2"
)

var mongoGridFS mgo.GridFS
var mongoPrinterCollection mgo.Collection

func main() {
	var api api.API
	err := api.InitConfig()
	if err != nil {
		fmt.Println(err)
	}
	err = api.GetNewToken()
	if err != nil {
		fmt.Println(err)
	}
	err, files := api.GetFileList()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Files = ", files)
	// err := initAll()
	// if err != nil {
	// 	fmt.Println(err)
	// }

}
func initMongoDb(conf config.MongodbConf) {
	session, err := mgo.Dial(conf.Host)
	if err != nil {
		fmt.Println("Don't connect to mongodb")
	}
	session.SetMode(mgo.Monotonic, true)
	cP := session.DB(conf.DatabaseName).C("printers")
	grfs := session.DB(conf.DatabaseName).GridFS("fs")
	mongoGridFS = *grfs
	mongoPrinterCollection = *cP
}

func initAll() error {
	var conf config.Configuration
	err := conf.ParseConfig()
	if err != nil {
		return err
	}
	initMongoDb(conf.Databases.MongoDb)
	return nil

}
