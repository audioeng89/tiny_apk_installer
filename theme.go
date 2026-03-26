package main

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/charmtone"
)

type Theme struct {
	Primary      lipgloss.Color
	Success      lipgloss.Color
	Error        lipgloss.Color
	Warning      lipgloss.Color
	Help         lipgloss.Color
	Muted        lipgloss.Color
	Selected     lipgloss.Color
	Header       lipgloss.Color
	Border       lipgloss.Color
	BorderActive lipgloss.Color
	TabBorder    lipgloss.Color
	Background   lipgloss.Color
	ButtonBg     lipgloss.Color
	Disabled     lipgloss.Color
}

var defaultTheme = Theme{
	Primary:      lipgloss.Color(charmtone.Guac.Hex()),
	Success:      lipgloss.Color(charmtone.Guac.Hex()),
	Error:        lipgloss.Color(charmtone.Cherry.Hex()),
	Warning:      lipgloss.Color(charmtone.Tang.Hex()),
	Help:         lipgloss.Color(charmtone.Lichen.Hex()),
	Muted:        lipgloss.Color("250"),
	Selected:     lipgloss.Color(charmtone.Guac.Hex()),
	Header:       lipgloss.Color("250"),
	Border:       lipgloss.Color("250"),
	BorderActive: lipgloss.Color(charmtone.Guac.Hex()),
	TabBorder:    lipgloss.Color(charmtone.Pickle.Hex()),
	Background:   lipgloss.Color("236"),
	ButtonBg:     lipgloss.Color("236"),
	Disabled:     lipgloss.Color("240"),
}

type styles struct {
	DocStyle            lipgloss.Style
	SplashDocStyle      lipgloss.Style
	FileBrowserDocStyle lipgloss.Style
	Title               lipgloss.Style
	SplashTitle         lipgloss.Style
	SplashHeader        lipgloss.Style
	SplashTitleBig      lipgloss.Style
	SplashDivider       lipgloss.Style
	SplashButton        lipgloss.Style
	SplashButtonActive  lipgloss.Style
	Header              lipgloss.Style
	Item                lipgloss.Style
	SelectedItem        lipgloss.Style
	Pagination          lipgloss.Style
	Help                lipgloss.Style
	QuitText            lipgloss.Style
	InactiveTab         lipgloss.Style
	ActiveTab           lipgloss.Style
	Content             lipgloss.Style
	InputLabel          lipgloss.Style
	InputField          lipgloss.Style
	Placeholder         lipgloss.Style
	InputError          lipgloss.Style
	InputValid          lipgloss.Style
	Button              lipgloss.Style
	ButtonActive        lipgloss.Style
	Divider             lipgloss.Style
	DividerLine         lipgloss.Style

	SuccessMsg lipgloss.Style
	ErrorMsg   lipgloss.Style
	Warning    lipgloss.Style

	InactiveOption lipgloss.Style
	ActiveOption   lipgloss.Style
	DisabledOption lipgloss.Style
	SuccessStatus  lipgloss.Style
	Muted          lipgloss.Style
	Spinner        lipgloss.Style
}

func NewStyles(t Theme) styles {
	return styles{
		DocStyle: lipgloss.NewStyle().
			Width(uiWidth).
			Height(uiHeight).
			Align(lipgloss.Center, lipgloss.Center),

		SplashDocStyle: lipgloss.NewStyle().
			Width(uiWidth).
			Height(uiHeight).
			Padding(1, 2).
			Align(lipgloss.Center, lipgloss.Center),

		FileBrowserDocStyle: lipgloss.NewStyle().
			Width(uiWidth).
			Height(uiHeight).
			Padding(1, 2).
			Align(lipgloss.Left, lipgloss.Top),

		Title: lipgloss.NewStyle().
			Foreground(t.Primary).
			Bold(true).
			Align(lipgloss.Center).
			Width(uiWidth),

		SplashTitle: lipgloss.NewStyle().
			Foreground(t.Primary).
			Bold(true).
			Width(uiWidth),

		SplashHeader: lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Bold(true),

		SplashTitleBig: lipgloss.NewStyle().
			Foreground(t.Primary).
			Bold(true).
			Align(lipgloss.Left).
			Width(uiWidth),

		SplashDivider: lipgloss.NewStyle().
			Foreground(t.Primary),

		SplashButton: lipgloss.NewStyle().
			Foreground(t.Muted).
			Background(t.ButtonBg),

		SplashButtonActive: lipgloss.NewStyle().
			Foreground(t.Primary).
			Background(t.ButtonBg).
			Bold(true),

		Header: lipgloss.NewStyle().
			Foreground(t.Header),

		Item: lipgloss.NewStyle().
			Foreground(t.Muted),

		SelectedItem: lipgloss.NewStyle().
			Foreground(t.Selected).
			Bold(true).
			Underline(true),

		Pagination: lipgloss.NewStyle().
			Foreground(t.Muted),

		Help: lipgloss.NewStyle().
			Foreground(t.Help).
			Faint(true),

		InactiveTab: lipgloss.NewStyle().
			Border(tabBorderWithBottom("┴", "─", "┴"), true).
			BorderForeground(t.TabBorder).
			Foreground(t.Disabled).
			Padding(0, 1),

		ActiveTab: lipgloss.NewStyle().
			Border(tabBorderWithBottom("┘", " ", "└"), true).
			BorderForeground(t.TabBorder).
			Foreground(lipgloss.Color("255")).
			Padding(0, 1),

		Content: lipgloss.NewStyle().
			Width(uiWidth).
			Height(uiHeight-2).
			Padding(1, 2),

		InputLabel: lipgloss.NewStyle().
			Foreground(t.Muted),

		InputField: lipgloss.NewStyle().
			Foreground(t.Primary),

		Placeholder: lipgloss.NewStyle().
			Foreground(t.Disabled),

		InputError: lipgloss.NewStyle().
			Foreground(t.Error),

		InputValid: lipgloss.NewStyle().
			Foreground(t.Success),

		Button: lipgloss.NewStyle().
			Foreground(t.Muted).
			Background(t.ButtonBg),

		ButtonActive: lipgloss.NewStyle().
			Foreground(t.Selected).
			Background(t.ButtonBg).
			Bold(true),

		Divider: lipgloss.NewStyle().
			Foreground(t.Border),

		DividerLine: lipgloss.NewStyle().
			Foreground(t.Primary),

		SuccessMsg: lipgloss.NewStyle().
			Foreground(t.Success),

		ErrorMsg: lipgloss.NewStyle().
			Foreground(t.Error),

		Warning: lipgloss.NewStyle().
			Foreground(t.Warning).
			Bold(true),

		InactiveOption: lipgloss.NewStyle().
			Foreground(t.Muted),

		ActiveOption: lipgloss.NewStyle().
			Foreground(t.Selected).
			Bold(true),

		DisabledOption: lipgloss.NewStyle().
			Foreground(t.Disabled),

		SuccessStatus: lipgloss.NewStyle().
			Foreground(t.Success),

		Muted: lipgloss.NewStyle().
			Foreground(t.Muted),

		Spinner: lipgloss.NewStyle().
			Foreground(t.Primary),
	}
}

func NewSpinner(t Theme) spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(t.Primary)
	return s
}
