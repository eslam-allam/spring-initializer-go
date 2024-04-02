package buttons

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eslam-allam/spring-initializer-go/constants"
)

type Action int

const (
	DOWNLOAD Action = iota
	DOWNLOAD_EXTRACT
)

type ActionState int

const (
	ACTION_IDOL ActionState = iota
	ACTION_SUCCESS
	ACTION_FAILED
	ACTION_RESET
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
	keys        KeyMap
	buttons     []Button
	spinner     spinner.Model
	cursor      int
	width       int
	height      int
	actionIndex int
	actionState ActionState
	inAction    bool
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
				BorderForeground(lipgloss.Color(constants.SecondaryColour)).Foreground(lipgloss.Color(constants.SecondaryColour))
	successMessageStyle lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7ae878"))
	failureMessageStyle lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
)

func (m Model) View() string {
	var s string

	switch m.actionState {
	case ACTION_SUCCESS:

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, successMessageStyle.Render("Download Successful!"))
	case ACTION_FAILED:

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, failureMessageStyle.Render("Download Failed, check logs for more info."))
	}

	if m.inAction {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, lipgloss.JoinHorizontal(lipgloss.Left, m.spinner.View(), "Downloading..."))
	}

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

	case ActionState:
		m.inAction = false
		m.actionState = msg
		switch msg {
		case ACTION_SUCCESS, ACTION_FAILED:
			cmd = func() tea.Msg {
				time.Sleep(2 * time.Second)
				return ACTION_RESET
			}
		case ACTION_RESET:
			m.actionState = ACTION_IDOL
		}
	case spinner.TickMsg:
		if m.inAction {
			m.spinner, cmd = m.spinner.Update(msg)
		}
	case tea.KeyMsg:
		if m.inAction {
			return m, cmd
		}
		m.actionState = ACTION_IDOL
		switch {

		case key.Matches(msg, m.keys.NEXT):
			if m.cursor < len(m.buttons)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.PREV):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.SUBMIT):
			cmd = tea.Batch(getCmd(m.buttons[m.cursor].Action), m.spinner.Tick)
			m.inAction = true
			m.actionIndex = m.cursor
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
		spinner: spinner.New(spinner.WithSpinner(spinner.Dot)),
	}
}
