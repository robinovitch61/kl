package dev

import (
	"fmt"
	"github.com/charmbracelet/bubbles/v2/cursor"
	tea "github.com/charmbracelet/bubbletea/v2"
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

func DebugUpdateMsg(component string, msg tea.Msg) {
	switch msg.(type) {
	case message.BatchUpdateLogsMsg, message.UpdateSinceTimeTextMsg, cursor.BlinkMsg:
	// skip logging messages that are too frequent
	default:
		Debug("--")
		Debug(fmt.Sprintf("Update %s: %T", component, msg))
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			Debug(fmt.Sprintf("  Key: '%v'", keyMsg.String()))
		}
	}
}
