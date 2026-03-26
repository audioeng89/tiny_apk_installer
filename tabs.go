package main

import (
	"sort"

	tea "github.com/charmbracelet/bubbletea"
)

type Tab interface {
	Name() string
	Order() int
	Init(m *model) []tea.Cmd
	Update(m *model, msg tea.Msg) (tea.Model, []tea.Cmd)
	View(m *model) string
	NavHint(m *model) string
}

var tabs []Tab

func RegisterTab(t Tab) {
	tabs = append(tabs, t)
}

func GetTabs() []Tab {
	sort.Slice(tabs, func(i, j int) bool {
		return tabs[i].Order() < tabs[j].Order()
	})
	return tabs
}
