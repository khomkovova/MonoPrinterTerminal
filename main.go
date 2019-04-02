package main

import (
	"MonoPrinterTerminal/api"
	"MonoPrinterTerminal/config"
	"MonoPrinterTerminal/constant"
	"MonoPrinterTerminal/db"
	"MonoPrinterTerminal/uploadFile"
	"fmt"
	"io"
	"io/ioutil"
	"sync"
	"time"
	// "google.golang.org/api/file/v1beta1"
	"gopkg.in/mgo.v2"
	"os"
)

var mongoGridFS mgo.GridFS
var mongoPrinterCollection mgo.Collection
var wg sync.WaitGroup

func main() {

	err := initAll()
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Println("asdfasdfasdfasdfasdf")
	// testDb()
	wg.Add(1)
	go addNewFileCycle()
	go priningFileCycle()
	wg.Wait()
}

func addNewFileCycle() {

	defer wg.Done()
	var api api.API
	err := api.InitConfig()
	if err != nil {
		fmt.Println(err)
	}

	var database db.DB
	database.MongoGridFS = mongoGridFS
	database.MongoPrinterCollection = mongoPrinterCollection
	var uploadFile uploadFile.UploadFile

	for true {
		time.Sleep(constant.TIME_OFTEN_REPEAT_ADD_NEW_FILE)
		err, files := api.GetFileList()
		if err != nil {
			fmt.Println(err)
		}
		if len(files) < 1 {
			fmt.Println("Not files for this terminal")
			continue
		}
		fmt.Println("Files = ", files[0])
		uploadFile.Info = files[0]
		uploadFile.Info.Status = constant.STATUS_WAITING_FOR_PRINTING
		err, fileData := api.DownloadFile(files[0])
		if err != nil {
			fmt.Println(err)
			continue
		}

		newFile, err := os.Create(uploadFile.Info.UniqueId)
		defer newFile.Close()
		if err != nil {
			fmt.Println(err)
			continue
		}
		_, err = newFile.Write(fileData)
		if err != nil {
			fmt.Println(err)
			continue
		}

		newFile2, _ := os.Open(uploadFile.Info.UniqueId)
		defer newFile2.Close()
		uploadFile.File = newFile2
		err = database.AddFile(uploadFile)
		if err != nil {
			fmt.Println(err)
			continue
		}

		err = api.ChangeFileStatus(uploadFile.Info)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

}

func priningFileCycle() {
	layout := "2006-01-02T15:04:05"
	var api api.API
	err := api.InitConfig()
	if err != nil {
		fmt.Println(err)
	}
	var database db.DB
	database.MongoGridFS = mongoGridFS
	database.MongoPrinterCollection = mongoPrinterCollection
	err = api.GetNewToken()
	if err != nil {
		fmt.Println(err)
	}
	for true {
		time.Sleep(constant.TIME_OFTEN_REPEAT_PRINT_FILE)
		fmt.Println("Start priningFileCycle")
		err, files := database.GetFilesWaitingForPrinting()
		if err != nil {
			fmt.Println(err)
		}
		// fmt.Println(files)
		for _, file := range files {
			printingTime, err := time.Parse(layout, file.PrintingDate)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if time.Now().Add(constant.TIME_OVERSIGHT_FOR_PRINTING).Before(printingTime) {
				continue
			}
			if time.Now().After(printingTime.Add(constant.TIME_RETIRED_OLD_FILE)) {
				err = retiringFile(file) // TOO DO
				if err != nil {
					fmt.Println(err)
				}
				continue
			}
			err, gridFile := database.GetFile(file)
			if err != nil {
				fmt.Println(err)
			}

			// b, err := ioutil.ReadAll(gridFile)
			// if err != nil {
			// 	fmt.Println(err)
			// }
			prinitingFile, err := os.Create(file.UniqueId)
			if err != nil {
				fmt.Println(err)
			}
			_, err = io.Copy(prinitingFile, gridFile)
			if err != nil {
				fmt.Println(err)
			}
			prinitingFile.Close()
			err = printFile(file.UniqueId)
			if err != nil {
				fmt.Println(err)
			}
			_ = os.Remove(file.UniqueId)

			err = database.DeleteFile(file)
			if err != nil {
				fmt.Println(err)
			}
			file.Status = constant.STATUS_SUCCESSFUL_PRINTED
			err = api.ChangeFileStatus(file)
			if err != nil {
				fmt.Println(err)
			}

		}
	}

}



func retiringFile(fileInfo uploadFile.FileInfo) error {
	var database db.DB
	database.MongoGridFS = mongoGridFS
	database.MongoPrinterCollection = mongoPrinterCollection
	err := database.DeleteFile(fileInfo)
	if err != nil {
		fmt.Println(err)
	}
	var api api.API
	err = api.InitConfig()
	if err != nil {
		fmt.Println(err)
	}
	fileInfo.Status = constant.STATUS_WAITING_FOR_RETURN_PAGES
	err = api.ChangeFileStatus(fileInfo)
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

func printFile(fileName string) error {
	return nil
}


func initMongoDb(conf config.MongodbConf) {
	session, err := mgo.Dial(conf.Host)
	if err != nil {
		fmt.Println("Don't connect to mongodb")
	}
	session.SetMode(mgo.Monotonic, true)
	cP := session.DB(conf.DatabaseName).C("terminal")
	grfs := session.DB(conf.DatabaseName).GridFS("terminalFs")
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

func testApi() {
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

	err, fileData := api.DownloadFile(files[0])
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(fileData))

	fileNew := files[0]
	fileNew.Status = "STATUS_WAITING_FOR_PRINTING"
	err = api.ChangeFileStatus(fileNew)
	fmt.Println(err)
}

func testDb() {
	var api api.API
	err := api.InitConfig()
	if err != nil {
		fmt.Println(err)
	}
	var database db.DB
	database.MongoGridFS = mongoGridFS
	database.MongoPrinterCollection = mongoPrinterCollection
	var uploadFile uploadFile.UploadFile
	err = api.GetNewToken()
	if err != nil {
		fmt.Println(err)
	}
	err, files := api.GetFileList()
	if err != nil {
		fmt.Println(err)
	}
	if len(files) < 1 {
		fmt.Println("Not files for this terminal")
	}
	fmt.Println("Files = ", files[0])
	uploadFile.Info = files[0]
	err, fileData := api.DownloadFile(files[0])
	if err != nil {
		fmt.Println(err)
	}
	newFile, err := os.Create(uploadFile.Info.UniqueId)
	if err != nil {
		fmt.Println(err)
	}
	_, err = newFile.Write(fileData)
	if err != nil {
		fmt.Println(err)
	}

	newFile2, _ := os.Open(uploadFile.Info.UniqueId)
	defer newFile2.Close()
	uploadFile.File = newFile2
	database.AddFile(uploadFile)

	err, files = database.GetFilesWaitingForPrinting()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(files)

	err, gridFile := database.GetFile(files[len(files)-1])
	if err != nil {
		fmt.Println(err)
	}

	b, err := ioutil.ReadAll(gridFile)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(b))
	// err = database.DeleteFile(uploadFile.Info)

	// if err != nil {
	// 	fmt.Println(err)
	// }
	newStatus := uploadFile.Info
	newStatus.Status = "HAHAHAHAHAHA"
	database.UpdateFileStatus(newStatus)
	if err != nil {
		fmt.Println(err)
	}
}