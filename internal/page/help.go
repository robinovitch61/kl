package page

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/kl/internal/style"
)

func makePageHelp(pageName string, globalBindings, pageBindings []key.Binding) string {
	title := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1).Render("Help (press any key to hide)")
	rowsPerCol := 6

	pageHelp := lipgloss.JoinVertical(
		lipgloss.Center,
		style.Underline.Render("Current View: "+pageName),
		"",
		formatKeyBindings(pageBindings, rowsPerCol),
	)

	generalHelp := lipgloss.JoinVertical(
		lipgloss.Center,
		"",
		style.Underline.Render("Anywhere in Application"),
		"",
		formatKeyBindings(globalBindings, rowsPerCol),
	)

	res := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		pageHelp,
		generalHelp,
	)
	return res
}

func formatKeyBindings(bindings []key.Binding, maxRowsPerCol int) string {
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
		formattedCol := formatColumn(col)
		if i != len(columns)-1 {
			formattedCol = formattedCol + "   "
		}
		formattedCols = append(formattedCols, formattedCol)
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, formattedCols...)
}

func formatColumn(bindings []key.Binding) string {
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
			help = append(help, "->"+d)
		} else {
			help = append(help, "")
		}
	}
	keyCol := style.InverseUnderline.Render(lipgloss.JoinVertical(lipgloss.Right, keys...))
	helpCol := lipgloss.JoinVertical(lipgloss.Left, help...)
	return lipgloss.JoinHorizontal(lipgloss.Left, keyCol, helpCol)
}
