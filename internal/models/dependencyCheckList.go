package models

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

var hoverStyle lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

type model struct {
	Selected      map[string]struct{}
	filter        string
	dependencies  []Dependency
	filteredDeps  []Dependency
	filterField   textinput.Model
	paginate      paginator.Model
	cursor        int
	footerToggled bool
}

type Dependency struct {
	Id        string
	Name      string
	GroupName string
}

func (m model) View() string {
	body := m.bodyView()
	body = lipgloss.Place(100, m.paginate.PerPage, lipgloss.Left, lipgloss.Top, body)
	body = lipgloss.JoinVertical(lipgloss.Center, body, m.paginate.View())

	footer := "press / to filter, ↑↓ to navigate"

	if m.footerToggled {
		footer = m.filterField.View()
	}
	return lipgloss.JoinVertical(lipgloss.Left, body, footer)
}

func (m model) bodyView() string {
	body := strings.Builder{}

	start, end := m.paginate.GetSliceBounds(len(m.filteredDeps))
	for i, item := range m.filteredDeps[start:end] {
		currentIndex := i + start
		if _, ok := m.Selected[item.Id]; ok {
			body.WriteString("[✓] ")
		} else {
			body.WriteString("[ ] ")
		}

		itemDisplay := item.Name
		if currentIndex == m.cursor {
			itemDisplay = hoverStyle.Render(itemDisplay)
		}
		body.WriteString(itemDisplay)

		if i != m.paginate.PerPage-1 {
			body.WriteString("\n")
		}
	}
	return body.String()
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Is it a key press?
	case tea.KeyMsg:

		if m.footerToggled {

			switch msg.String() {

			case "enter":
				m.footerToggled = false

			case "esc", "ctrl+c":
				m.footerToggled = false
				m.filter = ""
				m.filterField.Reset()
				m.cursor = 0
				m.filteredDeps = m.dependencies
				m.paginate.SetTotalPages(len(m.filteredDeps))
				m.paginate.Page = 0
			}

			m.filterField, _ = m.filterField.Update(msg)
			newFilter := m.filterField.Value()

			if newFilter == m.filter {
				return m, nil
			}

			m.filter = newFilter

			m.filteredDeps = filterDeps(m.dependencies, m.filter)
			totalItems := len(m.filteredDeps)

			if totalItems == 0 {
				m.paginate.TotalPages = 1
			} else {
				m.paginate.SetTotalPages(len(m.filteredDeps))
			}
			m.paginate.Page = 0
			m.cursor = 0

			return m, nil
		}
		// Cool, what was the actual key pressed?
		switch msg.String() {

		case "/":
			m.footerToggled = !m.footerToggled
			if m.footerToggled {
				return m, m.filterField.Focus()
			}
			m.filterField.Blur()

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.paginate.Page = m.cursor / m.paginate.PerPage
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.filteredDeps)-1 {
				m.cursor++
				m.paginate.Page = m.cursor / m.paginate.PerPage
			}

		case "left", "h":
			m.paginate.PrevPage()
			m.cursor = m.paginate.Page * m.paginate.PerPage

		case "right", "l":
			m.paginate.NextPage()
			m.cursor = m.paginate.Page * m.paginate.PerPage

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			currentId := m.filteredDeps[m.cursor].Id
			if _, ok := m.Selected[currentId]; ok {
				delete(m.Selected, currentId)
			} else {
				m.Selected[currentId] = struct{}{}
			}
			sort.Slice(m.dependencies, func(i, j int) bool {
				if _, ok1 := m.Selected[m.dependencies[i].Id]; ok1 {
					if _, ok2 := m.Selected[m.dependencies[j].Id]; ok2 {
						return false
					} else {
						return true
					}
				}
				return false
			})
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func filterDeps(deps []Dependency, value string) []Dependency {
	filtered := make([]Dependency, 0)
	for _, dep := range deps {
		if fuzzy.MatchFold(value, dep.Name) {
			filtered = append(filtered, dep)
		}
	}
	return filtered
}

func NewModel(dependencies ...Dependency) tea.Model {
	filterField := textinput.New()
	filterField.Placeholder = "press esc to stop filtering"

	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = 10
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")
	p.SetTotalPages(len(dependencies))

	model := model{
		Selected:     make(map[string]struct{}),
		filterField:  filterField,
		dependencies: dependencies,
		filteredDeps: dependencies,
		paginate:     p,
	}

	return model
}
