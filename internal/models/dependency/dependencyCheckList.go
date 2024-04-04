package dependency

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eslam-allam/spring-initializer-go/constants"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/muesli/reflow/truncate"
)

var (
	itemStyle  lipgloss.Style = lipgloss.NewStyle()
	hoverStyle lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(constants.SecondaryColour))
)

type Model struct {
	Selected      map[string]struct{}
	filter        string
	mainKeys      MainKeyMap
	filterKeys    FilterKeyMap
	dependencies  []Dependency
	filteredDeps  []Dependency
	filterField   textinput.Model
	paginate      paginator.Model
	cursor        int
	filterToggled bool
	width         int
	height        int
}

func (m *Model) SetSize(h, v int) {
	m.width = h
	m.height = v

	if v < 3 {
		v = 3
	}

	m.paginate.PerPage = v - 2
	m.paginate.SetTotalPages(len(m.filteredDeps))
}

func (m Model) GetSize() (h, v int) {
	return m.width, m.height
}

func (m *Model) GetSelectedIds() []string {
	ids := make([]string, len(m.Selected))
	i := 0

	for id := range m.Selected {
		ids[i] = id
		i++
	}
	return ids
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
}

func (k MainKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

func (k MainKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.PagePrev, k.PageNext},
		{k.ToggleSelect, k.Filter},
	}
}

var defaultMainKeys = MainKeyMap{
	Up:           key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
	Down:         key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
	PagePrev:     key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "previous page")),
	PageNext:     key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "next page")),
	ToggleSelect: key.NewBinding(key.WithKeys("enter", " "), key.WithHelp("enter/space", "toggle selection")),
	Filter:       key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
}

type FilterKeyMap struct {
	Submit key.Binding
	Cancel key.Binding
}

func (k FilterKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

func (k FilterKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Submit, k.Cancel}}
}

var defaultFilterKeys = FilterKeyMap{
	Submit: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
	Cancel: key.NewBinding(key.WithKeys("esc", "ctrl+c"), key.WithHelp("esc", "cancel")),
}

func (m Model) View() string {
	body := m.bodyView()
	body = lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, body)

	body = lipgloss.JoinVertical(lipgloss.Center, body, truncate.StringWithTail(m.paginate.View(), uint(m.width), "…"))

	filter := m.filterField.View()

	if lipgloss.Width(filter) > m.width {
		filter = truncate.StringWithTail(filter, uint(m.width), "…")
	}
	return lipgloss.JoinVertical(lipgloss.Left, body, filter)
}

func (m Model) ShortHelp() []key.Binding {
	if m.filterToggled {
		return m.filterKeys.ShortHelp()
	}
	return m.mainKeys.ShortHelp()
}

func (m Model) FullHelp() [][]key.Binding {
	if m.filterToggled {
		return m.filterKeys.FullHelp()
	}
	return m.mainKeys.FullHelp()
}

func (m Model) bodyView() string {
	body := strings.Builder{}
	start, end := m.paginate.GetSliceBounds(len(m.filteredDeps))
	for i, item := range m.filteredDeps[start:end] {
		currentIndex := i + start
		if _, ok := m.Selected[item.Id]; ok {
			body.WriteString("[✓] ")
		} else {
			body.WriteString("[ ] ")
		}

		itemDisplay := itemStyle.Render(item.Name)
		if currentIndex == m.cursor {
			itemDisplay = hoverStyle.Render(itemDisplay)
		}

		if lipgloss.Width(itemDisplay) > m.width-4 {
			itemDisplay = truncate.StringWithTail(itemDisplay, uint(m.width-4), "…")
		}
		body.WriteString(itemDisplay)

		if i != m.paginate.PerPage-1 {
			body.WriteString("\n")
		}
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, body.String())
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) updateMain(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {

	case key.Matches(msg, m.mainKeys.Filter):
		m.filterToggled = !m.filterToggled
		if m.filterToggled {
			return m, m.filterField.Focus()
		}

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
			_, ok1 := m.Selected[m.dependencies[i].Id]
			_, ok2 := m.Selected[m.dependencies[j].Id]
			if ok1 && !ok2 {
				return true
			}
			if !ok1 && ok2 {
				return false
			}
			return m.dependencies[i].Name < m.dependencies[j].Name
		})
	}
	return m, nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:

		if m.filterToggled {
			return updateFilter(msg, m)
		}

		return m.updateMain(msg)
	}
	return m, nil
}

func updateFilter(msg tea.KeyMsg, m Model) (Model, tea.Cmd) {
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

func filterDeps(deps []Dependency, value string) []Dependency {
	filtered := make([]Dependency, 0)
	for _, dep := range deps {
		if fuzzy.MatchFold(value, dep.Name) {
			filtered = append(filtered, dep)
		}
	}
	return filtered
}

func New(dependencies ...Dependency) Model {
	sort.Slice(dependencies, func(i, j int) bool {
		return dependencies[i].Name < dependencies[j].Name
	})

	filterField := textinput.New()
	filterField.Placeholder = "Type here to filter dependencies..."

	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = 20
	p.InactiveDot = lipgloss.NewStyle().Render("•")
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color(constants.SecondaryColour)).Render(p.ActiveDot)
	p.SetTotalPages(len(dependencies))

	model := Model{
		Selected:     make(map[string]struct{}),
		filterField:  filterField,
		dependencies: dependencies,
		filteredDeps: dependencies,
		paginate:     p,
		mainKeys:     defaultMainKeys,
		filterKeys:   defaultFilterKeys,
	}

	return model
}
