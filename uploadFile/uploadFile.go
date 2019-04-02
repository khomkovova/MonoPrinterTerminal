package uploadFile

import (
	"os"
)

type FileInfo struct {
	UniqueId     string
	Filename     string
	PrintingDate string

	UploadDate string
	NumberPage int
	Size       string

	IdPrinter int
	Status    string
}

type UploadFile struct {
	Info FileInfo
	File *os.File
}
