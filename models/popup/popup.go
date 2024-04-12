package popup

import (
	"log"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eslam-allam/spring-initializer-go/constants"
	"github.com/eslam-allam/spring-initializer-go/models/overlay"
)

var logger *log.Logger = log.Default()

type PopupModule interface {
	Update(msg tea.Msg) (PopupModule, tea.Cmd)
	View() string
	SetSize(h, v int) PopupModule
	GetSize() (h, v int)
	ShortHelp() []key.Binding
	FullHelp() [][]key.Binding
}

type PopUpMessage struct {
	Content PopupModule
	Title   string
}

type PopupKeys struct {
	DISMISS        key.Binding
	innerKeysShort []key.Binding
	innerKeysLong  [][]key.Binding
}

func (p PopupKeys) ShortHelp() []key.Binding {
	return append(p.innerKeysShort, p.DISMISS)
}

func (p PopupKeys) FullHelp() [][]key.Binding {
	return append(p.innerKeysLong, []key.Binding{p.DISMISS})
}

var defaultPopupKeys = PopupKeys{
	DISMISS: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "dismiss")),
}

type Model struct {
	title      string
	innerModel PopupModule
	keys       PopupKeys
	active     bool
	height     int
	width      int
}

func (m Model) IsActive() bool {
	return m.active
}

func (m *Model) SetSize(h, v int) {
	m.width = h
	m.height = v
	hp, vp := popupStyle.GetFrameSize()

	if m.innerModel != nil {
		m.innerModel.SetSize(h-hp, v-vp)
	}
}

func (m Model) ShortHelp() []key.Binding {
	return m.keys.ShortHelp()
}

func (m Model) FullHelp() [][]key.Binding {
	return m.keys.FullHelp()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case PopUpMessage:
		logger.Println("PopUpMessage")
		m.active = true
		m.innerModel = msg.Content
		m.title = msg.Title
		m.keys.innerKeysShort = m.innerModel.ShortHelp()
		m.keys.innerKeysLong = m.innerModel.FullHelp()
		m.SetSize(m.width, m.height)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.DISMISS):
			m.active = false
			return m, cmd
		}
	}
	m.innerModel, cmd = m.innerModel.Update(msg)
	return m, cmd
}

var popupStyle lipgloss.Style = lipgloss.NewStyle().Padding(1).
	Border(lipgloss.ThickBorder(), true).
	BorderForeground(lipgloss.Color(constants.HighlightColour))

func (m Model) View() string {
	body := popupStyle.Render(m.innerModel.View())
	x := popupStyle.GetHorizontalFrameSize() / 2
	return overlay.PlaceTitle(m.title, body, 0, 0, x, 0)
}

func New() Model {
	return Model{
		keys:   defaultPopupKeys,
		active: false,
	}
}
