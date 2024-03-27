package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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

type model struct {
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
	dependencies  []checkListGroup
	javaVersion   []checkListItem
}

func initialModel() model {
	metaData, err := getMeta()
	if err != nil {
		panic(err)
	}
	bootVersions := make([]checkListItem, len(metaData.BootVersion.Values))
	dependencies := make([]checkListGroup, len(metaData.Dependencies.Values))
	javaVersions := make([]checkListItem, len(metaData.JavaVersion.Values))

	for i, field := range metaData.BootVersion.Values {
		bootVersions[i] = checkListItem{
			id:   field.Id,
			name: field.Name,
		}
	}

	for i, dependencyGroup := range metaData.Dependencies.Values {
		g := checkListGroup{
			groupId: dependencyGroup.Name,
			items:   make([]checkListItem, len(dependencyGroup.Values)),
		}
		for j, groupItem := range dependencyGroup.Values {
			g.items[j] = checkListItem{
				id:   groupItem.Id,
				name: groupItem.Name,
			}
		}
		dependencies[i] = g
	}

	for i, version := range metaData.JavaVersion.Values {
		javaVersions[i] = checkListItem{
			id:   version.Id,
			name: version.Name,
		}
	}

	return model{
		bootVersion:  bootVersions,
		dependencies: dependencies,
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

func (m model) View() string {
	s := strings.Builder{}

	for _, dependencyGroup := range m.dependencies {
		s.WriteString(dependencyGroup.groupId)
		s.WriteString("\n")
		for _, item := range dependencyGroup.items {
			if item.checked {
				s.WriteString("✓")
			} else {
				s.WriteString(" ")
			}
			s.WriteString(item.name)
			s.WriteString("\n")
		}
		s.WriteString("\n")
	}
	return s.String()
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
        }
    }

    // Return the updated model to the Bubble Tea runtime for processing.
    // Note that we're not returning a command.
    return m, nil
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
