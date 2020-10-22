package models

type Response struct {
	Status            string `json:"status"`
	StatusDescription string `json:"status_description"`
	Data              string `json:"data"`
}

type GCPFile struct {
	FileUrl string `json:"file_url"`
}


type FileStatus struct {
	FileStatus string `json:"status"`
}