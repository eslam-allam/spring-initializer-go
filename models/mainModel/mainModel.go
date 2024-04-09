package mainModel

import (
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eslam-allam/spring-initializer-go/constants"
	"github.com/eslam-allam/spring-initializer-go/models/buttons"
	"github.com/eslam-allam/spring-initializer-go/models/dependency"
	"github.com/eslam-allam/spring-initializer-go/models/metadata"
	"github.com/eslam-allam/spring-initializer-go/models/notification"
	"github.com/eslam-allam/spring-initializer-go/models/overlay"
	"github.com/eslam-allam/spring-initializer-go/models/radioList"
	"github.com/eslam-allam/spring-initializer-go/service/files"
	"github.com/eslam-allam/spring-initializer-go/service/springio"
)

var logger *log.Logger = log.Default()

var (
	docStyle          lipgloss.Style = lipgloss.NewStyle().Border(lipgloss.ThickBorder(), true).Padding(1)
	sectionTitleStyle lipgloss.Style = lipgloss.NewStyle().Bold(true).Border(lipgloss.NormalBorder(), true, true, false).
				PaddingBottom(1).Bold(true).PaddingLeft(1).BorderForeground(lipgloss.Color(constants.MainColour))
	currentSectionTitleStyle lipgloss.Style = sectionTitleStyle.Copy().BorderForeground(lipgloss.Color(constants.HighlightColour))
	sectionStyle             lipgloss.Style = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, true, true).
					PaddingLeft(1).BorderForeground(lipgloss.Color(constants.MainColour))
	currentSectionStyle lipgloss.Style = sectionStyle.Copy().BorderForeground(lipgloss.Color(constants.HighlightColour))
)

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

type appState int

const (
	LOADING appState = iota
	READY
)

type model struct {
	help              help.Model
	currentHelp       string
	targetDirectory   string
	keys              MainKeyMap
	spinner           spinner.Model
	metadata          metadata.Model
	notification      notification.Model
	dependencies      dependency.Model
	packaging         radioList.Model
	springBootVersion radioList.Model
	language          radioList.Model
	javaVersion       radioList.Model
	project           radioList.Model
	buttons           buttons.Model
	state             appState
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

func (m model) generateProject() (fullPath string, isZip bool, err error) {
	url, err := springio.GenerateDownloadRequest(m.project.GetSelected().Action,
		m.project.GetSelected().Id,
		m.language.GetSelected().Id,
		m.springBootVersion.GetSelected().Id,
		m.packaging.GetSelected().Id,
		m.javaVersion.GetSelected().Id,
		m.dependencies.GetSelectedIds(),
		m.metadata.GetValues(),
	)
	if err != nil {
		return fullPath, isZip, fmt.Errorf("error generating download request: %v", err)
	}

	baseName := path.Base(url.Path)
	fullPath = path.Join(m.targetDirectory, baseName)
	err = springio.DownloadGeneratedZip(url.String(), fullPath)
	if err != nil {
		return fullPath, isZip, fmt.Errorf("error downloading zip: %v", err)
	}
	return fullPath, strings.HasSuffix(baseName, "zip"), nil
}

func initialModel() model {
	metaData, err := springio.GetMeta()
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
			dependencies = append(dependencies,
				dependency.Dependency{
					GroupName:   dependencyGroup.Name,
					Id:          groupItem.Id,
					Name:        groupItem.Name,
					Description: groupItem.Description,
				})
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
	return tea.Batch(m.spinner.Tick, func() tea.Msg {
		return initialModel()
	})
}

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

func (m model) renderMain(content string) string {
	h, v := docStyle.GetFrameSize()
	return docStyle.Render(lipgloss.Place(m.width-h, m.height-v, lipgloss.Center, lipgloss.Center, content))
}

func (m model) View() string {
	if m.state == LOADING {
		return m.renderMain(lipgloss.JoinHorizontal(lipgloss.Center, m.spinner.View(), "Loading metadata from spring.io..."))
	}

	if m.width < constants.MinScreenWidth || m.height < constants.MinScreenHeight {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			fmt.Sprintf("This screen is too small. (Min: %dx%d) (Current: %dx%d)",
				constants.MinScreenWidth, constants.MinScreenHeight, m.width, m.height))
	}

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
	body := m.renderMain(
		lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.JoinHorizontal(lipgloss.Top, leftSection, rightSection),
			m.help.View(m.keys)))

	if m.notification.IsActive() {
		h, v := lipgloss.Size(body)
		notification := m.notification.View()
		hn, vn := lipgloss.Size(notification)
		verticalPos := math.Floor(float64(v) * 0.5 - float64(vn) * 0.5)
		body = overlay.PlaceOverlay(h/2-hn/2, int(verticalPos), notification, body)
	}

	return body
}

