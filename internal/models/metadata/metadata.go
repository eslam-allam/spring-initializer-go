package metadata

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eslam-allam/spring-initializer-go/constants"
	"github.com/muesli/reflow/truncate"
)

type Field struct {
	Name           string
	Id             string
	Default        string
	inputLastValue string
	input          textinput.Model
}

type Model struct {
	keys      KeyMap
	fieldKeys InputKeyMap
	fields    []Field
	cursor    int
	typing    bool
	width     int
	height    int
}

func (m *Model) SetSize(h, v int) {
	m.width = h
	m.height = v
}

type FieldValue struct {
	Id    string
	Value string
}

func (m Model) GetValues() []FieldValue {
	values := make([]FieldValue, len(m.fields))
	for i, field := range m.fields {
		value := field.inputLastValue
		if value == "" {
			value = field.Default
		}
		values[i] = FieldValue{
			Id:    field.Id,
			Value: value,
		}
	}
	return values
}

func (m Model) ShortHelp() []key.Binding {
	if m.typing {
		return m.fieldKeys.ShortHelp()
	}
	return m.keys.ShortHelp()
}

func (m Model) FullHelp() [][]key.Binding {
	if m.typing {
		return m.fieldKeys.FullHelp()
	}
	return m.keys.FullHelp()
}

type KeyMap struct {
	PREV  key.Binding
	NEXT  key.Binding
	FOCUS key.Binding
	CLEAR key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.PREV, k.NEXT}, {k.FOCUS, k.CLEAR}}
}

type InputKeyMap struct {
	SUBMIT key.Binding
	CANCEL key.Binding
}

func (k InputKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

func (k InputKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.SUBMIT, k.CANCEL}}
}

var DefaultKeyMap = KeyMap{
	PREV:  key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "previous")),
	NEXT:  key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "next")),
	FOCUS: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "focus")),
	CLEAR: key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("ctrl+l", "clear")),
}

var DefaultInputKeyMap = InputKeyMap{
	SUBMIT: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
	CANCEL: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.typing {
			field := &m.fields[m.cursor]
			switch {
			case key.Matches(msg, m.fieldKeys.SUBMIT):
				field.input.Blur()
				field.inputLastValue = field.input.Value()
				m.typing = false
			case key.Matches(msg, m.fieldKeys.CANCEL):
				field.input.SetValue(field.inputLastValue)
				field.input.Blur()
				m.typing = false
			default:
				field.input, cmd = field.input.Update(msg)

			}
		} else {
			switch {
			case key.Matches(msg, m.keys.PREV):
				if m.cursor > 0 {
					m.cursor--
				}
			case key.Matches(msg, m.keys.NEXT):
				if m.cursor < len(m.fields)-1 {
					m.cursor++
				}
			case key.Matches(msg, m.keys.CLEAR):
				m.fields[m.cursor].input.Reset()
				m.fields[m.cursor].inputLastValue = ""
			case key.Matches(msg, m.keys.FOCUS):
				cmd = m.fields[m.cursor].input.Focus()
				m.typing = true
			}
		}
	}
	return m, cmd
}

var hoverStyle lipgloss.Style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(constants.SecondaryColour))

func (m Model) View() string {
	s := strings.Builder{}

	for i, field := range m.fields {

		display := field.input.View()

		if i == m.cursor {
			display = hoverStyle.Render(display)
		}

        if lipgloss.Width(display) > m.width - 1 {
            display = truncate.StringWithTail(display, uint(m.width - 1), "…")
        }
		s.WriteString(display)

		if i < len(m.fields)-1 {
			s.WriteString("\n")
		}
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, s.String())
}

func New(fields ...Field) Model {
	newFields := make([]Field, len(fields))
	for i, field := range fields {
		input := textinput.New()
		input.Prompt = fmt.Sprintf("%s: ", strings.TrimSpace(field.Name))
		input.Placeholder = field.Default
		field.input = input
		newFields[i] = field
	}

	return Model{
		fields:    newFields,
		keys:      DefaultKeyMap,
		fieldKeys: DefaultInputKeyMap,
	}
}
