package main

import (
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	uiWidth  = 70
	uiHeight = 18
)

type model struct {
	tabs                []Tab
	tabIndex            int
	styles              styles
	adbPath             string
	spinner             spinner.Model
	isLoading           bool
	devices             []Device
	selectedDeviceIndex int
	selectedDevice      *Device
	errorMsg            string
	width               int
	height              int

	appState    string
	splash      splashModel
	fileBrowser *FileBrowser

	wirelessInputs []textinput.Model
	wirelessFocus  int
	isPairing      bool
	pairingMsg     string
	pairingSuccess bool
	pairingSpinner spinner.Model

	apkPath        string
	bundlePath     string
	apkHandler     *APKHandler
	isExtracting   bool
	isInstalling   bool
	installSpinner spinner.Model
	installMsg     string
	installSuccess bool

	isScanningDevices bool
}

func newModel() model {
	m := model{
		tabs:                GetTabs(),
		tabIndex:            0,
		styles:              NewStyles(defaultTheme),
		spinner:             NewSpinner(defaultTheme),
		isLoading:           false,
		devices:             []Device{},
		selectedDeviceIndex: -1,
		wirelessFocus:       0,
		appState:            "splash",
		splash:              newSplashModel(),
	}

	return m
}

func newTextInput(placeholder string, validator func(string) error) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 64
	ti.Prompt = "  "
	ti.Validate = validator
	return ti
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

func (m *model) Init() tea.Cmd {
	exeName := getAdbExeName()

	if path, err := exec.LookPath(exeName); err == nil {
		m.adbPath = path
		adbPath = path
		m.splash.adbFound = true
		m.splash.adbInTemp = false
	} else {
		m.splash.adbFound = false
		if path, found := CheckADBInTemp(); found {
			m.adbPath = path
			adbPath = path
			m.splash.adbInTemp = true
			m.splash.tempPath = GetTempADBDir()
		} else {
			m.splash.adbInTemp = false
		}
	}

	var cmds []tea.Cmd
	if m.appState == "splash" {
		cmds = append(cmds, m.splash.Init())
	} else {
		for _, tab := range m.tabs {
			cmds = append(cmds, tab.Init(m)...)
		}
	}
	if len(cmds) == 0 {
		cmds = append(cmds, m.spinner.Tick)
	}

	return tea.Batch(cmds...)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.appState == "splash" {
		switch msg := msg.(type) {
		case SplashContinueMsg:
			m.appState = "main"
			m.adbPath = adbPath
			cmds := []tea.Cmd{}
			for _, tab := range m.tabs {
				cmds = append(cmds, tab.Init(m)...)
			}
			if len(cmds) == 0 {
				cmds = append(cmds, m.spinner.Tick)
			}
			return m, tea.Batch(cmds...)

		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height

		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			}
		}

		var cmd tea.Cmd
		var splashView tea.Model
		splashView, cmd = m.splash.Update(msg)
		m.splash = *splashView.(*splashModel)
		return m, cmd
	}

	if m.appState == "filebrowser" {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.appState = "main"
				m.fileBrowser = nil
				return m, nil
			case "enter":
				if m.fileBrowser != nil {
					if path, ok := m.fileBrowser.SelectedFile(); ok {
						m.installMsg = ""
						m.installSuccess = false

						bundleType := DetectBundleType(path)
						if bundleType != BundleNone {
							m.bundlePath = path
							m.apkPath = ""
							m.apkHandler = nil
						} else {
							m.apkPath = path
							m.bundlePath = ""
						}

						m.appState = "main"
						m.fileBrowser = nil
						return m, nil
					}
				}
			}
		}
		var fbModel tea.Model
		var cmd tea.Cmd
		fbModel, cmd = m.fileBrowser.Update(msg)
		m.fileBrowser = fbModel.(*FileBrowser)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "right", "tab":
			if m.tabIndex < len(m.tabs)-1 {
				m.tabIndex++
			}

		case "left", "shift+tab":
			if m.tabIndex > 0 {
				m.tabIndex--
			}
		}
	}

	if m.tabIndex >= 0 && m.tabIndex < len(m.tabs) {
		_, cmds := m.tabs[m.tabIndex].Update(m, msg)
		return m, tea.Batch(cmds...)
	}

	return m, nil
}

func (m *model) View() string {
	if m.appState == "splash" {
		splashView := m.splash.View()
		styledDoc := m.styles.DocStyle.Render(splashView)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, styledDoc)
	}

	if m.appState == "filebrowser" {
		fbView := m.fileBrowser.View()
		styledDoc := m.styles.FileBrowserDocStyle.Render(fbView)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, styledDoc)
	}

	doc := strings.Builder{}

	row := m.renderTabs()
	content := m.styles.Content.Width(lipgloss.Width(row)).Render(m.renderContent())
	lineStyle := lipgloss.NewStyle().Foreground(defaultTheme.TabBorder)
	line := lineStyle.Render(strings.Repeat("─", lipgloss.Width(row)))

	var navHint string
	if m.tabIndex >= 0 && m.tabIndex < len(m.tabs) {
		navHint = m.tabs[m.tabIndex].NavHint(m)
	}

	doc.WriteString(row)
	doc.WriteString("\n")
	doc.WriteString(content)
	doc.WriteString("\n")
	doc.WriteString(line)
	if navHint != "" {
		doc.WriteString("\n")
		doc.WriteString(m.styles.Help.Render(navHint))
	}

	styledDoc := m.styles.DocStyle.Render(doc.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, styledDoc)
}

func (m *model) renderTabs() string {
	var renderedTabs []string
	for i, tab := range m.tabs {
		isFirst, isLast, isActive := i == 0, i == len(m.tabs)-1, i == m.tabIndex

		var style lipgloss.Style
		if isActive {
			style = m.styles.ActiveTab.Copy()
		} else {
			style = m.styles.InactiveTab.Copy()
		}

		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(" "+tab.Name()+" "))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
}

func (m *model) renderContent() string {
	if m.tabIndex >= 0 && m.tabIndex < len(m.tabs) {
		return m.tabs[m.tabIndex].View(m)
	}
	return ""
}

func (m *model) HasSelectedAPK() bool {
	return m.apkPath != "" || m.bundlePath != "" || m.apkHandler != nil
}

func Run() error {
	m := newModel()
	p := tea.NewProgram(&m, tea.WithAltScreen())

	defer func() {
		StopADBServer()
	}()

	_, err := p.Run()
	return err
}
