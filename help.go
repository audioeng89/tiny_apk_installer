package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type HelpTab struct {
	viewport viewport.Model
}

func init() {
	RegisterTab(&HelpTab{})
}

func (t *HelpTab) Name() string {
	return "Help"
}

func (t *HelpTab) Order() int {
	return 4
}

func (t *HelpTab) Init(m *model) []tea.Cmd {
	content := t.content(m)
	w := m.styles.Content.GetWidth()
	h := m.styles.Content.GetHeight() - 2
	t.viewport = viewport.New(w, h)
	t.viewport.YPosition = 0
	t.viewport.SetContent(content)
	return nil
}

func (t *HelpTab) Update(m *model, msg tea.Msg) (tea.Model, []tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.viewport.Width = m.styles.Content.GetWidth()
		t.viewport.Height = m.styles.Content.GetHeight() - 2
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			t.viewport.ScrollUp(1)
		case "down", "j":
			t.viewport.ScrollDown(1)
		}
	}
	return m, nil
}

func (t *HelpTab) content(m *model) string {
	var b strings.Builder
	b.WriteString("How to Use\n")
	b.WriteString(m.styles.DividerLine.Render(strings.Repeat("─", 24)) + "\n")
	b.WriteString("Enable Developer Options on your Android device.\n\n")
	b.WriteString("USB:\n")
	b.WriteString(m.styles.DividerLine.Render(strings.Repeat("─", 6)) + "\n")
	b.WriteString("1. Enable Developer Options → USB Debugging.\n")
	b.WriteString("2. Connect via USB cable.\n")
	b.WriteString("3. Accept the pop-up on your device.\n")
	b.WriteString("4. Select device from Devices tab.\n\n")
	b.WriteString("Wireless:\n")
	b.WriteString(m.styles.DividerLine.Render(strings.Repeat("─", 12)) + "\n")
	b.WriteString("1. Enable Developer Options → Wireless Debugging.\n")
	b.WriteString("2. Use Pairing device with pairing code.\n")
	b.WriteString("3. Go to Pairing tab.\n")
	b.WriteString("4. Enter IP, PORT, and CODE.\n")
	b.WriteString("5. Accept the pop-up on your device.\n")
	b.WriteString("6. Select device from Devices tab.\n")
	return b.String()
}

func (t *HelpTab) View(m *model) string {
	return t.viewport.View()
}

func (t *HelpTab) NavHint(m *model) string {
	if !t.viewport.AtTop() || !t.viewport.AtBottom() {
		return "←/→ tabs  •  ↑/↓ scroll"
	}
	return ""
}
