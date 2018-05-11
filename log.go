package main

import (
	"io"
	"log"
	"os"
	"time"
)

// Compare is instance of the log
var (
	Compare *log.Logger
)

// LogInit initialises the log file to detail comparison differences
func LogInit(
	compareHandle io.Writer) {

	file, err := os.OpenFile(LogName(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", "changes.log", ":", err)
	}

	Compare = log.New(file,
		"",
		log.Ldate|log.Ltime|log.Lshortfile)

	Compare.SetFlags(0)
}

// LogName returns the folder and file name for creating a new log
func LogName() string {
	return "compare/changes_" + time.Now().Format("20060102150405") + ".log"
}
