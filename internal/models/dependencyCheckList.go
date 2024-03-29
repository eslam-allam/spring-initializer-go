package models

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

var hoverStyle lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

type model struct {
	mainKeys      MainKeyMap
	filterKeys    FilterKeyMap
	Selected      map[string]struct{}
	filter        string
	dependencies  []Dependency
	filteredDeps  []Dependency
	filterField   textinput.Model
	paginate      paginator.Model
	cursor        int
	filterToggled bool
	help          help.Model
}

type Dependency struct {
	Id        string
	Name      string
	GroupName string
}

type MainKeyMap struct {
	Up           key.Binding
	Down         key.Binding
	PagePrev     key.Binding
	PageNext     key.Binding
	ToggleSelect key.Binding
	Filter       key.Binding
	Help         key.Binding
	Quit         key.Binding
}

func (k MainKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k MainKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.PagePrev, k.PageNext},
		{k.ToggleSelect, k.Help},
		{k.Filter, k.Quit},
	}
}

var defaultMainKeys = MainKeyMap{
	Up:           key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
	Down:         key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
	PagePrev:     key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "previous page")),
	PageNext:     key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "next page")),
	ToggleSelect: key.NewBinding(key.WithKeys("enter", " "), key.WithHelp("enter/space", "toggle selection")),
	Filter:       key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
	Help:         key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Quit:         key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q/ctrl+c", "quit")),
}

type FilterKeyMap struct {
	Submit key.Binding
	Cancel key.Binding
}

func (k FilterKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Submit, k.Cancel}
}

func (k FilterKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}

var defaultFilterKeys = FilterKeyMap{
	Submit: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
	Cancel: key.NewBinding(key.WithKeys("esc", "ctrl+c"), key.WithHelp("esc", "cancel")),
}

func (m model) View() string {
	body := m.bodyView()
	body = lipgloss.Place(100, m.paginate.PerPage, lipgloss.Left, lipgloss.Top, body)
	body = lipgloss.JoinVertical(lipgloss.Center, body, m.paginate.View())

	footer := m.help.View(m.mainKeys)
	filter := m.filterField.View()
	if m.filterToggled {
		footer = m.help.View(m.filterKeys)
	}
	return lipgloss.JoinVertical(lipgloss.Left, body, filter, footer)
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

		if m.filterToggled {

			switch {

			case key.Matches(msg, m.filterKeys.Submit):
				m.filterToggled = false
				m.filterField.Blur()

			case key.Matches(msg, m.filterKeys.Cancel):
				m.filterToggled = false
				m.filter = ""
				m.filterField.Reset()
				m.cursor = 0
				m.filteredDeps = m.dependencies
				m.paginate.SetTotalPages(len(m.filteredDeps))
				m.paginate.Page = 0
				m.filterField.Blur()
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
		switch {

		case key.Matches(msg, m.mainKeys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.mainKeys.Filter):
			m.help.ShowAll = false
			m.filterToggled = !m.filterToggled
			if m.filterToggled {
				return m, m.filterField.Focus()
			}

		// These keys should exit the program.
		case key.Matches(msg, m.mainKeys.Quit):
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case key.Matches(msg, m.mainKeys.Up):
			if m.cursor > 0 {
				m.cursor--
				m.paginate.Page = m.cursor / m.paginate.PerPage
			}

		// The "down" and "j" keys move the cursor down
		case key.Matches(msg, m.mainKeys.Down):
			if m.cursor < len(m.filteredDeps)-1 {
				m.cursor++
				m.paginate.Page = m.cursor / m.paginate.PerPage
			}

		case key.Matches(msg, m.mainKeys.PagePrev):
			m.paginate.PrevPage()
			m.cursor = m.paginate.Page * m.paginate.PerPage

		case key.Matches(msg, m.mainKeys.PageNext):
			m.paginate.NextPage()
			m.cursor = m.paginate.Page * m.paginate.PerPage

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case key.Matches(msg, m.mainKeys.ToggleSelect):
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
	filterField.Placeholder = "Type here to filter dependencies..."

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
		mainKeys:     defaultMainKeys,
		filterKeys:   defaultFilterKeys,
		help:         help.New(),
	}

	return model
}
