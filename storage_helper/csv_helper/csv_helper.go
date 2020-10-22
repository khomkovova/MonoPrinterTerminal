package csv_helper
import (
	"encoding/csv"
	"fmt"

	"os"
	"github.com/khomkovova/MonoPrinterTerminal/uploadFile"
)


var CSV_FILENAME = "files_db.csv"
type CSV struct {
	//FileName string
}

func (csvfile * CSV) DeleteRow(filename string) ( error) {
	if _, err := os.Stat(CSV_FILENAME)
		err != nil {
		_, err := os.Create(CSV_FILENAME)
		if err != nil {
			fmt.Println(err)
		}
	}
	err, filesInfo := csvfile.GetFilesInfo()
	if err != nil{
		return err
	}
	for i, fileInfo := range(filesInfo){
		if fileInfo.UniqueId == filename {
			filesInfo = append(filesInfo[:i], filesInfo[i+1:]...)
		}
	}
	err = csvfile.UpdateFilesInfo(filesInfo)
	if err != nil{
		return err
	}
	return nil
}
func (csvfile * CSV) AddRow(fileInfoCSV uploadFile.FileInfo) ( error) {
	if _, err := os.Stat(CSV_FILENAME)
	err != nil {
		_, err := os.Create(CSV_FILENAME)
		if err != nil {
			fmt.Println(err)
		}
	}
	err, filesInfoCSV := csvfile.GetFilesInfo()
	if err != nil{
		return err
	}
	filesInfoCSV = append(filesInfoCSV, fileInfoCSV)
	err = csvfile.UpdateFilesInfo(filesInfoCSV)
	if err != nil{
		return err
	}
	return nil
}

func (csvfile * CSV) GetFilesInfo() (error, []uploadFile.FileInfo) {
	var filesInfo []uploadFile.FileInfo
	// Open CSV file
	f, err := os.Open(CSV_FILENAME)
	if err != nil {
		return err, filesInfo
	}
	defer f.Close()

	// Read File into a Variable
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return err, filesInfo
	}

	var fileInfo uploadFile.FileInfo
	for _, line := range lines {
		fileInfo.UniqueId = line[0]
		fileInfo.PrintingDate = line[1]
		filesInfo = append(filesInfo, fileInfo)
	}
	return nil, filesInfo
}

func (csvfile * CSV) UpdateFilesInfo(filesInfoCSV []uploadFile.FileInfo) ( error) {
	var csvData [][]string
	for _,fileInfoCSV := range(filesInfoCSV){
		var rowData []string
		rowData = append(rowData, fileInfoCSV.UniqueId)
		rowData = append(rowData, fileInfoCSV.PrintingDate)
		csvData = append(csvData, rowData)
	}
	f, err := os.Create(CSV_FILENAME)
	if err != nil {
		fmt.Println(err)
	}
	err = csv.NewWriter(f).WriteAll(csvData)
	f.Close()
	if err != nil {
		return err
	}
	return nil
}