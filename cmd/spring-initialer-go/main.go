package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	models "github.com/eslam-allam/spring-initializer-go/internal/models"
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

const (
	DEPENDENCIES section = iota
)

type model struct {
	currentSection section

	artifactId    string
	description   string
	groupId       string
	language      string
	name          string
	packageName   string
	packaging     string
	packagingType string
	version       string
	bootVersion   []checkListItem
	dependencies  tea.Model
	javaVersion   []checkListItem
}

func initialModel() model {
	metaData, err := getMeta()
	if err != nil {
		panic(err)
	}
	bootVersions := make([]checkListItem, len(metaData.BootVersion.Values))
	dependencies := make([]models.Dependency, len(metaData.Dependencies.Values))
	javaVersions := make([]checkListItem, len(metaData.JavaVersion.Values))

	for i, field := range metaData.BootVersion.Values {
		bootVersions[i] = checkListItem{
			id:   field.Id,
			name: field.Name,
		}
	}
	for _, dependencyGroup := range metaData.Dependencies.Values {
		for _, groupItem := range dependencyGroup.Values {
			dependencies = append(dependencies, models.Dependency{GroupName: dependencyGroup.Name, Id: groupItem.Id, Name: groupItem.Name})
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
		dependencies: models.NewModel(dependencies...),
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

var hoverStyle lipgloss.Style = lipgloss.NewStyle().Background(lipgloss.Color("#FFFF00"))

func (m model) View() string {
	return m.dependencies.View()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.dependencies.Update(msg)
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
