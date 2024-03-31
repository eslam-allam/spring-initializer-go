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
	project           radioList.Model
	language          radioList.Model
	springBootVersion radioList.Model
	packaging         radioList.Model
	javaVersion       radioList.Model
	metadata          metadata.Model
	dependencies      dependency.Model
	currentSection    section
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
	renderer := m.iteratingRenderer()

	leftSection := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Center, renderer(m.project.View()),
			lipgloss.JoinVertical(lipgloss.Center, renderer(m.language.View()), renderer(m.packaging.View()))),
		lipgloss.JoinHorizontal(lipgloss.Center,
			renderer(m.javaVersion.View()), renderer(m.springBootVersion.View())),
		renderer(m.metadata.View()),
	)
	rightSection := lipgloss.JoinVertical(lipgloss.Center, renderer(m.dependencies.View()), renderer("Buttons"))

	return docStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, leftSection, rightSection))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		hs, hv := sectionStyle.GetFrameSize()
		m.dependencies.SetSize((msg.Width-h)/2-hs, (msg.Height-v)/2-hv)
		m.project.SetSize((msg.Width-h)/4-hs, (msg.Height-v)/5-hv)
		m.language.SetSize((msg.Width-h)/4-hs, (msg.Height-v)/5-hv-3)
		m.springBootVersion.SetSize((msg.Width-h)/4-hs, (msg.Height-v)/5-hv)
		m.javaVersion.SetSize((msg.Width-h)/4-hs, (msg.Height-v)/5-hv)
		m.packaging.SetSize((msg.Width-h)/4-hs, 1)
		m.metadata.SetSize((msg.Width-h)/2-hs, (msg.Height-v)/2-hv)
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
