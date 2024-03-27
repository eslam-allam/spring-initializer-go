package models

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var hoverStyle lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

type item struct {
    id string
    name string
    checked string
}

type itemGroup struct {
    name string
    items []item
}

type dependencyList struct {
	Selected   map[int]struct{}
	Groups     []itemGroup
	Cursor     int
	TotalItems int
}

type Dependency struct {
    Id string
    Name string
    GroupName string
}

func (m dependencyList) View() string {
	s := strings.Builder{}
    
    counter := 0
	for i, dependencyGroup := range m.Groups {
        if i > 3 {
            break
        }
		s.WriteString(dependencyGroup.name)
		s.WriteString("\n")
		for _, item := range dependencyGroup.items {

			if _, ok := m.Selected[counter]; ok{
				s.WriteString("[✓] ")
			} else {
				s.WriteString("[ ] ")
			}
            itemDisplay := item.name
            if counter == m.Cursor {
                itemDisplay = hoverStyle.Render(itemDisplay)
            }
			s.WriteString(itemDisplay)
			s.WriteString("\n")
            counter++
		}
		s.WriteString("\n")
	}
	return s.String()
}

func (m dependencyList) Init() tea.Cmd  {
    return nil
}


func (m dependencyList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
            id: o.Id,
            name: o.Name,
        })
    }

    model := dependencyList{TotalItems: len(items), Selected: make(map[int]struct{})}

    for key, val := range itemMap {
        model.Groups = append(model.Groups, itemGroup{name: key, items: val})
    }

    return model
}
