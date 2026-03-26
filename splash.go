package main

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SplashOption int

const (
	SplashOptionUsePath SplashOption = iota
	SplashOptionUseTemp
	SplashOptionDownload
)

type splashModel struct {
	adbFound         bool
	adbInTemp        bool
	tempPath         string
	selectedOption   SplashOption
	isDownloading    bool
	downloadComplete bool
	downloadSpinner  spinner.Model
	downloadError    string
	browseError      string
	isStartingADB    bool
	startADBError    string
}

func newSplashModel() splashModel {
	s := splashModel{
		selectedOption:  SplashOptionUsePath,
		downloadSpinner: NewSpinner(defaultTheme),
		isStartingADB:   false,
		startADBError:   "",
	}
	s.selectedOption = s.firstAvailableOption()
	return s
}

func (s *splashModel) Init() tea.Cmd {
	return nil
}

func (s *splashModel) canSelect(opt SplashOption) bool {
	if opt == SplashOptionUsePath && !s.adbFound {
		return false
	}
	if opt == SplashOptionUseTemp && !s.adbInTemp {
		return false
	}
	return true
}

func (s *splashModel) firstAvailableOption() SplashOption {
	for opt := SplashOption(0); opt <= SplashOptionDownload; opt++ {
		if s.canSelect(opt) {
			return opt
		}
	}
	return SplashOptionDownload
}

func (s *splashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if s.isDownloading || s.isStartingADB {
			return s, nil
		}

		switch msg.String() {
		case "up", "k":
			for {
				if s.selectedOption > 0 {
					s.selectedOption--
				}
				if s.canSelect(s.selectedOption) {
					break
				}
				if s.selectedOption == SplashOptionUsePath {
					break
				}
			}
		case "down", "j":
			for {
				if s.selectedOption < SplashOptionDownload {
					s.selectedOption++
				}
				if s.canSelect(s.selectedOption) {
					break
				}
				if s.selectedOption == SplashOptionDownload {
					break
				}
			}
		case "enter":
			return s, s.handleSelect()
		}

	case spinner.TickMsg:
		if s.isDownloading || s.isStartingADB {
			s.downloadSpinner, _ = s.downloadSpinner.Update(msg)
			return s, s.downloadSpinner.Tick
		}

	case ADBPathSelectedMsg:
		s.browseError = ""
		if msg.Err != nil {
			s.browseError = msg.Err.Error()
			return s, nil
		}
		if err := ValidateADBPath(msg.Path); err != nil {
			s.browseError = err.Error()
			return s, nil
		}
		exeName := "adb"
		if runtime.GOOS == "windows" {
			exeName = "adb.exe"
		}
		adbPath = msg.Path + string(filepath.Separator) + exeName
		return s, func() tea.Msg { return SplashContinueMsg{} }

	case ADBDownloadResultMsg:
		s.isDownloading = false
		if msg.Err != nil {
			s.downloadError = msg.Err.Error()
			return s, nil
		}
		adbPath = msg.Path
		s.downloadComplete = true
		s.tempPath = GetTempADBDir()
		s.adbInTemp = true
		return s, nil
	}

	return s, nil
}

func (s *splashModel) handleSelect() tea.Cmd {
	s.downloadSpinner = NewSpinner(defaultTheme)

	switch s.selectedOption {
	case SplashOptionUsePath:
		if s.adbFound {
			s.isStartingADB = true
			s.startADBError = ""
			return tea.Batch(
				s.downloadSpinner.Tick,
				func() tea.Msg {
					if err := StartADBServer(); err != nil {
						s.isStartingADB = false
						s.startADBError = err.Error()
						return nil
					}
					return SplashContinueMsg{}
				},
			)
		}
		return nil
	case SplashOptionUseTemp:
		if s.adbInTemp {
			adbPath = s.tempPath + string(filepath.Separator) + "adb"
			if runtime.GOOS == "windows" {
				adbPath = s.tempPath + string(filepath.Separator) + "adb.exe"
			}
			s.isStartingADB = true
			s.startADBError = ""
			return tea.Batch(
				s.downloadSpinner.Tick,
				func() tea.Msg {
					if err := StartADBServer(); err != nil {
						s.isStartingADB = false
						s.startADBError = err.Error()
						return nil
					}
					return SplashContinueMsg{}
				},
			)
		}
		return nil
	case SplashOptionDownload:
		s.isDownloading = true
		s.downloadError = ""
		s.downloadSpinner = NewSpinner(defaultTheme)
		return tea.Batch(
			s.downloadSpinner.Tick,
			func() tea.Msg {
				path, err := DownloadADB()
				return ADBDownloadResultMsg{Path: path, Err: err}
			},
		)
	}
	return nil
}

