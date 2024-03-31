package radioList

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Item struct {
	Name   string
	Id     string
	Action string
}

type Model struct {
	keys      KeyMap
	choices   []Item
	cursor    int
	selected  int
	height    int
	width     int
	direction direction
}

func (m *Model) SetSize(h, v int) {
	m.height = v
	m.width = h
}

func (m Model) GetSelected() Item {
	return m.choices[m.selected]
}

func (m Model) ShortHelp() []key.Binding {
    return m.keys.ShortHelp()
}
func (m Model) FullHelp() [][]key.Binding {
    return m.keys.FullHelp()
}

type KeyMap struct {
	PREV   key.Binding
	NEXT   key.Binding
	SELECT key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
    return []key.Binding{}
}
 
func (k KeyMap) FullHelp() [][]key.Binding {
    return [][]key.Binding{{k.PREV, k.NEXT}, {k.SELECT}}
}
    
var defaultKeys = KeyMap{
	PREV:   key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "previous")),
	NEXT:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "next")),
	SELECT: key.NewBinding(key.WithKeys("enter", " "), key.WithHelp("enter/space", "select")),
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.PREV):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.NEXT):
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.SELECT):
			m.selected = m.cursor
		}
	}
	return m, nil
}

var hoverStyle lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("62"))

func (m Model) View() string {
	s := strings.Builder{}
	for i, choice := range m.choices {
		if i == m.selected {
			s.WriteString("(*) ")
		} else {
			s.WriteString("( ) ")
		}

		choiceDisplay := choice.Name
		if m.cursor == i {
			choiceDisplay = hoverStyle.Render(choice.Name)
		}
		s.WriteString(choiceDisplay)
		if i != len(m.choices)-1 {
			if m.direction == HORIZONTAL {
				s.WriteString(" ")
			} else {
				s.WriteString("\n")
			}
		}
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, s.String())
}

type direction int

const (
	HORIZONTAL direction = iota
	VERTICAL
)

func New(d direction, choices ...Item) Model {
	keys := defaultKeys

	if d == HORIZONTAL {
		keys.PREV = key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "prev"))
		keys.NEXT = key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "next"))
	}
	return Model{
		choices:   choices,
		keys:      keys,
		direction: d,
	}
}
