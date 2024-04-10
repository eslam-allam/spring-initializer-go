package notification

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eslam-allam/spring-initializer-go/constants"
	"github.com/eslam-allam/spring-initializer-go/models/overlay"
	"github.com/muesli/reflow/wordwrap"
	"golang.design/x/clipboard"
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
				Border(lipgloss.ThickBorder(), true).
				BorderForeground(lipgloss.Color(constants.HighlightColour))
	notificationTextStyle = lipgloss.NewStyle()
)

type Model struct {
	message     string
	keys        NotificationKeyMap
	level       NotificationLevel
	width       int
	height      int
	active      bool
	copyAllowed bool
	copied      bool
}

func (m Model) IsActive() bool {
	return m.active
}

func (m *Model) Activate() {
	m.active = true
}

type CopyDone struct{}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case CopyDone:
		m.copied = false
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.DISMISS):
			m.active = false
		case m.copyAllowed && key.Matches(msg, m.keys.COPY):
			clipboard.Write(clipboard.FmtText, []byte(m.message))
			m.copied = true
			cmd = func() tea.Msg {
				time.Sleep(2 * time.Second)
				return CopyDone{}
			}

		}
	}
	return m, cmd
}

func (m Model) UpdateMessage(msg NotificationMsg) Model {
	m.message = msg.Message
	m.level = msg.Level
	return m
}

func (m Model) ShortHelp() []key.Binding {
	return m.keys.ShortHelp(m.copyAllowed)
}

func (m Model) FullHelp() [][]key.Binding {
	return m.keys.FullHelp(m.copyAllowed)
}

func (m *Model) SetSize(h, v int) {
	m.width = h
	m.height = v
}

func (m Model) View() string {
	var currentColor string
	var title string

	switch m.level {
	case INFO:
		currentColor = infoColor
		title = "INFO"
	case WARNING:
		currentColor = warningColor
		title = "WARNING"
	case ERROR:
		currentColor = errorColor
		title = "ERROR"
	}

	currentNotificationStyle := notificationStyle.Copy()

	if m.copied {
		title = "COPIED"
		currentNotificationStyle.BorderForeground(lipgloss.Color(infoColor))
	}

	body := currentNotificationStyle.
		Render(
			notificationTextStyle.MaxWidth(m.width).MaxHeight(m.height).
				Foreground(lipgloss.Color(currentColor)).
				Render(wordwrap.String(m.message, m.width-notificationStyle.GetHorizontalFrameSize())))
	x := notificationStyle.GetHorizontalFrameSize()/2 + notificationTextStyle.GetHorizontalFrameSize()/2
	return overlay.PlaceTitle(title, body, 0, 0, x, 0)
}

type NotificationKeyMap struct {
	DISMISS key.Binding
	COPY    key.Binding
}

func (k NotificationKeyMap) ShortHelp(copyAllowed bool) []key.Binding {
	keys := []key.Binding{k.DISMISS}
	if copyAllowed {
		keys = append(keys, k.COPY)
	}
	return keys
}

func (k NotificationKeyMap) FullHelp(copyAllowed bool) [][]key.Binding {
	keys := [][]key.Binding{{k.DISMISS}}
	if copyAllowed {
		keys[0] = append(keys[0], k.COPY)
	}
	return keys
}

var defaultNotificationKeys NotificationKeyMap = NotificationKeyMap{
	DISMISS: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "dismiss")),
	COPY:    key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "copy to clipboard")),
}

func New() Model {
	copyAllowed := false
	err := clipboard.Init()
	if err == nil {
		copyAllowed = true
	}
	return Model{
		keys:        defaultNotificationKeys,
		copyAllowed: copyAllowed,
	}
}
