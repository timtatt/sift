package sift

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
)

type keyMap struct {
	viewport               viewport.KeyMap
	Up                     key.Binding
	Down                   key.Binding
	PrevTest               key.Binding
	NextTest               key.Binding
	ToggleTestsRecursively key.Binding
	ExpandAllTests         key.Binding
	CollapseAllTests       key.Binding
	ToggleTest             key.Binding
	ExpandTest             key.Binding
	CollapseTest           key.Binding
	Search                 key.Binding
	ClearSearch            key.Binding
	Help                   key.Binding
	Quit                   key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PrevTest, k.NextTest},
		{k.viewport.Up, k.viewport.Down, k.viewport.HalfPageUp, k.viewport.HalfPageDown},
		{k.ToggleTest, k.ExpandTest, k.CollapseTest},
		{k.ToggleTestsRecursively, k.ExpandAllTests, k.CollapseAllTests},
		{k.Search, k.ClearSearch, k.Help, k.Quit},
	}
}

var (
	keys = keyMap{
		viewport: viewport.KeyMap{
			Down: key.NewBinding(
				key.WithKeys("ctrl+e"),
				key.WithHelp("ctrl+y", "scroll down"),
			),
			Up: key.NewBinding(
				key.WithKeys("ctrl+y"),
				key.WithHelp("ctrl+y", "scroll down"),
			),
			HalfPageUp: key.NewBinding(
				key.WithKeys("ctrl+u"),
				key.WithHelp("ctrl+u", "half page up"),
			),
			HalfPageDown: key.NewBinding(
				key.WithKeys("ctrl+d"),
				key.WithHelp("ctrl+d", "half page down"),
			),
		},
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		PrevTest: key.NewBinding(
			key.WithKeys("{"),
			key.WithHelp("{", "previous test"),
		),
		NextTest: key.NewBinding(
			key.WithKeys("}"),
			key.WithHelp("}", "next test"),
		),
		ToggleTestsRecursively: key.NewBinding(
			key.WithKeys("zA"),
			key.WithHelp("zA", "toggle recursively"),
		),
		ExpandAllTests: key.NewBinding(
			key.WithKeys("zR"),
			key.WithHelp("zR", "expand all"),
		),
		CollapseAllTests: key.NewBinding(
			key.WithKeys("zM"),
			key.WithHelp("zM", "collapse all"),
		),
		ToggleTest: key.NewBinding(
			key.WithKeys("za", "enter", " "),
			key.WithHelp("za/enter/space", "toggle test"),
		),
		ExpandTest: key.NewBinding(
			key.WithKeys("zo"),
			key.WithHelp("zo", "expand test"),
		),
		CollapseTest: key.NewBinding(
			key.WithKeys("zc"),
			key.WithHelp("zc", "collapse test"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search tests"),
		),
		ClearSearch: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear search"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
)
