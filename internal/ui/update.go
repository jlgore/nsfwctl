package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jlgore/nsfwctl/internal/git"
	"github.com/jlgore/nsfwctl/internal/terraform"
)

type fetchBranchesWithDescriptionsMsg []git.BranchInfo
type statusMsg string

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.selectedBranch = i.title
				return m, switchBranchCmd(m.repoPath, m.selectedBranch)
			}
		}

	case statusMsg:
		m.status = string(msg)

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
}

func fetchBranchesWithDescriptionsCmd(repoPath string) tea.Cmd {
	return func() tea.Msg {
		branchInfos, err := git.FetchBranchesWithDescriptions(repoPath)
		if err != nil {
			return statusMsg(fmt.Sprintf("Error fetching branches: %v", err))
		}
		return fetchBranchesWithDescriptionsMsg(branchInfos)
	}
}

func switchBranchCmd(repoPath, branch string) tea.Cmd {
	return func() tea.Msg {
		if err := git.SwitchBranch(repoPath, branch); err != nil {
			return statusMsg(err.Error())
		}
		output, err := terraform.InitTerraform(repoPath)
		if err != nil {
			return statusMsg(err.Error())
		}
		return statusMsg("Switched to branch: " + branch + "\n" + output)
	}
}
