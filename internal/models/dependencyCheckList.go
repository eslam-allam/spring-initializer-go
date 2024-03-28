package models

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var hoverStyle lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

type item struct {
	id      string
	name    string
	checked string
}

type itemGroup struct {
	name  string
	items []item
}

type model struct {
	Selected    map[int]struct{}
	Groups      []itemGroup
	Cursor      int
	TotalItems  int
	currentPage int
	pageSize    int
}

type Dependency struct {
	Id        string
	Name      string
	GroupName string
}

func (m model) View() string {
	s := strings.Builder{}
	currentPage := m.Cursor / m.pageSize
	startingIndex := currentPage * m.pageSize
	lastIndex := startingIndex + m.pageSize - 1
	itemIndex := 0
	for _, page := range m.Groups {
		hasItems := false
		firstItem := true
		for _, item := range page.items {
			if itemIndex < startingIndex {
				itemIndex++
				continue
			}

			if itemIndex > lastIndex {
				break
			}

			if firstItem {
				s.WriteString(page.name)
				s.WriteString("\n")
				firstItem = false
			}
			hasItems = true

			if _, ok := m.Selected[itemIndex]; ok {
				s.WriteString("[✓] ")
			} else {
				s.WriteString("[ ] ")
			}
			itemDisplay := item.name
			if itemIndex == m.Cursor {
				itemDisplay = hoverStyle.Render(itemDisplay)
			}
			s.WriteString(itemDisplay)
			s.WriteString("\n")
			itemIndex++
		}
		if hasItems {
			s.WriteRune('\n')
		}
	}
	return s.String()
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

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.Cursor < m.TotalItems-1 {
				m.Cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":

			if _, ok := m.Selected[m.Cursor]; ok {
				delete(m.Selected, m.Cursor)
			} else {
				m.Selected[m.Cursor] = struct{}{}
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func NewModel(items ...Dependency) tea.Model {
	itemMap := make(map[string][]item)

	for _, o := range items {
		itemMap[o.GroupName] = append(itemMap[o.GroupName], item{
			id:   o.Id,
			name: o.Name,
		})
	}

	model := model{
		TotalItems: len(items),
		Selected:   make(map[int]struct{}),
		pageSize:   5,
		Groups:     make([]itemGroup, 0),
	}

	for group, items := range itemMap {

		if group == "" || len(items) == 0 {
			continue
		}
		model.Groups = append(model.Groups, itemGroup{name: group, items: items})
	}

	return model
}
