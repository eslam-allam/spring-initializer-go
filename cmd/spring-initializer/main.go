package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eslam-allam/spring-initializer-go/constants"
	"github.com/eslam-allam/spring-initializer-go/internal/models/buttons"
	"github.com/eslam-allam/spring-initializer-go/internal/models/dependency"
	"github.com/eslam-allam/spring-initializer-go/internal/models/metadata"
	"github.com/eslam-allam/spring-initializer-go/internal/models/radioList"
)

var logger *log.Logger = log.Default()

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
	javaVersion       radioList.Model
	packaging         radioList.Model
	project           radioList.Model
	springBootVersion radioList.Model
	language          radioList.Model
	buttons           buttons.Model
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

	metaDisplayFields := []metadata.Field{
		metadata.NewField("Group", "groupId", metaData.GroupId.Default, metadata.WithLink(4)),
		metadata.NewField("Artifact", "artifactId", metaData.ArtifactId.Default, metadata.WithLink(4, 2)),
		metadata.NewField("Name", "name", metaData.Name.Default, metadata.UpdatesFrom(' ', 1)),
		metadata.NewField("Description", "description", metaData.Description.Default),
		metadata.NewField("Package Name", "packageName", metaData.PackageName.Default, metadata.UpdatesFrom('.', 0, 1)),
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
			Id:   sanitizeId(field.Name),
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
		buttons: buttons.New([]buttons.Button{
			{Name: "Download", Action: buttons.DOWNLOAD},
			{Name: "Download and Extract", Action: buttons.DOWNLOAD_EXTRACT},
		}...),
	}
}

func sanitizeId(s string) string {
	sanitized := strings.TrimSpace(s)
	sanitized = strings.ReplaceAll(sanitized, " ", "-")
	sanitized = strings.ReplaceAll(sanitized, "(", "")
	sanitized = strings.ReplaceAll(sanitized, ")", "")

	return sanitized
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func main() {
	tmpDir := os.TempDir()
	f, err := tea.LogToFile(path.Join(tmpDir, "spring-init.log"), "Main loop")
	if err != nil {
		fmt.Printf("Failed to start logger: %v", err)
		os.Exit(1)
	}
	defer f.Close()
	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		logger.Fatalf("Error occurred in main loop: %v", err)
	}
}

var (
	docStyle          lipgloss.Style = lipgloss.NewStyle().Border(lipgloss.ThickBorder(), true).Padding(1)
	sectionTitleStyle lipgloss.Style = lipgloss.NewStyle().Bold(true).Border(lipgloss.NormalBorder(), true, true, false).
				PaddingBottom(1).Bold(true).PaddingLeft(1)
	currentSectionTitleStyle lipgloss.Style = sectionTitleStyle.Copy().BorderForeground(lipgloss.Color(constants.MainColour))
	sectionStyle             lipgloss.Style = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, true, true).PaddingLeft(1)
	currentSectionStyle      lipgloss.Style = sectionStyle.Copy().BorderForeground(lipgloss.Color(constants.MainColour))
)

func renderSection(title, s string, isCurrent bool) string {
	section := sectionStyle.Render(s)
	paddedTitle := lipgloss.PlaceHorizontal(lipgloss.Width(section)-3, lipgloss.Left, title)
	sectionTitle := sectionTitleStyle.Render(paddedTitle)
	if isCurrent {
		section = currentSectionStyle.Render(s)
		sectionTitle = currentSectionTitleStyle.Render(paddedTitle)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sectionTitle, section)
}

func (m model) iteratingRenderer() func(title, s string) string {
	i := 0
	return func(title, s string) string {
		section := renderSection(title, s, i == int(m.currentSection))
		i++
		return section
	}
}

func (m model) View() string {
	m.updateHelp()
	renderer := m.iteratingRenderer()
	leftSection := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Center, renderer("Project", m.project.View()),
			lipgloss.JoinVertical(lipgloss.Center, renderer("Language", m.language.View()), renderer("Packaging", m.packaging.View()))),
		lipgloss.JoinHorizontal(lipgloss.Center,
			renderer("Java", m.javaVersion.View()), renderer("Spring Boot", m.springBootVersion.View())),
		renderer("Project Metadata", m.metadata.View()),
	)
	rightSection := lipgloss.JoinVertical(lipgloss.Center, renderer("Dependencies", m.dependencies.View()), renderer("Generate", m.buttons.View()))
	h, v := docStyle.GetFrameSize()
	return docStyle.Render(lipgloss.Place(m.width-h, m.height-v, lipgloss.Center, lipgloss.Center,
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
		m.keys.SectionShortKeys = m.metadata.ShortHelp()
		m.keys.SectionFullKeys = m.metadata.FullHelp()
	case PACKAGING:
		m.keys.SectionShortKeys = m.packaging.ShortHelp()
		m.keys.SectionFullKeys = m.packaging.FullHelp()
	case JAVA:
		m.keys.SectionShortKeys = m.javaVersion.ShortHelp()
		m.keys.SectionFullKeys = m.javaVersion.FullHelp()
	case DEPENDENCIES:
		m.keys.SectionShortKeys = m.dependencies.ShortHelp()
		m.keys.SectionFullKeys = m.dependencies.FullHelp()
	case BUTTONS:
		m.keys.SectionShortKeys = m.buttons.ShortHelp()
		m.keys.SectionFullKeys = m.buttons.FullHelp()
	}
}

