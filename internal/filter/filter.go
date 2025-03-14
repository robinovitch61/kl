package filter

import (
	"fmt"
	"github.com/charmbracelet/bubbles/v2/cursor"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/kl/internal/dev"
	"github.com/robinovitch61/kl/internal/keymap"
	"github.com/robinovitch61/kl/internal/style"
	"github.com/robinovitch61/kl/internal/textinput"
	"regexp"
	"strings"
)

type Model struct {
	KeyMap                filterKeyMap
	ShowContext           bool
	canToggleShowContext  bool
	isRegex               bool
	regexp                *regexp.Regexp
	currentMatchNum       int
	indexesMatchingFilter []int
	textinput             textinput.Model
	suffix                string
	styles                style.Styles
}

func New(km keymap.KeyMap) Model {
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
		textinput: ti,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	dev.DebugUpdateMsg("Filter", msg)
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.textinput, cmd = m.textinput.Update(msg)
	cmds = append(cmds, cmd)

	// update regexp based on filter text
	m.updateRegexp()

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.textinput.Focused() {
		m.textinput.PromptStyle = m.styles.Inverse
		m.textinput.TextStyle = m.styles.Inverse
		m.textinput.Cursor.Style = lipgloss.NewStyle()
		m.textinput.Cursor.TextStyle = lipgloss.NewStyle()
		if !m.textinput.IsEmpty() {
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
				m.textinput.SetValue("type to regex filter ")
			} else {
				m.textinput.SetValue("type to filter ")
			}
		}
	} else {
		if !m.textinput.IsEmpty() {
			// filter applied, not editing
			if m.isRegex {
				m.textinput.Prompt = "regex filter: "
				if m.regexp == nil {
					m.textinput.Prompt = "invalid regex: "
				}
			} else {
				m.textinput.Prompt = "filter: "
			}
			m.textinput.PromptStyle = m.styles.AltInverse
			m.textinput.TextStyle = m.styles.AltInverse
			m.textinput.Cursor.Style = m.styles.AltInverse
			m.textinput.Cursor.TextStyle = m.styles.AltInverse
		} else {
			// no filter, not editing
			m.textinput.Prompt = ""
			m.textinput.SetValue(fmt.Sprintf("'%s' or '%s' to filter", m.KeyMap.Filter.Help().Key, m.KeyMap.FilterRegex.Help().Key))
			m.SetSuffix("")
		}
	}
	m.textinput.SetValue(m.textinput.Value() + m.suffix)
	filterString := m.textinput.View()
	filterStringStyle := m.textinput.TextStyle.PaddingLeft(1)

	return filterStringStyle.Render(filterString)
}

func (m Model) Matches(s string) bool {
	if m.IsEmpty() {
		return true
	}
	// if invalid regexp, fallback to string matching
	if m.isRegex && m.regexp != nil {
		return m.regexp.MatchString(s)
	} else {
		// TODO LEO: reduce number of calls to .Value() here
		return strings.Contains(s, m.Value())
	}
}

func (m Model) Value() string {
	return m.textinput.Value()
}

func (m Model) IsEmpty() bool {
	return m.textinput.IsEmpty()
}

func (m Model) GetContextualMatchIdx() int {
	if !m.ShowContext {
		return 0
	}
	if m.currentMatchNum < 0 || m.currentMatchNum >= len(m.indexesMatchingFilter) {
		return 0
	}
	return m.indexesMatchingFilter[m.currentMatchNum]
}

func (m Model) HasContextualMatches() bool {
	return m.ShowContext && len(m.indexesMatchingFilter) > 0
}

func (m Model) HasFilterText() bool {
	return !m.IsEmpty()
}

func (m Model) Focused() bool {
	return m.textinput.Focused()
}

func (m Model) IsRegex() bool {
	return m.isRegex
}

func (m *Model) SetIsRegex(isRegex bool) {
	m.isRegex = isRegex
	m.updateRegexp()
}

func (m *Model) SetValue(value string) {
	m.textinput.SetValue(value)
}

func (m *Model) SetSuffix(suffix string) {
	m.suffix = suffix
}

func (m *Model) SetStyles(styles style.Styles) {
	m.styles = styles
}

func (m *Model) SetShowContext(showContext bool, canToggleShowContext bool) {
	m.ShowContext = showContext
	m.canToggleShowContext = canToggleShowContext
	m.UpdateLabelAndSuffix()
}

func (m *Model) ResetContextualFilterMatchNum() {
	if !m.ShowContext {
		return
	}
	m.currentMatchNum = 0
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

func (m *Model) IncrementFilteredSelectionNum() {
	m.changeFilteredSelectionNum(1)
}

func (m *Model) DecrementFilteredSelectionNum() {
	m.changeFilteredSelectionNum(-1)
}

func (m *Model) changeFilteredSelectionNum(delta int) {
	if len(m.indexesMatchingFilter) == 0 {
		return
	}
	m.currentMatchNum += delta
	if m.currentMatchNum >= len(m.indexesMatchingFilter) {
		m.currentMatchNum = 0
	} else if m.currentMatchNum < 0 {
		m.currentMatchNum = len(m.indexesMatchingFilter) - 1
	}
}

func (m *Model) UpdateLabelAndSuffix() {
	// don't show the label "matches only" if you can't toggle show context
	if !m.canToggleShowContext && !m.ShowContext {
		return
	}

	if !m.ShowContext {
		m.SetSuffix(" (matches only) ")
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
