package notification

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eslam-allam/spring-initializer-go/constants"
	"github.com/muesli/reflow/wordwrap"
)

type NotificationMsg struct {
	Message string
	Level   NotificationLevel
}

type NotificationLevel int

const (
	INFO NotificationLevel = iota
	WARNING
	ERROR
)

const (
	infoColor    = "#7ae878"
	warningColor = "#e6db74"
	errorColor   = "#f84841"
)

var (
	notificationStyle lipgloss.Style = lipgloss.NewStyle().Padding(1).
				Border(lipgloss.NormalBorder(), true).
				BorderForeground(lipgloss.Color(constants.HighlightColour))
	notificationTextStyle = lipgloss.NewStyle()
)

type Model struct {
	message string
	keys    NotificationKeyMap
	level   NotificationLevel
	active  bool
	width   int
	height  int
}

func (m Model) IsActive() bool {
	return m.active
}

func (m *Model) Activate() {
	m.active = true
}

func (m Model) Update(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.DISMISS):
		m.active = false
	}
	return m, nil
}

func (m Model) UpdateMessage(msg NotificationMsg) Model {
	m.message = msg.Message
	m.level = msg.Level
	return m
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

func (m Model) View() string {
	var currentColor string

	switch m.level {
	case INFO:
		currentColor = infoColor
	case WARNING:
		currentColor = warningColor
	case ERROR:
		currentColor = errorColor
	}
	return notificationStyle.
		Render(
			notificationTextStyle.MaxWidth(m.width).MaxHeight(m.height).
				Foreground(lipgloss.Color(currentColor)).
				Render(wordwrap.String(m.message, m.width-notificationStyle.GetHorizontalFrameSize())))
}

type NotificationKeyMap struct {
	DISMISS key.Binding
}

func (k NotificationKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.DISMISS}
}

func (k NotificationKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.DISMISS}}
}

var defaultNotificationKeys NotificationKeyMap = NotificationKeyMap{
	DISMISS: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "dismiss")),
}

func New() Model {
	return Model{
		keys: defaultNotificationKeys,
	}
}
