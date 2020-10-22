package gcp_helper

import (
	"cloud.google.com/go/storage"
	//"io"
	"io/ioutil"
	"context"
	"github.com/khomkovova/MonoPrinterTerminal/config"
	//"log"
	"time"
)

func GCP_download_file(filename string)  (error,  []byte) {
	var conf config.Configuration
	err := conf.ParseConfig()
	if err != nil {
		return err, nil
	}
	bucketName := conf.GCP.BucketUsersFiles

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err, nil
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	if err != nil {
		return  err, nil
	}
	rc, err := client.Bucket(bucketName).Object(filename).NewReader(ctx)

	if err != nil {
		return  err, nil
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return err, nil
	}
	return nil, data

}