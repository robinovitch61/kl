package help

import (
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/kl/internal/keymap"
)

func MakeHelp(keyMap keymap.KeyMap, keyColStyle lipgloss.Style) string {
	title := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1).Render("Help (press any key to hide)")
	rowsPerCol := 11

	generalHelp := lipgloss.JoinVertical(
		lipgloss.Center,
		"",
		formatKeyBindings(keymap.DescriptiveKeyBindings(keyMap), rowsPerCol, keyColStyle),
		"",
		formatKeyBindings(keymap.LookbackKeyBindings(keyMap), 6, keyColStyle),
	)

	res := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		generalHelp,
	)
	return res
}

func formatKeyBindings(bindings []key.Binding, maxRowsPerCol int, keyColStyle lipgloss.Style) string {
	if len(bindings) == 0 {
		return ""
	}
	numColumns := (len(bindings) + maxRowsPerCol - 1) / maxRowsPerCol
	var columns [][]key.Binding
	for colIndex := 0; colIndex < numColumns; colIndex++ {
		start := colIndex * maxRowsPerCol
		end := start + maxRowsPerCol
		if end > len(bindings) {
			end = len(bindings)
		}
		columns = append(columns, bindings[start:end])
	}
	var formattedCols []string
	for i, col := range columns {
		formattedCol := formatColumn(col, keyColStyle)
		if i != len(columns)-1 {
			formattedCol = formattedCol + "   "
		}
		formattedCols = append(formattedCols, formattedCol)
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, formattedCols...)
}

func formatColumn(bindings []key.Binding, keyColStyle lipgloss.Style) string {
	var keys []string
	var help []string
	for _, b := range bindings {
		k := b.Help().Key
		if len(k) > 0 {
			keys = append(keys, " "+k+" ")
		} else {
			keys = append(keys, "")
		}

		d := b.Help().Desc
		if len(d) > 0 {
			help = append(help, " "+d)
		} else {
			help = append(help, "")
		}
	}
	keyCol := keyColStyle.Render(lipgloss.JoinVertical(lipgloss.Right, keys...))
	helpCol := lipgloss.JoinVertical(lipgloss.Left, help...)
	return lipgloss.JoinHorizontal(lipgloss.Left, keyCol, helpCol)
}
