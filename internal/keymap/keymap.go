package keymap

import (
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/robinovitch61/kl/internal/viewport"
)

type KeyMap struct {
	Clear               key.Binding
	Copy                key.Binding
	Context             key.Binding
	Enter               key.Binding
	Filter              key.Binding
	FilterRegex         key.Binding
	FilterNextRow       key.Binding
	FilterPrevRow       key.Binding
	Fullscreen          key.Binding
	Help                key.Binding
	Logs                key.Binding
	LogsFullScreen      key.Binding
	Name                key.Binding
	NextLog             key.Binding
	PrevLog             key.Binding
	Quit                key.Binding
	ReverseOrder        key.Binding
	Save                key.Binding
	Selection           key.Binding
	SelectionFullScreen key.Binding
	SinceTime           key.Binding
	Timestamps          key.Binding
	TogglePause         key.Binding
	Wrap                key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Clear: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "discard filter"),
		),
		Copy: key.NewBinding(
			key.WithKeys("ctrl+y"),
			key.WithHelp("ctrl+y", "copy zoomed log"),
		),
		Context: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "filter matches only"),
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
			key.WithKeys("n"),
			key.WithHelp("n", "next filter match"),
		),
		FilterPrevRow: key.NewBinding(
			key.WithKeys("shift+n"),
			key.WithHelp("N", "prev filter match"),
		),
		Fullscreen: key.NewBinding(
			key.WithKeys("shift+f"),
			key.WithHelp("F", "toggle fullscreen"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "show/hide help"),
		),
		Logs: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "focus logs"),
		),
		LogsFullScreen: key.NewBinding(
			key.WithKeys("shift+l"),
			key.WithHelp("L", "logs fullscreen"),
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
			key.WithHelp("ctrl+s", "save focus to file"),
		),
		Selection: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "focus selection"),
		),
		SelectionFullScreen: key.NewBinding(
			key.WithKeys("shift+s"),
			key.WithHelp("S", "selection fullscreen"),
		),
		SinceTime: key.NewBinding(
			key.WithKeys("0", "1", "2", "3", "4", "5", "6", "7", "8", "9"),
			key.WithHelp("0-9", "change log start time"),
		),
		Timestamps: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "show short/full/no timestamps"),
		),
		TogglePause: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pause/resume logs"),
		),
		Wrap: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "toggle line wrap"),
		),
	}
}

func LookbackKeyBindings(km KeyMap) []key.Binding {
	return []key.Binding{
		km.SinceTime,
		WithDesc(WithKeys(km.SinceTime, "0"), "now onwards"),
		WithDesc(WithKeys(km.SinceTime, "1"), "1m"),
		WithDesc(WithKeys(km.SinceTime, "2"), "5m"),
		WithDesc(WithKeys(km.SinceTime, "3"), "15m"),
		WithDesc(WithKeys(km.SinceTime, "4"), "30m"),
		WithDesc(WithKeys(km.SinceTime, "5"), "1h"),
		WithDesc(WithKeys(km.SinceTime, "6"), "3h"),
		WithDesc(WithKeys(km.SinceTime, "7"), "12h"),
		WithDesc(WithKeys(km.SinceTime, "8"), "1d"),
		WithDesc(WithKeys(km.SinceTime, "9"), "all time"),
	}
}

func GlobalKeyBindings(km KeyMap) []key.Binding {
	// available from anywhere on the app
	return []key.Binding{
		km.Logs,
		km.LogsFullScreen,
		km.Selection,
		km.SelectionFullScreen,
		km.Fullscreen,
		km.Wrap,
		km.Timestamps,
		km.Name,
		km.ReverseOrder,
		km.Filter,
		km.FilterRegex,
		km.Clear,
		WithDesc(km.Enter, "apply filter"),
		km.FilterNextRow,
		km.FilterPrevRow,
		km.Context,
		viewport.DefaultKeyMap().Up,
		viewport.DefaultKeyMap().Down,
		viewport.DefaultKeyMap().PageUp,
		viewport.DefaultKeyMap().PageDown,
		viewport.DefaultKeyMap().Top,
		viewport.DefaultKeyMap().Bottom,
		km.Save,
		km.TogglePause,
		WithDesc(km.Enter, "zoom on log"),
		WithDesc(km.Clear, "back to all logs"),
		km.Copy,
		km.Quit,
		km.Help,
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
