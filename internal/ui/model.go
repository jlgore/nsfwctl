package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle   = lipgloss.NewStyle().Padding(1, 2)
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)
	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
)

type item struct {
	title       string
	description string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

type Model struct {
	list           list.Model
	repoPath       string
	status         string
	selectedBranch string
}

func NewModel(repoPath string) Model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select a branch"
	l.SetShowTitle(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	return Model{
		list:           l,
		repoPath:       repoPath,
		status:         "Initializing...",
		selectedBranch: "main",
	}
}

func (m Model) Init() tea.Cmd {
	return fetchBranchesWithDescriptionsCmd(m.repoPath)
}
