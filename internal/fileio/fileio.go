package fileio

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

type SaveCompleteMsg struct {
	FullPath, SuccessMessage, ErrMessage string
}

func GetSaveCommand(fileName string, content []string) tea.Cmd {
	return func() tea.Msg {
		savePathWithFileName, err := saveToFile(fileName, content)
		if err != nil {
			return SaveCompleteMsg{ErrMessage: err.Error()}
		}
		return SaveCompleteMsg{
			FullPath:       savePathWithFileName,
			SuccessMessage: fmt.Sprintf("Saved to %s", savePathWithFileName),
		}
	}
}

func saveToFile(fileName string, fileContent []string) (string, error) {
	now := time.Now().UTC().Format("20060102T150405Z")
	path := "."
	if fileName == "" {
		fileName = now
	} else {
		if strings.Contains(fileName, "~") {
			currUser, err := user.Current()
			if err != nil {
				return "", err
			}
			fileName = strings.ReplaceAll(fileName, "~", currUser.HomeDir)
		}

		if strings.Contains(fileName, string(os.PathSeparator)) {
			path = filepath.Dir(fileName)
			fileName = filepath.Base(fileName)
		}
	}

	// unless otherwise specified, make extension .txt
	if filepath.Ext(fileName) == "" {
		fileName += ".txt"
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	if exists, err := fileOrDirectoryExists(absPath); err == nil {
		if !exists {
			if mkdirErr := os.MkdirAll(absPath, 0755); mkdirErr != nil {
				return "", mkdirErr
			}
		}
	} else {
		return "", err
	}

	pathWithFileName := fmt.Sprintf("%s%s%s", absPath, string(os.PathSeparator), fileName)

	// if file already exists at specified location, append timestamp to filename
	if exists, err := fileOrDirectoryExists(pathWithFileName); err == nil {
		if exists {
			extension := filepath.Ext(pathWithFileName)
			if extension == "" {
				// /home/test -> /home/test_2021-01-01T12:00:00
				pathWithFileName += "_" + now
			} else {
				// /home/test.txt -> /home/test_2021-01-01T12:00:00.txt
				pathWithFileName = strings.ReplaceAll(pathWithFileName, extension, "_"+now+extension)
			}
		}
	} else {
		return "", err
	}

	f, err := os.Create(pathWithFileName)
	if err != nil {
		return "", err
	}
	defer f.Close()

	for _, line := range fileContent {
		_, writeErr := f.WriteString(line + "\n")
		if writeErr != nil {
			return "", writeErr
		}
	}
	return pathWithFileName, nil
}

func fileOrDirectoryExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
