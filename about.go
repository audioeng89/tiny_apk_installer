package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type AboutTab struct {
	viewport viewport.Model
}

func init() {
	RegisterTab(&AboutTab{})
}

func (t *AboutTab) Name() string {
	return "About"
}

func (t *AboutTab) Order() int {
	return 5
}

func (t *AboutTab) Init(m *model) []tea.Cmd {
	content := t.content(m)
	w := m.styles.Content.GetWidth()
	h := m.styles.Content.GetHeight() - 2
	t.viewport = viewport.New(w, h)
	t.viewport.YPosition = 0
	t.viewport.SetContent(content)
	return nil
}

func (t *AboutTab) Update(m *model, msg tea.Msg) (tea.Model, []tea.Cmd) {
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

func (t *AboutTab) content(m *model) string {
	var b strings.Builder
	b.WriteString("Tiny APK Installer v" + Version + "\n")
	b.WriteString(m.styles.DividerLine.Render(strings.Repeat("─", 24)) + "\n\n")
	b.WriteString("A simple TUI for installing APKs and XAPK/APKM bundles.\n")
	b.WriteString("Written by R. Triggs, assistance from MiniMax M2.7.\n\n")
	b.WriteString("Built with Go and Bubble Tea\n")
	b.WriteString(m.styles.Help.Render("https://github.com/charmbracelet/bubbletea") + "\n\n")
	b.WriteString("Android Debug Bridge is licensed under Apache 2.0.\n")
	b.WriteString(m.styles.Help.Render("https://source.android.com/license") + "\n\n")
	b.WriteString("This project is licensed under MIT.\n")
	return b.String()
}

func (t *AboutTab) View(m *model) string {
	return t.viewport.View()
}

func (t *AboutTab) NavHint(m *model) string {
	if !t.viewport.AtTop() || !t.viewport.AtBottom() {
		return "←/→ tabs  •  ↑/↓ scroll"
	}
	return ""
}
