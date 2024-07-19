package ui

import (
	"fmt"
)

func (m Model) View() string {
	return appStyle.Render(fmt.Sprintf(
		"nsfwctl\n\n"+
			"Repository: %s\n"+
			"Current Branch: %s\n"+
			"Status: %s\n\n%s",
		m.repoPath,
		m.selectedBranch,
		m.status,
		m.list.View(),
	))
}
