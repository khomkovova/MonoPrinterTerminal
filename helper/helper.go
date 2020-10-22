package helper

import (
	"log"
	"runtime"
)

func LogErrorMsg(err error, errorComment string) {
	pc, _, line, _ := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	log.Printf("Error in function: %s\n", details.Name())
	log.Printf("Line: %d\n", line)
	log.Printf("Error: %s\n", err)
	log.Printf("Error comment: %s\n\n", errorComment)

}

func LogInfoMsg(infoDescription string) {
	pc, _, line, _ := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	log.Printf("Info in function: %s\n", details.Name())
	log.Printf("Line: %d\n\n", line)
	log.Printf("Info comment: %s\n\n", infoDescription)

}