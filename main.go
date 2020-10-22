package main

import (
	"errors"
	"github.com/khomkovova/MonoPrinterTerminal/api"
	"github.com/khomkovova/MonoPrinterTerminal/constant"
	"github.com/khomkovova/MonoPrinterTerminal/db"
	"github.com/khomkovova/MonoPrinterTerminal/helper"
	"github.com/khomkovova/MonoPrinterTerminal/storage_helper"
	"github.com/khomkovova/MonoPrinterTerminal/storage_helper/csv_helper"
	"github.com/khomkovova/MonoPrinterTerminal/uploadFile"
	"os/exec"
	"strings"
	"sync"
	"time"
	// "google.golang.org/api/file/v1beta1"
	"gopkg.in/mgo.v2"
	"os"
)

var mongoGridFS mgo.GridFS
var mongoPrinterCollection mgo.Collection
var wg sync.WaitGroup
var loggerCritical = helper.InitLogger("log/criticalLog.txt")
var loggerAddNewFileCycle =  helper.InitLogger("log/addNewFileCycleLog.txt")
var loggerPriningFileCycle =  helper.InitLogger("log/priningFileCycle.txt")


func main() {
	wg.Add(1)
	go addNewFileCycle()
	go priningFileCycle()
	wg.Wait()
}

func addNewFileCycle() {
	helper.LogInfoMsg("Start addNewFileCycle()", loggerAddNewFileCycle)
	defer wg.Done()
	var api api.API
	err := api.InitConfig()
	if err != nil {
		helper.LogErrorMsg(err, "", loggerAddNewFileCycle)
	}

	var database db.DB
	database.MongoGridFS = mongoGridFS
	database.MongoPrinterCollection = mongoPrinterCollection
	var uploadFile uploadFile.UploadFile

	for true {
		time.Sleep(constant.TIME_OFTEN_REPEAT_ADD_NEW_FILE)
		err, files := api.GetFileList()
		if err != nil {
			continue
		}
		if len(files) < 1 {
			//helper.LogInfoMsg("Aren't files for this terminal", logger)
			continue
		}
		//logger.Println("Files = ", files)
		uploadFile.Info = files[0]
		uploadFile.Info.Status = constant.STATUS_WAITING_FOR_PRINTING
		err, fileData := api.DownloadFile(files[0])
		if err != nil {
			helper.LogErrorMsg(err, "", loggerAddNewFileCycle)
			continue
		}

		newFile, err := os.Create(uploadFile.Info.UniqueId)
		if err != nil {
			helper.LogErrorMsg(err, "", loggerAddNewFileCycle)
			continue
		}
		defer newFile.Close()

		_, err = newFile.Write(fileData)
		if err != nil {
			helper.LogErrorMsg(err, "", loggerAddNewFileCycle)
			continue
		}

		//newFile2, err := os.Open(uploadFile.Info.UniqueId)
		//if err != nil {
		//	helper.LogErrorMsg(err, "")
		//	continue
		//}
		//defer newFile2.Close()
		//uploadFile.File = newFile2

		var storage storage_helper.Storage
		err = storage.AddFile(uploadFile)
		if err != nil {
			helper.LogErrorMsg(err, "", loggerAddNewFileCycle)
			continue
		}
		//err = os.Remove(uploadFile.Info.UniqueId)
		//if err != nil {
		//	helper.LogErrorMsg(err, "")
		//	continue
		//}

		err = api.ChangeFileStatus(uploadFile.Info)
		if err != nil {
			helper.LogErrorMsg(err, "Critical", loggerCritical)
			continue
		}
		helper.LogInfoMsg("Successfully added file to queue :" + uploadFile.Info.UniqueId, loggerAddNewFileCycle)
		printingTime, err := time.Parse(constant.TIME_LAYOUT, uploadFile.Info.PrintingDate)
		if err != nil {
			helper.LogErrorMsg(err, "", loggerAddNewFileCycle)
			continue
		}
		helper.LogInfoMsg("Printing time for " + uploadFile.Info.UniqueId + " is " +  printingTime.Local().Format("2006-01-02 15:04:05"), loggerAddNewFileCycle)
	}

}

