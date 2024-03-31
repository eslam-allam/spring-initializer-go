package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eslam-allam/spring-initializer-go/internal/models/dependency"
	"github.com/eslam-allam/spring-initializer-go/internal/models/metadata"
	"github.com/eslam-allam/spring-initializer-go/internal/models/radioList"
)

const springUrl = "https://start.spring.io"

type metaFieldType string

const (
	TEXT          metaFieldType = "text"
	SINGLE_SELECT metaFieldType = "single-select"
	MULTI_SELECT  metaFieldType = "heirarchal-multi-select"
	ACTION        metaFieldType = "action"
)

type metaField struct {
	Id          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Type        metaFieldType `json:"type"`
	Default     string        `json:"default"`
	Action      string        `json:"action"`
	Values      []metaField   `json:"values"`
}

type springInitMeta struct {
	ArtifactId   metaField `json:"artifactId"`
	BootVersion  metaField `json:"bootVersion"`
	Dependencies metaField `json:"dependencies"`
	Description  metaField `json:"description"`
	GroupId      metaField `json:"groupId"`
	JavaVersion  metaField `json:"javaVersion"`
	Language     metaField `json:"language"`
	Name         metaField `json:"name"`
	PackageName  metaField `json:"packageName"`
	Packaging    metaField `json:"packaging"`
	Type         metaField `json:"type"`
	Version      metaField `json:"version"`
}

type checkListItem struct {
	id      string
	name    string
	checked bool
}

type checkListGroup struct {
	groupId string
	items   []checkListItem
}

type section int

const NSECTIONS = 8

const (
	PROJECT section = iota
	LANGUAGE
	PACKAGING
	JAVA
	SPRING_BOOT
	METADATA
	DEPENDENCIES
	BUTTONS
)

type model struct {
	help              help.Model
	currentHelp       string
	keys              MainKeyMap
	metadata          metadata.Model
	dependencies      dependency.Model
	packaging         radioList.Model
	javaVersion       radioList.Model
	project           radioList.Model
	springBootVersion radioList.Model
	language          radioList.Model
	currentSection    section
	width             int
	height            int
}

type MainKeyMap struct {
	NEXT_SECTION     key.Binding
	PREV_SECTION     key.Binding
	HELP             key.Binding
	QUIT             key.Binding
	SectionShortKeys []key.Binding
	SectionFullKeys  [][]key.Binding
}

func (k MainKeyMap) ShortHelp() []key.Binding {
	return append([]key.Binding{k.HELP, k.QUIT}, k.SectionShortKeys...)
}

func (k MainKeyMap) FullHelp() [][]key.Binding {
	return append([][]key.Binding{{k.NEXT_SECTION, k.PREV_SECTION}, {k.HELP, k.QUIT}}, k.SectionFullKeys...)
}

var defaultKeys MainKeyMap = MainKeyMap{
	NEXT_SECTION: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next section")),
	PREV_SECTION: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "previous section")),
	HELP:         key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
	QUIT:         key.NewBinding(key.WithKeys("ctrl+q"), key.WithHelp("ctrl+q", "quit")),
}

func initialModel() model {
	metaData, err := getMeta()
	if err != nil {
		panic(err)
	}
	bootVersions := make([]radioList.Item, len(metaData.BootVersion.Values))
	dependencies := make([]dependency.Dependency, 0)
	javaVersions := make([]radioList.Item, len(metaData.JavaVersion.Values))
	projects := make([]radioList.Item, len(metaData.Type.Values))
	language := make([]radioList.Item, len(metaData.Language.Values))
	packaging := make([]radioList.Item, len(metaData.Packaging.Values))
	metadataFields := []metaField{metaData.GroupId, metaData.ArtifactId, metaData.Name, metaData.Description, metaData.PackageName}
	metadataFieldNames := []string{"Group", "Artifact", "Name", "Description", "Package Name"}
	metaDisplayFields := make([]metadata.Field, len(metadataFields))

	for i, field := range metadataFields {
		metaDisplayFields[i] = metadata.Field{
			Id:      field.Id,
			Name:    metadataFieldNames[i],
			Default: field.Default,
		}
	}
	for i, field := range metaData.Packaging.Values {
		packaging[i] = radioList.Item{
			Id:   field.Id,
			Name: field.Name,
		}
	}

	for i, field := range metaData.Language.Values {
		language[i] = radioList.Item{
			Id:   field.Id,
			Name: field.Name,
		}
	}

	for i, field := range metaData.Type.Values {
		projects[i] = radioList.Item{
			Id:     field.Id,
			Name:   field.Name,
			Action: field.Action,
		}
	}

	for i, field := range metaData.BootVersion.Values {
		bootVersions[i] = radioList.Item{
			Id:   field.Id,
			Name: field.Name,
		}
	}
	for _, dependencyGroup := range metaData.Dependencies.Values {
		for _, groupItem := range dependencyGroup.Values {
			dependencies = append(dependencies, dependency.Dependency{GroupName: dependencyGroup.Name, Id: groupItem.Id, Name: groupItem.Name})
		}
	}

	for i, version := range metaData.JavaVersion.Values {
		javaVersions[i] = radioList.Item{
			Id:   version.Id,
			Name: version.Name,
		}
	}

	return model{
		project:           radioList.New(radioList.VERTICAL, projects...),
		language:          radioList.New(radioList.VERTICAL, language...),
		springBootVersion: radioList.New(radioList.VERTICAL, bootVersions...),
		dependencies:      dependency.New(dependencies...),
		javaVersion:       radioList.New(radioList.VERTICAL, javaVersions...),
		packaging:         radioList.New(radioList.HORIZONTAL, packaging...),
		metadata:          metadata.New(metaDisplayFields...),
		help:              help.New(),
		keys:              defaultKeys,
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

var (
	docStyle            lipgloss.Style = lipgloss.NewStyle().Margin(2, 1)
	hoverStyle          lipgloss.Style = lipgloss.NewStyle().Background(lipgloss.Color("#FFFF00"))
	sectionStyle        lipgloss.Style = lipgloss.NewStyle().Margin(0, 0, 0).Border(lipgloss.RoundedBorder(), true)
	currentSectionStyle lipgloss.Style = lipgloss.NewStyle().Inherit(sectionStyle).BorderForeground(lipgloss.Color("205"))
)

func renderSection(s string, isCurrent bool) string {
	if isCurrent {
		return currentSectionStyle.Render(s)
	}
	return sectionStyle.Render(s)
}

func (m model) iteratingRenderer() func(s string) string {
	i := 0
	return func(s string) string {
		section := renderSection(s, i == int(m.currentSection))
		i++
		return section
	}
}

func (m model) View() string {
	m.updateHelp()
	renderer := m.iteratingRenderer()

	leftSection := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Center, renderer(m.project.View()),
			lipgloss.JoinVertical(lipgloss.Center, renderer(m.language.View()), renderer(m.packaging.View()))),
		lipgloss.JoinHorizontal(lipgloss.Center,
			renderer(m.javaVersion.View()), renderer(m.springBootVersion.View())),
		renderer(m.metadata.View()),
	)
	rightSection := lipgloss.JoinVertical(lipgloss.Center, renderer(m.dependencies.View()), renderer("Buttons"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, docStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.JoinHorizontal(lipgloss.Top, leftSection, rightSection),
			m.help.View(m.keys))))
}

