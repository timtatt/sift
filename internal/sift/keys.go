package sift

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/timtatt/sift/pkg/viewport2"
)

type keyMap struct {
	viewport               viewport2.KeyMap
	Up                     key.Binding
	Down                   key.Binding
	PrevTest               key.Binding
	NextTest               key.Binding
	PrevFailingTest        key.Binding
	NextFailingTest        key.Binding
	ToggleTestsRecursively key.Binding
	ExpandAllTests         key.Binding
	CollapseAllTests       key.Binding
	ToggleTest             key.Binding
	ToggleTestAlt          key.Binding
	ExpandTest             key.Binding
	CollapseTest           key.Binding
	Search                 key.Binding
	ClearSearch            key.Binding
	Help                   key.Binding
	Quit                   key.Binding
	ChangeMode             key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.ChangeMode},
		{k.PrevTest, k.NextTest, k.PrevFailingTest, k.NextFailingTest},
		{k.viewport.Up, k.viewport.Down, k.viewport.HalfPageUp, k.viewport.HalfPageDown},
		{k.ToggleTest, k.ExpandTest, k.CollapseTest},
		{k.ToggleTestsRecursively, k.ExpandAllTests, k.CollapseAllTests},
		{k.Search, k.ClearSearch, k.Help, k.Quit},
	}
}

var (
	keys = keyMap{
		viewport: viewport2.KeyMap{
			Down: key.NewBinding(
				key.WithKeys("ctrl+e"),
				key.WithHelp("ctrl+e", "scroll down"),
			),
			Up: key.NewBinding(
				key.WithKeys("ctrl+y"),
				key.WithHelp("ctrl+y", "scroll up"),
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
		ChangeMode: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "change mode"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k", "ctrl+p"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j", "ctrl+n"),
			key.WithHelp("↓/j", "move down"),
		),
		PrevTest: key.NewBinding(
			key.WithKeys("{"),
			key.WithHelp("{", "prev test"),
		),
		NextTest: key.NewBinding(
			key.WithKeys("}"),
			key.WithHelp("}", "next test"),
		),
		PrevFailingTest: key.NewBinding(
			key.WithKeys("["),
			key.WithHelp("[", "prev failed test"),
		),
		NextFailingTest: key.NewBinding(
			key.WithKeys("]"),
			key.WithHelp("]", "next failed test"),
		),
		ToggleTestsRecursively: key.NewBinding(
			key.WithKeys("zA"),
			key.WithHelp("zA", "toggle children"),
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
			key.WithKeys("za"),
			key.WithHelp("za", "toggle test"),
		),
		ToggleTestAlt: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "toggle test"),
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
