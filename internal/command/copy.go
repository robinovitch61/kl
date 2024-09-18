package command

import (
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

type ContentCopiedToClipboardMsg struct {
	Content string
	Err     error
}

func CopyContentToClipboardCmd(content string) tea.Cmd {
	return func() tea.Msg {
		err := clipboard.WriteAll(content)
		return ContentCopiedToClipboardMsg{Content: content, Err: err}
	}
}
