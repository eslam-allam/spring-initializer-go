package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eslam-allam/spring-initializer-go/internal/models/dependency"
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
    SPRING_BOOT
    METADATA
    PACKAGING
    JAVA
	DEPENDENCIES
    BUTTONS
)

type model struct {
	packaging      string
	packageName    string
	description    string
	groupId        string
	language       string
	name           string
	packagingType  string
	version        string
	artifactId     string
	bootVersion    []checkListItem
	javaVersion    []checkListItem
	dependencies   dependency.Model
	currentSection section
}

func initialModel() model {
	metaData, err := getMeta()
	if err != nil {
		panic(err)
	}
	bootVersions := make([]checkListItem, len(metaData.BootVersion.Values))
	dependencies := make([]dependency.Dependency, 0)
	javaVersions := make([]checkListItem, len(metaData.JavaVersion.Values))

	for i, field := range metaData.BootVersion.Values {
		bootVersions[i] = checkListItem{
			id:   field.Id,
			name: field.Name,
		}
	}
	for _, dependencyGroup := range metaData.Dependencies.Values {
		for _, groupItem := range dependencyGroup.Values {
			dependencies = append(dependencies, dependency.Dependency{GroupName: dependencyGroup.Name, Id: groupItem.Id, Name: groupItem.Name})
		}
	}

	for i, version := range metaData.JavaVersion.Values {
		javaVersions[i] = checkListItem{
			id:   version.Id,
			name: version.Name,
		}
	}

	return model{
		bootVersion:  bootVersions,
		dependencies: dependency.NewModel(dependencies...),
		javaVersion:  javaVersions,
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

var docStyle lipgloss.Style = lipgloss.NewStyle().Margin(2, 1)
var hoverStyle lipgloss.Style = lipgloss.NewStyle().Background(lipgloss.Color("#FFFF00"))
var sectionStyle lipgloss.Style = lipgloss.NewStyle().Margin(0, 0, 0).Border(lipgloss.RoundedBorder(), true)
var currentSectionStyle lipgloss.Style = lipgloss.NewStyle().Inherit(sectionStyle).BorderForeground(lipgloss.Color("205"))

func renderSection(s string, width, height int, isCurrent bool) string {
    block := lipgloss.Place(width, height, lipgloss.Center, lipgloss.Top, s, lipgloss.WithWhitespaceChars("-"))

    if isCurrent {
        return currentSectionStyle.Render(block)
    }
    return sectionStyle.Render(block)
        
}

func (m model) iteratingRenderer() func(s string, width, height int) string {
    i := 0
    return func(s string, width, height int) string {
        section := renderSection(s, width, height, i == int(m.currentSection))
        i++
        return section
    }
}

func (m model) View() string {
    renderer := m.iteratingRenderer()

    leftSection := lipgloss.JoinVertical(lipgloss.Center,
        renderer("Project",40, 4),
        renderer("Language",40, 4),
        renderer("Spring Boot",40, 4),
        renderer("ProjectMetadata",40, 12),
        renderer("Packaging",40, 2),
        renderer("Java",40, 2),
		)
    rightSection := lipgloss.JoinVertical(lipgloss.Center, renderer(m.dependencies.View(), 20 ,20), renderer("Buttons", 10, 2))

    return docStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, leftSection, rightSection))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        h, v := docStyle.GetFrameSize()
        hs, hv := sectionStyle.GetFrameSize()
        m.dependencies.SetSize((msg.Width-h)/2 - hs, ( msg.Height-v ) /2 -hv )
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.currentSection = (m.currentSection + 1) % NSECTIONS
		case "shift+tab":
			m.currentSection = (m.currentSection - 1 + NSECTIONS) % NSECTIONS

		case "q":
			return m, tea.Quit
		}
		switch m.currentSection {

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
