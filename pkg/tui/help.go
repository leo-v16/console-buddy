package tui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// helpKeyMap defines the key bindings for the help view.
// It is used to navigate the help view and to close it.
type helpKeyMap struct {
	help key.Binding
	quit key.Binding
}

// ShortHelp returns a slice of key bindings to be displayed in the short help view.
func (k helpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.help, k.quit}
}

// FullHelp returns a slice of key bindings to be displayed in the full help view.
func (k helpKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.help, k.quit},
	}
}

// newHelpKeyMap creates a new helpKeyMap with default key bindings.
func newHelpKeyMap() *helpKeyMap {
	return &helpKeyMap{
		help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q", "quit"),
		),
	}
}

// newHelp creates a new help model with the given key map.
func newHelp(keys *helpKeyMap) help.Model {
	h := help.New()
	h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
	h.Styles.FullDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
	return h
}
