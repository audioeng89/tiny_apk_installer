package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type PairingTab struct{}

func init() {
	RegisterTab(&PairingTab{})
}

func (t *PairingTab) Name() string {
	return "Pairing"
}

func (t *PairingTab) Order() int {
	return 2
}

func (t *PairingTab) Init(m *model) []tea.Cmd {
	m.wirelessInputs = []textinput.Model{
		newTextInput("Device IP (e.g. 192.168.1.100)", ipValidator),
		newTextInput("Pairing Port (e.g. 45678)", portValidator),
		newTextInput("Pairing Code", codeValidator),
	}

	m.wirelessFocus = 0
	m.pairingSpinner = NewSpinner(defaultTheme)

	return []tea.Cmd{m.wirelessInputs[0].Focus()}
}

func (t *PairingTab) Update(m *model, msg tea.Msg) (tea.Model, []tea.Cmd) {
	switch msg := msg.(type) {
	case PairingResultMsg:
		m.isPairing = false
		if msg.Success {
			m.pairingSuccess = true
			m.pairingMsg = "Pairing successful!"
			for i := range m.wirelessInputs {
				m.wirelessInputs[i].SetValue("")
			}
		} else {
			m.pairingSuccess = false
			m.pairingMsg = "Pairing failed: " + msg.Error
		}
		return m, []tea.Cmd{
			func() tea.Msg {
				devices, err := DiscoverDevices()
				return DiscoveredDevicesMsg{Devices: devices, Err: err}
			},
		}

	case spinner.TickMsg:
		m.pairingSpinner, _ = m.pairingSpinner.Update(msg)
		if m.isPairing {
			return m, []tea.Cmd{m.pairingSpinner.Tick}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.wirelessFocus >= 0 && m.wirelessFocus < len(m.wirelessInputs) {
				m.wirelessInputs[m.wirelessFocus].Blur()
			}
			m.wirelessFocus--
			if m.wirelessFocus < 0 {
				m.wirelessFocus = len(m.wirelessInputs)
			}
			if m.wirelessFocus >= 0 && m.wirelessFocus < len(m.wirelessInputs) {
				return m, []tea.Cmd{m.wirelessInputs[m.wirelessFocus].Focus()}
			}

		case "down":
			if m.wirelessFocus >= 0 && m.wirelessFocus < len(m.wirelessInputs) {
				m.wirelessInputs[m.wirelessFocus].Blur()
			}
			m.wirelessFocus++
			if m.wirelessFocus > len(m.wirelessInputs) {
				m.wirelessFocus = 0
			}
			if m.wirelessFocus >= 0 && m.wirelessFocus < len(m.wirelessInputs) {
				return m, []tea.Cmd{m.wirelessInputs[m.wirelessFocus].Focus()}
			}

		case "enter":
			if m.wirelessFocus == len(m.wirelessInputs) {
				ip := m.wirelessInputs[0].Value()
				port := m.wirelessInputs[1].Value()
				code := m.wirelessInputs[2].Value()

				ipErr := ipValidator(ip)
				portErr := portValidator(port)
				codeErr := codeValidator(code)

				if ipErr != nil {
					m.pairingMsg = "Invalid IP: " + ipErr.Error()
				} else if portErr != nil {
					m.pairingMsg = "Invalid Port: " + portErr.Error()
				} else if codeErr != nil {
					m.pairingMsg = "Invalid Code: " + codeErr.Error()
				} else if ip == "" || port == "" || code == "" {
					m.pairingMsg = "Please fill in all fields"
				} else {
					return m, t.doPairing(m, ip, port, code)
				}
			} else {
				if m.wirelessFocus >= 0 && m.wirelessFocus < len(m.wirelessInputs) {
					m.wirelessInputs[m.wirelessFocus].Blur()
				}
				m.wirelessFocus++
				if m.wirelessFocus > len(m.wirelessInputs) {
					m.wirelessFocus = 0
				}
				if m.wirelessFocus < len(m.wirelessInputs) {
					return m, []tea.Cmd{m.wirelessInputs[m.wirelessFocus].Focus()}
				}
			}
		}

		if !m.isPairing && m.wirelessFocus >= 0 && m.wirelessFocus < len(m.wirelessInputs) {
			m.wirelessInputs[m.wirelessFocus], _ = m.wirelessInputs[m.wirelessFocus].Update(msg)
		}
	}
	return m, nil
}

func (t *PairingTab) doPairing(m *model, ip, port, code string) []tea.Cmd {
	m.isPairing = true
	m.pairingMsg = "Pairing..."
	m.pairingSpinner = NewSpinner(defaultTheme)

	return []tea.Cmd{
		m.pairingSpinner.Tick,
		func() tea.Msg {
			result := PairDevice(ip, port, code)
			if result.Success {
				ConnectDevice(result.ConnectAddr)
			}
			return PairingResultMsg{
				Success: result.Success,
				Error:   result.Error,
			}
		},
	}
}

func (t *PairingTab) View(m *model) string {
	var b strings.Builder

	b.WriteString("Wireless Pairing\n")
	b.WriteString(m.styles.DividerLine.Render(strings.Repeat("─", 24)) + "\n\n")

	if m.isPairing {
		b.WriteString(fmt.Sprintf("%s %s\n\n", m.pairingSpinner.View(), m.pairingMsg))
	} else {
		if m.pairingMsg != "" {
			if m.pairingSuccess {
				b.WriteString(m.styles.InputValid.Render(m.pairingMsg) + "\n\n")
			} else {
				b.WriteString(m.styles.InputError.Render(m.pairingMsg) + "\n\n")
			}
		}

		labels := []string{"Device IP", "Pairing Port", "Pairing Code"}
		for i, input := range m.wirelessInputs {
			b.WriteString(m.styles.InputLabel.Render(labels[i]) + "\n")
			b.WriteString(input.View() + "\n")

			if input.Value() != "" {
				if input.Err != nil {
					b.WriteString(m.styles.InputError.Render("  "+input.Err.Error()) + "\n")
				} else {
					b.WriteString(m.styles.InputValid.Render("  OK") + "\n")
				}
			}
			b.WriteString("\n")
		}

		buttonText := "[ Pair & Connect ]"
		if m.wirelessFocus == len(m.wirelessInputs) {
			b.WriteString(m.styles.ButtonActive.Render("> " + buttonText))
		} else {
			b.WriteString(m.styles.Button.Render("  " + buttonText))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (t *PairingTab) NavHint(m *model) string {
	return "←/→ tabs  •  ↑/↓ navigate  •  Enter submit"
}

func ipValidator(s string) error {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return fmt.Errorf("ip must have 4 parts")
	}
	for _, part := range parts {
		if len(part) == 0 {
			return fmt.Errorf("ip octet cannot be empty")
		}
		var n int
		if _, err := fmt.Sscanf(part, "%d", &n); err != nil {
			return fmt.Errorf("ip octet must be numeric")
		}
		if n < 0 || n > 255 {
			return fmt.Errorf("ip octet must be 0-255")
		}
	}
	return nil
}

func portValidator(s string) error {
	if s == "" {
		return nil
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return fmt.Errorf("port must be numeric")
		}
	}
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return fmt.Errorf("port must be numeric")
	}
	if n < 1 || n > 65535 {
		return fmt.Errorf("port must be 1-65535")
	}
	return nil
}

func codeValidator(s string) error {
	if s == "" {
		return nil
	}
	if len(s) != 6 {
		return fmt.Errorf("code must be exactly 6 digits")
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return fmt.Errorf("code must be numeric only")
		}
	}
	return nil
}
