package output

import (
	"fmt"
	"log"
	"os"
)

var logFile *os.File

func InitLogger() {
	var err error
	logFile, err = os.OpenFile("daxwalkerfix.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Warning: could not open log file: %v", err)
	}
}

func Info(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	log.Printf("[INFO] %s", msg)
	if logFile != nil {
		fmt.Fprintf(logFile, "[INFO] %s\n", msg)
	}
}

func Warn(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	log.Printf("[WARN] %s", msg)
	if logFile != nil {
		fmt.Fprintf(logFile, "[WARN] %s\n", msg)
	}
}

func Error(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	log.Printf("[ERROR] %s", msg)
	if logFile != nil {
		fmt.Fprintf(logFile, "[ERROR] %s\n", msg)
	}
}