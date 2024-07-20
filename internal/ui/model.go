package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ModelState int

const (
	StateSelectingBranch ModelState = iota
	StateViewingSlides
	StateDeploymentOptions
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
	slideModel     SlideModel
	repoPath       string
	status         string
	selectedBranch string
	state          ModelState
	err            error
}

var (
	appStyle   = lipgloss.NewStyle().Padding(1, 2)
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)
)

func NewModel(repoPath string) Model {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(lipgloss.Color("205"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(lipgloss.Color("240"))

	l := list.New([]list.Item{}, delegate, 20, 10) // Set initial width to 20 and height to 10
	l.Title = "Select a branch"
	l.SetShowTitle(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	return Model{
		list:     l,
		repoPath: repoPath,
		status:   "Initializing...",
		state:    StateSelectingBranch,
	}
}

func (m Model) Init() tea.Cmd {
	return fetchBranchesWithDescriptionsCmd(m.repoPath)
}
