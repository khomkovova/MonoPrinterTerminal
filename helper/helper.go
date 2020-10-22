package helper

import (
	"log"
	"os"
	"runtime"
)

func LogErrorMsg(err error, errorComment string, logger *log.Logger) {
	pc, _, line, _ := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	logger.Printf("Error in function: %s\n", details.Name())
	logger.Printf("Line: %d\n", line)
	logger.Printf("Error: %s\n", err)
	logger.Printf("Error comment: %s\n\n", errorComment)

}

func LogInfoMsg(infoDescription string, logger *log.Logger) {
	pc, _, line, _ := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	logger.Printf("Info in function: %s\n", details.Name())
	logger.Printf("Line: %d\n", line)
	logger.Printf("Info comment: %s\n\n", infoDescription)

}

func InitLogger(filename string) *log.Logger {
	logfile, err := os.OpenFile(filename, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Println("Error", err)
		return nil
	}
	return log.New(logfile, "", log.LstdFlags|log.Lshortfile)
}