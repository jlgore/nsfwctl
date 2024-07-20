package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	switch m.state {
	case StateSelectingBranch:
		return m.viewBranchSelection()
	case StateViewingSlides:
		return m.viewSlides()
	case StateDeploymentOptions:
		return m.viewDeploymentOptions()
	default:
		return "Unknown state"
	}
}

func (m Model) viewBranchSelection() string {
	title := titleStyle.Render("nsfwctl")
	repoInfo := fmt.Sprintf("Repository: %s", m.repoPath)
	statusInfo := statusStyle.Render(m.status)
	listView := m.list.View()

	if m.err != nil {
		errorView := errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
		return lipgloss.JoinVertical(lipgloss.Left,
			title,
			repoInfo,
			statusInfo,
			errorView,
			listView,
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		repoInfo,
		statusInfo,
		listView,
	)
}

func (m Model) viewSlides() string {
	if m.err != nil {
		errorMsg := fmt.Sprintf("Error fetching slides: %v", m.err)
		helpMsg := "Press 'q' to return to branch selection"
		return lipgloss.JoinVertical(lipgloss.Left,
			errorStyle.Render(errorMsg),
			"\n",
			subtle.Render(helpMsg),
		)
	}

	title := titleStyle.Render(fmt.Sprintf("Slides for branch: %s", m.selectedBranch))
	slideContent := m.slideModel.View()
	navigationHelp := subtle.Render("← → to navigate • q to quit • d for deployment options")

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"\n",
		slideContent,
		"\n",
		navigationHelp,
	)
}

func (m Model) viewDeploymentOptions() string {
	title := titleStyle.Render("Deployment Options")
	options := []string{
		"1. Deploy this branch",
		"2. Return to branch selection",
	}
	optionsView := strings.Join(options, "\n")

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"\n",
		optionsView,
		"\n",
		subtle.Render("Enter your choice (1 or 2)"),
	)
}

var (
	subtle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
)
