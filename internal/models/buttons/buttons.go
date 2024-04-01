package buttons

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Action int

const (
	DOWNLOAD Action = iota
	DOWNLOAD_EXTRACT
)

var (
	downloadCmd tea.Cmd = func() tea.Msg {
		return DOWNLOAD
	}
	downloadExtractCmd tea.Cmd = func() tea.Msg {
		return DOWNLOAD_EXTRACT
	}
)

type Button struct {
	Name   string
	Action Action
}

type Model struct {
	keys    KeyMap
	buttons []Button
	cursor  int
	width   int
	height  int
}

func (m Model) ShortHelp() []key.Binding {
	return m.keys.ShortHelp()
}

func (m Model) FullHelp() [][]key.Binding {
	return m.keys.FullHelp()
}

func (m *Model) SetSize(h, v int) {
	m.width = h
	m.height = v
}

var (
	buttonStyle lipgloss.Style = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).
			Margin(0, 1).Padding(0, 1)
	currentButtonStyle lipgloss.Style = lipgloss.NewStyle().Inherit(buttonStyle).Margin(0, 1).
				Padding(0, 1).
				BorderForeground(lipgloss.Color("205")).Foreground(lipgloss.Color("205"))
)

func (m Model) View() string {
	var s string

	for i, b := range m.buttons {
		buttonDisplay := buttonStyle.Render(b.Name)

		if i == m.cursor {
			buttonDisplay = currentButtonStyle.Render(b.Name)
		}

		if i == 0 {
			s = buttonDisplay
		} else {
			s = lipgloss.JoinHorizontal(lipgloss.Left, s, buttonDisplay)
		}
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Center, s)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {

		case key.Matches(msg, m.keys.NEXT):
			if m.cursor < len(m.buttons) {
				m.cursor++
			}
		case key.Matches(msg, m.keys.PREV):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.SUBMIT):
			cmd = getCmd(m.buttons[m.cursor].Action)
		}
	}
	return m, cmd
}

func getCmd(action Action) tea.Cmd {
	var cmd tea.Cmd
	switch action {
	case DOWNLOAD:
		cmd = downloadCmd
	case DOWNLOAD_EXTRACT:
		cmd = downloadExtractCmd
	}
	return cmd
}

type KeyMap struct {
	NEXT   key.Binding
	PREV   key.Binding
	SUBMIT key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.NEXT, k.PREV}, {k.SUBMIT}}
}

var defaultKeyMap = KeyMap{
	NEXT:   key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "next")),
	PREV:   key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "previous")),
	SUBMIT: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
}

func New(buttons ...Button) Model {
	return Model{
		keys:    defaultKeyMap,
		buttons: buttons,
	}
}