func (m *model) updateHelp() {
	switch m.currentSection {
	case PROJECT:
		m.keys.SectionShortKeys = m.project.ShortHelp()
		m.keys.SectionFullKeys = m.project.FullHelp()
	case LANGUAGE:
		m.keys.SectionShortKeys = m.language.ShortHelp()
		m.keys.SectionFullKeys = m.language.FullHelp()
	case SPRING_BOOT:
		m.keys.SectionShortKeys = m.springBootVersion.ShortHelp()
		m.keys.SectionFullKeys = m.springBootVersion.FullHelp()
	case METADATA:

	case PACKAGING:
		m.keys.SectionShortKeys = m.packaging.ShortHelp()
		m.keys.SectionFullKeys = m.packaging.FullHelp()
	case JAVA:
		m.keys.SectionShortKeys = m.javaVersion.ShortHelp()
		m.keys.SectionFullKeys = m.javaVersion.FullHelp()
	case DEPENDENCIES:
		m.keys.SectionShortKeys = m.dependencies.ShortHelp()
		m.keys.SectionFullKeys = m.dependencies.FullHelp()
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.width, m.height = msg.Width, msg.Height
		hs, hv := sectionStyle.GetFrameSize()
		m.dependencies.SetSize((msg.Width-h)/2-hs, (msg.Height-v)/2-hv)
		m.project.SetSize((msg.Width-h)/4-hs, (msg.Height-v)/5-hv)
		m.language.SetSize((msg.Width-h)/4-hs, (msg.Height-v)/5-hv-3)
		m.springBootVersion.SetSize((msg.Width-h)/4-hs, (msg.Height-v)/5-hv)
		m.javaVersion.SetSize((msg.Width-h)/4-hs, (msg.Height-v)/5-hv)
		m.packaging.SetSize((msg.Width-h)/4-hs, 1)
		m.metadata.SetSize((msg.Width-h)/2-hs, (msg.Height-v)/2-hv)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.NEXT_SECTION):
			m.currentSection = (m.currentSection + 1) % NSECTIONS
		case key.Matches(msg, m.keys.PREV_SECTION):
			m.currentSection = (m.currentSection - 1 + NSECTIONS) % NSECTIONS
		case key.Matches(msg, m.keys.HELP):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.QUIT):
			return m, tea.Quit
		}
		switch m.currentSection {

		case PROJECT:
			m.project, cmd = m.project.Update(msg)
		case LANGUAGE:
			m.language, cmd = m.language.Update(msg)
		case PACKAGING:
			m.packaging, cmd = m.packaging.Update(msg)
		case SPRING_BOOT:
			m.springBootVersion, cmd = m.springBootVersion.Update(msg)
		case JAVA:
			m.javaVersion, cmd = m.javaVersion.Update(msg)
		case METADATA:
			m.metadata, cmd = m.metadata.Update(msg)
		case DEPENDENCIES:
			m.dependencies, cmd = m.dependencies.Update(msg)
		}
	}
	return m, cmd
}

func getMeta() (springInitMeta, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", springUrl, nil)
	req.Header.Set("Accept", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return springInitMeta{}, err
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return springInitMeta{}, err
	}

	var responseObject springInitMeta
	err = json.Unmarshal(body, &responseObject)
	if err != nil {
		return springInitMeta{}, err
	}
	return responseObject, nil
}
