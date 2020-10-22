package main

import (
	"github.com/khomkovova/MonoPrinterTerminal/api"
	"github.com/khomkovova/MonoPrinterTerminal/config"
	"github.com/khomkovova/MonoPrinterTerminal/constant"
	"github.com/khomkovova/MonoPrinterTerminal/db"
	"github.com/khomkovova/MonoPrinterTerminal/helper"
	"github.com/khomkovova/MonoPrinterTerminal/uploadFile"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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

func main() {
	err := initAll()
	if err != nil {
		fmt.Println(err)
	}
	// testDb()
	// testApi()
	wg.Add(1)
	go addNewFileCycle()
	//go priningFileCycle()
	wg.Wait()
}

func addNewFileCycle() {
	logfile, err := os.OpenFile("addNewFileCycleLog.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Println("error opening file: %v", err)
	}
	defer logfile.Close()

	Log := log.New(logfile, "", log.LstdFlags|log.Lshortfile)

	Log.Println("Start addNewFileCycle()")
	defer wg.Done()
	var api api.API
	err = api.InitConfig()
	if err != nil {
		Log.Println("Error in api.InitConfig()")
		Log.Println(err)
	}

	var database db.DB
	database.MongoGridFS = mongoGridFS
	database.MongoPrinterCollection = mongoPrinterCollection
	var uploadFile uploadFile.UploadFile

	for true {
		Log.Println("Start cycle")
		time.Sleep(constant.TIME_OFTEN_REPEAT_ADD_NEW_FILE)
		err, files := api.GetFileList()
		if err != nil {
			continue
		}
		if len(files) < 1 {
			helper.LogInfoMsg("Aren't files for this terminal")
			continue
		}
		Log.Println("Files = ", files)
		uploadFile.Info = files[0]
		uploadFile.Info.Status = constant.STATUS_WAITING_FOR_PRINTING
		err, fileData := api.DownloadFile(files[0])
		if err != nil {
			log.Println(err)
			continue
		}

		newFile, err := os.Create(uploadFile.Info.UniqueId)
		if err != nil {
			helper.LogErrorMsg(err, "")
			continue
		}
		defer newFile.Close()

		_, err = newFile.Write(fileData)
		if err != nil {
			helper.LogErrorMsg(err, "")
			continue
		}

		newFile2, err := os.Open(uploadFile.Info.UniqueId)
		if err != nil {
			helper.LogErrorMsg(err, "")
			continue
		}
		defer newFile2.Close()
		uploadFile.File = newFile2
		err = database.AddFile(uploadFile)
		if err != nil {
			helper.LogErrorMsg(err, "")
			continue
		}
		//err = os.Remove(uploadFile.Info.UniqueId)
		if err != nil {
			helper.LogErrorMsg(err, "")
			continue
		}

		err = api.ChangeFileStatus(uploadFile.Info)
		if err != nil {
			helper.LogErrorMsg(err, "")
			continue
		}
		Log.Println("End cycle\n\n\n")
	}

}

func priningFileCycle() {
	logfile, err := os.OpenFile("priningFileCycle.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Println("error opening file: %v", err)
	}
	defer logfile.Close()

	Log := log.New(logfile, "", log.LstdFlags|log.Lshortfile)

	Log.Println("Start priningFileCycle()")

	layout := "2006-01-02T15:04:05"
	var api api.API
	err = api.InitConfig()
	if err != nil {
		Log.Println("Error in api.InitConfig()")
		Log.Println(err)
	}
	var database db.DB
	database.MongoGridFS = mongoGridFS
	database.MongoPrinterCollection = mongoPrinterCollection
	err = api.GetNewToken()
	if err != nil {
		Log.Println("Error in api.GetNewToken()")
		Log.Println(err)
	}
	for true {
		time.Sleep(constant.TIME_OFTEN_REPEAT_PRINT_FILE)
		Log.Println("Start priningFileCycle")
		err, files := database.GetFilesWaitingForPrinting()
		if err != nil {
			Log.Println("Error in  database.GetFilesWaitingForPrinting()")
			Log.Println(err)
		}
		// Log.Println(files)
		for _, file := range files {
			printingTime, err := time.Parse(layout, file.PrintingDate)
			if err != nil {
				Log.Println(err)
				continue
			}
			if time.Now().Add(constant.TIME_OVERSIGHT_FOR_PRINTING).Before(printingTime) {

				continue
			}
			if time.Now().After(printingTime.Add(constant.TIME_RETIRED_OLD_FILE)) {
				err = retiringFile(file) // TOO DO
				if err != nil {
					Log.Println("Error in retiringFile(file)")
					Log.Println(err)
				}
				continue
			}

			delta := time.Now().Sub(printingTime).Seconds()
			if delta < -10 || delta > 10 {
				//Log.Println(delta)
				//Log.Println(time.Now().UTC().Format("2006-01-02T15:04:05"))
				//Log.Println(printingTime.UTC().Format("2006-01-02T15:04:05"))
				//Log.Println("Wait for a time")
				continue
			}
			err, gridFile := database.GetFile(file)
			if err != nil {
				Log.Println("Error in database.GetFile(file)")
				Log.Println(err)
				continue
			}

			// b, err := ioutil.ReadAll(gridFile)
			// if err != nil {
			// 	Log.Println(err)
			// }
			prinitingFile, err := os.Create(file.UniqueId)
			if err != nil {
				Log.Println(err)
				continue
			}
			_, err = io.Copy(prinitingFile, gridFile)
			if err != nil {
				Log.Println(err)
				continue
			}
			prinitingFile.Close()
			Log.Println("*************************************************************")
			Log.Println("Start Print File")
			Log.Println("*************************************************************")
			Log.Println("File name for printing: ", file.UniqueId)
			err = printFile(file.UniqueId)
			if err != nil {
				Log.Println("Error in printFile(file.UniqueId); ", file.UniqueId)
				Log.Println(err)
				continue
			}
			_ = os.Remove(file.UniqueId)

			err = database.DeleteFile(file)
			if err != nil {
				Log.Println(err)
				continue
			}
			file.Status = constant.STATUS_SUCCESSFUL_PRINTED
			err = api.ChangeFileStatus(file)
			if err != nil {
				Log.Println("Error in api.ChangeFileStatus()")
				Log.Println(err)
				continue
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
	cmd := exec.Command("bash", "-c",  "./print_file.sh " + fileName)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New("cmd.Run() failed with: " + err.Error())
	}
	//fmt.Printf("combined out:\n%s\n", string(out))
	if strings.Contains(string(out), "Successful"){
		//fmt.Printf("combined out:\n%s\n", string(out))
		return nil
	}
	return errors.New("Can't print file")
}


func initMongoDb(conf config.MongodbConf) error {
	session, err := mgo.Dial(conf.Host)
	if err != nil {
		return err
	}
	session.SetMode(mgo.Monotonic, true)
	cP := session.DB(conf.DatabaseName).C("terminal")
	grfs := session.DB(conf.DatabaseName).GridFS("terminalFs")
	mongoGridFS = *grfs
	mongoPrinterCollection = *cP
	return nil
}

func initAll() error {
	var conf config.Configuration
	err := conf.ParseConfig()
	if err != nil {
		return err
	}
	err = initMongoDb(conf.Databases.MongoDb)
	if err != nil {
		return err
	}
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