func (m model) generateDownloadRequest() (*url.URL, error) {
	form := url.Values{}

	for _, m := range m.metadata.GetValues() {
		form.Add(m.Id, m.Value)
	}

	form.Add("type", m.project.GetSelected().Id)
	form.Add("language", m.language.GetSelected().Id)
	form.Add("bootVersion", m.springBootVersion.GetSelected().Id)
	form.Add("packaging", m.packaging.GetSelected().Id)
	form.Add("javaVersion", m.javaVersion.GetSelected().Id)

	for _, d := range m.dependencies.GetSelectedIds() {
		form.Add("dependencies", d)
	}

	url, error := url.Parse(fmt.Sprintf("%s?%s", springUrl, form.Encode()))

	if error != nil {
		return url, error
	}

	url = url.JoinPath(m.project.GetSelected().Action)

	return url, nil
}

func downloadGeneratedZip(url string, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error downloading file: %s, %s", resp.Status, body)
	}
	defer resp.Body.Close()
	// Create the output file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	// Copy the response body to the output file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func unzipFile(zipFile, destDir string) error {
	// Open the zip file for reading
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()
	// Create the destination directory if it doesn't exist
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		os.MkdirAll(destDir, os.ModePerm)
	}
	// Extract each file from the zip archive
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		path := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, os.ModePerm)
		} else {
			outFile, err := os.Create(path)
			if err != nil {
				return err
			}
			defer outFile.Close()
			_, err = io.Copy(outFile, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func cellDimentions(h, v int, mh, mv float64, cieling ...bool) (int, int) {
	rh, rv := float64(h)*mh, float64(v)*mv
	for i, c := range cieling {
		switch i {
		case 0:
			if c {
				rh = math.Ceil(rh)
			} else {
				rh = math.Floor(rh)
			}
		case 1:
			if c {
				rv = math.Ceil(rv)
			} else {
				rv = math.Floor(rv)
			}
		}
	}
	return int(rh), int(rv)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case spinner.TickMsg:
		m.buttons, cmd = m.buttons.Update(msg)

	case buttons.ActionState:
		m.buttons, cmd = m.buttons.Update(msg)

	case buttons.Action:
		switch msg {
		case buttons.DOWNLOAD:
			cmd = func() tea.Msg {
				url, _ := m.generateDownloadRequest()

				cwd, err := os.Getwd()
				if err != nil {
					return buttons.ACTION_FAILED
				}

				err = downloadGeneratedZip(url.String(), path.Join(cwd, path.Base(url.Path)))
				if err != nil {
					return buttons.ACTION_FAILED
				}
				return buttons.ACTION_SUCCESS
			}
		case buttons.DOWNLOAD_EXTRACT:
			cmd = func() tea.Msg {
				url, _ := m.generateDownloadRequest()

				cwd, err := os.Getwd()
				if err != nil {
					return buttons.ACTION_FAILED
				}

				base := path.Base(url.Path)
				assetPath := path.Join(cwd, base)
				err = downloadGeneratedZip(url.String(), assetPath)
				if err != nil {
					return buttons.ACTION_FAILED
				}

				if strings.HasSuffix(base, "zip") {
					err = unzipFile(assetPath, cwd)
					if err != nil {
						return buttons.ACTION_FAILED
					}
					os.Remove(assetPath)
				}

				return buttons.ACTION_SUCCESS
			}
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.width, m.height = msg.Width, msg.Height

		hs, vs := sectionStyle.GetFrameSize()
		hs, vs = hs+1, vs+sectionTitleStyle.GetVerticalFrameSize()+1
		cellDimentsionCalc := func(eh, ev int, mh, mv float64, cieling ...bool) (int, int) {
			return cellDimentions(m.width-h-(hs*eh), m.height-v-(vs*ev)-2, mh, mv, cieling...)
		}

		cw, cv := cellDimentsionCalc(3, 4, 0.25, 0.2, false, false)
		ph, pv := cw, cv+1+vs
		m.project.SetSize(ph, pv)
		m.language.SetSize(cw, cv)
		m.packaging.SetSize(cw, 1)

		_, cmv := cellDimentsionCalc(3, 3, 1, 1, false, false)
		m.springBootVersion.SetSize(cw, cmv-5-pv)
		m.javaVersion.SetSize(cw, cmv-5-pv)

		// c2w, _ := cellDimentsionCalc(2, 3, 0.5, 0.25, false, false)

		c2w := cw*2 + hs - 1
		m.metadata.SetSize(c2w, 5)

		m.dependencies.SetSize(c2w, pv+(cmv-5-pv)+2)
		m.buttons.SetSize(c2w, 5)
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
		case BUTTONS:
			m.buttons, cmd = m.buttons.Update(msg)
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