func (m *model) updateHelp() {
	if m.notification.IsActive() {
		m.keys.SectionShortKeys = m.notification.ShortHelp()
		m.keys.SectionFullKeys = m.notification.FullHelp()
		return
	}
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

	case notification.NotificationMsg:
		m.notification.Activate()
		m.notification = m.notification.UpdateMessage(msg)
		m.updateHelp()

	case model:
		msg.height = m.height
		msg.width = m.width
		msg.project.SetSize(m.project.GetSize())
		msg.language.SetSize(m.language.GetSize())
		msg.springBootVersion.SetSize(m.springBootVersion.GetSize())
		msg.javaVersion.SetSize(m.javaVersion.GetSize())
		msg.packaging.SetSize(m.packaging.GetSize())
		msg.metadata.SetSize(m.metadata.GetSize())
		msg.dependencies.SetSize(m.dependencies.GetSize())
		msg.buttons.SetSize(m.buttons.GetSize())
		msg.targetDirectory = m.targetDirectory
		msg.help.Width = m.help.Width
		msg.notification = m.notification
		m = msg
		m.state = READY

	case spinner.TickMsg:
		switch m.state {
		case LOADING:
			m.spinner, cmd = m.spinner.Update(msg)
		case READY:
			m.buttons, cmd = m.buttons.Update(msg)
		}

	case buttons.ActionState:
		m.buttons, cmd = m.buttons.Update(msg)

	case buttons.Action:
		switch msg {
		case buttons.DOWNLOAD:
			cmd = func() tea.Msg {
				_, _, err := m.generateProject()
				if err != nil {
					logger.Printf("%v", err)
					return buttons.ACTION_FAILED
				}
				return buttons.ACTION_SUCCESS
			}
		case buttons.DOWNLOAD_EXTRACT:
			cmd = func() tea.Msg {
				fullPath, isZip, err := m.generateProject()
				if err != nil {
					logger.Printf("%v", err)
					return buttons.ACTION_FAILED
				}
				if isZip {
					err = files.UnzipFile(fullPath, m.targetDirectory)
					if err != nil {
						logger.Printf("Error unzipping file: %v", err)
						return buttons.ACTION_FAILED
					}
					err = os.Remove(fullPath)
					if err != nil {
						logger.Printf("Error deleting zip file: %v", err)
					}
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

		c2w := cw*2 + hs - 1
		m.metadata.SetSize(c2w, 5)

		m.dependencies.SetSize(c2w, pv+(cmv-5-pv)+2)
		m.buttons.SetSize(c2w, 5)

		m.help.Width = c2w*2 - h - hs

		m.notification.SetSize(c2w, cmv)
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

		if m.notification.IsActive() {
			m.notification, cmd = m.notification.Update(msg)
			return m, cmd
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

type modelOption func(m *model)

func WithSpinner(spinner spinner.Model) modelOption {
	return func(m *model) {
		m.spinner = spinner
	}
}

func WithTargetDir(targetDirectory string) modelOption {
	return func(m *model) {
		m.targetDirectory = targetDirectory
	}
}

func New(options ...modelOption) model {
	model := model{
		notification: notification.New(),
	}

	for _, opt := range options {
		opt(&model)
	}
	return model
}