func (s *splashModel) View() string {
	sty := NewStyles(defaultTheme)

	banner := `   ░░░        ░░░   
  TINY APK INSTALLER  
  ░░░░░░░░░░░░░░░░  
 ░░░░█░░░░░░░░█░░░░ 
░░░░░░░░░░░░░░░░░░░░`

	configureContent := sty.SplashHeader.Render("🖳  CONFIGURE ADB") + "\n\n"

	options := []struct {
		opt      SplashOption
		label    string
		disabled bool
		button   bool
	}{
		{SplashOptionUsePath, "Use PATH", !s.adbFound, false},
		{SplashOptionUseTemp, "Use APP folder", !s.adbInTemp, false},
		{SplashOptionDownload, "[ ↓ Download ADB from Google ]", false, true},
	}

	for i, opt := range options {
		if i == 2 {
			configureContent += "\n"
		}
		prefix := "  "
		var line string

		if s.selectedOption == opt.opt && !opt.disabled {
			prefix = "> "
			if opt.button {
				line = prefix + sty.SplashButtonActive.Render(opt.label)
			} else {
				line = sty.ActiveOption.Render(prefix + opt.label)
			}
		} else if opt.disabled {
			prefix = "  "
			if opt.button {
				line = prefix + sty.DisabledOption.Render(opt.label)
			} else {
				line = sty.DisabledOption.Render(prefix + opt.label + " [not configured]")
			}
		} else {
			if opt.button {
				line = prefix + sty.SplashButton.Render(opt.label)
			} else {
				line = sty.InactiveOption.Render(prefix + opt.label)
			}
		}

		configureContent += line + "\n"
	}
	configureContent += "\n"

	if s.isStartingADB {
		configureContent += s.downloadSpinner.View() + " Starting ADB..."
	} else if s.isDownloading {
		configureContent += s.downloadSpinner.View() + " Downloading ADB..."
	} else if s.adbFound && s.adbInTemp {
		configureContent += sty.Help.Render("✓ ADB found in PATH and APP folder")
	} else if s.adbFound {
		configureContent += sty.Help.Render("✓ ADB found in PATH")
	} else if s.adbInTemp {
		configureContent += sty.Help.Render("✓ ADB found in APP folder: ")
		configureContent += sty.Muted.Render(s.tempPath)
	}

	if s.startADBError != "" {
		configureContent += "\n" + sty.ErrorMsg.Render("Error starting ADB: "+s.startADBError)
	}

	if s.downloadError != "" {
		configureContent += "\n" + sty.ErrorMsg.Render("Download error: "+s.downloadError)
	}

	securityContent := sty.Warning.Render("⚠︎  SECURITY WARNING") + "\n\n" +
		sty.Muted.Render("Beware malware! Never install apps under duress.") + "\n\n" +
		sty.Muted.Render("Only download from trusted sources.") + "\n\n" +
		sty.Muted.Render("Use at your own risk.\n")

	configureWidth := 42
	securityWidth := 23

	configureStyle := lipgloss.NewStyle().Width(configureWidth)
	securityStyle := lipgloss.NewStyle().Width(securityWidth)

	configureStyled := configureStyle.Render(configureContent)
	securityStyled := securityStyle.Render(securityContent)

	configureLines := strings.Count(configureContent, "\n") + 1
	securityLines := strings.Count(securityContent, "\n") + 1
	separatorLines := configureLines
	if securityLines > separatorLines {
		separatorLines = securityLines
	}
	separator := sty.SplashDivider.Render(strings.Repeat("│\n", separatorLines+2) + "│")

	leftSection := lipgloss.JoinHorizontal(lipgloss.Top, configureStyled, separator, securityStyled)

	bannerStyle := lipgloss.NewStyle().Foreground(defaultTheme.Primary).Align(lipgloss.Center).Width(66).Bold(true)
	bannerStyled := bannerStyle.Render(banner)

	row := lipgloss.JoinHorizontal(lipgloss.Top, leftSection)
	lineStyle := lipgloss.NewStyle().Foreground(defaultTheme.TabBorder)
	divider := lineStyle.Render(strings.Repeat("─", lipgloss.Width(row)))

	navHint := sty.Help.Render("↑/↓ select  •  Enter: confirm  •  Q: quit")

	return sty.SplashDocStyle.Render(bannerStyled + "\n" + divider + "\n" + leftSection + "\n" + divider + "\n" + navHint)
}

var (
	_ tea.Model = (*splashModel)(nil)
)
