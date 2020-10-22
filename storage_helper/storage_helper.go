package storage_helper

import (
	"github.com/khomkovova/MonoPrinterTerminal/uploadFile"

	"os"
)
import "github.com/khomkovova/MonoPrinterTerminal/storage_helper/csv_helper"
type Storage struct {
	CSV csv_helper.CSV

}

func (storage *Storage) AddFile(newFile uploadFile.UploadFile) error {
	err := storage.CSV.AddRow(newFile.Info)
	if err != nil{
		return err
	}
	return nil
}


func (storage *Storage) DeleteFile(filename string) error {
	_ = os.Remove(filename)
	err := storage.CSV.DeleteRow(filename)
	if err != nil{
		return err
	}
	return nil
}