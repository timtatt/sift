package sift

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
)

type keyMap struct {
	viewport viewport.KeyMap
	Help     key.Binding
	Quit     key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.viewport.Up, k.viewport.Down, k.viewport.HalfPageUp, k.viewport.HalfPageDown},
		{k.Help, k.Quit},
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
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
)
