package output

import (
	"fmt"
	"log"
	"os"
	"time"
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
		fmt.Fprintf(logFile, "[%s] [INFO] %s\n", time.Now().Format("15:04:05"), msg)
	}
}

func Warn(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	log.Printf("[WARN] %s", msg)
	if logFile != nil {
		fmt.Fprintf(logFile, "[%s] [WARN] %s\n", time.Now().Format("15:04:05"), msg)
	}
}

func Error(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	log.Printf("[ERROR] %s", msg)
	if logFile != nil {
		fmt.Fprintf(logFile, "[%s] [ERROR] %s\n", time.Now().Format("15:04:05"), msg)
	}
}