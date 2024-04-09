package shared

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

var notificationStyle lipgloss.Style = lipgloss.NewStyle().Padding(1).
	Border(lipgloss.NormalBorder(), true).
	BorderForeground(lipgloss.Color(constants.HighlightColour))

type Notification struct {
	message string
	keys    NotificationKeyMap
	level   NotificationLevel
	active  bool
	width   int
	height  int
}

func (n Notification) IsActive() bool {
	return n.active
}

func (n *Notification) Activate() {
	n.active = true
}

func (n Notification) Update(msg tea.KeyMsg) (Notification, tea.Cmd) {
	switch {
	case key.Matches(msg, n.keys.DISMISS):
		n.active = false
	}
	return n, nil
}

func (n Notification) UpdateMessage(msg NotificationMsg) Notification {
	n.message = msg.Message
	n.level = msg.Level
	return n
}

func (n Notification) ShortHelp() []key.Binding {
	return n.keys.ShortHelp()
}

func (n Notification) FullHelp() [][]key.Binding {
	return n.keys.FullHelp()
}

func (n *Notification) SetSize(h, v int) {
	n.width = h
	n.height = v
}

func (n Notification) View() string {
	var currentColor string

	switch n.level {
	case INFO:
		currentColor = infoColor
	case WARNING:
		currentColor = warningColor
	case ERROR:
		currentColor = errorColor
	}
	return notificationStyle.Copy().
		Foreground(lipgloss.Color(currentColor)).Render(
		lipgloss.Place(n.width, n.height, lipgloss.Center, lipgloss.Center,
			wordwrap.String(n.message, n.width-2)))
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

func NewNotification() Notification {
	return Notification{
		keys: defaultNotificationKeys,
	}
}
