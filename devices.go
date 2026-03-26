package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type DevicesTab struct{}

func init() {
	RegisterTab(&DevicesTab{})
}

func (t *DevicesTab) Name() string {
	return "Devices"
}

func (t *DevicesTab) Order() int {
	return 0
}

func (t *DevicesTab) Init(m *model) []tea.Cmd {
	if m.isScanningDevices {
		return []tea.Cmd{m.spinner.Tick}
	}
	m.isLoading = true
	m.errorMsg = ""
	m.devices = []Device{}
	m.selectedDeviceIndex = -1
	m.isScanningDevices = true

	return []tea.Cmd{
		m.spinner.Tick,
		func() tea.Msg {
			devices, err := DiscoverDevices()
			return DiscoveredDevicesMsg{Devices: devices, Err: err}
		},
	}
}

func (t *DevicesTab) Update(m *model, msg tea.Msg) (tea.Model, []tea.Cmd) {
	switch msg := msg.(type) {
	case DiscoveredDevicesMsg:
		m.isLoading = false
		m.isScanningDevices = false
		if msg.Err != nil {
			m.errorMsg = msg.Err.Error()
		} else {
			m.devices = msg.Devices
			if len(m.devices) > 0 && m.selectedDeviceIndex == -1 {
				m.selectedDeviceIndex = 0
			}
		}

	case spinner.TickMsg:
		m.spinner, _ = m.spinner.Update(msg)
		if m.isLoading {
			return m, []tea.Cmd{m.spinner.Tick}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "r", "R":
			return m, t.Init(m)
		case "up":
			if len(m.devices) > 0 {
				if m.selectedDeviceIndex > 0 {
					m.selectedDeviceIndex--
				} else if m.selectedDeviceIndex == -1 && len(m.devices) > 0 {
					m.selectedDeviceIndex = len(m.devices) - 1
				}
			}
		case "down":
			if len(m.devices) > 0 {
				if m.selectedDeviceIndex < len(m.devices)-1 {
					m.selectedDeviceIndex++
				} else if m.selectedDeviceIndex == -1 {
					m.selectedDeviceIndex = 0
				}
			}
		case "enter":
			if m.selectedDeviceIndex >= 0 {
				d := m.devices[m.selectedDeviceIndex]
				m.selectedDevice = &d

				if d.Type == "wireless" && !d.IsConnected {
					ipPort := d.IPPort
					if idx := strings.LastIndex(ipPort, ":"); idx > 0 {
						ip := ipPort[:idx]
						m.wirelessInputs[0].SetValue(ip)
						m.wirelessInputs[0].Blur()
						m.tabIndex = 2
						m.wirelessFocus = 1
						m.wirelessInputs[1].Focus()
						return m, []tea.Cmd{m.wirelessInputs[1].Focus()}
					}
				}
			}
		}
	}
	return m, nil
}

func (t *DevicesTab) View(m *model) string {
	var b strings.Builder

	if m.adbPath == "" {
		b.WriteString("Devices\n\n")
		b.WriteString("Error: ADB not found\n")
		b.WriteString("Place adb in APP folder or add to PATH")
		return b.String()
	}

	b.WriteString("Available Devices\n")
	b.WriteString(m.styles.DividerLine.Render(strings.Repeat("─", 24)) + "\n\n")

	if m.isLoading {
		b.WriteString(fmt.Sprintf("%s Discovering devices...\n\n", m.spinner.View()))
	}

	if m.errorMsg != "" {
		b.WriteString(fmt.Sprintf("Error: %s\n\n", m.errorMsg))
	}

	if len(m.devices) == 0 && !m.isLoading && m.errorMsg == "" {
		b.WriteString("No devices found\n")
		b.WriteString("Press R to refresh\n")
		return b.String()
	}

	for i, d := range m.devices {
		typeLabel := "USB"
		if d.Type == "wireless" {
			typeLabel = "WiFi"
		}

		status := ""
		if m.selectedDevice != nil && m.selectedDevice.Name == d.Name && d.IsConnected {
			status = " [selected]"
		} else if d.Type == "wireless" && !d.IsConnected {
			status = " [not paired yet]"
		}

		line := fmt.Sprintf("%d. %s [%s]%s", i+1, d.Name, typeLabel, status)

		if i == m.selectedDeviceIndex {
			b.WriteString(m.styles.SelectedItem.Render("> " + line))
		} else {
			b.WriteString(m.styles.Item.Render("  " + line))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (t *DevicesTab) NavHint(m *model) string {
	return "←/→ tabs  •  ↑/↓ select  •  R refresh"
}
