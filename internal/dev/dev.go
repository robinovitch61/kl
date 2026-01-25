package dev

import (
	"fmt"
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

// DebugUpdateMsg logs a tea.Msg for debugging
func DebugUpdateMsg(component string, msg any) {
	if debugSet != "" {
		Debug(component + ": " + msgToString(msg))
	}
}

func msgToString(msg any) string {
	if msg == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%T", msg)
}
