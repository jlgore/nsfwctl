package ui

import (
	"fmt"
	"strings"
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

// Helper function to truncate long descriptions
func truncateDescription(desc string, maxLen int) string {
	if len(desc) <= maxLen {
		return desc
	}
	return strings.TrimSpace(desc[:maxLen-3]) + "..."
}
