package filter

import (
	"fmt"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/keymap"
	"github.com/robinovitch61/kl/internal/style"
	"regexp"
	"strings"
)

type Renderable interface {
	Render() string
}

type Model struct {
	KeyMap                filterKeyMap
	FilteringWithContext  bool
	width                 int
	isRegex               bool
	regexp                *regexp.Regexp
	currentMatchNum       int
	indexesMatchingFilter []int
	label                 string
	textinput             textinput.Model
	suffix                string
}

func New(label string, width int, km keymap.KeyMap) Model {
	ti := textinput.New()
	ti.Prompt = ""
	ti.Cursor.SetMode(cursor.CursorHide)

	fkm := filterKeyMap{
		Forward:       km.Enter,
		Back:          km.Clear,
		Filter:        km.Filter,
		FilterRegex:   km.FilterRegex,
		FilterNextRow: km.FilterNextRow,
		FilterPrevRow: km.FilterPrevRow,
	}

	return Model{
		KeyMap:    fkm,
		width:     width,
		label:     label,
		textinput: ti,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	dev.DebugMsg("Filter", msg)
	var cmd tea.Cmd
	var cmds []tea.Cmd

	m.textinput, cmd = m.textinput.Update(msg)
	cmds = append(cmds, cmd)

	// update regexp based on filter text
	m.updateRegexp()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.FilteringWithContext && !m.Focused() && m.HasFilterText() && len(m.indexesMatchingFilter) > 0 {
			if key.Matches(msg, m.KeyMap.FilterNextRow) {
				m.incrementFilteredSelectionNum()
			} else if key.Matches(msg, m.KeyMap.FilterPrevRow) {
				m.decrementFilteredSelectionNum()
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.textinput.Focused() {
		m.textinput.PromptStyle = style.FilterEditing
		m.textinput.TextStyle = style.FilterEditing
		m.textinput.Cursor.Style = style.FilterCursor
		m.textinput.Cursor.TextStyle = style.FilterEditing
		if len(m.textinput.Value()) > 0 {
			// editing existing filter
			if m.isRegex {
				m.textinput.Prompt = "regex filter: "
				if m.regexp == nil {
					m.textinput.Prompt = "invalid regex: "
				}
			} else {
				m.textinput.Prompt = "filter: "
			}
		} else {
			// editing but no filter value yet
			m.textinput.Prompt = ""
			m.SetSuffix("")
			m.textinput.Cursor.SetMode(cursor.CursorHide)
			if m.isRegex {
				m.textinput.SetValue("type to regex filter")
			} else {
				m.textinput.SetValue("type to filter")
			}
		}
	} else {
		if len(m.textinput.Value()) > 0 {
			// filter applied, not editing
			if m.isRegex {
				m.textinput.Prompt = "regex filter: "
				if m.regexp == nil {
					m.textinput.Prompt = "invalid regex: "
				}
			} else {
				m.textinput.Prompt = "filter: "
			}
			m.textinput.PromptStyle = style.FilterApplied
			m.textinput.TextStyle = style.FilterApplied
			m.textinput.Cursor.Style = style.FilterApplied
			m.textinput.Cursor.TextStyle = style.FilterApplied
		} else {
			// no filter, not editing
			m.textinput.Prompt = ""
			m.textinput.PromptStyle = style.Regular
			m.textinput.TextStyle = style.Regular
			m.textinput.Cursor.Style = style.Regular
			m.textinput.Cursor.TextStyle = style.Regular
			m.textinput.SetValue(fmt.Sprintf("'%s' or '%s' to filter", m.KeyMap.Filter.Help().Key, m.KeyMap.FilterRegex.Help().Key))
			m.SetSuffix("")
		}
	}
	m.textinput.SetValue(m.textinput.Value() + m.suffix)
	filterString := m.textinput.View()
	filterStringStyle := m.textinput.TextStyle.Copy().PaddingLeft(1)

	view := style.FilterPrefix.Copy().PaddingRight(1).Underline(true).Render(m.label) + filterStringStyle.Render(filterString)
	widthView := lipgloss.Width(view)
	if widthView < m.width {
		view += style.Regular.Render(strings.Repeat(" ", m.width-widthView))
	}
	return view
}

func (m Model) Matches(renderable Renderable) bool {
	if m.Value() == "" {
		return true
	}
	// if invalid regexp, fallback to string matching
	if m.isRegex && m.regexp != nil {
		return m.regexp.MatchString(renderable.Render())
	} else {
		return strings.Contains(renderable.Render(), m.Value())
	}
}

func (m Model) Value() string {
	return m.textinput.Value()
}

func (m Model) GetContextualMatchIdx() int {
	if !m.FilteringWithContext {
		return 0
	}
	if m.currentMatchNum < 0 || m.currentMatchNum >= len(m.indexesMatchingFilter) {
		return 0
	}
	return m.indexesMatchingFilter[m.currentMatchNum]
}

func (m Model) ViewHeight() int {
	return lipgloss.Height(m.View())
}

func (m Model) HasFilterText() bool {
	return m.Value() != ""
}

func (m Model) Focused() bool {
	return m.textinput.Focused()
}

func (m Model) IsRegex() bool {
	return m.isRegex
}

func (m *Model) SetWidth(width int) {
	m.width = width
}

func (m *Model) SetIsRegex(isRegex bool) {
	m.isRegex = isRegex
	m.updateRegexp()
}

func (m *Model) SetValue(value string) {
	m.textinput.SetValue(value)
}

func (m *Model) SetLabel(label string) {
	m.label = label
}

func (m *Model) SetSuffix(suffix string) {
	m.suffix = suffix
}

func (m *Model) SetFilteringWithContext(filteringWithContext bool) {
	m.FilteringWithContext = filteringWithContext
	m.UpdateLabelAndSuffix()
}

func (m *Model) SetIndexesMatchingFilter(indexes []int) {
	m.indexesMatchingFilter = indexes
	m.UpdateLabelAndSuffix()
}

func (m *Model) Focus() {
	m.textinput.Cursor.SetMode(cursor.CursorBlink)
	m.textinput.Focus()
}

func (m *Model) Blur() {
	// move cursor to end of word so right padding shows up even if cursor not at end when blurred
	m.textinput.SetCursor(len(m.textinput.Value()))

	m.textinput.Cursor.SetMode(cursor.CursorHide)
	m.textinput.Blur()
}

func (m *Model) BlurAndClear() {
	m.Blur()
	m.indexesMatchingFilter = []int{}
	m.currentMatchNum = 0
	m.textinput.SetValue("")
}

func (m *Model) updateRegexp() {
	if m.isRegex {
		regex, err := regexp.Compile(m.textinput.Value())
		if err == nil {
			m.regexp = regex
		} else {
			m.regexp = nil
		}
	}
}

func (m *Model) incrementFilteredSelectionNum() {
	if len(m.indexesMatchingFilter) == 0 {
		return
	}
	m.currentMatchNum++
	if m.currentMatchNum >= len(m.indexesMatchingFilter) {
		m.currentMatchNum = 0
	} else if m.currentMatchNum < 0 {
		m.currentMatchNum = 0
	}
}

func (m *Model) decrementFilteredSelectionNum() {
	if len(m.indexesMatchingFilter) == 0 {
		return
	}
	m.currentMatchNum--
	if m.currentMatchNum < 0 {
		m.currentMatchNum = len(m.indexesMatchingFilter) - 1
	} else if m.currentMatchNum >= len(m.indexesMatchingFilter) {
		m.currentMatchNum = len(m.indexesMatchingFilter) - 1
	}
}

func (m *Model) UpdateLabelAndSuffix() {
	if !m.FilteringWithContext {
		m.SetSuffix("")
		return
	}

	if len(m.indexesMatchingFilter) == 0 {
		m.SetSuffix(" (no matches) ")
	} else if m.Focused() {
		m.SetSuffix(
			fmt.Sprintf(
				" (%d/%d, enter to apply) ",
				m.currentMatchNum+1,
				len(m.indexesMatchingFilter),
			),
		)
	} else {
		m.SetSuffix(
			fmt.Sprintf(
				" (%d/%d, %s/%s to cycle) ",
				m.currentMatchNum+1,
				len(m.indexesMatchingFilter),
				m.KeyMap.FilterNextRow.Help().Key,
				m.KeyMap.FilterPrevRow.Help().Key,
			),
		)
	}
}
