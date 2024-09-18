package dev

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/kl/internal/message"
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
		defer file.Close()
		logger := log.New(file, "", log.Ldate|log.Lmicroseconds)
		logger.Printf("%q", msg)
	}
}

func DebugMsg(component string, msg tea.Msg) {
	switch msg.(type) {
	case message.BatchUpdateLogsMsg, message.UpdateSinceTimeTextMsg:
	// skip logging messages that are too frequent
	default:
		Debug(fmt.Sprintf("Update %s: %T", component, msg))
	}
}
