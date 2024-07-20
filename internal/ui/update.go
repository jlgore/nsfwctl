package ui

import (
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jlgore/nsfwctl/internal/git"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	switch m.state {
	case StateSelectingBranch:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					m.selectedBranch = i.title
					return m, fetchSlidesCmd(m.repoPath, m.selectedBranch)
				}
			}
		case fetchBranchesWithDescriptionsMsg:
			items := make([]list.Item, len(msg))
			for i, branchInfo := range msg {
				items[i] = item{title: branchInfo.Name, description: branchInfo.Description}
			}
			m.list.SetItems(items)
		}
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd

	case StateViewingSlides:
		var cmd tea.Cmd
		m.slideModel, cmd = m.slideModel.Update(msg)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "d" {
				m.state = StateDeploymentOptions
				return m, nil
			}
		case slideModelMsg:
			m.slideModel = msg.model
			m.state = StateViewingSlides
		}
		return m, cmd

	case StateDeploymentOptions:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "1":
				// Implement deployment logic here
				return m, nil
			case "2":
				m.state = StateSelectingBranch
				return m, nil
			}
		}
	}

	return m, nil
}

func fetchBranchesWithDescriptionsCmd(repoPath string) tea.Cmd {
	return func() tea.Msg {
		log.Printf("Executing fetchBranchesWithDescriptionsCmd for repo: %s", repoPath)
		branchInfos, err := git.FetchBranchesWithDescriptions(repoPath)
		if err != nil {
			log.Printf("Error fetching branches: %v", err)
			return errMsg{err}
		}
		log.Printf("Fetched %d branch infos", len(branchInfos))
		return fetchBranchesWithDescriptionsMsg(branchInfos)
	}
}

func fetchSlidesCmd(repoPath, branchName string) tea.Cmd {
	return func() tea.Msg {
		content, err := git.FetchSlides(repoPath, branchName)
		if err != nil {
			return errMsg{err}
		}

		slideModel, err := NewSlideModel(content)
		if err != nil {
			return errMsg{err}
		}

		return slideModelMsg{slideModel}
	}
}

type fetchBranchesWithDescriptionsMsg []git.BranchInfo
type slideModelMsg struct {
	model SlideModel
}
type errMsg struct{ err error }
