package dev

import (
	"log"
	"os"
)

var debugSet = os.Getenv("KL_DEBUG")
var debugPath = os.Getenv("KL_DEBUG_PATH")

func Debug(msg string) {
	if debugPath == "" {
		debugPath = "kl.log"
	}
	if debugSet != "" {
		file, err := os.OpenFile(debugPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
		defer func() { _ = file.Close() }()
		logger := log.New(file, "", log.Ldate|log.Lmicroseconds)
		logger.Printf("%q", msg)
	}
}
