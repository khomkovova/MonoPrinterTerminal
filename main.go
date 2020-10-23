package main

import (
	"errors"
	"github.com/khomkovova/MonoPrinterTerminal/api"
	"github.com/khomkovova/MonoPrinterTerminal/constant"
	"github.com/khomkovova/MonoPrinterTerminal/helper"
	"github.com/khomkovova/MonoPrinterTerminal/storage_helper"
	"github.com/khomkovova/MonoPrinterTerminal/storage_helper/csv_helper"
	"github.com/khomkovova/MonoPrinterTerminal/uploadFile"
	"gopkg.in/mgo.v2"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
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

	var newFile uploadFile.UploadFile

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
		newFile.Info = files[0]
		newFile.Info.Status = constant.STATUS_WAITING_FOR_PRINTING
		err, fileData := api.DownloadFile(files[0])
		if err != nil {
			helper.LogErrorMsg(err, "", loggerAddNewFileCycle)
			continue
		}

		file, err := os.Create(newFile.Info.UniqueId)
		if err != nil {
			helper.LogErrorMsg(err, "", loggerAddNewFileCycle)
			continue
		}
		defer file.Close()

		_, err = file.Write(fileData)
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
		err = storage.AddFile(newFile)
		if err != nil {
			helper.LogErrorMsg(err, "", loggerAddNewFileCycle)
			continue
		}
		//err = os.Remove(uploadFile.Info.UniqueId)
		//if err != nil {
		//	helper.LogErrorMsg(err, "")
		//	continue
		//}

		err = api.ChangeFileStatus(newFile.Info)
		if err != nil {
			helper.LogErrorMsg(err, "Critical", loggerCritical)
			continue
		}
		helper.LogInfoMsg("Successfully added file to queue :" + newFile.Info.UniqueId, loggerAddNewFileCycle)
		printingTime, err := time.Parse(constant.TIME_LAYOUT, newFile.Info.PrintingDate)
		if err != nil {
			helper.LogErrorMsg(err, "", loggerAddNewFileCycle)
			continue
		}
		helper.LogInfoMsg("Printing time for " + newFile.Info.UniqueId + " is " +  printingTime.Local().Format("2006-01-02 15:04:05"), loggerAddNewFileCycle)
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
	var api api.API
	err := api.InitConfig()
	if err != nil {
		return err
	}
	fileInfo.Status = constant.STATUS_WAITING_FOR_RETURN_PAGES
	err = api.ChangeFileStatus(fileInfo)
	if err != nil {
		return err
	}
	err = storage.DeleteFile(fileInfo.UniqueId)
	if err != nil {
		return err
	}
	err = os.Remove(fileInfo.UniqueId)
	if err != nil {
		helper.LogErrorMsg(err, "Critical", loggerCritical)
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