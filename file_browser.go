package main

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const maxVisibleItems = 10

type FileBrowser struct {
	currentDir string
	files      []os.DirEntry
	selected   int
	scroll     int
	sty        styles
}

func NewFileBrowser(startDir string) *FileBrowser {
	dir := startDir
	if dir == "." || dir == "" {
		usr, err := user.Current()
		if err == nil {
			dir = filepath.Join(usr.HomeDir, "Documents")
		}
	}
	fb := &FileBrowser{
		currentDir: dir,
		sty:        NewStyles(defaultTheme),
	}
	fb.readDir()
	return fb
}

func (fb *FileBrowser) readDir() {
	entries, err := os.ReadDir(fb.currentDir)
	if err != nil {
		fb.files = nil
		return
	}

	canGoUp := fb.currentDir != filepath.VolumeName(fb.currentDir)+string(filepath.Separator)

	var dirs []os.DirEntry
	var files []os.DirEntry

	for _, e := range entries {
		// Skip hidden files/folders
		if len(e.Name()) > 0 && e.Name()[0] == '.' {
			continue
		}
		if e.IsDir() {
			dirs = append(dirs, e)
		} else if isAPKFile(e.Name()) || isBundleFile(e.Name()) {
			files = append(files, e)
		}
	}

	// Sort alphabetically
	sortDirs(dirs)
	sortFiles(files)

	var filtered []os.DirEntry
	if canGoUp {
		filtered = append(filtered, nil)
	}
	filtered = append(filtered, dirs...)
	filtered = append(filtered, files...)

	fb.files = filtered
	fb.selected = 0
	fb.scroll = 0
}

func sortDirs(dirs []os.DirEntry) {
	for i := 0; i < len(dirs)-1; i++ {
		for j := i + 1; j < len(dirs); j++ {
			if strings.ToLower(dirs[i].Name()) > strings.ToLower(dirs[j].Name()) {
				dirs[i], dirs[j] = dirs[j], dirs[i]
			}
		}
	}
}

func sortFiles(files []os.DirEntry) {
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if strings.ToLower(files[i].Name()) > strings.ToLower(files[j].Name()) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}
}

func isAPKFile(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), ".apk")
}

func isBundleFile(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".apkm") ||
		strings.HasSuffix(lower, ".xapk") ||
		strings.HasSuffix(lower, ".apks")
}

func (fb *FileBrowser) Init() tea.Cmd {
	return nil
}

func (fb *FileBrowser) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if fb.selected > 0 {
				fb.selected--
				if fb.selected < fb.scroll {
					fb.scroll = fb.selected
				}
			}
		case "down", "j":
			if fb.selected < len(fb.files)-1 {
				fb.selected++
				if fb.selected >= fb.scroll+maxVisibleItems {
					fb.scroll = fb.selected - maxVisibleItems + 1
				}
			}
		case "enter", "right", "l":
			if fb.selected < len(fb.files) && fb.selected >= 0 {
				if fb.files[fb.selected] == nil {
					fb.currentDir = filepath.Dir(fb.currentDir)
					fb.readDir()
				} else if fb.files[fb.selected].IsDir() {
					fb.currentDir = filepath.Join(fb.currentDir, fb.files[fb.selected].Name())
					fb.readDir()
				}
			}
		case "backspace", "left", "h":
			if fb.currentDir != filepath.VolumeName(fb.currentDir)+string(filepath.Separator) {
				fb.currentDir = filepath.Dir(fb.currentDir)
				fb.readDir()
			}
		}
	}
	return fb, nil
}

func (fb *FileBrowser) View() string {
	primary := lipgloss.NewStyle().Foreground(defaultTheme.Primary)
	muted := lipgloss.NewStyle().Foreground(defaultTheme.Muted)
	selected := lipgloss.NewStyle().Foreground(defaultTheme.Selected).Bold(true)

	var b strings.Builder
	b.WriteString(primary.Render("Select APK File"))
	b.WriteString("\n\n")
	b.WriteString(muted.Render(fb.currentDir))
	b.WriteString("\n\n")

	end := fb.scroll + maxVisibleItems
	if end > len(fb.files) {
		end = len(fb.files)
	}

	for i := fb.scroll; i < end; i++ {
		f := fb.files[i]
		var name string
		if f == nil {
			name = "🗁  ../"
		} else if f.IsDir() {
			name = "🗁  " + f.Name() + "/"
		} else if isBundleFile(f.Name()) {
			name = "▤ " + f.Name()
		} else {
			name = "🗎 " + f.Name()
		}
		if f != nil && f.IsDir() {
			// Folders always use muted color
			if i == fb.selected {
				b.WriteString(muted.Render("> " + name))
			} else {
				b.WriteString(muted.Render(name))
			}
		} else if i == fb.selected {
			// APK files use selected style when highlighted
			b.WriteString(selected.Render("> " + name))
		} else {
			// APK files plain when not selected
			b.WriteString(name)
		}
		b.WriteString("\n")
	}

	if len(fb.files) > maxVisibleItems {
		b.WriteString(muted.Render("... more items ..."))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(fb.sty.DividerLine.Render(strings.Repeat("─", 24)))
	b.WriteString("\n")
	b.WriteString(muted.Render("↑/↓ navigate • ←/→: back/forward • Enter: select • Esc: cancel"))
	return b.String()
}

func (fb *FileBrowser) SelectedFile() (string, bool) {
	if fb.selected >= 0 && fb.selected < len(fb.files) {
		entry := fb.files[fb.selected]
		if entry == nil {
			return "", false
		}
		if !entry.IsDir() {
			return filepath.Join(fb.currentDir, entry.Name()), true
		}
	}
	return "", false
}

var _ tea.Model = (*FileBrowser)(nil)