func priningFileCycle() {

	helper.LogInfoMsg("Start addNewFileCycle()", loggerPriningFileCycle)


	var api api.API
	err := api.InitConfig()
	if err != nil {
		helper.LogErrorMsg(err, "", loggerPriningFileCycle)
		return
	}

	var csvFilesInfo csv_helper.CSV


	for true {
		time.Sleep(constant.TIME_OFTEN_REPEAT_PRINT_FILE)
		err = api.GetNewToken()
		if err != nil {
			helper.LogErrorMsg(err, "", loggerPriningFileCycle)
			continue
		}
		err, filesInfo := csvFilesInfo.GetFilesInfo()
		if err != nil {
			helper.LogErrorMsg(err, "", loggerPriningFileCycle)
			continue
		}
		for _, fileInfo := range filesInfo {
			printingTime, err := time.Parse(constant.TIME_LAYOUT, fileInfo.PrintingDate)
			if err != nil {
				helper.LogErrorMsg(err, "", loggerPriningFileCycle)
				continue
			}
			if time.Now().Add(constant.TIME_OVERSIGHT_FOR_PRINTING).Before(printingTime) {
				continue
			}
			if time.Now().After(printingTime.Add(constant.TIME_RETIRED_OLD_FILE)) {
				err = retiringFile(fileInfo) // TOO DO
				if err != nil {
					helper.LogErrorMsg(err, "Critical", loggerCritical)
				}
				continue
			}

			delta := time.Now().Sub(printingTime).Seconds()
			if delta < -10 || delta > 10 {
				//fmt.Println(delta)
				//fmt.Println(time.Now().UTC().Format("2006-01-02T15:04:05"))
				//log.Printf("Printing time for %s is %s", fileInfo.UniqueId, printingTime.Local().Format("2006-01-02 15:04:05"))
				//fmt.Println("Wait for a time")
				continue
			}

			helper.LogInfoMsg("Start Print File: " + fileInfo.UniqueId, loggerPriningFileCycle)
			err = printFile(fileInfo.UniqueId)
			if err != nil {
				helper.LogErrorMsg(err, "Critical", loggerCritical)
				fileInfo.Status = constant.STATUS_ERROR_WITH_PRINTING
			}else{
				fileInfo.Status = constant.STATUS_SUCCESSFUL_PRINTED

			}
			err = os.Remove(fileInfo.UniqueId)
			if err != nil {
				helper.LogErrorMsg(err, "Critical", loggerCritical)
			}
			err = csvFilesInfo.DeleteRow(fileInfo.UniqueId)
			if err != nil {
				helper.LogErrorMsg(err, "Critical", loggerCritical)
				continue
			}

			err = api.ChangeFileStatus(fileInfo)
			if err != nil {
				helper.LogErrorMsg(err, "Critical", loggerCritical)
				continue
			}

		}
	}

}



func retiringFile(fileInfo uploadFile.FileInfo) error {
	var storage storage_helper.Storage
	err := storage.DeleteFile(fileInfo.UniqueId)
	if err != nil {
		return err
	}
	var api api.API
	err = api.InitConfig()
	if err != nil {
		return err
	}
	fileInfo.Status = constant.STATUS_WAITING_FOR_RETURN_PAGES
	err = api.ChangeFileStatus(fileInfo)
	if err != nil {
		return err
	}
	return nil
}

func printFile(fileName string) error {
	//return nil
	cmd := exec.Command("bash", "-c",  "./print_file.sh '" + fileName + "'")

	out, err := cmd.CombinedOutput()
	helper.LogInfoMsg("print_file.sh output: " +  string(out), loggerPriningFileCycle)
	if err != nil {
		helper.LogErrorMsg(err, "Critical", loggerPriningFileCycle)
		return err
	}

	if strings.Contains(string(out), "Successful"){
		return nil
	}
	helper.LogErrorMsg(errors.New("print_file.sh can't print file"), "Critical", loggerPriningFileCycle)
	return errors.New("print_file.sh can't print file")
}