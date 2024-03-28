package models

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	fuzzy "github.com/lithammer/fuzzysearch/fuzzy"
)

var hoverStyle lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

type model struct {
	Selected      map[int]struct{}
	filter        string
	depIds        []string
	depGroups     []string
	depNames      []string
	filteredNames []string
	filterField   textinput.Model
	Cursor        int
	currentPage   int
	pageSize      int
	footerToggled bool
}

type Dependency struct {
	Id        string
	Name      string
	GroupName string
}

func (m model) View() string {
	body := m.bodyView()

	footer := "press / to filter, ↑↓ to navigate"

	if m.footerToggled {
		footer = m.filterField.View()
	}
	return lipgloss.JoinVertical(lipgloss.Left, body.String(), footer)
}

func (m model) bodyView() strings.Builder {
	body := strings.Builder{}

	pageNumber := m.Cursor / m.pageSize
	startingIndex := pageNumber * m.pageSize
	for i, item := range m.filteredNames {
        if i < startingIndex {
            continue
        }

        if i > startingIndex + m.pageSize - 1 {
            return body
        }

		if _, ok := m.Selected[i]; ok {
			body.WriteString("[✓] ")
		} else {
			body.WriteString("[ ] ")
		}

		itemDisplay := item
		if i == m.Cursor {
			itemDisplay = hoverStyle.Render(itemDisplay)
		}
		body.WriteString(itemDisplay)

		body.WriteString("\n")
	}
	return body
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		case "/":
			m.footerToggled = !m.footerToggled
            if m.footerToggled {
                return m, m.filterField.Focus()
            }
            m.filterField.Blur()
			return m, nil

		// These keys should exit the program.
		case "ctrl+c":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
            return m, nil

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.Cursor < len(m.filteredNames)-1 {
				m.Cursor++
			}
            return m, nil

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":

			if _, ok := m.Selected[m.Cursor]; ok {
				delete(m.Selected, m.Cursor)
			} else {
				m.Selected[m.Cursor] = struct{}{}
			}
            return m, nil

        case "esc":
            if m.footerToggled{
                m.footerToggled = false
                m.filter = ""
                m.filterField.Reset()
                m.Cursor = 0
                m.filteredNames = m.depNames
            }
		}

		if m.footerToggled {
			m.filterField, _ = m.filterField.Update(msg)
			m.filter = m.filterField.Value()
			m.filteredNames = fuzzy.FindFold(m.filter, m.depNames)
			m.Cursor = 0
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func NewModel(dependencies ...Dependency) tea.Model {
	filterField := textinput.New()
	filterField.Placeholder = "press esc to stop filtering"
	depNames := make([]string, len(dependencies))
	depIds := make([]string, len(dependencies))
	depGroups := make([]string, len(dependencies))

	for i, d := range dependencies {
		depNames[i] = d.Name
		depIds[i] = d.Id
		depGroups[i] = d.GroupName
	}
    filterField.SetSuggestions(depNames)

	model := model{
		Selected:    make(map[int]struct{}),
		pageSize:    5,
		depIds:      depIds,
		depNames:    depNames,
		depGroups:   depGroups,
		filterField: filterField,
        filteredNames: depNames,
	}

	return model
}
