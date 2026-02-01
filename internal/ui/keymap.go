package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines the key bindings for the entire application
type KeyMap struct {
	// Global keys (available in most states)
	Quit           key.Binding
	Help           key.Binding
	Back           key.Binding

	// Navigation
	Up             key.Binding
	Down           key.Binding
	Left           key.Binding
	Right          key.Binding
	PageUp         key.Binding
	PageDown       key.Binding
	Home           key.Binding
	End            key.Binding

	// Vim-style navigation
	VimUp          key.Binding
	VimDown        key.Binding
	VimLeft        key.Binding
	VimRight       key.Binding

	// Text editing
	Enter          key.Binding
	Tab            key.Binding
	ShiftTab       key.Binding
	Delete         key.Binding
	Backspace      key.Binding

	// HTTP Request specific
	ExecuteRequest key.Binding
	SaveRequest    key.Binding
	CopyURL        key.Binding
	CopyCurl       key.Binding
	SwitchMethod   key.Binding
	EditHeaders    key.Binding
	EditBody       key.Binding
	EditQuery      key.Binding

	// Database specific
	ExecuteQuery   key.Binding
	SaveQuery      key.Binding
	ExportResults  key.Binding
	ConnectDB      key.Binding
	ShowSchema     key.Binding
	QueryHistory   key.Binding

	// List navigation
	SelectItem     key.Binding
	DeleteItem     key.Binding
	SearchToggle   key.Binding

	// Environment management
	AddEnv         key.Binding
	EditEnv        key.Binding
	DeleteEnv      key.Binding
	SwitchEnv      key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Global keys
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "ctrl+q"),
			key.WithHelp("ctrl+c/ctrl+q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?", "h"),
			key.WithHelp("?/h", "help"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),

		// Navigation
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("←", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("→", "right"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("pgdown", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home"),
			key.WithHelp("home", "go to start"),
		),
		End: key.NewBinding(
			key.WithKeys("end"),
			key.WithHelp("end", "go to end"),
		),

		// Vim-style navigation
		VimUp: key.NewBinding(
			key.WithKeys("k"),
			key.WithHelp("k", "up"),
		),
		VimDown: key.NewBinding(
			key.WithKeys("j"),
			key.WithHelp("j", "down"),
		),
		VimLeft: key.NewBinding(
			key.WithKeys("h"),
			key.WithHelp("h", "left"),
		),
		VimRight: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "right"),
		),

		// Text editing
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select/confirm"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next field"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "previous field"),
		),
		Delete: key.NewBinding(
			key.WithKeys("delete"),
			key.WithHelp("del", "delete"),
		),
		Backspace: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("backspace", "delete backwards"),
		),

		// HTTP Request specific
		ExecuteRequest: key.NewBinding(
			key.WithKeys("ctrl+enter", "ctrl+r"),
			key.WithHelp("ctrl+enter/ctrl+r", "execute request"),
		),
		SaveRequest: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save request"),
		),
		CopyURL: key.NewBinding(
			key.WithKeys("ctrl+y"),
			key.WithHelp("ctrl+y", "copy URL"),
		),
		CopyCurl: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "copy curl"),
		),
		SwitchMethod: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "switch method"),
		),
		EditHeaders: key.NewBinding(
			key.WithKeys("ctrl+h"),
			key.WithHelp("ctrl+h", "edit headers"),
		),
		EditBody: key.NewBinding(
			key.WithKeys("ctrl+b"),
			key.WithHelp("ctrl+b", "edit body"),
		),
		EditQuery: key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("ctrl+p", "edit query params"),
		),

		// Database specific
		ExecuteQuery: key.NewBinding(
			key.WithKeys("ctrl+k", "ctrl+enter"),
			key.WithHelp("ctrl+k/ctrl+enter", "execute query"),
		),
		SaveQuery: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save query"),
		),
		ExportResults: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "export results"),
		),
		ConnectDB: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "connect to database"),
		),
		ShowSchema: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "show schema"),
		),
		QueryHistory: key.NewBinding(
			key.WithKeys("ctrl+h"),
			key.WithHelp("ctrl+h", "query history"),
		),

		// List navigation
		SelectItem: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "select"),
		),
		DeleteItem: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		SearchToggle: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),

		// Environment management
		AddEnv: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add environment"),
		),
		EditEnv: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit environment"),
		),
		DeleteEnv: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete environment"),
		),
		SwitchEnv: key.NewBinding(
			key.WithKeys("ctrl+e"),
			key.WithHelp("ctrl+e", "switch environment"),
		),
	}
}

