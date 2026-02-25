package command

import (
	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
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
