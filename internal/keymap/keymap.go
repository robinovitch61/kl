package keymap

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/kl/internal/viewport"
)

type KeyMap struct {
	Clear         key.Binding
	Copy          key.Binding
	Context       key.Binding
	Enter         key.Binding
	Filter        key.Binding
	FilterRegex   key.Binding
	FilterNextRow key.Binding
	FilterPrevRow key.Binding
	Help          key.Binding
	Logs          key.Binding
	Lookback      key.Binding
	Name          key.Binding
	NextLog       key.Binding
	PrevLog       key.Binding
	Quit          key.Binding
	ReverseOrder  key.Binding
	Save          key.Binding
	Timestamps    key.Binding
	TogglePause   key.Binding
	Wrap          key.Binding
}

// DefaultKeyMap for app. General principles:
// - toggles and actions on the same page start with ctrl
// - pages are single letters
// - selecting or drilling down is enter
var DefaultKeyMap = KeyMap{
	Clear: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "discard filter"),
	),
	Copy: key.NewBinding(
		key.WithKeys("ctrl+y"),
		key.WithHelp("ctrl+y", "copy log to clipboard"),
	),
	Context: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "toggle filtered logs only"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", ""), // means different things on different pages
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "edit filter"),
	),
	FilterRegex: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "regex filter"),
	),
	FilterNextRow: key.NewBinding(
		key.WithKeys("n", tea.KeyShiftUp.String()),
		key.WithHelp("n", "next filter match"),
	),
	FilterPrevRow: key.NewBinding(
		key.WithKeys("N"),
		key.WithHelp("N", "prev filter match"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "show help"),
	),
	Logs: key.NewBinding(
		key.WithKeys("l", "L"),
		key.WithHelp("L", "go to logs view"),
	),
	Lookback: key.NewBinding(
		key.WithKeys("0", "1", "2", "3", "4", "5", "6", "7", "8", "9"),
		key.WithHelp("0-9", "change log start time"),
	),
	Name: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "show short/full/no container names"),
	),
	NextLog: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("↓/j", "next log"),
	),
	PrevLog: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("↑/k", "previous log"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	ReverseOrder: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "reverse timestamp order"),
	),
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save view to file"),
	),
	Timestamps: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "show short/full/no timestamps"),
	),
	TogglePause: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "pause/resume log stream"),
	),
	Wrap: key.NewBinding(
		key.WithKeys("w"),
		key.WithHelp("w", "toggle line wrap"),
	),
}

func LookbackKeyBindings(km KeyMap) []key.Binding {
	return []key.Binding{
		km.Lookback,
		WithDesc(WithKeys(km.Lookback, "0"), "now onwards"),
		WithDesc(WithKeys(km.Lookback, "1"), "1m"),
		WithDesc(WithKeys(km.Lookback, "2"), "5m"),
		WithDesc(WithKeys(km.Lookback, "3"), "15m"),
		WithDesc(WithKeys(km.Lookback, "4"), "30m"),
		WithDesc(WithKeys(km.Lookback, "5"), "1h"),
		WithDesc(WithKeys(km.Lookback, "6"), "3h"),
		WithDesc(WithKeys(km.Lookback, "7"), "12h"),
		WithDesc(WithKeys(km.Lookback, "8"), "1d"),
		WithDesc(WithKeys(km.Lookback, "9"), "all time"),
	}
}

func GlobalKeyBindings(km KeyMap) []key.Binding {
	// available from anywhere on the app
	return []key.Binding{
		km.Quit,
		km.Help,
		km.Save,
		km.Wrap,
		viewport.DefaultKeyMap().Up,
		viewport.DefaultKeyMap().Down,
		viewport.DefaultKeyMap().PageUp,
		viewport.DefaultKeyMap().PageDown,
		viewport.DefaultKeyMap().Top,
		viewport.DefaultKeyMap().Bottom,
		km.Filter,
		km.FilterRegex,
		km.Clear,
		WithDesc(km.Enter, "apply filter"),
		km.FilterNextRow,
		km.FilterPrevRow,
	}
}

func WithKeys(k key.Binding, keys string) key.Binding {
	newK := k
	newK.SetHelp(keys, k.Help().Desc)
	return newK
}

func WithDesc(k key.Binding, d string) key.Binding {
	newK := k
	newK.SetHelp(newK.Help().Key, d)
	return newK
}