// StateSpecificKeys returns keys that are available for a specific state
func (k KeyMap) StateSpecificKeys(state AppState) []key.Binding {
	common := []key.Binding{k.Quit, k.Help, k.Back}

	switch state {
	case StateHome:
		return append(common, []key.Binding{
			k.Up, k.Down, k.VimUp, k.VimDown,
			k.Enter, k.SelectItem,
		}...)

	case StateRequestBuilder:
		return append(common, []key.Binding{
			k.ExecuteRequest, k.SaveRequest, k.CopyURL, k.CopyCurl,
			k.SwitchMethod, k.EditHeaders, k.EditBody, k.EditQuery,
			k.Tab, k.ShiftTab,
		}...)

	case StateRequestList:
		return append(common, []key.Binding{
			k.Up, k.Down, k.VimUp, k.VimDown,
			k.Enter, k.SelectItem, k.DeleteItem, k.SearchToggle,
		}...)

	case StateDatabase:
		return append(common, []key.Binding{
			k.Up, k.Down, k.VimUp, k.VimDown,
			k.Enter, k.ConnectDB, k.ShowSchema, k.QueryHistory,
		}...)

	case StateDatabaseQueryEditor:
		return append(common, []key.Binding{
			k.ExecuteQuery, k.SaveQuery, k.Tab, k.ShiftTab,
		}...)

	case StateDatabaseResult:
		return append(common, []key.Binding{
			k.Left, k.Right, k.VimLeft, k.VimRight,
			k.SaveQuery, k.ExportResults,
		}...)

	case StateDatabaseQueryList:
		return append(common, []key.Binding{
			k.Up, k.Down, k.VimUp, k.VimDown,
			k.Enter, k.SelectItem, k.DeleteItem,
		}...)

	case StateEnvironments:
		return append(common, []key.Binding{
			k.Up, k.Down, k.VimUp, k.VimDown,
			k.AddEnv, k.EditEnv, k.DeleteEnv, k.SwitchEnv,
		}...)

	default:
		return common
	}
}

// GetHelpText returns a formatted help string for the current state
func (k KeyMap) GetHelpText(state AppState) string {
	keys := k.StateSpecificKeys(state)

	var helpItems []string
	for _, binding := range keys {
		helpItems = append(helpItems, binding.Help().Desc+": "+binding.Help().Key)
	}

	return strings.Join(helpItems, " • ")
}

// KeyMatches checks if a key string matches any of the provided key bindings
func KeyMatches(keyStr string, bindings ...key.Binding) bool {
	for _, binding := range bindings {
		for _, k := range binding.Keys() {
			if k == keyStr {
				return true
			}
		}
	}
	return false
}

// IsNavigation checks if a key is a navigation key (arrow keys or vim keys)
func (k KeyMap) IsNavigation(keyStr string) bool {
	return KeyMatches(keyStr,
		k.Up, k.Down, k.Left, k.Right,
		k.VimUp, k.VimDown, k.VimLeft, k.VimRight,
		k.PageUp, k.PageDown, k.Home, k.End,
	)
}

// IsGlobal checks if a key is a global key available in all states
func (k KeyMap) IsGlobal(keyStr string) bool {
	return KeyMatches(keyStr, k.Quit, k.Help, k.Back)
}

// IsTextEditing checks if a key is related to text editing
func (k KeyMap) IsTextEditing(keyStr string) bool {
	return KeyMatches(keyStr,
		k.Enter, k.Tab, k.ShiftTab, k.Delete, k.Backspace,
	)
}