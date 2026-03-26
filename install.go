package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type InstallTab struct{}

func init() {
	RegisterTab(&InstallTab{})
}

func (t *InstallTab) Name() string {
	return "Install APK"
}

func (t *InstallTab) Order() int {
	return 1
}

func (t *InstallTab) Init(m *model) []tea.Cmd {
	m.installSpinner = NewSpinner(defaultTheme)
	return nil
}

func (t *InstallTab) Update(m *model, msg tea.Msg) (tea.Model, []tea.Cmd) {
	switch msg := msg.(type) {
	case InstallResultMsg:
		m.isInstalling = false
		if msg.Success {
			m.installSuccess = true
			m.installMsg = "Install successful!"
			m.apkPath = ""
			m.bundlePath = ""
			if m.apkHandler != nil {
				m.apkHandler.Cleanup()
				m.apkHandler = nil
			}
		} else {
			m.installSuccess = false
			m.installMsg = "Install failed: " + msg.Error
		}

	case BundleExtractedMsg:
		m.isExtracting = false
		if msg.Err != nil {
			m.installMsg = "Extract failed: " + msg.Err.Error()
			m.installSuccess = false
		} else {
			m.apkHandler = msg.Handler
			m.apkPath = "Bundle extracted"
			m.installMsg = "Installing..."
			m.isInstalling = true
			return m, []tea.Cmd{
				m.installSpinner.Tick,
				func() tea.Msg {
					config := ADBConfig{
						Serial: m.selectedDevice.IPPort,
					}
					result := m.apkHandler.Install(config)
					return InstallResultMsg{
						Success: result.Success,
						Error:   result.Error,
					}
				},
			}
		}

	case spinner.TickMsg:
		m.installSpinner, _ = m.installSpinner.Update(msg)
		if m.isInstalling || m.isExtracting {
			return m, []tea.Cmd{m.installSpinner.Tick}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "b", "B":
			t.browseForAPK(m)
		case "enter":
			if m.selectedDevice != nil && m.selectedDevice.IsConnected && m.HasSelectedAPK() && !m.isInstalling && !m.isExtracting {
				return m, t.doInstall(m)
			}
		}
	}
	return m, nil
}

func (t *InstallTab) doInstall(m *model) []tea.Cmd {
	if m.selectedDevice == nil || !m.selectedDevice.IsConnected {
		m.installMsg = "No device selected"
		m.installSuccess = false
		return nil
	}
	if !m.HasSelectedAPK() {
		m.installMsg = "No APK file selected"
		m.installSuccess = false
		return nil
	}

	m.isInstalling = true
	m.installSpinner = NewSpinner(defaultTheme)

	if m.bundlePath != "" {
		m.installMsg = "Extracting bundle..."
		tempBaseDir := filepath.Join(filepath.Dir(m.adbPath), "temp")
		return []tea.Cmd{
			m.installSpinner.Tick,
			func() tea.Msg {
				handler, err := ExtractBundle(m.bundlePath, tempBaseDir)
				return BundleExtractedMsg{Handler: handler, Err: err}
			},
		}
	}

	if m.apkHandler != nil {
		m.installMsg = "Installing..."
		return []tea.Cmd{
			m.installSpinner.Tick,
			func() tea.Msg {
				config := ADBConfig{
					Serial: m.selectedDevice.IPPort,
				}
				result := m.apkHandler.Install(config)
				return InstallResultMsg{
					Success: result.Success,
					Error:   result.Error,
				}
			},
		}
	}

	m.installMsg = "Installing..."
	return []tea.Cmd{
		m.installSpinner.Tick,
		func() tea.Msg {
			config := ADBConfig{
				Serial:  m.selectedDevice.IPPort,
				APKPath: m.apkPath,
			}
			result := config.Install()
			return InstallResultMsg{
				Success: result.Success,
				Error:   result.Error,
			}
		},
	}
}

func (t *InstallTab) browseForAPK(m *model) {
	if m.apkHandler != nil {
		m.apkHandler.Cleanup()
		m.apkHandler = nil
	}
	m.apkPath = ""
	m.bundlePath = ""
	m.fileBrowser = NewFileBrowser(".")
	m.appState = "filebrowser"
}

func (t *InstallTab) View(m *model) string {
	var b strings.Builder
	b.WriteString("Select APK to Install\n")
	b.WriteString(m.styles.DividerLine.Render(strings.Repeat("─", 24)) + "\n\n")

	b.WriteString("Device:\n")
	if m.selectedDevice != nil && m.selectedDevice.IsConnected {
		typeLabel := "USB"
		if m.selectedDevice.Type == "wireless" {
			typeLabel = "WiFi"
		}
		b.WriteString(m.styles.SelectedItem.Render("  " + m.selectedDevice.Name + " [" + typeLabel + "]"))
	} else {
		b.WriteString(m.styles.Placeholder.Render("  No device selected"))
	}
	b.WriteString("\n\n")

	if m.apkHandler != nil {
		b.WriteString(fmt.Sprintf("APKs: %d files\n", len(m.apkHandler.apkPaths)))
		if m.apkHandler.obbPath != "" {
			b.WriteString(m.styles.Muted.Render("  (includes OBB data)"))
			b.WriteString("\n")
		}
	} else if m.bundlePath != "" {
		b.WriteString("Bundle:\n")
		b.WriteString(m.styles.SuccessMsg.Render("  " + filepath.Base(m.bundlePath)))
		b.WriteString(m.styles.Muted.Render(" (will extract on install)"))
		b.WriteString("\n")
	} else {
		b.WriteString("APK:\n")
		if m.apkPath != "" {
			b.WriteString(m.styles.SuccessMsg.Render("  " + m.apkPath))
		} else {
			b.WriteString(m.styles.Placeholder.Render("  No file selected"))
		}
	}
	b.WriteString("\n\n")

	if m.isExtracting {
		b.WriteString(fmt.Sprintf("%s Extracting bundle...\n", m.installSpinner.View()))
		b.WriteString("\n")
	} else if m.isInstalling {
		b.WriteString(fmt.Sprintf("%s %s\n", m.installSpinner.View(), m.installMsg))
		b.WriteString("\n")
	} else if m.installMsg != "" {
		if m.installSuccess {
			b.WriteString(m.styles.InputValid.Render("  "+m.installMsg) + "\n\n")
		} else {
			b.WriteString(m.styles.InputError.Render("  "+m.installMsg) + "\n\n")
		}
	}

	return b.String()
}

func (t *InstallTab) NavHint(m *model) string {
	if m.isInstalling || m.isExtracting {
		return ""
	} else if m.selectedDevice != nil && m.selectedDevice.IsConnected && m.HasSelectedAPK() {
		return "←/→ tabs  •  Enter: install  •  B: browse files"
	} else if m.selectedDevice == nil || !m.selectedDevice.IsConnected {
		return "←/→ tabs  •  Select a connected device first"
	} else if !m.HasSelectedAPK() {
		return "←/→ tabs  •  B: browse files"
	}
	return ""
}
