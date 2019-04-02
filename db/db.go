package db

import (
	"MonoPrinterTerminal/uploadFile"
	"MonoPrinterTerminal/constant"
	"errors"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io"
	// "io/ioutil"
	"os"
)

type DB struct {
	MongoGridFS            mgo.GridFS
	MongoPrinterCollection mgo.Collection
}

func (db *DB) GetFilesWaitingForPrinting() (err error, files []uploadFile.FileInfo) {
	db.MongoPrinterCollection.Find(bson.M{"status":constant.STATUS_WAITING_FOR_PRINTING }).All(&files)
	return nil, files
}

func (db *DB) GetFile(fileInfo uploadFile.FileInfo) (err error, gridFile *mgo.GridFile) {
	fileFs, err := db.MongoGridFS.Open(fileInfo.UniqueId)
	if err != nil {
		fmt.Println("Not found file in mongo")
		return err, fileFs
	}
	return nil, fileFs
}

func (db *DB) AddFile(newFile uploadFile.UploadFile) (err error) {
	mongoFile, err := db.MongoGridFS.Create(newFile.Info.UniqueId)
	if err != nil {
		return errors.New("Not create file")
	}
	_, err = io.Copy(mongoFile, newFile.File)
	if err != nil {
		fmt.Println("Not create new file")
		return err
	}

	_ = mongoFile.Close()
	_ = newFile.File.Close()
	_ = os.Remove(newFile.Info.UniqueId)
	err = db.MongoPrinterCollection.Insert(newFile.Info)
	if err != nil {
		fmt.Println("Not insert new file info")
		return err
	}
	return nil
}

func (db *DB) DeleteFile(fileInfo uploadFile.FileInfo) (err error) {
	changeInfo, err := db.MongoPrinterCollection.RemoveAll(bson.M{"uniqueid": fileInfo.UniqueId})
	if err != nil {
		return errors.New("Not deleted file from printer collection")
	}
	fmt.Println(changeInfo)

	err = db.MongoGridFS.Remove(fileInfo.UniqueId)
	if err != nil {
		return errors.New("Not deleted file from printerfs collection")
	}
	return nil
}

func (db *DB) UpdateFileStatus(fileInfo uploadFile.FileInfo) (err error) {
	_, err = db.MongoPrinterCollection.UpdateAll(bson.M{"uniqueid": fileInfo.UniqueId}, bson.M{"$set": bson.M{"status": fileInfo.Status}})
	if err != nil {
		return errors.New("Not update file status")
	}
	return nil
}
